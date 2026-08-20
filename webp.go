package main

import (
	"crypto/sha1"
	"encoding/hex"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	webp "github.com/HugoSmits86/nativewebp"
)

// PNG images are served as lossless WebP (nativewebp, pure Go) with a cache
// in <datadir>/webp-cache. Originals on disk are never touched; the cache is
// derived data and can be deleted at any time.
//
// The original PNG is served instead when:
//   - the request asks for it explicitly (?raw=1)
//   - the request comes from the editor (Referer /editor/), which must edit
//     and re-save original image references
//   - decoding/encoding fails for any reason

// serveResourceImage serves a note resource file, converting png to webp
// (cached) when appropriate. Returns false if the file does not exist.
func serveResourceImage(w http.ResponseWriter, r *http.Request, imageFile string) bool {
	if _, err := os.Stat(imageFile); err != nil {
		// the view renders local png links as *.webp: if the requested webp
		// is not on disk, serve the converted counterpart of the png
		if strings.HasSuffix(strings.ToLower(imageFile), ".webp") {
			pngFile := imageFile[:len(imageFile)-len(".webp")] + ".png"
			if _, err := os.Stat(pngFile); err == nil {
				serveWebPConverted(w, r, pngFile)
				return true
			}
		}
		return false
	}
	if !strings.HasSuffix(strings.ToLower(imageFile), ".png") ||
		r.URL.Query().Get("raw") == "1" ||
		strings.Contains(r.Referer(), "/editor") {
		http.ServeFile(w, r, imageFile)
		return true
	}
	serveWebPConverted(w, r, imageFile)
	return true
}

// serveWebPConverted serves the cached webp conversion of a png file,
// converting on first request. Falls back to the original png on any
// conversion error.
func serveWebPConverted(w http.ResponseWriter, r *http.Request, pngFile string) {
	st, err := os.Stat(pngFile)
	if err != nil {
		http.ServeFile(w, r, pngFile)
		return
	}
	cacheFile := webpCachePath(pngFile, st)
	if _, err := os.Stat(cacheFile); err != nil {
		if err := webpConvert(pngFile, cacheFile); err != nil {
			// conversion failed — fall back to the original file
			http.ServeFile(w, r, pngFile)
			return
		}
	}
	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Cache-Control", "public, max-age=604800")
	http.ServeFile(w, r, cacheFile)
}

// webpCachePath builds the cache file name from the source path, size and
// mtime, so a changed source file automatically invalidates its cache entry.
func webpCachePath(imageFile string, st os.FileInfo) string {
	sum := sha1.Sum([]byte(imageFile + "|" + strconv.FormatInt(st.Size(), 10) + "|" + st.ModTime().String()))
	return filepath.Join(configGlobal.dataDir, "webp-cache", hex.EncodeToString(sum[:])+".webp")
}

func webpConvert(srcFile string, cacheFile string) error {
	f, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0755); err != nil {
		return err
	}
	// write via temp file + rename so concurrent/interrupted runs never
	// leave a truncated cache entry
	tmp := cacheFile + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	err = webp.Encode(out, img, nil)
	out.Close()
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, cacheFile)
}

// clearWebPCache removes the whole webp cache directory.
func clearWebPCache() error {
	dir := filepath.Join(configGlobal.dataDir, "webp-cache")
	if _, err := os.Stat(dir); err != nil {
		return nil // nothing to clear
	}
	return os.RemoveAll(dir)
}
