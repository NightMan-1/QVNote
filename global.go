package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/dustin/go-humanize"
	jsoniter "github.com/json-iterator/go"
	"github.com/ledisdb/ledisdb/ledis"

)

type configGlobalStruct struct {
	sourceFolder         string
	timeStart            time.Time
	execDir              string
	dataDir              string
	appInstalled         bool
	requestIndexing      bool //необходимость запустить переиндексацию поиска
	atStartCheckNewNotes bool
	consolePresent       bool
	atStartShowConsole   bool
	postEditor           string
	cmdPort              string
	cmdPortable          bool
	appStartingMode      string
	appStartingModeForce bool
	cmdServerMode        bool
	atStartOpenBrowser   bool
}

var configGlobal (configGlobalStruct)
var ConfigDB, NoteBookDB, NoteDB, TagsDB, FavoritesDB *ledis.DB //nolint:golint

type SearchService struct {
	index bleve.Index
	batch *bleve.Batch
}

var ss SearchService

var searchStatus struct {
	Status       string `json:"status"`
	NotesTotal   int    `json:"notesTotal"`
	NotesCurrent int    `json:"notesCurrent"`
}

var optimizationStatus struct {
	Status       string `json:"status"`
	NotesTotal   int    `json:"notesTotal"`
	NotesCurrent int    `json:"notesCurrent"`
}

type SearchContent struct {
	UUID  string             `json:"uuid"`
	Title string             `json:"title"`
	Cells []ContentCellsType `json:"cells"`
}

type SearchResult struct {
	Title        string `json:"title"`
	UUID         string `json:"uuid"`
	NoteBookUUID string `json:"NoteBookUUID"`
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type NoteBookType struct {
	UUID       string
	Name       string
	Notes      map[string]int64
	NotesCount int
}

type NoteBookTypeAPI struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	NotesCount int    `json:"notesCount"`
}

type NoteType struct {
	CreatedAt    int32    `json:"created_at"`
	UpdatedAt    int32    `json:"updated_at"`
	Tags         []string `json:"tags"`
	Title        string   `json:"title"`
	UUID         string   `json:"uuid"`
	URL          string   `json:"url_src"`
	NoteBookUUID string
	SearchIndex  bool
}

type NoteTypeWithContentAPI struct {
	CreatedAt    int32    `json:"created_at"`
	UpdatedAt    int32    `json:"updated_at"`
	Tags         []string `json:"tags"`
	Title        string   `json:"title"`
	UUID         string   `json:"uuid"`
	URL          string   `json:"url_src"`
	NoteBookUUID string
	SearchIndex  bool
	Content      string `json:"content"`
	ContentType  string `json:"type"`
	Favorites    bool   `json:"favorites"`
}

type NoteTypeAPI struct {
	UpdatedAt    int32  `json:"updated_at"`
	Title        string `json:"title"`
	UUID         string `json:"uuid"`
	NoteBookUUID string `json:"NoteBookUUID"`
}

var NoteBook = make(map[string]NoteBookType)
var TagsCloud = make(map[string][]string)

type TagsListStruct struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

type ContentCellsType struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type FilesForIndexType struct {
	Patch string
	UUID  string
}

var FilesForIndex = []FilesForIndexType{}

var systrayProcess *exec.Cmd

func BytesToString(data []byte) string {
	return string(data)
}

func RandStringBytes(n int) string {
	const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func MemStat() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	//fmt.Printf("\nAlloc = %v\nTotalAlloc = %v\nSys = %v\nNumGC = %v\n\n", humanize.Bytes(mem.Alloc), humanize.Bytes(mem.TotalAlloc), humanize.Bytes(mem.Sys), mem.NumGC)
	fmt.Printf("\nSys = %v\n\n", humanize.Bytes(mem.Sys))
}

//https://www.codesd.com/item/golang-how-to-get-the-total-size-of-the-directory.html
func DirSize2(path string) (int64, error) {
	var size int64
	adjSize := func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	}
	err := filepath.Walk(path, adjSize)

	return size, err
}

// https://codereview.stackexchange.com/questions/60074/in-array-in-go
func inArray(val string, array []string) (exists bool) {
	exists = false
	for _, v := range array {
		if val == v {
			exists = true
			return
		}
	}
	return
}
