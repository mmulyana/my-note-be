package notes

import (
	"encoding/json"
	"time"

	"my-note-be/internal/modules/categories"

	"github.com/google/uuid"
)

type TodoPriority string

const (
	PriorityLow    TodoPriority = "low"
	PriorityMedium TodoPriority = "medium"
	PriorityHigh   TodoPriority = "high"
)

type Note struct {
	ID        string     `gorm:"type:text;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null" json:"userId"`
	FolderID  *uuid.UUID `gorm:"type:uuid;null" json:"folderId"`
	Title     string    `gorm:"size:200;not null;default:''" json:"title"`
	Preview   string    `gorm:"type:text;not null;default:''" json:"preview"`
	Text      string    `gorm:"type:text;not null;default:''" json:"-"`
	Content   string    `gorm:"type:text;not null;default:''" json:"content"`
	TodoTotal int       `gorm:"column:todo_total;not null;default:0" json:"todoTotal"`
	TodoDone  int       `gorm:"column:todo_done;not null;default:0" json:"todoDone"`
	Archived  bool      `gorm:"not null;default:false" json:"archived"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Todos      []Todo                `gorm:"foreignKey:NoteID;constraint:OnDelete:CASCADE" json:"-"`
	Categories []categories.Category `gorm:"many2many:note_categories;" json:"-"`
}

type Todo struct {
	ID        string          `gorm:"type:text;primaryKey" json:"id"`
	NoteID    string          `gorm:"type:text;not null;index" json:"noteId"`
	Text      string          `gorm:"type:text;not null;default:''" json:"text"`
	Checked   bool            `gorm:"not null;default:false" json:"checked"`
	Deadline  *time.Time      `gorm:"type:date" json:"-"`
	Priority  TodoPriority    `gorm:"type:text;not null;default:'medium'" json:"priority"`
	Tags      json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"-"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}
