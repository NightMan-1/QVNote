package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/ru"
	"github.com/blevesearch/bleve/index/store/goleveldb"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/snowballstem"
	"github.com/blevesearch/snowballstem/russian"
	"github.com/dustin/go-humanize"
	"github.com/go-ini/ini"
	"github.com/google/uuid"
	"github.com/imroc/req"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/marcsauter/single"
	lediscfg "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

func check(e error, message string) {
	if e != nil {
		fmt.Println(message)
		showNotification(message, "dialog_warning")
		panic(e)
	}
}

func checkQuiet(e error) { //check_no_exit
	if e != nil {
		fmt.Println(e)
		showNotification(fmt.Sprintf("%s", e), "notify")
	}
}

func queryStem(query string) string {
	queryTMP := query
	queryArray := strings.Split(query, " ")
	env := snowballstem.NewEnv("")
	for _, word := range queryArray {
		env.SetCurrent(word)
		russian.Stem(env)
		queryTMP = strings.Replace(queryTMP, word, env.Current(), -1)
	}
	if len(queryArray) == 1 {
		queryTMP += "*"
	}
	return queryTMP

}

func (ss *SearchService) buildMapping() *mapping.IndexMappingImpl {
	mappingTMP := bleve.NewIndexMapping()
	mappingTMP.DefaultAnalyzer = ru.AnalyzerName
	return mappingTMP
}

func (ss *SearchService) IndexMessage(data SearchContent) error {
	err := ss.index.Index(data.UUID, data)
	checkQuiet(err)
	return nil
}

func (ss *SearchService) Search(query string) (*bleve.SearchResult, error) {
	qsq := bleve.NewQueryStringQuery(query)
	search := bleve.NewSearchRequest(qsq)
	search.Fields = []string{"UUID"}
	return ss.index.Search(search)
}

func initSystem() {
	configGlobal.timeStart = time.Now()

	//program folder
	ex, _ := os.Executable()
	configGlobal.execDir, _ = filepath.Abs(path.Dir(ex) + "/")

	switch runtime.GOOS {
	case "windows":
		configGlobal.dataDir = os.Getenv("USERPROFILE")
	case "darwin":
		configGlobal.dataDir = os.Getenv("HOME")
	case "linux":
		configGlobal.dataDir = os.Getenv("HOME")
	default:
		fmt.Println("Sorry, can not run on your OS.")
		os.Exit(1)
	}
	configGlobal.dataDir += "/.config/QVNote"

	portTMP := 8000
	configGlobal.cmdPortable = false
	configGlobal.cmdServerMode = false
	configGlobal.appStartingMode = "independent"
	configGlobal.appStartingModeForce = false

	//read configuration file
	cfgFile := configGlobal.execDir + "/config.ini"
	if _, err := os.Stat(cfgFile); err == nil {
		cfg, err := ini.Load(cfgFile)
		if err != nil {
			fmt.Printf("Fail to read file: %v", err)
			os.Exit(1)
		}

		if cfg.Section("").Key("port").MustInt(9999) > 0 && cfg.Section("").Key("port").MustInt(9999) < 65535 {
			portTMP = cfg.Section("").Key("port").MustInt(9999)
		}
		if runtime.GOOS == "windows" {
			configGlobal.cmdPortable = cfg.Section("").Key("portable").MustBool(false)
		}

		configGlobal.cmdServerMode = cfg.Section("").Key("servermode").MustBool(false)

		if cfg.Section("").Key("datadir").String() != "" {
			if _, err := os.Stat(cfgFile); err == nil {
				configGlobal.dataDir = cfg.Section("").Key("datadir").String()
			}
		}

		tM := cfg.Section("").Key("startingmode").String()
		if tM != "" {
			configGlobal.appStartingModeForce = true
			if tM == "independent" {
				configGlobal.appStartingMode = "independent"
			} else {
				configGlobal.appStartingMode = "browser"
			}
		}
	}

	//get command line flags
	flag.IntVar(&portTMP, "port", portTMP, "port number")
	configGlobal.cmdPort = strconv.Itoa(portTMP)
	flag.BoolVar(&configGlobal.cmdPortable, "portable", configGlobal.cmdPortable, "portable flag for Windows OS")
	flag.BoolVar(&configGlobal.cmdServerMode, "server", false, "server mode")
	flag.StringVar(&configGlobal.dataDir, "datadir", configGlobal.dataDir, "data folder")
	flag.Parse()

	if configGlobal.cmdPortable {
		configGlobal.dataDir, _ = filepath.Abs(configGlobal.execDir + "/data")
	}
	configGlobal.dataDir, _ = filepath.Abs(configGlobal.dataDir)

	//open database
	cfg := lediscfg.NewConfigDefault()
	os.MkdirAll(configGlobal.dataDir, 0760)
	cfg.DataDir = configGlobal.dataDir
	LedisDB, err := ledis.Open(cfg)
	check(err, "Error open data file")
	ConfigDB, err = LedisDB.Select(0)
	check(err, "Error open data file")
	NoteBookDB, err = LedisDB.Select(1)
	check(err, "Error open data file")
	NoteDB, err = LedisDB.Select(2)
	check(err, "Error open data file")
	TagsDB, err = LedisDB.Select(3)
	check(err, "Error open data file")
	FavoritesDB, err = LedisDB.Select(4)
	check(err, "Error open data file")

	//search db
	indexName, _ := filepath.Abs(configGlobal.dataDir + "/search.bleve")
	if _, err := os.Stat(indexName + "/index_meta.json"); err != nil {
		os.RemoveAll(indexName)
		// time.Sleep(1 * time.Second)
	}
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

	searchStatus.Status = "idle"

	//читаем настройки
	data, _ := ConfigDB.Get([]byte("appInstalled"))
	if BytesToString(data) != "" && BytesToString(data) == "true" {
		configGlobal.appInstalled = true
	} else {
		configGlobal.appInstalled = false
	}

	data, _ = ConfigDB.Get([]byte("requestIndexing"))
	if BytesToString(data) != "" && BytesToString(data) == "true" {
		configGlobal.appInstalled = true
	} else {
		configGlobal.requestIndexing = false
	}

	data, _ = ConfigDB.Get([]byte("sourceFolder"))
	if BytesToString(data) == "" {
		configGlobal.appInstalled = false
		switch runtime.GOOS {
		case "windows":
			configGlobal.sourceFolder = "./notes"
		default:
			configGlobal.sourceFolder = os.Getenv("HOME") + "/notes"
		}
	} else {
		configGlobal.sourceFolder = BytesToString(data)
	}
	if !CheckNotebooksFolderStructure(configGlobal.sourceFolder) {
		configGlobal.appInstalled = false
	}

	data, _ = ConfigDB.Get([]byte("atStartOpenBrowser"))
	if BytesToString(data) == "false" {
		configGlobal.atStartOpenBrowser = false
	} else {
		configGlobal.atStartOpenBrowser = true
	}

	data, _ = ConfigDB.Get([]byte("atStartCheckNewNotes"))
	if BytesToString(data) != "" && BytesToString(data) == "true" {
		configGlobal.atStartCheckNewNotes = true
	} else {
		configGlobal.atStartCheckNewNotes = false
	}
	data, _ = ConfigDB.Get([]byte("atStartShowConsole"))
	if BytesToString(data) != "" && BytesToString(data) == "true" {
		configGlobal.atStartShowConsole = true
	} else {
		configGlobal.atStartShowConsole = false
	}

	data, _ = ConfigDB.Get([]byte("postEditor"))
	if BytesToString(data) != "" {
		configGlobal.postEditor = BytesToString(data)
	} else {
		configGlobal.postEditor = "quill"
	}

	if !configGlobal.appStartingModeForce {
		data, _ = ConfigDB.Get([]byte("startingMode"))
		if BytesToString(data) == "browser" { // independent by default
			configGlobal.appStartingMode = "browser"
		}
	}
}

func addToIndex(path string, uuid string) error {
	path = strings.Replace(path, "meta.json", "content.json", -1)

	if _, err := os.Stat(path); err == nil {

		jsonFile, err := os.Open(path)
		if err != nil {
			return err
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		var content SearchContent
		json.Unmarshal(byteValue, &content)
		content.UUID = uuid
		jsonFile.Close()
		err = ss.IndexMessage(content)
		return err
	}
	return nil
}

func FindAllNotes() {
	fmt.Println("Preparing the list of categories...")
	metaNoteBookRE := regexp.MustCompile(`.*\.qvnotebook[\\|/]meta.json$`)
	if _, err := os.Stat(configGlobal.sourceFolder); err == nil {
		filepath.Walk(configGlobal.sourceFolder, func(path string, info os.FileInfo, err error) error {

			//list of categories
			if metaNoteBookRE.MatchString(path) {
				jsonFile, err := os.Open(path)
				if err != nil {
					fmt.Println(err)
				}

				byteValue, _ := ioutil.ReadAll(jsonFile)
				var notebook NoteBookType
				json.Unmarshal(byteValue, &notebook)
				notebook.Notes = make(map[string]int64)
				NoteBook[notebook.UUID] = notebook
				jsonFile.Close()

			}
			return nil
		})
	}

	fmt.Println("Preparing the list of notes...")

	//list of notes
	cursor := []byte(nil)
	NoteOld := make(map[string]NoteType)
	for {
		allDBData, err := NoteDB.Scan(ledis.KV, cursor, 0, false, "")
		if err != nil || len(allDBData) == 0 {
			break
		}
		for _, NoteID := range allDBData {
			cursor = NoteID
			data, _ := NoteDB.Get(NoteID)
			var note NoteType
			err = json.Unmarshal(data, &note)
			//checkQuiet(err)
			check(err, "Ошибка:")
			NoteOld[BytesToString(NoteID)] = note
		}
	}

	NoteDB.FlushAll()
	metaNoteRE := regexp.MustCompile(`.*[\\|/](.*)\.qvnotebook[\\|/](.*)\.qvnote[\\|/]meta.json$`)

	if _, err := os.Stat(configGlobal.sourceFolder); err == nil {
		filepath.Walk(configGlobal.sourceFolder, func(path string, info os.FileInfo, err error) error {

			noteFile := metaNoteRE.FindAllStringSubmatch(path, -1)
			if len(noteFile) == 1 {
				notebookID := noteFile[0][1]

				jsonFile, err := os.Open(path)
				checkQuiet(err)
				byteValue, _ := ioutil.ReadAll(jsonFile)
				var note NoteType
				json.Unmarshal(byteValue, &note)
				jsonFile.Close()

				note.NoteBookUUID = notebookID

				NoteBook[notebookID].Notes[note.UUID] = time.Now().Unix()

				if value, ok := NoteOld[note.UUID]; ok {
					if note.UUID != value.UUID ||
						note.Title != value.Title ||
						note.NoteBookUUID != value.NoteBookUUID ||
						note.CreatedAt != value.CreatedAt ||
						note.UpdatedAt != value.UpdatedAt {
						note.SearchIndex = false
						configGlobal.requestIndexing = true
					} else {
						note.SearchIndex = value.SearchIndex
					}

				} else {
					note.SearchIndex = false
					configGlobal.requestIndexing = true
				}

				enc, err := json.Marshal(note)
				checkQuiet(err)
				err = NoteDB.Set([]byte(note.UUID), enc)
				checkQuiet(err)

				if len(note.Tags) > 0 {
					for _, TagName := range note.Tags {
						TagsCloud[TagName] = append(TagsCloud[TagName], note.UUID)

					}

				}

			}

			return nil
		})
	}

	fmt.Println("Updating the database...")

	NoteBookDB.FlushAll()
	for k, v := range NoteBook {
		enc, err := json.Marshal(v)
		checkQuiet(err)
		NoteBookDB.Set([]byte(k), enc)
	}
	TagsDB.FlushAll()
	for k, v := range TagsCloud {
		enc, err := json.Marshal(v)
		checkQuiet(err)
		TagsDB.Set([]byte(k), enc)
	}

	SaveConfig()
	fmt.Println("Done!")
	showNotification("The list of notes has been updated.", "notify")
}

//creating a structure for new notes
func CreateNewNotebooksFolder(folder string) bool {
	err := os.MkdirAll(folder+"/Inbox.qvnotebook", 0777)
	if err != nil {
		return false
	}
	err = os.MkdirAll(folder+"/Trash.qvnotebook", 0777)
	if err != nil {
		return false
	}

	content := "{\"children\" : [], \"uuid\" : \"Notebooks\" }\n"
	err = ioutil.WriteFile(folder+"/meta.json", []byte(content), 0644)
	if err != nil {
		return false
	}

	content = "{ \"name\" : \"Inbox\", \"uuid\" : \"Inbox\"}\n"
	err = ioutil.WriteFile(folder+"/Inbox.qvnotebook/meta.json", []byte(content), 0644)
	if err != nil {
		return false
	}

	content = "{ \"name\" : \"Trash\", \"uuid\" : \"Trash\"}\n"
	err = ioutil.WriteFile(folder+"/Trash.qvnotebook/meta.json", []byte(content), 0644)
	if err != nil { //nolint:gosimple
		return false
	}

	return true
}

func CheckNotebooksFolderStructure(folder string) bool {
	if _, err := os.Stat(folder); err != nil {
		return false
	}
	if _, err := os.Stat(folder + "/meta.json"); err != nil {
		return false
	}
	if _, err := os.Stat(folder + "/Inbox.qvnotebook/meta.json"); err != nil {
		return false
	}
	if _, err := os.Stat(folder + "/Trash.qvnotebook/meta.json"); err != nil {
		return false
	}
	return true
}

func SaveConfig() bool {
	err := ConfigDB.Set([]byte("sourceFolder"), []byte(configGlobal.sourceFolder))
	if err != nil {
		return false
	}
	err = ConfigDB.Set([]byte("postEditor"), []byte(configGlobal.postEditor))
	if err != nil {
		return false
	}
	err = ConfigDB.Set([]byte("startingMode"), []byte(configGlobal.appStartingMode))
	if err != nil {
		return false
	}
	tmp := "false"
	if configGlobal.appInstalled {
		tmp = "true"
	}
	err = ConfigDB.Set([]byte("appInstalled"), []byte(tmp))
	if err != nil {
		return false
	}

	tmp = "false"
	if configGlobal.requestIndexing {
		tmp = "true"
	}
	err = ConfigDB.Set([]byte("requestIndexing"), []byte(tmp))
	if err != nil {
		return false
	}

	tmp = "false"
	if configGlobal.atStartOpenBrowser {
		tmp = "true"
	}
	err = ConfigDB.Set([]byte("atStartOpenBrowser"), []byte(tmp))
	if err != nil {
		return false
	}

	tmp = "false"
	if configGlobal.atStartCheckNewNotes {
		tmp = "true"
	}
	err = ConfigDB.Set([]byte("atStartCheckNewNotes"), []byte(tmp))
	if err != nil {
		return false
	}

	tmp = "false"
	if configGlobal.atStartShowConsole {
		tmp = "true"
	}
	err = ConfigDB.Set([]byte("atStartShowConsole"), []byte(tmp))
	if err != nil { //nolint:gosimple
		return false
	}
	return true
}

func FixNoteImagesLinks(note NoteTypeWithContentAPI, content string, ctx iris.Context) string {
	ImageURL := "/resources/" + note.NoteBookUUID + "/" + note.UUID + ""
	content = strings.Replace(content, "quiver-image-url", ImageURL, -1)
	content = strings.Replace(content, "quiver-file-url", ImageURL, -1)
	content = strings.Replace(content, "//"+ctx.Host()+"/resources/", "/resources/", -1) // fix for old cleanup
	return content
}

//clear HTML
func ClearHTML(content string) string {
	r := regexp.MustCompile(`<pre (.*?)>`)
	content = r.ReplaceAllString(content, "<pre>")
	r = regexp.MustCompile(`<code (.*?)>`)
	content = r.ReplaceAllString(content, "<code>")

	r = regexp.MustCompile(`(?m)<pre>(?s).*?</pre>`)
	matchData := r.FindAllStringSubmatch(content, -1)
	savePRE := make(map[string]string)
	for _, match := range matchData {
		preIndex := RandStringBytes(64)
		savePRE[preIndex] = match[0]
		content = strings.Replace(content, match[0], preIndex, 1)
	}

	r = regexp.MustCompile(`\s{2,}`)
	content = r.ReplaceAllString(content, " ")

	r = regexp.MustCompile(`(?m)<code>(?s).*?</code>`)
	matchData = r.FindAllStringSubmatch(content, -1)
	saveCODE := make(map[string]string)
	for _, match := range matchData {
		preIndex := RandStringBytes(64)
		saveCODE[preIndex] = match[0]
		content = strings.Replace(content, match[0], preIndex, 1)
	}

	content = strings.Replace(content, "\n", "", -1)

	r = regexp.MustCompile(`<div`)
	content = r.ReplaceAllString(content, "<p")
	r = regexp.MustCompile(`</div>`)
	content = r.ReplaceAllString(content, "</p>")

	r = regexp.MustCompile(`<h(.).*?>`)
	content = r.ReplaceAllString(content, `<h$1>`)

	r = regexp.MustCompile(`<(p|br|hr).*?>`)
	content = r.ReplaceAllString(content, "<$1>")

	r = regexp.MustCompile(`<(p|h1|h2|h3|h4|h5|h6)>\s+`)
	content = r.ReplaceAllString(content, "<$1>")
	r = regexp.MustCompile(`\s+<(/p|/h1|/h2|/h3|/h4|/h5|/h6)>`)
	content = r.ReplaceAllString(content, "<$1>")

	r = regexp.MustCompile("(<p>){2,}")
	content = r.ReplaceAllString(content, `<p>`)
	r = regexp.MustCompile("(</p>){2,}")
	content = r.ReplaceAllString(content, `</p>`)

	r = regexp.MustCompile("<(p|h1|h2|h3|h4|h5|h6)><br>")
	content = r.ReplaceAllString(content, `<$1>`)

	r = regexp.MustCompile("<p>&nbsp;</p>")
	content = r.ReplaceAllString(content, ``)

	r = regexp.MustCompile(`<(span|p|h1|h2|h3|h4|h5|h6)></(span|p|h1|h2|h3|h4|h5|h6)>`)
	content = r.ReplaceAllString(content, "")

	// r = regexp.MustCompile(`<img.*?src=["|'](.*?)["|'].*?>`)
	// content = r.ReplaceAllString(content, `<img src="$1">`)

	// r = regexp.MustCompile(`class=["|'](.*?)["|']`)
	// content = r.ReplaceAllString(content, "")
	r = regexp.MustCompile(`id=["|'](.*?)["|']`)
	content = r.ReplaceAllString(content, "")

	r = regexp.MustCompile(`data-\w*?=["|'](.*?)["|']`)
	content = r.ReplaceAllString(content, "")
	r = regexp.MustCompile(`data-\w*?-\w*?=["|'](.*?)["|']`)
	content = r.ReplaceAllString(content, "")

	r = regexp.MustCompile(`font-family:(.*?);`)
	content = r.ReplaceAllString(content, "")
	r = regexp.MustCompile(`font-size:(.*?);`)
	content = r.ReplaceAllString(content, "")

	r = regexp.MustCompile(`position:(.*?);`)
	content = r.ReplaceAllString(content, "")

	r = regexp.MustCompile(`<table>`)
	content = r.ReplaceAllString(content, `<table class="table table-sm">`)

	//r = regexp.MustCompile(`width:(.*?);`)
	//content = r.ReplaceAllString(content, "")
	//r = regexp.MustCompile(`width:(.*?)px`)
	//content = r.ReplaceAllString(content, "")
	//r = regexp.MustCompile(`max-width:(.*?);`)
	//content = r.ReplaceAllString(content, "")
	//r = regexp.MustCompile(`padding-bottom:(.*?)%;`)
	//content = r.ReplaceAllString(content, "")

	//r = regexp.MustCompile(`style=["|']\s*["|']`)
	//content = r.ReplaceAllString(content, "")

	r = regexp.MustCompile(`<font.*?>(.*?)</font>`)
	content = r.ReplaceAllString(content, "$1")

	for index, code := range saveCODE {
		content = strings.Replace(content, index, code, 1)
	}
	for index, code := range savePRE {
		content = strings.Replace(content, index, code, 1)
	}

	//r = regexp.MustCompile(`<(p|pre|h1|h2|h3|h4|h5|ul|ol|/ul|/ol)>`)
	//content = r.ReplaceAllString(content, "\n<$1>")
	//r = regexp.MustCompile(`<li>`)
	//content = r.ReplaceAllString(content, "\n    <li>")
	//r = regexp.MustCompile(`^\n`)
	//content = r.ReplaceAllString(content, "")
	//r = regexp.MustCompile(`(\n|\s+)\n`)
	//content = r.ReplaceAllString(content, "\n")

	//content += `<div class="clearfix"></div>`

	//fmt.Println(content)

	return content
}

func OptimizeResources(uuid string) {
	if uuid == "" {
		return
	}

	data, _ := NoteDB.Get([]byte(uuid))
	var noteData NoteType
	err := json.Unmarshal(data, &noteData)
	checkQuiet(err)

	contentDir := configGlobal.sourceFolder + "/" + noteData.NoteBookUUID + ".qvnotebook/" + noteData.UUID + ".qvnote"
	contentPath := contentDir + "/content.json"
	if _, err := os.Stat(contentPath); err == nil {
		os.MkdirAll(contentDir+"/resources", 0755)

		jsonFile, err := os.Open(contentPath)
		checkQuiet(err)
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var contentData SearchContent
		json.Unmarshal(byteValue, &contentData)
		jsonFile.Close()

		content := ""
		contentType := ""
		for _, text := range contentData.Cells {
			content += text.Data
			contentType = text.Type
		}

		r := regexp.MustCompile(`<img.*?src=["|'](http.*?)["|'].*?>`)
		matchData := r.FindAllStringSubmatch(content, -1)
		if len(matchData) > 0 {
			fmt.Println("Optimization start for", uuid)

		}
		for _, match := range matchData {
			ImageURL := match[1]
			fmt.Println("\tdownloading", ImageURL)
			r, err := req.Get(ImageURL)
			if err == nil {
				resp := r.Response()
				if _, ok := resp.Header["Content-Type"]; ok {
					ImageType := ""
					ContentTypeTrue := strings.Split(resp.Header["Content-Type"][0], ";")
					switch ContentTypeTrue[0] {
					case "image/png":
						ImageType = ".png"
					case "image/gif":
						ImageType = ".gif"
					case "image/jpeg":
						ImageType = ".jpg"
					case "image/webp":
						ImageType = ".webp"
					case "image/svg+xml":
						ImageType = ".svg"
					}

					if ImageType != "" {
						FileName := RandStringBytes(32) + ImageType
						FileNameFull, _ := filepath.Abs(contentDir + "/resources/" + FileName)
						err = r.ToFile(FileNameFull)
						if err == nil {
							content = strings.Replace(content, ImageURL, "quiver-image-url/"+FileName, 1)
						} else {
							checkQuiet(err)
						}
					} else {
						fmt.Println("\t\tError: wrong image type:", ContentTypeTrue[0])
					}
				} else {
					fmt.Println("\t\tError: wrong headers:", resp.Header)
				}
			} else {
				// checkQuiet(err) // disabled, too many messages
				fmt.Println(err)
			}
		}
		var ContentFile struct {
			Title string             `json:"title"`
			Cells []ContentCellsType `json:"cells"`
		}
		ContentFile.Title = noteData.Title
		ContentFile.Cells = make([]ContentCellsType, 1)
		ContentFile.Cells[0] = ContentCellsType{contentType, content}

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.Encode(ContentFile)

		err = ioutil.WriteFile(contentDir+"/content.json", buf.Bytes(), 0644)
		checkQuiet(err)

		//delete unnecessary files
		r = regexp.MustCompile(`["|']quiver-image-url/(.*?)["|']`)
		matchData = r.FindAllStringSubmatch(content, -1)
		InternalImages := make(map[string]bool)
		for _, match := range matchData {
			InternalImages[match[1]] = true
		}
		r = regexp.MustCompile(`["|']quiver-file-url/(.*?)["|']`)
		matchData = r.FindAllStringSubmatch(content, -1)
		for _, match := range matchData {
			InternalImages[match[1]] = true
		}

		filepath.Walk(contentDir+"/resources", func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if _, ok := InternalImages[filepath.Base(path)]; ok {
			} else {
				fmt.Println("\tdeleting file", path)
				os.Remove(path)
			}
			return nil
		})

	}

}

func indexingAllNotes() {
	searchStatus.Status = "indexing"
	FilesForIndex = []FilesForIndexType{}
	cursor := []byte(nil)
	for {
		allDBData, err := NoteDB.Scan(ledis.KV, cursor, 0, false, "")
		if err != nil || len(allDBData) == 0 {
			break
		}
		for _, NoteID := range allDBData {
			cursor = NoteID
			//fmt.Println(BytesToString(NoteID))
			data, _ := NoteDB.Get(NoteID)
			var note NoteType
			err := json.Unmarshal(data, &note)
			checkQuiet(err)
			if !note.SearchIndex {
				noteFilePath, _ := filepath.Abs(configGlobal.sourceFolder + "/" + note.NoteBookUUID + ".qvnotebook/" + note.UUID + ".qvnote/meta.json")
				FilesForIndex = append(FilesForIndex, FilesForIndexType{noteFilePath, note.UUID})
			}
		}
	}

	if len(FilesForIndex) > 0 {
		searchStatus.NotesTotal = len(FilesForIndex)
		searchStatus.NotesCurrent = 0

		for _, item := range FilesForIndex {
			searchStatus.NotesCurrent++
			addToIndex(item.Patch, item.UUID)

			//обновляем данные об индексации
			data, _ := NoteDB.Get([]byte(item.UUID))
			var note NoteType
			json.Unmarshal(data, &note)
			note.SearchIndex = true
			enc, _ := json.Marshal(note)
			NoteDB.Set([]byte(item.UUID), enc)

		}
	}
	configGlobal.requestIndexing = false
	SaveConfig()
	searchStatus.Status = "done"

}

func optimizeAllNotes() {
	optimizationStatus.Status = "processing"

	var NotesForOptimization []string
	cursor := []byte(nil)
	for {
		allDBData, err := NoteDB.Scan(ledis.KV, cursor, 0, false, "")
		if err != nil || len(allDBData) == 0 {
			break
		}
		for _, NoteID := range allDBData {
			cursor = NoteID
			NotesForOptimization = append(NotesForOptimization, BytesToString(NoteID))
		}
	}

	if len(NotesForOptimization) > 0 {
		optimizationStatus.NotesTotal = len(NotesForOptimization)
		optimizationStatus.NotesCurrent = 0

		for _, item := range NotesForOptimization {
			optimizationStatus.NotesCurrent++
			OptimizeResources(item)

		}
	}
	optimizationStatus.Status = "done"
	showNotification("Optimization is complete.", "notify")

}

func WebServer(webserverChan chan bool) { //nolint:gocyclo
	app := iris.New()
	app.Use(iris.Gzip)
	app.Use(recover.New())
	app.Use(logger.New())

	app.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "OPTIONS", "POST", "PATCH", "PUT", "DELETE"},
	}))

	//app.StaticWeb("/static", configGlobal.execDir + "/templates/static")
	//app.StaticEmbedded("/", "./templates", Asset, AssetNames)
	app.HandleDir("/", "./templates", iris.DirOptions{Asset: Asset, AssetInfo: AssetInfo, AssetNames: AssetNames, Gzip: false})

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

	app.Handle("GET", "/", func(ctx iris.Context) {
		//indexHTML, err := ioutil.ReadFile(configGlobal.execDir + "/templates/index.html")
		//check(err, "Error loading index.html")
		//ctx.HTML(string(indexHTML))
		data, _ := Asset("templates/index.html")
		ctx.HTML(string(data))
	})

	app.OnErrorCode(404, func(ctx iris.Context) {
		data, _ := Asset("templates/index.html")
		ctx.StatusCode(200)
		ctx.HTML(string(data))
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
		showNotification("Good buy!", "notify")
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
			"consolePresent":       configGlobal.consolePresent,
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
				favoritesList = append(favoritesList, BytesToString(FavoriteID))

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
			ctx.ServeFile(imageFile, false)
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
				TagsCloud = append(TagsCloud, TagsListStruct{len(tagsData), strings.Trim(BytesToString(TagID), " "), url.PathEscape(BytesToString(TagID))})

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
			indexName, _ := filepath.Abs(configGlobal.dataDir + "/data/search.bleve")
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

			NoteDB.FlushAll()
			NoteBookDB.FlushAll()

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

		//TODO add image processing

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
			dataString := BytesToString(data)
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
			if BytesToString(data) != "" {
				var tagsData []string
				err := json.Unmarshal(data, &tagsData)
				checkQuiet(err)
				for _, noteID := range tagsData {
					//change files
					dataNote, _ := NoteDB.Get([]byte(noteID))
					if BytesToString(dataNote) != "" {
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
					if BytesToString(data) != "" {
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
					//TODO показывать ошибку о переносе папки
				}

			} else { //nolint:staticcheck
				//TODO показывать ошибку о несуществующем блокноте
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
					//TODO показывать ошибку о переносе папки
				}

			}
		default:
			ctx.JSON(iris.Map{})
		}
	})

	app.Run(iris.Addr(":"+configGlobal.cmdPort), iris.WithOptimizations)

	webserverChan <- true
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == string("--systray") && runtime.GOOS == string("darwin") {
		runSystray()
		os.Exit(0)
	}

	s := single.New("QVNote")
	if err := s.CheckLock(); err != nil && err == single.ErrAlreadyRunning {
		showNotification("another instance of the app is already running, exiting", "dialog_warning")
		log.Fatal("another instance of the app is already running, exiting")
		os.Exit(1)
	} else if err != nil {
		// Another error occurred, might be worth handling it as well
		showNotification("failed to acquire exclusive app lock", "dialog_warning")
		log.Fatalf("failed to acquire exclusive app lock: %v", err)
		os.Exit(1)
	}
	defer s.TryUnlock()

	start := time.Now()
	println("Initializing...")
	initSystem()
	initPlatformSpecific()

	//update the list of notes
	if configGlobal.appInstalled {
		if configGlobal.atStartCheckNewNotes {
			FindAllNotes()
		}
	}

	//start web server
	println("Starting web server...")
	if configGlobal.atStartOpenBrowser && !configGlobal.cmdServerMode && configGlobal.appStartingMode != "independent" {
		go openBrowser("http://localhost:" + configGlobal.cmdPort + "/")
	}
	webserverChan := make(chan bool)
	go WebServer(webserverChan)

	if configGlobal.appStartingMode == "independent" {
		startStadaloneGUI()
	} else {
		<-webserverChan
	}

	MemStat()
	fmt.Printf("page took %s\n", time.Since(start))

}
