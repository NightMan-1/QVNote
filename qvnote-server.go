package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/ru"
	"github.com/blevesearch/bleve/index/store/goleveldb"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/snowballstem"
	"github.com/blevesearch/snowballstem/russian"
)

func check(e error, message string) {
	if e != nil {
		fmt.Println(message)
		panic(e)
	}
}

func checkQuiet(e error) { //check_no_exit
	if e != nil {
		fmt.Println(e)
	}
}

func queryStem(query string) string {
	query = strings.ToLower(strings.TrimSpace(query))
	words := strings.Fields(query)
	env := snowballstem.NewEnv("")
	for i, word := range words {
		env.SetCurrent(word)
		russian.Stem(env)
		words[i] = env.Current()
	}
	result := strings.Join(words, " ")
	if len(words) == 1 {
		result += "*"
	}
	return result
}

func (ss *SearchService) buildMapping() *mapping.IndexMappingImpl {
	mappingTMP := bleve.NewIndexMapping()
	mappingTMP.DefaultAnalyzer = ru.AnalyzerName
	return mappingTMP
}

func (ss *SearchService) IndexMessage(data SearchContent) error {
	err := ss.index.Delete(data.UUID)
	err = ss.index.Index(data.UUID, data)
	checkQuiet(err)
	return nil
}

func (ss *SearchService) Search(query string) (*bleve.SearchResult, error) {
	qsq := bleve.NewQueryStringQuery(query)
	search := bleve.NewSearchRequest(qsq)
	search.Fields = []string{"UUID"}
	return ss.index.Search(search)
}

func readINI(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cfg := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "[") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			cfg[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return cfg, scanner.Err()
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

	//read configuration file
	cfgFile := configGlobal.execDir + "/config.ini"
	if _, err := os.Stat(cfgFile); err == nil {
		cfg, err := readINI(cfgFile)
		if err != nil {
			fmt.Printf("Fail to read file: %v", err)
			os.Exit(1)
		}

		if port, ok := cfg["port"]; ok {
			if p, err := strconv.Atoi(port); err == nil && p > 0 && p < 65535 {
				portTMP = p
			}
		}
		if v, ok := cfg["datadir"]; ok && v != "" {
			configGlobal.dataDir = v
		}
	}

	//get command line flags
	flag.IntVar(&portTMP, "port", portTMP, "port number")
	flag.StringVar(&configGlobal.dataDir, "datadir", configGlobal.dataDir, "data folder")
	flag.Parse()
	configGlobal.cmdPort = strconv.Itoa(portTMP)

	configGlobal.dataDir, _ = filepath.Abs(configGlobal.dataDir)

	//open database
	os.MkdirAll(configGlobal.dataDir, 0760)
	store, err := OpenStore(configGlobal.dataDir)
	check(err, "Error open data file")
	if err := migrateFromLedisIfNeeded(configGlobal.dataDir, store); err != nil {
		check(err, "Error migrating legacy data")
	}
	ConfigDB = store.Config()
	NoteBookDB = store.NoteBook()
	NoteDB = store.Note()
	TagsDB = store.Tags()
	FavoritesDB = store.Favorites()

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

	//read the settings
	data, _ := ConfigDB.Get([]byte("appInstalled"))
	if string(data) != "" && string(data) == "true" {
		configGlobal.appInstalled = true
	} else {
		configGlobal.appInstalled = false
	}

	data, _ = ConfigDB.Get([]byte("requestIndexing"))
	if string(data) != "" && string(data) == "true" {
		configGlobal.requestIndexing = true
	} else {
		configGlobal.requestIndexing = false
	}

	data, _ = ConfigDB.Get([]byte("sourceFolder"))
	if string(data) == "" {
		configGlobal.appInstalled = false
		switch runtime.GOOS {
		case "windows":
			configGlobal.sourceFolder = "./notes"
		default:
			configGlobal.sourceFolder = os.Getenv("HOME") + "/notes"
		}
	} else {
		configGlobal.sourceFolder = string(data)
	}
	if !CheckNotebooksFolderStructure(configGlobal.sourceFolder) {
		configGlobal.appInstalled = false
	}

	data, _ = ConfigDB.Get([]byte("atStartCheckNewNotes"))
	if string(data) != "" && string(data) == "true" {
		configGlobal.atStartCheckNewNotes = true
	} else {
		configGlobal.atStartCheckNewNotes = false
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
	NoteOld := make(map[string]NoteType)
	if err := NoteDB.Scan(func(NoteID, data []byte) error {
		var note NoteType
		if err := json.Unmarshal(data, &note); err != nil {
			return err
		}
		NoteOld[string(NoteID)] = note
		return nil
	}); err != nil {
		check(err, "Ошибка:")
	}

	check(NoteDB.Flush(), "Error flush notes")
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
					} else if configGlobal.requestIndexing {
						// The application requested a full re-index; do not trust the previous index state.
						note.SearchIndex = false
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

	check(NoteBookDB.Flush(), "Error flush notebooks")
	for k, v := range NoteBook {
		enc, err := json.Marshal(v)
		checkQuiet(err)
		checkQuiet(NoteBookDB.Set([]byte(k), enc))
	}
	check(TagsDB.Flush(), "Error flush tags")
	for k, v := range TagsCloud {
		enc, err := json.Marshal(v)
		checkQuiet(err)
		checkQuiet(TagsDB.Set([]byte(k), enc))
	}

	SaveConfig()
	fmt.Println("Done!")
}

// creating a structure for new notes
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
	if err := ConfigDB.Set([]byte("sourceFolder"), []byte(configGlobal.sourceFolder)); err != nil {
		return false
	}
	tmp := "false"
	if configGlobal.appInstalled {
		tmp = "true"
	}
	if err := ConfigDB.Set([]byte("appInstalled"), []byte(tmp)); err != nil {
		return false
	}

	tmp = "false"
	if configGlobal.requestIndexing {
		tmp = "true"
	}
	if err := ConfigDB.Set([]byte("requestIndexing"), []byte(tmp)); err != nil {
		return false
	}

	tmp = "false"
	if configGlobal.atStartCheckNewNotes {
		tmp = "true"
	}
	if err := ConfigDB.Set([]byte("atStartCheckNewNotes"), []byte(tmp)); err != nil {
		return false
	}

	return true
}

// clear HTML
func ClearHTML(content string) string {
	r := regexp.MustCompile(`<pre([^>]*)>`)
	content = r.ReplaceAllStringFunc(content, func(s string) string {
		re := regexp.MustCompile(`class=["'](language-[^"']+)["']`)
		if m := re.FindStringSubmatch(s); m != nil {
			return `<pre class="` + m[1] + `">`
		}
		return "<pre>"
	})
	r = regexp.MustCompile(`<code (.*?)>`)
	content = r.ReplaceAllString(content, "<code>")

	// Extract <pre> blocks first so nested <code> (and any other tags) survive cleanup.
	r = regexp.MustCompile(`(?s)<pre[^>]*>.*?</pre>`)
	matchData := r.FindAllStringSubmatch(content, -1)
	savePRE := make(map[string]string)
	for _, match := range matchData {
		preIndex := RandStringBytes(64)
		savePRE[preIndex] = match[0]
		content = strings.Replace(content, match[0], preIndex, 1)
	}

	r = regexp.MustCompile(`\s{2,}`)
	content = r.ReplaceAllString(content, " ")

	// Extract any <code> blocks that are not already inside a saved <pre> block.
	r = regexp.MustCompile(`(?s)<code>.*?</code>`)
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

	// Convert <br> to \n inside <pre> blocks before restoring
	for index, pre := range savePRE {
		pre = strings.Replace(pre, "<br>", "\n", -1)
		pre = strings.Replace(pre, "<br/>", "\n", -1)
		pre = strings.Replace(pre, "<br />", "\n", -1)
		savePRE[index] = pre
	}

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
			resp, err := http.Get(ImageURL)
			if err == nil {
				if ct := resp.Header.Get("Content-Type"); ct != "" {
					ImageType := ""
					ContentTypeTrue := strings.Split(ct, ";")
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
						out, err := os.Create(FileNameFull)
						if err == nil {
							_, err = io.Copy(out, resp.Body)
							out.Close()
							if err == nil {
								content = strings.Replace(content, ImageURL, "quiver-image-url/"+FileName, 1)
							} else {
								checkQuiet(err)
							}
						}
					} else {
						fmt.Println("\t\tError: wrong image type:", ContentTypeTrue[0])
					}
				} else {
					fmt.Println("\t\tError: wrong headers")
				}
				resp.Body.Close()
			} else {
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
	checkQuiet(NoteDB.Scan(func(NoteID, data []byte) error {
		var note NoteType
		if err := json.Unmarshal(data, &note); err != nil {
			return err
		}
		if !note.SearchIndex {
			noteFilePath, _ := filepath.Abs(configGlobal.sourceFolder + "/" + note.NoteBookUUID + ".qvnotebook/" + note.UUID + ".qvnote/meta.json")
			FilesForIndex = append(FilesForIndex, FilesForIndexType{noteFilePath, note.UUID})
		}
		return nil
	}))

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
	checkQuiet(NoteDB.Keys(func(NoteID []byte) error {
		NotesForOptimization = append(NotesForOptimization, string(NoteID))
		return nil
	}))

	if len(NotesForOptimization) > 0 {
		optimizationStatus.NotesTotal = len(NotesForOptimization)
		optimizationStatus.NotesCurrent = 0

		for _, item := range NotesForOptimization {
			optimizationStatus.NotesCurrent++
			OptimizeResources(item)

		}
	}
	optimizationStatus.Status = "done"

}

func main() {
	start := time.Now()

	fmt.Println("Initializing...")
	initSystem()

	//update the list of notes
	if configGlobal.appInstalled {
		if configGlobal.atStartCheckNewNotes || configGlobal.requestIndexing {
			FindAllNotes()
		}
		if configGlobal.requestIndexing {
			go indexingAllNotes()
		}
	}

	//start web server
	fmt.Println("Starting web server...")
	webserverChan := make(chan bool)
	go WebServer(webserverChan)

	<-webserverChan

	MemStat()
	fmt.Printf("Execution took %s\n", time.Since(start))
}
