package notes

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CreateNoteInput struct {
	ID       string      `json:"id" binding:"required"`
	Content  string      `json:"content"`
	TodoDiff TodoDiff    `json:"todoDiff"`
	LabelIDs []uuid.UUID `json:"labelIds"`
	FolderID *uuid.UUID  `json:"folderId"`
}

type SaveNoteInput struct {
	Content  string      `json:"content" binding:"required"`
	Preview  string      `json:"preview"`
	TodoDiff TodoDiff    `json:"todoDiff"`
	LabelIDs []uuid.UUID `json:"labelIds"`
	FolderID *uuid.UUID  `json:"folderId"`
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
	Priority TodoPriority `json:"priority"`
}

type TodoUpdate struct {
	ID     string                     `json:"id"`
	Fields map[string]json.RawMessage `json:"fields"`
}
