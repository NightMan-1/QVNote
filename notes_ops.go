package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Reusable note/notebook/tag operations shared by the HTTP API handlers
// (webserver.go) and the MCP tools (mcp.go). Extracted so both entry points
// run exactly the same logic.

// FixNoteImagesLinks rewrites stored image placeholders to browser-relative
// resource URLs. host is the request Host header ("" when there is no HTTP
// request, e.g. MCP).
func FixNoteImagesLinks(note NoteTypeWithContentAPI, content string, host string) string {
	ImageURL := "/resources/" + note.NoteBookUUID + "/" + note.UUID + ""
	content = strings.Replace(content, "quiver-image-url", ImageURL, -1)
	content = strings.Replace(content, "quiver-file-url", ImageURL, -1)
	if host != "" {
		content = strings.Replace(content, "//"+host+"/resources/", "/resources/", -1) // fix for old cleanup
	}
	return content
}

// searchNotes runs a Bleve search. Queries shorter than 3 characters return
// an empty list (same as the HTTP API).
func searchNotes(text string) []SearchResult {
	NotesList := make([]SearchResult, 0)
	if len(text) < 3 {
		return NotesList
	}
	NoteListDedup := make(map[string]bool)
	searchResult, err := ss.Search(text)
	if err != nil || searchResult == nil {
		return NotesList
	}
	var noteShort SearchResult
	for _, item := range searchResult.Hits {
		data, _ := NoteDB.Get([]byte(item.ID))
		if len(data) == 0 {
			continue
		}
		if err := json.Unmarshal(data, &noteShort); err != nil {
			checkQuiet(err)
			continue
		}
		if NoteListDedup[noteShort.UUID] {
			continue
		}
		NoteListDedup[noteShort.UUID] = true
		NotesList = append(NotesList, noteShort)
	}
	return NotesList
}

// listNotesAtNotebook returns notes for a special id (Favorites / Allnotes)
// or for a concrete notebook UUID. Returns nil for an empty id.
func listNotesAtNotebook(notebookID string) []NoteTypeAPI {
	var NotesList []NoteTypeAPI
	switch {
	case notebookID == "Favorites":
		checkQuiet(FavoritesDB.Keys(func(NoteID []byte) error {
			data, _ := NoteDB.Get(NoteID)
			var note NoteTypeAPI
			if err := json.Unmarshal(data, &note); err != nil {
				return err
			}
			note.NoteBookUUID = "Favorites"
			NotesList = append(NotesList, note)
			return nil
		}))
	case notebookID == "Allnotes":
		checkQuiet(NoteDB.Scan(func(NoteID, data []byte) error {
			var note NoteTypeAPI
			if err := json.Unmarshal(data, &note); err != nil {
				return err
			}
			NotesList = append(NotesList, note)
			return nil
		}))
	case len(notebookID) > 0:
		data, _ := NoteBookDB.Get([]byte(notebookID))
		var notebookData NoteBookType
		if err := json.Unmarshal(data, &notebookData); err != nil {
			checkQuiet(err)
			return NotesList
		}
		for NoteBookID := range notebookData.Notes {
			data, _ := NoteDB.Get([]byte(NoteBookID))
			var note NoteTypeAPI
			if err := json.Unmarshal(data, &note); err != nil {
				checkQuiet(err)
				continue
			}
			NotesList = append(NotesList, note)
		}
	default:
		return nil
	}
	sort.Slice(NotesList, func(i, j int) bool {
		return NotesList[i].UpdatedAt > NotesList[j].UpdatedAt
	})
	return NotesList
}

// listNotebooks returns all notebooks sorted by name.
func listNotebooks() []NoteBookTypeAPI {
	var noteBooksList []NoteBookTypeAPI
	checkQuiet(NoteBookDB.Scan(func(NoteBookID, data []byte) error {
		var notebookData NoteBookType
		if err := json.Unmarshal(data, &notebookData); err != nil {
			return err
		}
		noteBooksList = append(noteBooksList, NoteBookTypeAPI{notebookData.UUID, notebookData.Name, len(notebookData.Notes)})
		return nil
	}))
	sort.Slice(noteBooksList, func(i, j int) bool {
		return strings.ToLower(noteBooksList[i].Name) < strings.ToLower(noteBooksList[j].Name)
	})
	return noteBooksList
}

// listTags returns the tag cloud sorted by name.
func listTags() []TagsListStruct {
	var TagsCloud []TagsListStruct
	checkQuiet(TagsDB.Scan(func(TagID, data []byte) error {
		var tagsData []string
		if err := json.Unmarshal(data, &tagsData); err != nil {
			return err
		}
		TagsCloud = append(TagsCloud, TagsListStruct{len(tagsData), strings.Trim(string(TagID), " "), string(TagID)})
		return nil
	}))
	sort.Slice(TagsCloud, func(i, j int) bool {
		return strings.ToLower(TagsCloud[i].Name) < strings.ToLower(TagsCloud[j].Name)
	})
	return TagsCloud
}

// listNotesByTag returns notes carrying the given tag, newest first.
func listNotesByTag(tagName string) []NoteTypeAPI {
	var NotesList []NoteTypeAPI
	if tagName == "" {
		return NotesList
	}
	data, _ := TagsDB.Get([]byte(tagName))
	var notesListTMP []string
	if err := json.Unmarshal(data, &notesListTMP); err != nil {
		checkQuiet(err)
		return NotesList
	}
	for _, noteID := range notesListTMP {
		data, _ := NoteDB.Get([]byte(noteID))
		var note NoteTypeAPI
		if err := json.Unmarshal(data, &note); err != nil {
			checkQuiet(err)
			continue
		}
		NotesList = append(NotesList, note)
	}
	sort.Slice(NotesList, func(i, j int) bool {
		return NotesList[i].UpdatedAt > NotesList[j].UpdatedAt
	})
	return NotesList
}

// loadNote returns a full note (metadata + content) or nil when the note or
// its content file does not exist. raw skips normalization; host is used to
// rewrite image URLs ("" when there is no HTTP request).
func loadNote(noteID string, raw bool, host string) *NoteTypeWithContentAPI {
	if noteID == "" {
		return nil
	}
	data, _ := NoteDB.Get([]byte(noteID))
	if len(data) == 0 {
		return nil
	}
	var noteData NoteTypeWithContentAPI
	if err := json.Unmarshal(data, &noteData); err != nil {
		checkQuiet(err)
		return nil
	}

	contentDir := configGlobal.sourceFolder + "/" + noteData.NoteBookUUID + ".qvnotebook/" + noteData.UUID + ".qvnote"
	contentPath := contentDir + "/content.json"
	if _, err := os.Stat(contentPath); err != nil {
		return nil
	}
	jsonFile, err := os.Open(contentPath)
	if err != nil {
		checkQuiet(err)
		return nil
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var contentFile SearchContent
	json.Unmarshal(byteValue, &contentFile)
	jsonFile.Close()

	noteData.Content = ""
	for _, text := range contentFile.Cells {
		noteData.Content += text.Data
		noteData.ContentType = text.Type
	}

	// Legacy notes (no content_state) are normalized on read.
	// refetched/edited notes and raw requests are served as stored.
	if !raw && noteData.ContentType != "code" && noteData.ContentState == "" {
		noteData.Content = ClearHTML(noteData.Content, noteData.Title)
		// Old notes often start with the bare source URL as the first
		// line; lift it into url_src (response only, stored data
		// untouched) so it doesn't clutter the content.
		cleaned, src := stripLeadingSourceLink(noteData.Content)
		if src != "" {
			noteData.Content = cleaned
			if noteData.URL == "" {
				noteData.URL = src
			}
		}
	}

	// CodePen markers are expanded into iframes only for display;
	// the editor (raw) and the stored content keep the original markup.
	if !raw && noteData.ContentType != "code" {
		noteData.Content = RenderCodePenEmbeds(noteData.Content)
	}

	noteData.Content = FixNoteImagesLinks(noteData, noteData.Content, host)

	dataExists, _ := FavoritesDB.Exists([]byte(noteID))
	noteData.Favorites = dataExists

	return &noteData
}

// saveNoteParams is the input for saveNote. UUID empty creates a new note.
type saveNoteParams struct {
	UUID         string
	Title        string
	URL          string
	Type         string
	Content      string
	Tags         []string
	ContentState string
	// NotebookUUID is only used when creating (UUID == ""); empty means Inbox.
	NotebookUUID string
}

type saveNoteResult struct {
	NoteBookUUID string `json:"NoteBookUUID"`
	UUID         string `json:"uuid"`
}

// saveNote creates or updates a note (files + bbolt + tag cloud + search index).
func saveNote(p saveNoteParams) (saveNoteResult, error) {
	var noteUUID string
	var notebookUUID string
	var noteData NoteType
	if p.UUID == "" {
		noteUUID = strings.ToUpper(generateUUID())
		notebookUUID = p.NotebookUUID
		if notebookUUID == "" {
			notebookUUID = "Inbox"
		}
		// Validate target notebook exists (Inbox always does).
		if notebookUUID != "Inbox" {
			data, _ := NoteBookDB.Get([]byte(notebookUUID))
			var nb NoteBookType
			if err := json.Unmarshal(data, &nb); err != nil || nb.UUID == "" {
				return saveNoteResult{}, errors.New("notebook not found: " + notebookUUID)
			}
		}
		noteData.NoteBookUUID = notebookUUID
		noteData.UUID = noteUUID
	} else {
		noteUUID = p.UUID
		data, _ := NoteDB.Get([]byte(noteUUID))
		if len(data) == 0 {
			return saveNoteResult{}, errors.New("note not found: " + p.UUID)
		}
		if err := json.Unmarshal(data, &noteData); err != nil {
			return saveNoteResult{}, err
		}
		notebookUUID = noteData.NoteBookUUID
	}

	noteData.Title = normalizeWhitespace(p.Title)
	noteData.URL = p.URL
	noteData.SearchIndex = false
	// A note saved through the editor or confirmed after a defuddle
	// refetch is never auto-normalized again.
	if p.ContentState == "refetched" {
		noteData.ContentState = "refetched"
	} else {
		noteData.ContentState = "edited"
	}

	if p.UUID == "" {
		noteData.CreatedAt = int32(time.Now().Unix())
		noteData.UpdatedAt = noteData.CreatedAt
	} else {
		noteData.UpdatedAt = int32(time.Now().Unix())
	}
	if p.Type == "tinymce" {
		p.Type = "text"
	}

	// update file
	noteDir, _ := filepath.Abs(configGlobal.sourceFolder + "/" + notebookUUID + ".qvnotebook/" + noteUUID + ".qvnote")
	os.MkdirAll(noteDir, 0755)
	var meta struct {
		CreatedAt    int32    `json:"created_at"`
		UpdatedAt    int32    `json:"updated_at"`
		Tags         []string `json:"tags"`
		Title        string   `json:"title"`
		UUID         string   `json:"uuid"`
		URL          string   `json:"url_src"`
		ContentState string   `json:"content_state,omitempty"`
	}
	meta.CreatedAt = noteData.CreatedAt
	meta.UpdatedAt = noteData.UpdatedAt
	meta.Title = noteData.Title
	meta.UUID = noteData.UUID
	meta.URL = noteData.URL
	meta.Tags = p.Tags
	meta.ContentState = noteData.ContentState
	metaJSON, _ := json.MarshalIndent(meta, "", "  ")
	err := ioutil.WriteFile(noteDir+"/meta.json", metaJSON, 0644)
	checkQuiet(err)

	var content struct {
		Title string             `json:"title"`
		Cells []ContentCellsType `json:"cells"`
	}
	content.Title = noteData.Title
	// External images are downloaded into resources/ on save, so notes
	// never depend on third-party hotlinks.
	if p.Type != "code" {
		p.Content = downloadNoteImages(noteDir, p.Content)
	}
	content.Cells = make([]ContentCellsType, 1)
	content.Cells[0] = ContentCellsType{Type: p.Type, Data: p.Content}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.Encode(content)

	err = ioutil.WriteFile(noteDir+"/content.json", buf.Bytes(), 0644)
	checkQuiet(err)

	// remove old tags from cloud
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
			if err := TagsDB.Del([]byte(tagID)); err != nil {
				checkQuiet(err)
			}
		} else {
			enc, err := json.Marshal(notesListNew)
			checkQuiet(err)
			TagsDB.Set([]byte(tagID), enc)
		}
	}

	// Add new tags to cloud
	for _, tagID := range p.Tags {
		data, _ := TagsDB.Get([]byte(tagID))
		dataString := string(data)
		var notesList []string
		if dataString != "" {
			err := json.Unmarshal(data, &notesList)
			checkQuiet(err)
		}
		notesList = append(notesList, noteUUID)

		enc, err := json.Marshal(notesList)
		checkQuiet(err)
		err = TagsDB.Set([]byte(tagID), enc)
		checkQuiet(err)
	}

	// add to search index
	addToIndex(noteDir+"/content.json", noteUUID)
	noteData.SearchIndex = true

	// update database
	noteData.Tags = p.Tags
	encNote, _ := json.Marshal(noteData)
	err = NoteDB.Set([]byte(noteUUID), encNote)
	checkQuiet(err)

	// add new note to its notebook
	if p.UUID == "" {
		data, _ := NoteBookDB.Get([]byte(notebookUUID))
		var notebookData NoteBookType
		json.Unmarshal(data, &notebookData)
		if notebookData.Notes == nil {
			notebookData.Notes = make(map[string]int64)
		}
		notebookData.Notes[noteUUID] = time.Now().Unix()
		encData, _ := json.Marshal(notebookData)
		NoteBookDB.Set([]byte(notebookUUID), encData)
	}

	SaveConfig()

	return saveNoteResult{NoteBookUUID: notebookUUID, UUID: noteUUID}, nil
}

// moveNote moves a note between notebooks.
func moveNote(uuid, target string) error {
	if uuid == "" || target == "" {
		return errors.New("uuid and target are required")
	}
	var note NoteType
	data, _ := NoteDB.Get([]byte(uuid))
	if len(data) == 0 {
		return errors.New("note not found: " + uuid)
	}
	json.Unmarshal(data, &note)

	var notebookSRC NoteBookType
	data, _ = NoteBookDB.Get([]byte(note.NoteBookUUID))
	json.Unmarshal(data, &notebookSRC)

	var notebookDST NoteBookType
	data, _ = NoteBookDB.Get([]byte(target))
	json.Unmarshal(data, &notebookDST)
	if notebookDST.UUID == "" {
		return errors.New("target notebook not found: " + target)
	}

	noteDirSrc, _ := filepath.Abs(configGlobal.sourceFolder + "/" + note.NoteBookUUID + ".qvnotebook/" + note.UUID + ".qvnote")
	noteDirDst, _ := filepath.Abs(configGlobal.sourceFolder + "/" + notebookDST.UUID + ".qvnotebook/" + note.UUID + ".qvnote")

	if err := CopyDir(noteDirSrc, noteDirDst); err != nil {
		return err
	}

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
	return nil
}

// deleteNote moves a note to Trash; if it is already in Trash it is removed
// permanently (files, DB, search index).
func deleteNote(uuid string) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}
	var note NoteType
	data, _ := NoteDB.Get([]byte(uuid))
	if len(data) == 0 {
		return errors.New("note not found: " + uuid)
	}
	json.Unmarshal(data, &note)

	var notebookSRC NoteBookType
	data, _ = NoteBookDB.Get([]byte(note.NoteBookUUID))
	json.Unmarshal(data, &notebookSRC)

	if notebookSRC.UUID == "Trash" {
		// permanent delete
		delete(notebookSRC.Notes, note.UUID)
		encSRC, _ := json.Marshal(notebookSRC)
		NoteBookDB.Set([]byte(notebookSRC.UUID), encSRC)

		NoteDB.Del([]byte(note.UUID))

		noteDirSrc, _ := filepath.Abs(configGlobal.sourceFolder + "/" + note.NoteBookUUID + ".qvnotebook/" + note.UUID + ".qvnote")
		os.RemoveAll(noteDirSrc)
		ss.index.Delete(note.UUID)
		return nil
	}

	// move to trash
	var notebookDST NoteBookType
	data, _ = NoteBookDB.Get([]byte("Trash"))
	json.Unmarshal(data, &notebookDST)
	noteDirSrc, _ := filepath.Abs(configGlobal.sourceFolder + "/" + note.NoteBookUUID + ".qvnotebook/" + note.UUID + ".qvnote")
	noteDirDst, _ := filepath.Abs(configGlobal.sourceFolder + "/Trash.qvnotebook/" + note.UUID + ".qvnote")
	if err := CopyDir(noteDirSrc, noteDirDst); err != nil {
		return err
	}

	note.NoteBookUUID = "Trash"
	enc, _ := json.Marshal(note)
	NoteDB.Set([]byte(note.UUID), enc)

	delete(notebookSRC.Notes, note.UUID)
	encSRC, _ := json.Marshal(notebookSRC)
	NoteBookDB.Set([]byte(notebookSRC.UUID), encSRC)

	notebookDST.Notes[note.UUID] = time.Now().Unix()
	encDST, _ := json.Marshal(notebookDST)
	NoteBookDB.Set([]byte(notebookDST.UUID), encDST)

	os.RemoveAll(noteDirSrc)
	return nil
}

// editNotebook handles rename / new / remove. remove moves notes to Trash
// first and refuses to delete Inbox or Trash themselves.
func editNotebook(action, uuid, title string) error {
	switch {
	case action == "rename" && uuid != "":
		var meta struct {
			Name string `json:"name"`
			UUID string `json:"uuid"`
		}
		meta.Name = title
		meta.UUID = uuid
		metaJSON, _ := json.Marshal(meta)
		jsonFile, _ := filepath.Abs(configGlobal.sourceFolder + "/" + uuid + ".qvnotebook/meta.json")
		if err := ioutil.WriteFile(jsonFile, metaJSON, 0644); err != nil {
			return err
		}

		data, _ := NoteBookDB.Get([]byte(uuid))
		var notebookData NoteBookType
		json.Unmarshal(data, &notebookData)
		notebookData.Name = title
		enc, err := json.Marshal(notebookData)
		if err != nil {
			return err
		}
		return NoteBookDB.Set([]byte(uuid), enc)

	case action == "new" && uuid == "":
		u1 := strings.ToUpper(generateUUID())

		notebookDir, _ := filepath.Abs(configGlobal.sourceFolder + "/" + u1 + ".qvnotebook")
		metaFile, _ := filepath.Abs(notebookDir + "/meta.json")
		var meta struct {
			Name string `json:"name"`
			UUID string `json:"uuid"`
		}
		meta.Name = title
		meta.UUID = u1
		metaJSON, _ := json.MarshalIndent(meta, "", "  ")
		os.MkdirAll(notebookDir, 0755)
		if err := ioutil.WriteFile(metaFile, metaJSON, 0644); err != nil {
			return err
		}

		var notebookNew NoteBookType
		notebookNew.Name = title
		notebookNew.UUID = u1
		notebookNew.Notes = make(map[string]int64)
		enc, err := json.Marshal(notebookNew)
		if err != nil {
			return err
		}
		return NoteBookDB.Set([]byte(u1), enc)

	case action == "remove" && uuid != "" && uuid != "Inbox" && uuid != "Trash":
		data, _ := NoteBookDB.Get([]byte("Trash"))
		var notebookDataTrash NoteBookType
		json.Unmarshal(data, &notebookDataTrash)

		data, _ = NoteBookDB.Get([]byte(uuid))
		var notebookData NoteBookType
		json.Unmarshal(data, &notebookData)
		if notebookData.UUID == "" {
			return errors.New("notebook not found: " + uuid)
		}
		canDelete := true
		for noteUUID := range notebookData.Notes {
			var note NoteType
			data, _ := NoteDB.Get([]byte(noteUUID))
			json.Unmarshal(data, &note)
			noteDirSrc, _ := filepath.Abs(configGlobal.sourceFolder + "/" + uuid + ".qvnotebook/" + noteUUID + ".qvnote")
			noteDirDst, _ := filepath.Abs(configGlobal.sourceFolder + "/Trash.qvnotebook/" + noteUUID + ".qvnote")

			if err := CopyDir(noteDirSrc, noteDirDst); err == nil {
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
			srcFolder, _ := filepath.Abs(configGlobal.sourceFolder + "/" + uuid + ".qvnotebook/")
			os.RemoveAll(srcFolder)
			NoteBookDB.Del([]byte(uuid))
			return nil
		}
		return errors.New("failed to move all notes to trash")
	}
	return nil
}

// editTag handles rename (with merge) / remove of a tag across all notes.
func editTag(action, urlName, title string) error {
	if urlName == "" {
		return errors.New("tag name is required")
	}
	if action != "rename" && action != "remove" {
		return errors.New("unknown action: " + action)
	}
	if action == "rename" && (urlName == "" || urlName == title) {
		return nil
	}

	data, _ := TagsDB.Get([]byte(urlName))
	if string(data) == "" {
		return errors.New("tag not found: " + urlName)
	}
	var tagsData []string
	if err := json.Unmarshal(data, &tagsData); err != nil {
		return err
	}

	for _, noteID := range tagsData {
		dataNote, _ := NoteDB.Get([]byte(noteID))
		if string(dataNote) == "" {
			continue
		}
		var note NoteType
		if err := json.Unmarshal(dataNote, &note); err != nil {
			checkQuiet(err)
			continue
		}

		metaFile, _ := filepath.Abs(configGlobal.sourceFolder + "/" + note.NoteBookUUID + ".qvnotebook/" + note.UUID + ".qvnote/meta.json")
		jsonFile, err := os.Open(metaFile)
		if err != nil {
			continue
		}
		byteValue, _ := ioutil.ReadAll(jsonFile)
		json.Unmarshal(byteValue, &note)
		jsonFile.Close()

		var tagsNew = make([]string, 0)
		for _, tagName := range note.Tags {
			if tagName != urlName && tagName != title {
				tagsNew = append(tagsNew, tagName)
			}
		}
		switch action {
		case "rename":
			tagsNew = append(tagsNew, title)
		case "remove":
			// do nothing
		}
		note.Tags = tagsNew

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
		if err := ioutil.WriteFile(metaFile, metaJSON, 0644); err != nil {
			checkQuiet(err)
		}

		enc, _ := json.Marshal(note)
		NoteDB.Set([]byte(note.UUID), enc)
	}

	if action == "remove" {
		TagsDB.Del([]byte(urlName))
	} else if action == "rename" {
		TagsDB.Del([]byte(urlName))

		// add new data (merge with an existing tag of the same name)
		data, _ := TagsDB.Get([]byte(title))
		if string(data) != "" {
			var tagsDataExist []string
			if err := json.Unmarshal(data, &tagsDataExist); err != nil {
				checkQuiet(err)
			}
			for _, tagName := range tagsDataExist {
				if !inArray(tagName, tagsData) {
					tagsData = append(tagsData, tagName)
				}
			}
		}

		enc, err := json.Marshal(tagsData)
		if err != nil {
			return err
		}
		TagsDB.Set([]byte(title), enc)
	}
	return nil
}

// setFavorite adds or removes a note UUID from the favorites bucket.
func setFavorite(action, uuid string) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}
	switch action {
	case "add":
		return FavoritesDB.Set([]byte(uuid), []byte(""))
	case "remove":
		return FavoritesDB.Del([]byte(uuid))
	default:
		return errors.New("unknown action: " + action)
	}
}
