package releases

import (
	"time"

	"github.com/google/uuid"
)

type Tag string

const (
	TagNew      Tag = "new"
	TagImproved Tag = "improved"
	TagFixed    Tag = "fixed"
)

func (t Tag) valid() bool {
	switch t {
	case TagNew, TagImproved, TagFixed:
		return true
	}
	return false
}

type Release struct {
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	Version string    `gorm:"type:text;not null;default:''" json:"version"`
	Title   string    `gorm:"type:text;not null" json:"title"`
	Tag     Tag       `gorm:"type:text;not null;default:'new'" json:"tag"`
	Summary string    `gorm:"type:text;not null;default:''" json:"summary"`
	Content string    `gorm:"type:text;not null;default:''" json:"content"`
	Image   string    `gorm:"type:text;not null;default:''" json:"image"`
	// note: NULL is draft
	PublishedAt *time.Time `gorm:"column:published_at" json:"publishedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
