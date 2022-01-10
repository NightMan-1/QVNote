package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/store/goleveldb"
	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/ledisdb/ledisdb/ledis"
	"github.com/mattn/go-colorable"
)

func WebServer(webserverChan chan bool) { //nolint:gocyclo
	app := iris.New()
	app.Use(iris.Compression)
	app.Use(recover.New())
	app.Use(logger.New())

	// fix console colors
	app.Logger().SetOutput(colorable.NewColorableStdout())

	app.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "OPTIONS", "POST", "PATCH", "PUT", "DELETE", "HEAD"},
	}))

	dirOptions := iris.DirOptions{ShowList: false, Compress: true, IndexName: "index.html", ShowHidden: false}

	//app.StaticWeb("/static", configGlobal.execDir + "/templates/static")
	//app.StaticEmbedded("/", "./templates", Asset, AssetNames)
	//app.HandleDir("/", "./templates", iris.DirOptions{Asset: Asset, AssetInfo: AssetInfo, AssetNames: AssetNames, Compress: false, ShowHidden: false, ShowList: false})
	app.HandleDir("/", AssetFile(), dirOptions)

	//app.HandleDir("/", iris.Dir("./templates"), dirOptions)

	var err error

	// Register custom handler for specific http errors.
	app.OnErrorCode(iris.StatusInternalServerError, func(ctx iris.Context) {
		// .Values are used to communicate between handlers, middleware.
		errMessage := ctx.Values().GetString("error")
		if errMessage != "" {
			ctx.Writef("Internal server error: %s", errMessage)
			return
		}

		ctx.Writef("(Unexpected) internal server error")
	})

	app.Handle("ANY", "/", func(ctx iris.Context) {
		data, _ := Asset("index.html")
		ctx.HTML(string(data))
	})

	app.Handle("ANY", "/notes/*", func(ctx iris.Context) {
		data, _ := Asset("index.html")
		ctx.HTML(string(data))
	})

	app.Handle("ANY", "/tags/*", func(ctx iris.Context) {
		data, _ := Asset("index.html")
		ctx.HTML(string(data))
	})

	app.OnErrorCode(404, func(ctx iris.Context) {
		//data, _ := Asset("index.html")
		//fmt.Println(string(data))
		//ctx.StatusCode(200)
		//ctx.HTML(string(data))
		//ctx.WriteString(string(data))
		ctx.Redirect("/", iris.StatusPermanentRedirect)
	})

	// for installation
	app.Handle("ANY", "/api/config.write", func(ctx iris.Context) {
		var config struct {
			Sourcefolder                 string `json:"sourceFolder"`
			SourceFolderCreateIfNotExist bool   `json:"sourceFolderCreateIfNotExist"`
			StartingMode                 string `json:"startingMode"`
		}
		ctx.ReadJSON(&config)
		if _, err := os.Stat(config.Sourcefolder); err == nil {
			if CheckNotebooksFolderStructure(config.Sourcefolder) {
				configGlobal.sourceFolder = config.Sourcefolder
				configGlobal.appInstalled = true
				configGlobal.appStartingMode = config.StartingMode
				fmt.Println(configGlobal.appStartingMode)
				if SaveConfig() {
					FindAllNotes()
					ctx.JSON(iris.Map{
						"error":     false,
						"errorText": "The source folder is successfully connected, you can use.",
					})
				} else {
					ctx.JSON(iris.Map{
						"error":     true,
						"errorText": "Error saving settings",
					})
				}

			} else {
				ctx.JSON(iris.Map{
					"error":     true,
					"errorText": "Invalid source data format",
				})
			}

		} else {
			if config.SourceFolderCreateIfNotExist {
				err = os.MkdirAll(config.Sourcefolder, 0777)
				if err != nil {
					ctx.JSON(iris.Map{
						"error":     true,
						"errorText": "Error creating directory",
					})
				} else {
					//создание структуры для новых заметок
					if CreateNewNotebooksFolder(config.Sourcefolder) {
						configGlobal.sourceFolder = config.Sourcefolder
						configGlobal.appInstalled = true
						if SaveConfig() {
							FindAllNotes()
							ctx.JSON(iris.Map{
								"error":     false,
								"errorText": "A new notebook was successfully created, you can use.",
							})

						} else {
							ctx.JSON(iris.Map{
								"error":     true,
								"errorText": "Error saving settings",
							})
						}

					} else {
						ctx.JSON(iris.Map{
							"error":     true,
							"errorText": "Error initializing a new notebook",
						})
					}

				}
			}
		}
	})

	app.Handle("ANY", "/api/exit", func(ctx iris.Context) {
		fmt.Println("Good buy!")
		if systrayProcess != nil {
			systrayProcess.Process.Kill()
		}
		os.Exit(0)
	})

	app.Handle("ANY", "/api/ping", func(ctx iris.Context) {
		ctx.JSON(iris.Map{"result": "pong"})
	})

	app.Handle("ANY", "/api/config.json", func(ctx iris.Context) {
		var request struct {
			OpenBrowser   string `json:"atStartOpenBrowser"`
			CheckNewNotes string `json:"atStartCheckNewNotes"`
			ShowConsole   string `json:"atStartShowConsole"`
			PostEditor    string `json:"postEditor"`
			StartingMode  string `json:"startingMode"`
		}
		ctx.ReadJSON(&request)

		switch request.OpenBrowser {
		case "true":
			configGlobal.atStartOpenBrowser = true
		case "false":
			configGlobal.atStartOpenBrowser = false
		}

		switch request.CheckNewNotes {
		case "true":
			configGlobal.atStartCheckNewNotes = true
		case "false":
			configGlobal.atStartCheckNewNotes = false
		}

		switch request.ShowConsole {
		case "true":
			configGlobal.atStartShowConsole = true
		case "false":
			configGlobal.atStartShowConsole = false
		}

		if request.PostEditor != "" {
			configGlobal.postEditor = request.PostEditor
		}

		if request.StartingMode != "" {
			configGlobal.appStartingMode = request.StartingMode
		}

		SaveConfig()

		ctx.JSON(iris.Map{
			"installed":            configGlobal.appInstalled,
			"sourceFolder":         configGlobal.sourceFolder,
			"requestIndexing":      configGlobal.requestIndexing,
			"atStartOpenBrowser":   configGlobal.atStartOpenBrowser,
			"atStartCheckNewNotes": configGlobal.atStartCheckNewNotes,
			"atStartShowConsole":   configGlobal.atStartShowConsole,
			"postEditor":           configGlobal.postEditor,
			"startingMode":         configGlobal.appStartingMode,
		})
	})

	app.Handle("ANY", "/api/favorites.json", func(ctx iris.Context) {

		var request struct {
			Action string `json:"action"`
			UUID   string `json:"UUID"`
		}
		ctx.ReadJSON(&request)

		switch request.Action {
		case "add":
			err = FavoritesDB.Set([]byte(request.UUID), []byte(""))
			checkQuiet(err)

		case "remove":
			FavoritesDB.Del([]byte(request.UUID))
		}
		cursor := []byte(nil)
		var favoritesList []string
		for {
			allDBData, err := FavoritesDB.Scan(ledis.KV, cursor, 0, false, "")
			if err != nil || len(allDBData) == 0 {
				break
			}
			for _, FavoriteID := range allDBData {
				cursor = FavoriteID
				favoritesList = append(favoritesList, string(FavoriteID))

			}
		}

		ctx.JSON(favoritesList)

	})

	//display images
	app.Get("/resources/{notebookUUID:string}/{noteUUID:string}/{image:string}", func(ctx iris.Context) {
		notebookUUID := ctx.Params().Get("notebookUUID")
		noteUUID := ctx.Params().Get("noteUUID")
		image := ctx.Params().Get("image")
		imageFile, _ := filepath.Abs(configGlobal.sourceFolder + "/" + notebookUUID + ".qvnotebook/" + noteUUID + ".qvnote/resources/" + image)
		if _, err := os.Stat(imageFile); err == nil {
			ctx.ServeFile(imageFile)
		} else {
			ctx.NotFound()
		}
	})

	app.Get("/api/notebooks.json", func(ctx iris.Context) {
		cursor := []byte(nil)
		var noteBooksList []NoteBookTypeAPI
		for {
			allDBData, err := NoteBookDB.Scan(ledis.KV, cursor, 0, false, "")
			if err != nil || len(allDBData) == 0 {
				break
			}
			for _, NoteBookID := range allDBData {
				cursor = NoteBookID
				data, _ := NoteBookDB.Get(NoteBookID)
				var notebookData NoteBookType
				err := json.Unmarshal(data, &notebookData)
				checkQuiet(err)
				noteBooksList = append(noteBooksList, NoteBookTypeAPI{notebookData.UUID, notebookData.Name, len(notebookData.Notes)})

			}
		}

		sort.Slice(noteBooksList, func(i, j int) bool {
			return strings.ToLower(noteBooksList[i].Name) < strings.ToLower(noteBooksList[j].Name)
		})
		ctx.JSON(noteBooksList)
	})

	app.Get("/api/tags.json", func(ctx iris.Context) {
		cursor := []byte(nil)
		var TagsCloud []TagsListStruct
		for {
			allDBData, err := TagsDB.Scan(ledis.KV, cursor, 0, false, "")
			if err != nil || len(allDBData) == 0 {
				break
			}
			for _, TagID := range allDBData {
				cursor = TagID
				data, _ := TagsDB.Get(TagID)
				var tagsData []string
				err := json.Unmarshal(data, &tagsData)
				checkQuiet(err)
				TagsCloud = append(TagsCloud, TagsListStruct{len(tagsData), strings.Trim(string(TagID), " "), url.PathEscape(string(TagID))})

			}
		}
		sort.Slice(TagsCloud, func(i, j int) bool {
			return strings.ToLower(TagsCloud[i].Name) < strings.ToLower(TagsCloud[j].Name)
		})

		ctx.JSON(TagsCloud)
	})

	app.Handle("ANY", "/api/notes_at_notebook.json", func(ctx iris.Context) {
		var request struct {
			NotebookID string `json:"NotebookID"`
		}
		ctx.ReadJSON(&request)
		switch {
		case request.NotebookID == "Favorites":
			var NotesList []NoteTypeAPI
			cursor := []byte(nil)
			for {
				favoritesDBData, err := FavoritesDB.Scan(ledis.KV, cursor, 0, false, "")
				if err != nil || len(favoritesDBData) == 0 {
					break
				}
				for _, NoteID := range favoritesDBData {
					cursor = NoteID
					data, _ := NoteDB.Get(NoteID)
					var note NoteTypeAPI
					err := json.Unmarshal(data, &note)
					checkQuiet(err)
					note.NoteBookUUID = "Favorites"
					NotesList = append(NotesList, note)
				}
			}
			sort.Slice(NotesList, func(i, j int) bool {
				return NotesList[i].UpdatedAt > NotesList[j].UpdatedAt
			})

			ctx.JSON(NotesList)
		case request.NotebookID == "Allnotes":
			var NotesList []NoteTypeAPI
			cursor := []byte(nil)
			for {
				allDBData, err := NoteDB.Scan(ledis.KV, cursor, 0, false, "")
				if err != nil || len(allDBData) == 0 {
					break
				}
				for _, NoteID := range allDBData {
					cursor = NoteID
					data, _ := NoteDB.Get(NoteID)
					var note NoteTypeAPI
					err := json.Unmarshal(data, &note)
					checkQuiet(err)
					NotesList = append(NotesList, note)
				}
			}
			sort.Slice(NotesList, func(i, j int) bool {
				return NotesList[i].UpdatedAt > NotesList[j].UpdatedAt
			})
			ctx.JSON(NotesList)
		case len(request.NotebookID) > 0:
			var NotesList []NoteTypeAPI
			data, _ := NoteBookDB.Get([]byte(request.NotebookID))
			var notebookData NoteBookType
			err := json.Unmarshal(data, &notebookData)
			checkQuiet(err)
			for NoteBookID := range notebookData.Notes {
				data, _ := NoteDB.Get([]byte(NoteBookID))
				var note NoteTypeAPI
				err := json.Unmarshal(data, &note)
				checkQuiet(err)
				NotesList = append(NotesList, note)
			}
			sort.Slice(NotesList, func(i, j int) bool {
				return NotesList[i].UpdatedAt > NotesList[j].UpdatedAt
			})
			ctx.JSON(NotesList)

		default:
			ctx.JSON(iris.Map{})
		}

	})

	app.Handle("ANY", "/api/statistic.json", func(ctx iris.Context) {
		var dateFirst int32 = 2147483647
		var dateLast int32
		var dateSkip = int32(time.Now().Unix()) - (60 * 60 * 24 * 365 * 2)
		var tagsCount = make(map[int]int)
		//var chartsCreatedDate  = make(map[string]int)
		var chartsUpdatedDate = make(map[string]int)
		cursor := []byte(nil)
		for {
			allDBData, err := NoteDB.Scan(ledis.KV, cursor, 0, false, "")
			if err != nil || len(allDBData) == 0 {
				break
			}
			for _, NoteID := range allDBData {
				cursor = NoteID
				data, _ := NoteDB.Get(NoteID)
				var note NoteType
				err := json.Unmarshal(data, &note)
				checkQuiet(err)
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

			}
		}

		//sourceSize, _ := DirSize2(configGlobal.sourceFolder) //take long time
		dataSize, _ := DirSize2(configGlobal.dataDir)

		ctx.JSON(iris.Map{
			"dateFirst": dateFirst,
			"dateLast":  dateLast,
			"tagsCount": tagsCount,
			//"chartsCreatedDate": chartsCreatedDate,
			"chartsUpdatedDate": chartsUpdatedDate,
			//"sourceSize": humanize.Bytes(uint64(sourceSize)),
			"dataSize": humanize.Bytes(uint64(dataSize)),
		})
	})

	//reload data
	app.Handle("ANY", "/api/refresh_data.json", func(ctx iris.Context) {
		var request struct {
			Action string `json:"action"`
		}
		ctx.ReadJSON(&request)
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
			if err == bleve.ErrorIndexPathDoesNotExist {
				mapping := ss.buildMapping()
				kvStore := goleveldb.Name
				kvConfig := map[string]interface{}{
					"create_if_missing": true,
					//	"write_buffer_size":         536870912,
					//	"lru_cache_capacity":        536870912,
					//	"bloom_filter_bits_per_key": 10,
				}

				index, err = bleve.NewUsing(indexName, mapping, "upside_down", kvStore, kvConfig)
			}
			check(err, "Can not initialize search database")
			ss.index = index
			ss.batch = index.NewBatch()

			FindAllNotes()

			configGlobal.requestIndexing = true
			SaveConfig()

			searchStatus.Status = "idle"

			go indexingAllNotes()
			time.Sleep(3 * time.Second)

		}
		ctx.JSON(iris.Map{"status": "done"})
	})

	//data optimization
	app.Handle("ANY", "/api/optimization.json", func(ctx iris.Context) {
		var request struct {
			Action string `json:"action"`
		}
		ctx.ReadJSON(&request)
		if request.Action == "start" && optimizationStatus.Status != "processing" {
			optimizationStatus.Status = "processing"
			go optimizeAllNotes()

		} else if optimizationStatus.Status == "" {
			optimizationStatus.Status = "idle"
		}
		ctx.JSON(iris.Map{"status": optimizationStatus.Status, "notesCurrent": optimizationStatus.NotesCurrent, "notesTotal": optimizationStatus.NotesTotal})

	})

	//search index
	app.Handle("ANY", "/api/search_index.json", func(ctx iris.Context) {
		var request struct {
			Action string `json:"action"`
		}
		ctx.ReadJSON(&request)
		if request.Action == "start" && searchStatus.Status != "indexing" {
			go indexingAllNotes()
			time.Sleep(3 * time.Second)
		}
		ctx.JSON(iris.Map{"status": searchStatus.Status, "notesCurrent": searchStatus.NotesCurrent, "notesTotal": searchStatus.NotesTotal})
	})

	//notebook_edit
	app.Handle("ANY", "/api/notebook_edit.json", func(ctx iris.Context) {
		var request struct {
			Action string `json:"action"`
			UUID   string `json:"uuid"`
			Title  string `json:"title"`
		}
		ctx.ReadJSON(&request)
		switch {
		case request.Action == "rename" && request.UUID != "":
			//update file
			var meta struct {
				Name string `json:"name"`
				UUID string `json:"uuid"`
			}
			meta.Name = request.Title
			meta.UUID = request.UUID
			metaJSON, _ := json.Marshal(meta)
			jsonFile, _ := filepath.Abs(configGlobal.sourceFolder + "/" + request.UUID + ".qvnotebook/meta.json")
			err = ioutil.WriteFile(jsonFile, metaJSON, 0644)
			checkQuiet(err)

			//update database
			data, _ := NoteBookDB.Get([]byte(request.UUID))
			var notebookData NoteBookType
			json.Unmarshal(data, &notebookData)
			notebookData.Name = request.Title
			enc, err := json.Marshal(notebookData)
			checkQuiet(err)
			NoteBookDB.Set([]byte(request.UUID), enc)
		case request.Action == "new" && request.UUID == "":
			u1 := strings.ToUpper(uuid.Must(uuid.NewRandom()).String())

			//new file
			notebookDir, _ := filepath.Abs(configGlobal.sourceFolder + "/" + u1 + ".qvnotebook")
			metaFile, _ := filepath.Abs(notebookDir + "/meta.json")
			var meta struct {
				Name string `json:"name"`
				UUID string `json:"uuid"`
			}
			meta.Name = request.Title
			meta.UUID = u1
			metaJSON, _ := json.MarshalIndent(meta, "", "  ")
			os.MkdirAll(notebookDir, 0755)
			err = ioutil.WriteFile(metaFile, metaJSON, 0644)
			checkQuiet(err)

			//update database
			var notebookNew NoteBookType
			notebookNew.Name = request.Title
			notebookNew.UUID = u1
			notebookNew.Notes = make(map[string]int64)
			enc, err := json.Marshal(notebookNew)
			checkQuiet(err)
			NoteBookDB.Set([]byte(u1), enc)
		case request.Action == "remove" && request.UUID != "" && request.UUID != "Inbox" && request.UUID != "Trash":
			data, _ := NoteBookDB.Get([]byte("Trash"))
			var notebookDataTrash NoteBookType
			json.Unmarshal(data, &notebookDataTrash)

			data, _ = NoteBookDB.Get([]byte(request.UUID))
			var notebookData NoteBookType
			json.Unmarshal(data, &notebookData)
			if notebookData.UUID != "" {
				canDelete := true
				for noteUUID := range notebookData.Notes {
					var note NoteType
					data, _ := NoteDB.Get([]byte(noteUUID))
					json.Unmarshal(data, &note)
					noteDirSrc, _ := filepath.Abs(configGlobal.sourceFolder + "/" + request.UUID + ".qvnotebook/" + noteUUID + ".qvnote")
					noteDirDst, _ := filepath.Abs(configGlobal.sourceFolder + "/Trash.qvnotebook/" + noteUUID + ".qvnote")

					err = CopyDir(noteDirSrc, noteDirDst)
					if err == nil {
						note.NoteBookUUID = "Trash"
						enc, _ := json.Marshal(note)
						NoteDB.Set([]byte(noteUUID), enc)

						notebookDataTrash.Notes[noteUUID] = time.Now().Unix()
						enc, _ = json.Marshal(notebookDataTrash)
						NoteBookDB.Set([]byte("Trash"), enc)

						os.RemoveAll(noteDirSrc)

					} else {
						canDelete = false
					}
				}
				if canDelete {
					srcFolder, _ := filepath.Abs(configGlobal.sourceFolder + "/" + request.UUID + ".qvnotebook/")
					os.RemoveAll(srcFolder)
					NoteBookDB.Del([]byte(request.UUID))
				}
			}
		}

		ctx.JSON(iris.Map{})
	})

	//search
	app.Handle("ANY", "/api/search.json", func(ctx iris.Context) {
		var request struct {
			Text string `json:"text"`
		}
		ctx.ReadJSON(&request)

		var NotesList []SearchResult
		NoteListDedup := make(map[string]bool)
		if len(request.Text) >= 3 {
			query := bleve.NewQueryStringQuery(queryStem(request.Text))
			searchRequest := bleve.NewSearchRequestOptions(query, 500, 0, false)
			searchResult, _ := ss.index.Search(searchRequest)
			var noteShort SearchResult
			for _, item := range searchResult.Hits {
				data, _ := NoteDB.Get([]byte(item.ID))
				err := json.Unmarshal(data, &noteShort)
				checkQuiet(err)
				if _, ok := NoteListDedup[noteShort.UUID]; ok {
					//duplicate detected
				} else {
					NoteListDedup[noteShort.UUID] = true
					NotesList = append(NotesList, noteShort)
				}

			}

		}
		ctx.JSON(NotesList)
	})

	app.Handle("ANY", "/api/notes_with_tag.json", func(ctx iris.Context) {
		var request struct {
			TagName string `json:"tag"`
		}
		ctx.ReadJSON(&request)
		if request.TagName != "" {
			var NotesList []NoteTypeAPI
			data, _ := TagsDB.Get([]byte(request.TagName))
			var notesListTMP []string
			err := json.Unmarshal(data, &notesListTMP)
			checkQuiet(err)
			for _, tagID := range notesListTMP {
				data, _ := NoteDB.Get([]byte(tagID))
				var note NoteTypeAPI
				err := json.Unmarshal(data, &note)
				checkQuiet(err)
				NotesList = append(NotesList, note)
			}
			sort.Slice(NotesList, func(i, j int) bool {
				return NotesList[i].UpdatedAt > NotesList[j].UpdatedAt
			})
			ctx.JSON(NotesList)

		} else {
			ctx.JSON(iris.Map{})
		}
	})

	app.Handle("ANY", "/api/note.json", func(ctx iris.Context) {
		var request struct {
			NoteID string `json:"NoteID"`
		}
		ctx.ReadJSON(&request)
		if request.NoteID != "" {
			//OptimizeResources(request.NoteID)

			data, _ := NoteDB.Get([]byte(request.NoteID))
			var noteData NoteTypeWithContentAPI
			err := json.Unmarshal(data, &noteData)
			checkQuiet(err)

			contentDir := configGlobal.sourceFolder + "/" + noteData.NoteBookUUID + ".qvnotebook/" + noteData.UUID + ".qvnote"
			contentPath := contentDir + "/content.json"
			if _, err := os.Stat(contentPath); err == nil {
				jsonFile, err := os.Open(contentPath)
				checkQuiet(err)
				byteValue, _ := ioutil.ReadAll(jsonFile)
				var contentFile SearchContent
				json.Unmarshal(byteValue, &contentFile)
				jsonFile.Close()

				noteData.Content = ""
				for _, text := range contentFile.Cells {
					noteData.Content += text.Data
					noteData.ContentType = text.Type
				}
				noteData.Content = ClearHTML(noteData.Content)

				noteData.Content = FixNoteImagesLinks(noteData, noteData.Content, ctx)

				dataExists, _ := FavoritesDB.Exists([]byte(request.NoteID))
				if dataExists == 1 {
					noteData.Favorites = true
				}

				ctx.JSON(noteData)

			} else {
				ctx.JSON(iris.Map{})
			}

		} else {
			ctx.JSON(iris.Map{})
		}
	})

	app.Handle("ANY", "/api/note_edit.json", func(ctx iris.Context) {
		var request struct {
			Title   string   `json:"title"`
			URL     string   `json:"url"`
			UUID    string   `json:"uuid"`
			Type    string   `json:"type"`
			Content string   `json:"content"`
			Tags    []string `json:"tags"`
		}
		ctx.ReadJSON(&request)

		var noteUUID string
		var notebookUUID string
		var noteData NoteType
		if request.UUID == "" {
			noteUUID = strings.ToUpper(uuid.Must(uuid.NewRandom()).String())
			notebookUUID = "Inbox"
			noteData.NoteBookUUID = notebookUUID
			noteData.UUID = noteUUID

		} else {
			noteUUID = request.UUID
			data, _ := NoteDB.Get([]byte(noteUUID))
			json.Unmarshal(data, &noteData)
			notebookUUID = noteData.NoteBookUUID

		}

		noteData.Title = request.Title
		noteData.URL = request.URL
		noteData.SearchIndex = false
		// configGlobal.requestIndexing = true

		if request.UUID == "" {
			noteData.CreatedAt = int32(time.Now().Unix())
			noteData.UpdatedAt = noteData.CreatedAt

		} else {
			noteData.UpdatedAt = int32(time.Now().Unix())
		}
		if request.Type == "tinymce" {
			request.Type = "text"
		}

		//request.Content = ClearHTML(request.Content)

		// update file
		noteDir, _ := filepath.Abs(configGlobal.sourceFolder + "/" + notebookUUID + ".qvnotebook/" + noteUUID + ".qvnote")
		os.MkdirAll(noteDir, 0755)
		var meta struct {
			CreatedAt int32    `json:"created_at"`
			UpdatedAt int32    `json:"updated_at"`
			Tags      []string `json:"tags"`
			Title     string   `json:"title"`
			UUID      string   `json:"uuid"`
			URL       string   `json:"url_src"`
		}
		meta.CreatedAt = noteData.CreatedAt
		meta.UpdatedAt = noteData.UpdatedAt
		meta.Title = noteData.Title
		meta.UUID = noteData.UUID
		meta.URL = noteData.URL
		meta.Tags = request.Tags
		metaJSON, _ := json.MarshalIndent(meta, "", "  ")
		err = ioutil.WriteFile(noteDir+"/meta.json", metaJSON, 0644)
		checkQuiet(err)

		var content struct {
			Title string             `json:"title"`
			Cells []ContentCellsType `json:"cells"`
		}
		content.Title = noteData.Title
		content.Cells = make([]ContentCellsType, 1)
		content.Cells[0] = ContentCellsType{Type: request.Type, Data: request.Content}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.Encode(content)

		err = ioutil.WriteFile(noteDir+"/content.json", buf.Bytes(), 0644)
		checkQuiet(err)

		//remove old tags from cloud
		for _, tagID := range noteData.Tags {
			data, _ := TagsDB.Get([]byte(tagID))
			var notesListOld []string
			var notesListNew []string
			err := json.Unmarshal(data, &notesListOld)
			checkQuiet(err)
			for _, noteID := range notesListOld {
				if noteID != noteUUID {
					notesListNew = append(notesListNew, noteID)
				}
			}
			if len(notesListNew) == 0 {
				idT, err := TagsDB.Del([]byte(tagID))
				checkQuiet(err)
				fmt.Println(idT)
			} else {
				enc, err := json.Marshal(notesListNew)
				checkQuiet(err)
				TagsDB.Set([]byte(tagID), enc)
			}
		}

		//Add new tags to cloud
		for _, tagID := range request.Tags {
			data, _ := TagsDB.Get([]byte(tagID))
			dataString := string(data)
			var notesList []string
			if dataString == "" {
				//new tag
			} else {
				//exist tag
				err := json.Unmarshal(data, &notesList)
				checkQuiet(err)
			}
			notesList = append(notesList, noteUUID)

			enc, err := json.Marshal(notesList)
			checkQuiet(err)
			err = TagsDB.Set([]byte(tagID), enc)
			checkQuiet(err)
		}

		//add to search index
		addToIndex(noteDir+"/content.json", noteUUID)
		noteData.SearchIndex = true

		// update database
		noteData.Tags = request.Tags
		encNote, _ := json.Marshal(noteData)
		err = NoteDB.Set([]byte(noteUUID), encNote)
		checkQuiet(err)

		// add new note to inbox
		if request.UUID == "" {
			data, _ := NoteBookDB.Get([]byte("Inbox"))
			var notebookDataInbox NoteBookType
			json.Unmarshal(data, &notebookDataInbox)
			notebookDataInbox.Notes[noteUUID] = time.Now().Unix()
			encData, _ := json.Marshal(notebookDataInbox)
			NoteBookDB.Set([]byte("Inbox"), encData)
		}

		SaveConfig()

		//ctx.JSON(iris.Map{"NoteBookUUID": notebookUUID, "uuid": noteUUID, "html": request.Content})
		ctx.JSON(iris.Map{"NoteBookUUID": notebookUUID, "uuid": noteUUID})

	})

	app.Handle("ANY", "/api/cleanup_html.json", func(ctx iris.Context) {
		var request struct {
			Content string `json:"content"`
		}
		ctx.ReadJSON(&request)
		ctx.JSON(iris.Map{"content": ClearHTML(request.Content)})
	})

	app.Handle("ANY", "/api/tag_edit.json", func(ctx iris.Context) {
		var request struct {
			Action string `json:"action"`
			URL    string `json:"url"`
			Title  string `json:"title"`
		}
		ctx.ReadJSON(&request)
		request.URL, _ = url.PathUnescape(request.URL)

		if request.URL != "" || (request.Action == "rename" && request.URL != "" && request.URL != request.Title) {
			data, _ := TagsDB.Get([]byte(request.URL))
			if string(data) != "" {
				var tagsData []string
				err := json.Unmarshal(data, &tagsData)
				checkQuiet(err)
				for _, noteID := range tagsData {
					//change files
					dataNote, _ := NoteDB.Get([]byte(noteID))
					if string(dataNote) != "" {
						var note NoteType
						err := json.Unmarshal(dataNote, &note)
						checkQuiet(err)

						metaFile, _ := filepath.Abs(configGlobal.sourceFolder + "/" + note.NoteBookUUID + ".qvnotebook/" + note.UUID + ".qvnote/meta.json")

						jsonFile, err := os.Open(metaFile)
						if err == nil {
							byteValue, _ := ioutil.ReadAll(jsonFile)
							json.Unmarshal(byteValue, &note)
							jsonFile.Close()
							var tagsNew = make([]string, 0)
							for _, tagName := range note.Tags {
								if tagName != request.URL && tagName != request.Title {
									tagsNew = append(tagsNew, tagName)
								}
							}
							switch request.Action {
							case "rename":
								tagsNew = append(tagsNew, request.Title) //add new nag name
							case "remove":
								// do nothing
							}
							note.Tags = tagsNew

							//save file with meta data
							var meta struct {
								CreatedAt int32    `json:"created_at"`
								UpdatedAt int32    `json:"updated_at"`
								Tags      []string `json:"tags"`
								Title     string   `json:"title"`
								UUID      string   `json:"uuid"`
							}
							meta.CreatedAt = note.CreatedAt
							meta.UpdatedAt = note.UpdatedAt
							meta.Title = note.Title
							meta.UUID = note.UUID
							meta.Tags = note.Tags

							metaJSON, _ := json.MarshalIndent(meta, "", "  ")
							err = ioutil.WriteFile(metaFile, metaJSON, 0644)
							checkQuiet(err)

							//update NoteDB
							enc, _ := json.Marshal(note)
							NoteDB.Set([]byte(note.UUID), enc)

						}
					}
				}

				//save tags
				if request.Action == "remove" {
					//remove old date
					TagsDB.Del([]byte(request.URL))

				} else if request.Action == "rename" {
					//remove old date
					TagsDB.Del([]byte(request.URL))

					//add new data
					data, _ := TagsDB.Get([]byte(request.Title)) //check the existence of a new tag (required for merging)
					if string(data) != "" {
						var tagsDataExist []string
						err := json.Unmarshal(data, &tagsDataExist)
						checkQuiet(err)

						for _, tagName := range tagsDataExist {
							if !inArray(tagName, tagsData) {
								tagsData = append(tagsData, tagName)
							}
						}

					}

					enc, err := json.Marshal(tagsData)
					checkQuiet(err)
					TagsDB.Set([]byte(request.Title), enc)

				}
			}
		}

		ctx.JSON(iris.Map{})
	})

	app.Handle("ANY", "/api/note_move.json", func(ctx iris.Context) {
		var request struct {
			Action string `json:"action"`
			UUID   string `json:"uuid"`
			Target string `json:"target"`
		}
		ctx.ReadJSON(&request)
		switch {
		case request.UUID != "" && request.Action == "move":
			//get note info
			var note NoteType
			data, _ := NoteDB.Get([]byte(request.UUID))
			json.Unmarshal(data, &note)

			//get source notebook info
			var notebookSRC NoteBookType
			data, _ = NoteBookDB.Get([]byte(note.NoteBookUUID))
			json.Unmarshal(data, &notebookSRC)

			//get target notebook info
			var notebookDST NoteBookType
			data, _ = NoteBookDB.Get([]byte(request.Target))
			json.Unmarshal(data, &notebookDST)
			if notebookDST.UUID != "" {

				//move folder
				noteDirSrc, _ := filepath.Abs(configGlobal.sourceFolder + "/" + note.NoteBookUUID + ".qvnotebook/" + note.UUID + ".qvnote")
				noteDirDst, _ := filepath.Abs(configGlobal.sourceFolder + "/" + notebookDST.UUID + ".qvnotebook/" + note.UUID + ".qvnote")

				err = CopyDir(noteDirSrc, noteDirDst)
				if err == nil {

					//update database
					note.NoteBookUUID = notebookDST.UUID
					enc, _ := json.Marshal(note)
					NoteDB.Set([]byte(note.UUID), enc)

					delete(notebookSRC.Notes, note.UUID)
					encSRC, _ := json.Marshal(notebookSRC)
					NoteBookDB.Set([]byte(notebookSRC.UUID), encSRC)

					notebookDST.Notes[note.UUID] = time.Now().Unix()
					encDST, _ := json.Marshal(notebookDST)
					NoteBookDB.Set([]byte(notebookDST.UUID), encDST)

					os.RemoveAll(noteDirSrc)

				} else { //nolint:staticcheck
					go showNotificationDialog("Error! Can not move folder " + noteDirSrc + " to " + noteDirDst)
				}

			} else { //nolint:staticcheck
				go showNotificationDialog("Error! Notebook " + notebookDST.UUID + " not exist")
			}
		case request.UUID != "" && request.Action == "delete":
			//get note info
			var note NoteType
			data, _ := NoteDB.Get([]byte(request.UUID))
			json.Unmarshal(data, &note)

			//get source notebook info
			var notebookSRC NoteBookType
			data, _ = NoteBookDB.Get([]byte(note.NoteBookUUID))
			json.Unmarshal(data, &notebookSRC)

			if notebookSRC.UUID == "Trash" {
				//delete
				delete(notebookSRC.Notes, note.UUID)
				encSRC, _ := json.Marshal(notebookSRC)
				NoteBookDB.Set([]byte(notebookSRC.UUID), encSRC)

				NoteDB.Del([]byte(note.UUID))

				noteDirSrc, _ := filepath.Abs(configGlobal.sourceFolder + "/" + note.NoteBookUUID + ".qvnotebook/" + note.UUID + ".qvnote")
				os.RemoveAll(noteDirSrc)
				ss.index.Delete(note.UUID) // delete from search index

			} else {
				//move to trash
				var notebookDST NoteBookType
				data, _ = NoteBookDB.Get([]byte("Trash"))
				json.Unmarshal(data, &notebookDST)
				noteDirSrc, _ := filepath.Abs(configGlobal.sourceFolder + "/" + note.NoteBookUUID + ".qvnotebook/" + note.UUID + ".qvnote")
				noteDirDst, _ := filepath.Abs(configGlobal.sourceFolder + "/" + notebookDST.UUID + ".qvnotebook/" + note.UUID + ".qvnote")
				err = CopyDir(noteDirSrc, noteDirDst)
				if err == nil {
					//update database
					note.NoteBookUUID = notebookDST.UUID
					enc, _ := json.Marshal(note)
					NoteDB.Set([]byte(note.UUID), enc)

					delete(notebookSRC.Notes, note.UUID)
					encSRC, _ := json.Marshal(notebookSRC)
					NoteBookDB.Set([]byte(notebookSRC.UUID), encSRC)

					notebookDST.Notes[note.UUID] = time.Now().Unix()
					encDST, _ := json.Marshal(notebookDST)
					NoteBookDB.Set([]byte(notebookDST.UUID), encDST)

					os.RemoveAll(noteDirSrc)

				} else { //nolint:staticcheck
					go showNotificationDialog("Error! Can not move folder " + noteDirSrc + " to " + noteDirDst)
				}

			}
		default:
			ctx.JSON(iris.Map{})
		}
	})

	app.Run(iris.Addr(":"+configGlobal.cmdPort), iris.WithOptimizations, iris.WithoutPathCorrection)

	webserverChan <- true
}

func FixNoteImagesLinks(note NoteTypeWithContentAPI, content string, ctx iris.Context) string {
	ImageURL := "/resources/" + note.NoteBookUUID + "/" + note.UUID + ""
	content = strings.Replace(content, "quiver-image-url", ImageURL, -1)
	content = strings.Replace(content, "quiver-file-url", ImageURL, -1)
	content = strings.Replace(content, "//"+ctx.Host()+"/resources/", "/resources/", -1) // fix for old cleanup
	return content
}
