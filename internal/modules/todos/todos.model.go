package todos

import (
	"encoding/json"
	"time"
)

type Priority string

const (
	PriorityNone   Priority = "none"
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Todo struct {
	ID        string          `gorm:"type:text;primaryKey" json:"id"`
	NoteID    string          `gorm:"type:text;not null;index" json:"noteId"`
	Text      string          `gorm:"type:text;not null;default:''" json:"text"`
	Checked   bool            `gorm:"not null;default:false" json:"checked"`
	Deadline  *time.Time      `gorm:"type:date" json:"-"`
	Today     *time.Time      `gorm:"type:date" json:"-"`
	Priority  Priority        `gorm:"type:text;not null;default:'none'" json:"priority"`
	Tags      json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"-"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}
