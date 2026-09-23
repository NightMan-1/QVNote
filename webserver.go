package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"uuid"
)

func generateUUID() string {
	return uuid.New().String()
}

func jsonResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func serveIndexHTML(w http.ResponseWriter, r *http.Request, status int) {
	data, err := templateFS.ReadFile("templates/index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	w.Write(data)
}

func hasFileExtension(path string) bool {
	segments := strings.Split(path, "/")
	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i] != "" {
			return strings.Contains(segments[i], ".")
		}
	}
	return false
}

func isFrontendRoute(path string) bool {
	var segments []string
	for _, s := range strings.Split(path, "/") {
		if s != "" {
			segments = append(segments, s)
		}
	}

	switch len(segments) {
	case 0:
		return true
	case 1:
		switch segments[0] {
		case "notes", "tags", "settings", "editor", "install", "error", "offline", "error404":
			return true
		}
	case 2, 3:
		if segments[0] == "notes" || segments[0] == "tags" {
			return true
		}
	}
	return false
}

func WebServer(webserverChan chan bool) { //nolint:gocyclo
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Compress(5))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\nDisallow: /\n"))
	})

	// MCP Streamable HTTP endpoint (bearer-token protected; 403 when disabled)
	registerMCPRoutes(r)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, "/api/") {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}

		if hasFileExtension(path) {
			http.NotFound(w, r)
			return
		}

		if !strings.HasSuffix(path, "/") {
			target := path + "/"
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}

		if !isFrontendRoute(path) {
			serveIndexHTML(w, r, http.StatusNotFound)
			return
		}

		// These frontend-only routes should not be opened directly while the
		// server is healthy; redirect them to the root application.
		switch path {
		case "/offline/", "/error/":
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		case "/install/":
			if configGlobal.appInstalled {
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}
		}

		serveIndexHTML(w, r, http.StatusOK)
	})

	// Resources route — display images
	r.Get("/resources/{notebookUUID}/{noteUUID}/{image}", func(w http.ResponseWriter, r *http.Request) {
		notebookUUID := chi.URLParam(r, "notebookUUID")
		noteUUID := chi.URLParam(r, "noteUUID")
		image := chi.URLParam(r, "image")
		imageFile, _ := filepath.Abs(configGlobal.sourceFolder + "/" + notebookUUID + ".qvnotebook/" + noteUUID + ".qvnote/resources/" + image)
		if !serveResourceImage(w, r, imageFile) {
			http.NotFound(w, r)
		}
	})

	// Static files from embedded filesystem
	fileServer := http.FileServer(templateFileSystem())
	r.Handle("/static/*", fileServer)

	// For installation
	r.HandleFunc("/api/config.write", func(w http.ResponseWriter, r *http.Request) {
		var config struct {
			Sourcefolder                 string `json:"sourceFolder"`
			SourceFolderCreateIfNotExist bool   `json:"sourceFolderCreateIfNotExist"`
		}
		readJSON(r, &config)
		if _, err := os.Stat(config.Sourcefolder); err == nil {
			if CheckNotebooksFolderStructure(config.Sourcefolder) {
				configGlobal.sourceFolder = config.Sourcefolder
				configGlobal.appInstalled = true
				if SaveConfig() {
					FindAllNotes()
					jsonResponse(w, map[string]interface{}{
						"error":     false,
						"errorText": "The source folder is successfully connected, you can use.",
					})
				} else {
					jsonResponse(w, map[string]interface{}{
						"error":     true,
						"errorText": "Error saving settings",
					})
				}
			} else {
				jsonResponse(w, map[string]interface{}{
					"error":     true,
					"errorText": "Invalid source data format",
				})
			}
		} else {
			if config.SourceFolderCreateIfNotExist {
				err := os.MkdirAll(config.Sourcefolder, 0777)
				if err != nil {
					jsonResponse(w, map[string]interface{}{
						"error":     true,
						"errorText": "Error creating directory",
					})
				} else {
					if CreateNewNotebooksFolder(config.Sourcefolder) {
						configGlobal.sourceFolder = config.Sourcefolder
						configGlobal.appInstalled = true
						if SaveConfig() {
							FindAllNotes()
							jsonResponse(w, map[string]interface{}{
								"error":     false,
								"errorText": "A new notebook was successfully created, you can use.",
							})
						} else {
							jsonResponse(w, map[string]interface{}{
								"error":     true,
								"errorText": "Error saving settings",
							})
						}
					} else {
						jsonResponse(w, map[string]interface{}{
							"error":     true,
							"errorText": "Error initializing a new notebook",
						})
					}
				}
			}
		}
	})

	r.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"result": "pong"})
	})

	r.HandleFunc("/api/config.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			CheckNewNotes  string `json:"atStartCheckNewNotes"`
			MCPEnabled     *bool  `json:"mcpEnabled"`
			MCPAllowWrite  *bool  `json:"mcpAllowWrite"`
			MCPRegenToken  bool   `json:"mcpRegenerateToken"`
		}
		readJSON(r, &request)

		switch request.CheckNewNotes {
		case "true":
			configGlobal.atStartCheckNewNotes = true
		case "false":
			configGlobal.atStartCheckNewNotes = false
		}
		if request.MCPEnabled != nil {
			configGlobal.mcpEnabled = *request.MCPEnabled
		}
		if request.MCPAllowWrite != nil {
			configGlobal.mcpAllowWrite = *request.MCPAllowWrite
		}
		// Token is generated on first enable (or explicit regeneration) so
		// MCP clients always have a non-empty bearer to send.
		if request.MCPRegenToken || (configGlobal.mcpEnabled && configGlobal.mcpToken == "") {
			configGlobal.mcpToken = RandStringBytes(32)
		}

		SaveConfig()

		jsonResponse(w, map[string]interface{}{
			"installed":            configGlobal.appInstalled,
			"sourceFolder":         configGlobal.sourceFolder,
			"requestIndexing":      configGlobal.requestIndexing,
			"atStartCheckNewNotes": configGlobal.atStartCheckNewNotes,
			"mcpEnabled":           configGlobal.mcpEnabled,
			"mcpAllowWrite":        configGlobal.mcpAllowWrite,
			"mcpToken":             configGlobal.mcpToken,
		})
	})

	r.HandleFunc("/api/favorites.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Action string `json:"action"`
			UUID   string `json:"UUID"`
		}
		readJSON(r, &request)

		switch request.Action {
		case "add":
			err := FavoritesDB.Set([]byte(request.UUID), []byte(""))
			checkQuiet(err)
		case "remove":
			FavoritesDB.Del([]byte(request.UUID))
		}
		var favoritesList []string
		checkQuiet(FavoritesDB.Keys(func(FavoriteID []byte) error {
			favoritesList = append(favoritesList, string(FavoriteID))
			return nil
		}))

		jsonResponse(w, favoritesList)
	})

	r.Get("/api/notebooks.json", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, listNotebooks())
	})

	r.Get("/api/tags.json", func(w http.ResponseWriter, r *http.Request) {
		// The HTTP API exposes URL-escaped tag names (the frontend uses them
		// as path segments); listTags returns raw names.
		TagsCloud := listTags()
		for i := range TagsCloud {
			TagsCloud[i].URL = url.PathEscape(TagsCloud[i].URL)
		}
		jsonResponse(w, TagsCloud)
	})

	r.HandleFunc("/api/notes_at_notebook.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			NotebookID string `json:"NotebookID"`
		}
		readJSON(r, &request)
		if request.NotebookID == "" {
			jsonResponse(w, map[string]interface{}{})
			return
		}
		jsonResponse(w, listNotesAtNotebook(request.NotebookID))
	})

	r.HandleFunc("/api/statistic.json", func(w http.ResponseWriter, r *http.Request) {
		var dateFirst int32 = 2147483647
		var dateLast int32
		var dateSkip = int32(time.Now().Unix()) - (60 * 60 * 24 * 365 * 2)
		var tagsCount = make(map[int]int)
		var chartsUpdatedDate = make(map[string]int)
		checkQuiet(NoteDB.Scan(func(NoteID, data []byte) error {
			var note NoteType
			if err := json.Unmarshal(data, &note); err != nil {
				return err
			}
			if note.CreatedAt < dateFirst {
				dateFirst = note.CreatedAt
			}
			if note.UpdatedAt > dateLast {
				dateLast = note.UpdatedAt
			}

			tagsCount[len(note.Tags)]++
			if note.UpdatedAt >= dateSkip {
				chartsUpdatedDate[time.Unix(int64(note.UpdatedAt), 0).Format("2006-01-02")]++
			}
			return nil
		}))

		dataSize, _ := DirSize2(configGlobal.dataDir)

		jsonResponse(w, map[string]interface{}{
			"dateFirst":         dateFirst,
			"dateLast":          dateLast,
			"tagsCount":         tagsCount,
			"chartsUpdatedDate": chartsUpdatedDate,
			"dataSize":          dataSize,
		})
	})

	// Reload data
	r.HandleFunc("/api/refresh_data.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Action string `json:"action"`
		}
		readJSON(r, &request)
		if request.Action == "reload" && searchStatus.Status != "indexing" && searchStatus.Status != "refresh" {
			searchStatus.Status = "refresh"
			FindAllNotes()
			searchStatus.Status = "idle"
		} else if request.Action == "reloadAll" && searchStatus.Status != "indexing" && searchStatus.Status != "refresh" {
			searchStatus.Status = "refresh"

			ss.index.Close()
			time.Sleep(1 * time.Second)
			indexName, _ := filepath.Abs(configGlobal.dataDir + "/search.bleve")
			os.RemoveAll(indexName)
			time.Sleep(1 * time.Second)
			index, err := bleve.Open(indexName)
			if errors.Is(err, bleve.ErrorIndexPathDoesNotExist) {
				index, err = bleve.New(indexName, ss.buildMapping())
			}
			check(err, "Can not initialize search database")
			ss.index = index
			ss.batch = index.NewBatch()

			// The search index was deleted, so every note must be re-indexed.
			// NB: bbolt deadlocks on a write inside a read transaction, so
			// collect the updates during Scan and apply them afterwards.
			type noteUpdate struct {
				id   []byte
				data []byte
			}
			var updates []noteUpdate
			checkQuiet(NoteDB.Scan(func(NoteID, data []byte) error {
				var note NoteType
				if err := json.Unmarshal(data, &note); err != nil {
					return nil
				}
				note.SearchIndex = false
				enc, err := json.Marshal(note)
				if err != nil {
					return nil
				}
				updates = append(updates, noteUpdate{append([]byte(nil), NoteID...), enc})
				return nil
			}))
			for _, u := range updates {
				checkQuiet(NoteDB.Set(u.id, u.data))
			}

			FindAllNotes()

			configGlobal.requestIndexing = true
			SaveConfig()

			searchStatus.Status = "idle"

			go indexingAllNotes()
			time.Sleep(3 * time.Second)
		}
		jsonResponse(w, map[string]string{"status": "done"})
	})

	// Data optimization
	r.HandleFunc("/api/optimization.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Action string `json:"action"`
		}
		readJSON(r, &request)
		if request.Action == "start" && optimizationStatus.Status != "processing" {
			optimizationStatus.Status = "processing"
			go optimizeAllNotes()
		} else if optimizationStatus.Status == "" {
			optimizationStatus.Status = "idle"
		}
		jsonResponse(w, map[string]interface{}{
			"status":       optimizationStatus.Status,
			"notesCurrent": optimizationStatus.NotesCurrent,
			"notesTotal":   optimizationStatus.NotesTotal,
		})
	})

	// Search index
	r.HandleFunc("/api/search_index.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Action string `json:"action"`
		}
		readJSON(r, &request)
		if request.Action == "start" && searchStatus.Status != "indexing" {
			go indexingAllNotes()
			time.Sleep(3 * time.Second)
		}
		jsonResponse(w, map[string]interface{}{
			"status":       searchStatus.Status,
			"notesCurrent": searchStatus.NotesCurrent,
			"notesTotal":   searchStatus.NotesTotal,
		})
	})

	// Notebook edit
	r.HandleFunc("/api/notebook_edit.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Action string `json:"action"`
			UUID   string `json:"uuid"`
			Title  string `json:"title"`
		}
		readJSON(r, &request)
		editNotebook(request.Action, request.UUID, request.Title)
		jsonResponse(w, map[string]interface{}{})
	})

	// Search
	r.HandleFunc("/api/search.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Text string `json:"text"`
		}
		readJSON(r, &request)
		jsonResponse(w, searchNotes(request.Text))
	})

	r.HandleFunc("/api/notes_with_tag.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			TagName string `json:"tag"`
		}
		readJSON(r, &request)
		if request.TagName != "" {
			jsonResponse(w, listNotesByTag(request.TagName))
		} else {
			jsonResponse(w, map[string]interface{}{})
		}
	})

	r.HandleFunc("/api/note.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			NoteID string `json:"NoteID"`
			Raw    bool   `json:"raw"`
		}
		readJSON(r, &request)
		noteData := loadNote(request.NoteID, request.Raw, r.Host)
		if noteData == nil {
			jsonResponse(w, map[string]interface{}{})
			return
		}
		jsonResponse(w, noteData)
	})

	r.HandleFunc("/api/note_edit.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Title        string   `json:"title"`
			URL          string   `json:"url"`
			UUID         string   `json:"uuid"`
			Type         string   `json:"type"`
			Content      string   `json:"content"`
			Tags         []string `json:"tags"`
			ContentState string   `json:"content_state"`
		}
		readJSON(r, &request)

		result, err := saveNote(saveNoteParams{
			UUID:         request.UUID,
			Title:        request.Title,
			URL:          request.URL,
			Type:         request.Type,
			Content:      request.Content,
			Tags:         request.Tags,
			ContentState: request.ContentState,
		})
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, result)
	})

	// Fetches a web page server-side so the frontend can run defuddle on it
	// (direct browser fetch is blocked by CORS on most sites).
	r.HandleFunc("/api/fetch.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			URL string `json:"url"`
		}
		readJSON(r, &request)
		target := strings.TrimSpace(request.URL)
		if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
			target = "https://" + target
		}
		parsed, err := url.Parse(target)
		if err != nil || parsed.Host == "" {
			jsonResponse(w, map[string]interface{}{"error": "bad url"})
			return
		}
		target = normalizeFetchURL(target)
		req, err := http.NewRequest("GET", target, nil)
		if err != nil {
			jsonResponse(w, map[string]interface{}{"error": err.Error()})
			return
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			jsonResponse(w, map[string]interface{}{"error": err.Error()})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			jsonResponse(w, map[string]interface{}{"error": "HTTP " + strconv.Itoa(resp.StatusCode), "status": resp.StatusCode})
			return
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 15*1024*1024))
		if err != nil {
			jsonResponse(w, map[string]interface{}{"error": err.Error()})
			return
		}
		jsonResponse(w, map[string]interface{}{"html": string(body), "url": resp.Request.URL.String()})
	})

	// Serves a CodePen pen as a standalone, framable HTML document, with a
	// two-level (memory + disk) cache. See codepen.go for the proxy details.
	r.HandleFunc("/api/codepen/{user}/{slug}", func(w http.ResponseWriter, r *http.Request) {
		serveCodePenEmbed(w, r, chi.URLParam(r, "user"), chi.URLParam(r, "slug"))
	})

	r.HandleFunc("/api/cleanup_html.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Content string `json:"content"`
			Title   string `json:"title"`
		}
		readJSON(r, &request)
		jsonResponse(w, map[string]string{"content": ClearHTML(request.Content, request.Title)})
	})

	// Drops the derived webp image cache (see webp.go); originals are untouched.
	r.HandleFunc("/api/webp_cache_clear.json", func(w http.ResponseWriter, r *http.Request) {
		err := clearWebPCache()
		if err != nil {
			jsonResponse(w, map[string]interface{}{"status": "error", "error": err.Error()})
			return
		}
		jsonResponse(w, map[string]interface{}{"status": "done"})
	})

	r.HandleFunc("/api/tag_edit.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Action string `json:"action"`
			URL    string `json:"url"`
			Title  string `json:"title"`
		}
		readJSON(r, &request)
		request.URL, _ = url.PathUnescape(request.URL)
		editTag(request.Action, request.URL, request.Title)
		jsonResponse(w, map[string]interface{}{})
	})

	r.HandleFunc("/api/note_move.json", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Action string `json:"action"`
			UUID   string `json:"uuid"`
			Target string `json:"target"`
		}
		readJSON(r, &request)
		switch {
		case request.UUID != "" && request.Action == "move":
			if err := moveNote(request.UUID, request.Target); err != nil {
				go showNotificationDialog("Error! Can not move note: " + err.Error())
			}
		case request.UUID != "" && request.Action == "delete":
			if err := deleteNote(request.UUID); err != nil {
				go showNotificationDialog("Error! Can not delete note: " + err.Error())
			}
		default:
			jsonResponse(w, map[string]interface{}{})
			return
		}
		jsonResponse(w, map[string]interface{}{})
	})

	fmt.Println("Server started on port " + configGlobal.cmdPort)
	log.Fatal(http.ListenAndServe(":"+configGlobal.cmdPort, r))
}

// normalizeFetchURL rewrites URLs of sites that moved domains, so server-side
// fetch can reach their new home. habrahabr.ru is DNS-dead (resolves to
// 0.0.0.0 on some networks) while the content lives on at habr.com/ru/<path>
// (habr's own redirects take it from there).
func normalizeFetchURL(target string) string {
	parsed, err := url.Parse(target)
	if err != nil {
		return target
	}
	switch strings.ToLower(parsed.Hostname()) {
	case "habrahabr.ru":
		parsed.Host = "habr.com"
		if !strings.HasPrefix(parsed.Path, "/ru/") && parsed.Path != "/ru" {
			parsed.Path = "/ru" + parsed.Path
		}
	}
	return parsed.String()
}
