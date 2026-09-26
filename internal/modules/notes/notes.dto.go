package notes

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CreateNoteInput struct {
	ID       string     `json:"id" binding:"required"`
	Content  string     `json:"content"`
	TodoDiff TodoDiff   `json:"todoDiff"`
	LinkDiff LinkDiff   `json:"linkDiff"`
	Labels   []string   `json:"labels"`
	FolderID *uuid.UUID `json:"folderId"`
}

type SaveNoteInput struct {
	Content  string     `json:"content" binding:"required"`
	Preview  string     `json:"preview"`
	TodoDiff TodoDiff   `json:"todoDiff"`
	LinkDiff LinkDiff   `json:"linkDiff"`
	Labels   []string   `json:"labels"`
	FolderID *uuid.UUID `json:"folderId"`
	Pinned   *bool      `json:"pinned"`
	Archived *bool      `json:"archived"`
	Secret   *bool      `json:"secret"`
}

type CountsResponse struct {
	Notes   int64            `json:"notes"`
	Todos   int64            `json:"todos"`
	Labels  int64            `json:"labels"`
	Archive int64            `json:"archive"`
	Folders map[string]int64 `gorm:"-" json:"folders"`
}

type TodoDiff struct {
	Added   []TodoInput  `json:"added"`
	Updated []TodoUpdate `json:"updated"`
	Removed []string     `json:"removed"`
}

type TodoInput struct {
	ID       string       `json:"id"`
	Checked  bool         `json:"checked"`
	Text     string       `json:"text"`
	Deadline *string      `json:"deadline"`
	Today    *string      `json:"today"`
	Priority TodoPriority `json:"priority"`
}

type TodoUpdate struct {
	ID     string                     `json:"id"`
	Fields map[string]json.RawMessage `json:"fields"`
}

// LinkDiff: kontraknya sama kayak TodoDiff — id dipegang client, yang dikirim cuma yang berubah
type LinkDiff struct {
	Added   []LinkInput  `json:"added"`
	Updated []LinkUpdate `json:"updated"`
	Removed []string     `json:"removed"`
}

type LinkInput struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Favicon     string `json:"favicon"`
	SiteName    string `json:"siteName"`
}

type LinkUpdate struct {
	ID     string                     `json:"id"`
	Fields map[string]json.RawMessage `json:"fields"`
}
