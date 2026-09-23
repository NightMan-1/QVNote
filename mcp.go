package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCP (Model Context Protocol) Streamable HTTP server on /mcp.
// Read tools are available whenever MCP is enabled; write tools are gated
// by configGlobal.mcpAllowWrite. Bearer token auth is mandatory.

const mcpServerName = "qvnote"
const mcpServerVersion = "1.0.0"

var (
	mcpServerOnce sync.Once
	mcpServerInst *mcp.Server
)

func mcpWriteAllowed() bool {
	return configGlobal.mcpAllowWrite
}

// mcpTextResult returns plain text content (LLM-friendly).
func mcpTextResult(text string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}

// mcpStructResult returns structured JSON (also serialized as Content by the SDK).
func mcpStructResult(v any) (*mcp.CallToolResult, any, error) {
	return nil, v, nil
}

// mcpJSON marshals v for a human-readable / LLM-readable text content.
func mcpJSON(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ---------- Read tool inputs ----------

type mcpSearchNotesIn struct {
	Query string `json:"query" jsonschema:"search query (3+ characters)"`
	Limit int    `json:"limit,omitempty" jsonschema:"max results, default 20"`
}

type mcpGetNoteIn struct {
	UUID   string `json:"uuid" jsonschema:"note UUID"`
	Format string `json:"format,omitempty" jsonschema:"content format: markdown (default) or html"`
}

type mcpListNotesIn struct {
	NotebookUUID string `json:"notebook_uuid,omitempty" jsonschema:"notebook UUID; special values: Allnotes, Favorites; empty = Allnotes"`
	Limit        int    `json:"limit,omitempty" jsonschema:"max results, default 50"`
}

type mcpGetNotesByTagIn struct {
	Tag   string `json:"tag" jsonschema:"tag name"`
	Limit int    `json:"limit,omitempty" jsonschema:"max results, default 50"`
}

// ---------- Write tool inputs ----------

type mcpCreateNoteIn struct {
	Title        string   `json:"title" jsonschema:"note title"`
	Content      string   `json:"content" jsonschema:"note content"`
	NotebookUUID string   `json:"notebook_uuid,omitempty" jsonschema:"target notebook UUID; empty = Inbox"`
	Format       string   `json:"format,omitempty" jsonschema:"content format: markdown (default) or html"`
	Tags         []string `json:"tags,omitempty" jsonschema:"optional tags"`
	URL          string   `json:"url,omitempty" jsonschema:"optional source URL"`
	Type         string   `json:"type,omitempty" jsonschema:"note type: text (default) or code"`
}

type mcpUpdateNoteIn struct {
	UUID    string   `json:"uuid" jsonschema:"note UUID"`
	Title   *string  `json:"title,omitempty" jsonschema:"new title; omit to keep"`
	Content *string  `json:"content,omitempty" jsonschema:"new content; omit to keep"`
	Format  string   `json:"format,omitempty" jsonschema:"content format for content field: markdown (default) or html"`
	Tags    []string `json:"tags,omitempty" jsonschema:"new tags list; omit to keep"`
	URL     *string  `json:"url,omitempty" jsonschema:"source URL; omit to keep"`
	Type    string   `json:"type,omitempty" jsonschema:"note type when content is set: text (default) or code"`
}

type mcpUUIDIn struct {
	UUID string `json:"uuid" jsonschema:"note UUID"`
}

type mcpMoveNoteIn struct {
	UUID         string `json:"uuid" jsonschema:"note UUID"`
	NotebookUUID string `json:"notebook_uuid" jsonschema:"target notebook UUID"`
}

type mcpCreateNotebookIn struct {
	Name string `json:"name" jsonschema:"notebook name"`
}

type mcpNotebookUUIDIn struct {
	UUID string `json:"uuid" jsonschema:"notebook UUID"`
}

type mcpRenameNotebookIn struct {
	UUID string `json:"uuid" jsonschema:"notebook UUID"`
	Name string `json:"name" jsonschema:"new notebook name"`
}

type mcpTagIn struct {
	Tag string `json:"tag" jsonschema:"tag name"`
}

type mcpRenameTagIn struct {
	Tag string `json:"tag" jsonschema:"current tag name"`
	New string `json:"new" jsonschema:"new tag name"`
}

// ---------- helpers ----------

// mcpContentToHTML converts incoming tool content to HTML for storage.
// format "html" is stored as-is; "markdown" (default) is converted via goldmark.
func mcpContentToHTML(content, format string) (string, error) {
	if strings.EqualFold(format, "html") {
		return content, nil
	}
	var buf strings.Builder
	if err := mdToHTML.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// mcpContentFromHTML converts stored note HTML for the requested read format.
// "html" returns as-is; "markdown" (default) converts via HTMLToMarkdown.
func mcpContentFromHTML(content, format string) (string, error) {
	if strings.EqualFold(format, "html") {
		return content, nil
	}
	return HTMLToMarkdown(content)
}

// mcpNotePayload is the JSON body returned by get_note and similar tools.
type mcpNotePayload struct {
	UUID         string   `json:"uuid"`
	NoteBookUUID string   `json:"notebook_uuid"`
	Title        string   `json:"title"`
	URL          string   `json:"url,omitempty"`
	Type         string   `json:"type"`
	Tags         []string `json:"tags,omitempty"`
	Content      string   `json:"content"`
	Format       string   `json:"format"`
	CreatedAt    int32    `json:"created_at"`
	UpdatedAt    int32    `json:"updated_at"`
	Favorites    bool     `json:"favorites"`
}

func mcpLoadNotePayload(uuid, format string) (*mcpNotePayload, error) {
	note := loadNote(uuid, false, "")
	if note == nil {
		return nil, fmt.Errorf("note not found: %s", uuid)
	}
	content, err := mcpContentFromHTML(note.Content, format)
	if err != nil {
		return nil, err
	}
	if format == "" || strings.EqualFold(format, "markdown") {
		format = "markdown"
	} else {
		format = "html"
	}
	return &mcpNotePayload{
		UUID:         note.UUID,
		NoteBookUUID: note.NoteBookUUID,
		Title:        note.Title,
		URL:          note.URL,
		Type:         note.ContentType,
		Tags:         note.Tags,
		Content:      content,
		Format:       format,
		CreatedAt:    note.CreatedAt,
		UpdatedAt:    note.UpdatedAt,
		Favorites:    note.Favorites,
	}, nil
}

func mcpShortNote(n NoteTypeAPI) map[string]any {
	return map[string]any{
		"uuid":          n.UUID,
		"notebook_uuid": n.NoteBookUUID,
		"title":         n.Title,
		"updated_at":    n.UpdatedAt,
	}
}

func mcpShortSearch(n SearchResult) map[string]any {
	return map[string]any{
		"uuid":          n.UUID,
		"notebook_uuid": n.NoteBookUUID,
		"title":         n.Title,
	}
}

// ---------- read tool handlers ----------

func mcpToolSearchNotes(ctx context.Context, req *mcp.CallToolRequest, in mcpSearchNotesIn) (*mcp.CallToolResult, any, error) {
	results := searchNotes(in.Query)
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	out := make([]map[string]any, 0, limit)
	for i, n := range results {
		if i >= limit {
			break
		}
		out = append(out, mcpShortSearch(n))
	}
	return mcpStructResult(out)
}

func mcpToolGetNote(ctx context.Context, req *mcp.CallToolRequest, in mcpGetNoteIn) (*mcp.CallToolResult, any, error) {
	payload, err := mcpLoadNotePayload(in.UUID, in.Format)
	if err != nil {
		return nil, nil, err
	}
	return mcpStructResult(payload)
}

func mcpToolListNotes(ctx context.Context, req *mcp.CallToolRequest, in mcpListNotesIn) (*mcp.CallToolResult, any, error) {
	nb := in.NotebookUUID
	if nb == "" {
		nb = "Allnotes"
	}
	notes := listNotesAtNotebook(nb)
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	out := make([]map[string]any, 0, limit)
	for i, n := range notes {
		if i >= limit {
			break
		}
		out = append(out, mcpShortNote(n))
	}
	return mcpStructResult(out)
}

func mcpToolListNotebooks(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
	nbs := listNotebooks()
	out := make([]map[string]any, 0, len(nbs))
	for _, nb := range nbs {
		out = append(out, map[string]any{
			"uuid":        nb.UUID,
			"name":        nb.Name,
			"notes_count": nb.NotesCount,
		})
	}
	return mcpStructResult(out)
}

func mcpToolListTags(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
	tags := listTags()
	out := make([]map[string]any, 0, len(tags))
	for _, t := range tags {
		out = append(out, map[string]any{
			"name":  t.Name,
			"count": t.Count,
		})
	}
	return mcpStructResult(out)
}

func mcpToolGetNotesByTag(ctx context.Context, req *mcp.CallToolRequest, in mcpGetNotesByTagIn) (*mcp.CallToolResult, any, error) {
	notes := listNotesByTag(in.Tag)
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	out := make([]map[string]any, 0, limit)
	for i, n := range notes {
		if i >= limit {
			break
		}
		out = append(out, mcpShortNote(n))
	}
	return mcpStructResult(out)
}

// ---------- write tool handlers ----------

func mcpRequireWrite() error {
	if !mcpWriteAllowed() {
		return fmt.Errorf("write operations are disabled (mcpAllowWrite is off)")
	}
	return nil
}

func mcpToolCreateNote(ctx context.Context, req *mcp.CallToolRequest, in mcpCreateNoteIn) (*mcp.CallToolResult, any, error) {
	if err := mcpRequireWrite(); err != nil {
		return nil, nil, err
	}
	noteType := in.Type
	if noteType == "" {
		noteType = "text"
	}
	content := in.Content
	if noteType != "code" {
		var err error
		content, err = mcpContentToHTML(in.Content, in.Format)
		if err != nil {
			return nil, nil, err
		}
	}
	res, err := saveNote(saveNoteParams{
		UUID:         "",
		Title:        in.Title,
		URL:          in.URL,
		Type:         noteType,
		Content:      content,
		Tags:         in.Tags,
		ContentState: "edited",
		NotebookUUID: in.NotebookUUID,
	})
	if err != nil {
		return nil, nil, err
	}
	return mcpStructResult(res)
}

func mcpToolUpdateNote(ctx context.Context, req *mcp.CallToolRequest, in mcpUpdateNoteIn) (*mcp.CallToolResult, any, error) {
	if err := mcpRequireWrite(); err != nil {
		return nil, nil, err
	}
	existing := loadNote(in.UUID, true, "") // raw: keep stored quiver-image-url placeholders
	if existing == nil {
		return nil, nil, fmt.Errorf("note not found: %s", in.UUID)
	}
	title := existing.Title
	if in.Title != nil {
		title = *in.Title
	}
	url := existing.URL
	if in.URL != nil {
		url = *in.URL
	}
	noteType := in.Type
	if noteType == "" {
		noteType = existing.ContentType
		if noteType == "" {
			noteType = "text"
		}
	}
	tags := existing.Tags
	if in.Tags != nil {
		tags = in.Tags
	}
	// Content: only rewritten when provided; otherwise keep stored HTML as-is.
	content := existing.Content
	if in.Content != nil {
		if noteType == "code" {
			content = *in.Content
		} else {
			html, err := mcpContentToHTML(*in.Content, in.Format)
			if err != nil {
				return nil, nil, err
			}
			content = html
		}
	}
	// contentState: "edited" so future reads don't re-normalize a write path
	// edit; for no-content-change updates we keep the existing state via a
	// temporary override — saveNote always sets edited/refetched, so we pass
	// "edited" only when content changed, else reuse existing.
	contentState := "edited"
	if in.Content == nil {
		// Preserve original state so pure title/tag edits don't flip
		// refetched notes into always-normalized path semantics.
		contentState = existing.ContentState
		if contentState == "" {
			contentState = "edited"
		}
	}
	res, err := saveNote(saveNoteParams{
		UUID:         in.UUID,
		Title:        title,
		URL:          url,
		Type:         noteType,
		Content:      content,
		Tags:         tags,
		ContentState: contentState,
	})
	if err != nil {
		return nil, nil, err
	}
	return mcpStructResult(res)
}

func mcpToolDeleteNote(ctx context.Context, req *mcp.CallToolRequest, in mcpUUIDIn) (*mcp.CallToolResult, any, error) {
	if err := mcpRequireWrite(); err != nil {
		return nil, nil, err
	}
	if err := deleteNote(in.UUID); err != nil {
		return nil, nil, err
	}
	return mcpStructResult(map[string]bool{"ok": true})
}

func mcpToolMoveNote(ctx context.Context, req *mcp.CallToolRequest, in mcpMoveNoteIn) (*mcp.CallToolResult, any, error) {
	if err := mcpRequireWrite(); err != nil {
		return nil, nil, err
	}
	if err := moveNote(in.UUID, in.NotebookUUID); err != nil {
		return nil, nil, err
	}
	return mcpStructResult(map[string]bool{"ok": true})
}

func mcpToolCreateNotebook(ctx context.Context, req *mcp.CallToolRequest, in mcpCreateNotebookIn) (*mcp.CallToolResult, any, error) {
	if err := mcpRequireWrite(); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if err := editNotebook("new", "", in.Name); err != nil {
		return nil, nil, err
	}
	return mcpStructResult(map[string]bool{"ok": true})
}

func mcpToolRenameNotebook(ctx context.Context, req *mcp.CallToolRequest, in mcpRenameNotebookIn) (*mcp.CallToolResult, any, error) {
	if err := mcpRequireWrite(); err != nil {
		return nil, nil, err
	}
	if in.UUID == "Inbox" || in.UUID == "Trash" {
		return nil, nil, fmt.Errorf("cannot rename system notebook")
	}
	if err := editNotebook("rename", in.UUID, in.Name); err != nil {
		return nil, nil, err
	}
	return mcpStructResult(map[string]bool{"ok": true})
}

func mcpToolDeleteNotebook(ctx context.Context, req *mcp.CallToolRequest, in mcpNotebookUUIDIn) (*mcp.CallToolResult, any, error) {
	if err := mcpRequireWrite(); err != nil {
		return nil, nil, err
	}
	if in.UUID == "Inbox" || in.UUID == "Trash" {
		return nil, nil, fmt.Errorf("cannot delete system notebook")
	}
	if err := editNotebook("remove", in.UUID, ""); err != nil {
		return nil, nil, err
	}
	return mcpStructResult(map[string]bool{"ok": true})
}

func mcpToolRenameTag(ctx context.Context, req *mcp.CallToolRequest, in mcpRenameTagIn) (*mcp.CallToolResult, any, error) {
	if err := mcpRequireWrite(); err != nil {
		return nil, nil, err
	}
	if in.Tag == "" || in.New == "" {
		return nil, nil, fmt.Errorf("tag and new are required")
	}
	if err := editTag("rename", in.Tag, in.New); err != nil {
		return nil, nil, err
	}
	return mcpStructResult(map[string]bool{"ok": true})
}

func mcpToolDeleteTag(ctx context.Context, req *mcp.CallToolRequest, in mcpTagIn) (*mcp.CallToolResult, any, error) {
	if err := mcpRequireWrite(); err != nil {
		return nil, nil, err
	}
	if in.Tag == "" {
		return nil, nil, fmt.Errorf("tag is required")
	}
	if err := editTag("remove", in.Tag, ""); err != nil {
		return nil, nil, err
	}
	return mcpStructResult(map[string]bool{"ok": true})
}

// ---------- resources ----------

func mcpReadNoteResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI
	// qvnote://note/{uuid}
	parts := strings.SplitN(strings.TrimPrefix(uri, "qvnote://note/"), "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	payload, err := mcpLoadNotePayload(parts[0], "markdown")
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	body, err := mcpJSON(payload)
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{URI: uri, Text: body, MIMEType: "application/json"}},
	}, nil
}

func mcpReadNotebookResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI
	id := strings.TrimPrefix(uri, "qvnote://notebook/")
	if id == "" {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	notes := listNotesAtNotebook(id)
	out := make([]map[string]any, 0, len(notes))
	for _, n := range notes {
		out = append(out, mcpShortNote(n))
	}
	body, err := mcpJSON(out)
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{URI: uri, Text: body, MIMEType: "application/json"}},
	}, nil
}

func mcpReadTagResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI
	tag := strings.TrimPrefix(uri, "qvnote://tag/")
	if tag == "" {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	notes := listNotesByTag(tag)
	out := make([]map[string]any, 0, len(notes))
	for _, n := range notes {
		out = append(out, mcpShortNote(n))
	}
	body, err := mcpJSON(out)
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{URI: uri, Text: body, MIMEType: "application/json"}},
	}, nil
}

// ---------- server construction ----------

func buildMCPServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: mcpServerName, Version: mcpServerVersion}, &mcp.ServerOptions{
		Instructions: "QVNote MCP server. Search and read notes, notebooks and tags. " +
			"Write operations (create/update/delete/move) are controlled by server settings.",
	})

	// Read tools (always registered; gate is mcpEnabled at HTTP layer)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_notes",
		Description: "Full-text search over note titles and content (query must be 3+ characters).",
	}, mcpToolSearchNotes)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_note",
		Description: "Read a single note by UUID. Content format: markdown (default) or html.",
	}, mcpToolGetNote)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_notes",
		Description: "List notes in a notebook, or all notes / favorites (special ids Allnotes, Favorites).",
	}, mcpToolListNotes)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_notebooks",
		Description: "List all notebooks with note counts.",
	}, mcpToolListNotebooks)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_tags",
		Description: "List all tags with usage counts.",
	}, mcpToolListTags)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_notes_by_tag",
		Description: "List notes that carry a given tag.",
	}, mcpToolGetNotesByTag)

	// Write tools (gated by mcpAllowWrite inside each handler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_note",
		Description: "Create a new note. Content is markdown by default (set format=html to store raw HTML).",
	}, mcpToolCreateNote)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_note",
		Description: "Update an existing note. Omitted fields are kept. Content format: markdown (default) or html.",
	}, mcpToolUpdateNote)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_note",
		Description: "Delete a note (moves to Trash; from Trash deletes permanently).",
	}, mcpToolDeleteNote)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "move_note",
		Description: "Move a note to another notebook.",
	}, mcpToolMoveNote)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_notebook",
		Description: "Create a new notebook.",
	}, mcpToolCreateNotebook)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rename_notebook",
		Description: "Rename a notebook (Inbox and Trash cannot be renamed).",
	}, mcpToolRenameNotebook)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_notebook",
		Description: "Delete a notebook (notes move to Trash; Inbox and Trash cannot be deleted).",
	}, mcpToolDeleteNotebook)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rename_tag",
		Description: "Rename a tag across all notes (merges if the new name already exists).",
	}, mcpToolRenameTag)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_tag",
		Description: "Delete a tag from all notes.",
	}, mcpToolDeleteTag)

	// Resources
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "qvnote://note/{uuid}",
		Name:        "note",
		Description: "A single note (metadata + markdown content).",
		MIMEType:    "application/json",
	}, mcpReadNoteResource)

	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "qvnote://notebook/{uuid}",
		Name:        "notebook",
		Description: "All notes in a notebook (or special id Allnotes / Favorites).",
		MIMEType:    "application/json",
	}, mcpReadNotebookResource)

	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "qvnote://tag/{name}",
		Name:        "tag",
		Description: "All notes carrying a given tag.",
		MIMEType:    "application/json",
	}, mcpReadTagResource)

	return s
}

func getMCPServer() *mcp.Server {
	mcpServerOnce.Do(func() {
		mcpServerInst = buildMCPServer()
	})
	return mcpServerInst
}

// mcpAuthMiddleware enforces mcpEnabled + bearer token.
func mcpAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !configGlobal.mcpEnabled {
			http.Error(w, "MCP is disabled", http.StatusForbidden)
			return
		}
		auth := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="qvnote-mcp"`)
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
		if configGlobal.mcpToken == "" || subtle.ConstantTimeCompare([]byte(token), []byte(configGlobal.mcpToken)) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="qvnote-mcp"`)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// registerMCPRoutes mounts the Streamable HTTP MCP endpoint on /mcp.
func registerMCPRoutes(r chi.Router) {
	h := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return getMCPServer()
	}, &mcp.StreamableHTTPOptions{})
	r.Handle("/mcp", mcpAuthMiddleware(h))
	r.Handle("/mcp/*", mcpAuthMiddleware(h))
}
