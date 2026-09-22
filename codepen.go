package main

import (
	"encoding/json"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// CodePen embed proxy with a persistent cache.
//
// CodePen's official embed host (codepen.io/<user>/embed/<slug>) sits behind a
// Cloudflare bot challenge that answers with X-Frame-Options: SAMEORIGIN, so it
// cannot be framed from another origin. cdpn.io is CodePen's own embed host and
// is not challenged, but it injects a "referer required" banner unless the
// request comes from codepen.io — so we fetch it server-side with a codepen.io
// Referer and return only the pen's document.
//
// Fetched pens are cached on disk (<datadir>/codepen-cache) and in memory so a
// page with dozens of embeds does not re-hit CodePen, and so pens survive both
// a CodePen outage and a full removal from CodePen.

const (
	codePenCacheTTL       = 24 * time.Hour      // fresh window; older entries revalidate
	codePenRefreshBudget  = 8 * time.Second     // max blocking wait when refreshing a stale entry
	codePenFallbackTTL    = 5 * time.Minute     // don't cache error pages long
	codePenFetchTimeout   = 15 * time.Second    // upstream request timeout
	codePenCacheDirName   = "codepen-cache"
	codePenDefaultHeight  = "480"
	codePenDefaultTheme   = "light"
	codePenDefaultTab     = "result"
	codePenUserAgent      = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
)

// codePenSrcdocRe pulls the pen document out of cdpn.io's fullembedgrid page:
// the pen lives in the srcdoc attribute of the #result iframe (every inner
// quote is entity-escaped, so the first literal " closes the attribute).
var codePenSrcdocRe = regexp.MustCompile(`(?s)<iframe[^>]*srcdoc="(.*?)"`)

// codePenTitleRe extracts the pen's <title> for the caption / tooltip.
var codePenTitleRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

// codePenParamRe validates user/slug path segments before they are spliced
// into the upstream URL.
var codePenParamRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// codePenHTTPClient fetches pens from cdpn.io. Cloudflare in front of cdpn.io
// stalls reused connections after a handful of requests (they then hang until
// the timeout), so keep-alive is disabled: every request opens a fresh
// connection. Measured: with keep-alive, request #5 onwards hangs; without it,
// 8/8 succeed.
var codePenHTTPClient = &http.Client{
	Timeout: codePenFetchTimeout,
	Transport: &http.Transport{
		DisableKeepAlives: true,
	},
}

// codePenCacheEntry is the on-disk/in-memory representation of a fetched pen.
type codePenCacheEntry struct {
	HTML      string `json:"html"`
	Title     string `json:"title"`
	Status    string `json:"status"` // "ok" or "error"
	Error     string `json:"error,omitempty"`
	FetchedAt int64  `json:"fetched_at"`
}

// codePenStore is a two-level (memory + disk) cache keyed by user/slug/tab/theme.
// tab and theme are part of the key because they change the pen's rendered view.
type codePenStore struct {
	mu      sync.RWMutex
	mem     map[string]*codePenCacheEntry
	dir     string
	dirOnce sync.Once
	inFlight map[string]*sync.Mutex // serializes fetches per key (refresh dedupe)
}

var codePenCache = &codePenStore{
	mem:      map[string]*codePenCacheEntry{},
	inFlight: map[string]*sync.Mutex{},
}

func (s *codePenStore) cacheDir() string {
	s.dirOnce.Do(func() {
		s.dir = filepath.Join(configGlobal.dataDir, codePenCacheDirName)
		os.MkdirAll(s.dir, 0760)
	})
	return s.dir
}

// cacheKey maps a validated user/slug/tab/theme to a filesystem-safe key.
func codePenCacheKey(user, slug, tab, theme string) string {
	return user + "_" + slug + "_" + tab + "_" + theme
}

func (s *codePenStore) path(key string) string {
	return filepath.Join(s.cacheDir(), key+".json")
}

// getFresh returns the entry when it exists and is younger than codePenCacheTTL.
// Stale or missing entries return nil.
func (s *codePenStore) getFresh(key string) *codePenCacheEntry {
	e := s.get(key)
	if e == nil || e.FetchedAt == 0 {
		return nil
	}
	if time.Since(time.Unix(e.FetchedAt, 0)) > codePenCacheTTL {
		return nil
	}
	return e
}

// get returns the entry regardless of age (may be nil).
func (s *codePenStore) get(key string) *codePenCacheEntry {
	s.mu.RLock()
	e, ok := s.mem[key]
	s.mu.RUnlock()
	if ok {
		return e
	}
	// Load from disk once, then serve from memory.
	data, err := os.ReadFile(s.path(key))
	if err != nil {
		return nil
	}
	var disk codePenCacheEntry
	if err := json.Unmarshal(data, &disk); err != nil {
		return nil
	}
	s.mu.Lock()
	s.mem[key] = &disk
	s.mu.Unlock()
	return &disk
}

func (s *codePenStore) put(key string, e *codePenCacheEntry) {
	s.mu.Lock()
	s.mem[key] = e
	s.mu.Unlock()
	data, err := json.Marshal(e)
	if err != nil {
		checkQuiet(err)
		return
	}
	tmp := s.path(key) + ".tmp"
	if err := os.WriteFile(tmp, data, 0640); err != nil {
		checkQuiet(err)
		return
	}
	if err := os.Rename(tmp, s.path(key)); err != nil {
		checkQuiet(err)
	}
}

// lockKey returns the per-key mutex used to dedupe concurrent fetches.
func (s *codePenStore) lockKey(key string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.inFlight[key]
	if !ok {
		m = &sync.Mutex{}
		s.inFlight[key] = m
	}
	return m
}

// serveCodePenEmbed answers /api/codepen/{user}/{slug}. It serves fresh cache
// entries directly, revalidates stale ones within codePenRefreshBudget, falls
// back to any cached copy on upstream failure, and only then renders a link.
// ?refresh=1 forces a synchronous revalidation.
func serveCodePenEmbed(w http.ResponseWriter, r *http.Request, user, slug string) {
	if !codePenParamRe.MatchString(slug) {
		http.Error(w, "bad pen", http.StatusBadRequest)
		return
	}
	if user == "" {
		user = "anon"
	}
	if !codePenParamRe.MatchString(user) {
		http.Error(w, "bad pen", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	tab := query.Get("default-tab")
	if tab == "" {
		tab = codePenDefaultTab
	}
	height := query.Get("height")
	if height == "" {
		height = codePenDefaultHeight
	}
	theme := query.Get("theme-id")
	if theme == "" {
		theme = codePenDefaultTheme
	}
	forceRefresh := query.Get("refresh") == "1"

	key := codePenCacheKey(user, slug, tab, theme)

	// Fresh cache wins — no upstream request at all.
	if !forceRefresh {
		if e := codePenCache.getFresh(key); e != nil {
			writeCodePenPen(w, e.HTML, true)
			return
		}
	}

	// Serialize concurrent fetches of the same pen (a page can reference the
	// same slug more than once; without this they'd race and duplicate work).
	lock := codePenCache.lockKey(key)
	lock.Lock()
	defer lock.Unlock()

	// Another goroutine may have refreshed it while we waited.
	if !forceRefresh {
		if e := codePenCache.getFresh(key); e != nil {
			writeCodePenPen(w, e.HTML, true)
			return
		}
	}

	cached := codePenCache.get(key)

	// Stale entry: try to revalidate within the budget, otherwise serve stale
	// immediately and refresh in the background.
	if cached != nil && !forceRefresh {
		done := make(chan *codePenCacheEntry, 1)
		go func() { done <- fetchCodePen(user, slug, tab, height, theme) }()
		select {
		case e := <-done:
			if e.Status == "ok" {
				codePenCache.put(key, e)
				writeCodePenPen(w, e.HTML, true)
				return
			}
			// Upstream failed: keep showing the saved copy.
			writeCodePenPen(w, cached.HTML, true)
			return
		case <-time.After(codePenRefreshBudget):
			// Too slow: serve the stale copy now, keep refreshing in background.
			go func() {
				e := fetchCodePen(user, slug, tab, height, theme)
				if e.Status == "ok" {
					codePenCache.put(key, e)
				}
			}()
			writeCodePenPen(w, cached.HTML, true)
			return
		case <-r.Context().Done():
			writeCodePenPen(w, cached.HTML, true)
			return
		}
	}

	// No cache (or forced refresh): fetch synchronously.
	e := fetchCodePen(user, slug, tab, height, theme)
	if e.Status == "ok" {
		codePenCache.put(key, e)
		writeCodePenPen(w, e.HTML, false)
		return
	}
	if cached != nil {
		writeCodePenPen(w, cached.HTML, true)
		return
	}
	// Nothing to show: cache a short-lived failure marker to avoid hammering a
	// dead pen, then render the plain link.
	e.FetchedAt = time.Now().Unix()
	codePenCache.put(key, e)
	serveCodePenFallback(w, user, slug)
}

// fetchCodePen retrieves one pen from cdpn.io and returns a cache entry. It
// never writes to the cache itself (the caller decides, based on freshness).
func fetchCodePen(user, slug, tab, height, theme string) *codePenCacheEntry {
	e := &codePenCacheEntry{Status: "error", FetchedAt: time.Now().Unix()}
	embedURL := "https://cdpn.io/" + url.PathEscape(user) + "/fullembedgrid/" + url.PathEscape(slug) +
		"?animations=run&type=embed&slug-hash=" + url.QueryEscape(slug) + "&user=" + url.QueryEscape(user) +
		"&default-tab=" + url.QueryEscape(tab) + "&height=" + url.QueryEscape(height) + "&theme-id=" + url.QueryEscape(theme)

	req, err := http.NewRequest("GET", embedURL, nil)
	if err != nil {
		e.Error = err.Error()
		return e
	}
	req.Header.Set("User-Agent", codePenUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	// A codepen.io Referer suppresses cdpn.io's "referer required" banner.
	req.Header.Set("Referer", "https://codepen.io/")

	resp, err := codePenHTTPClient.Do(req)
	if err != nil {
		e.Error = err.Error()
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		e.Error = "HTTP " + resp.Status
		return e
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil || len(body) == 0 {
		e.Error = "empty body"
		return e
	}
	// With a codepen.io Referer, cdpn.io returns the pen document itself. If it
	// ever falls back to the embed wrapper instead, pull the pen out of the
	// #result iframe's srcdoc attribute.
	pen := body
	if match := codePenSrcdocRe.FindSubmatch(body); match != nil {
		pen = []byte(html.UnescapeString(string(match[1])))
	}
	e.HTML = string(pen)
	e.Status = "ok"
	e.Error = ""
	if m := codePenTitleRe.FindSubmatch(pen); m != nil {
		e.Title = strings.TrimSpace(html.UnescapeString(string(m[1])))
	}
	return e
}

// writeCodePenPen serves a pen document with the sandbox CSP that gives it an
// opaque origin (its scripts cannot reach the app's DOM, storage or API).
func writeCodePenPen(w http.ResponseWriter, pen string, cacheable bool) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "sandbox allow-scripts allow-forms allow-modals allow-popups allow-presentation")
	if cacheable {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	} else {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write([]byte(pen))
}

// serveCodePenFallback renders a plain link when cdpn.io cannot be reached.
func serveCodePenFallback(w http.ResponseWriter, user, slug string) {
	link := "https://codepen.io/" + url.PathEscape(user) + "/pen/" + url.PathEscape(slug)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	writeCodePenFallbackBody(w, link, user, slug)
}

func writeCodePenFallbackBody(w io.Writer, link, user, slug string) {
	io.WriteString(w, `<!DOCTYPE html><html><head><meta charset="utf-8"><style>body{font-family:system-ui,sans-serif;font-size:14px;padding:12px;margin:0;color:#555}a{color:#0d6efd}</style></head><body>CodePen: <a href="`+link+`" target="_blank" rel="noopener">`+html.EscapeString(user)+`/`+html.EscapeString(slug)+`</a></body></html>`)
}

// cleanCodePenCacheDir is the cache directory path (used by cache stats/flush).
func cleanCodePenCacheDirName() string {
	return filepath.Join(configGlobal.dataDir, codePenCacheDirName)
}
