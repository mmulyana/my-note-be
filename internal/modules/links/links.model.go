package links

import "time"

type Link struct {
	ID          string    `gorm:"type:text;primaryKey" json:"id"`
	NoteID      string    `gorm:"type:text;not null;index" json:"noteId"`
	URL         string    `gorm:"column:url;type:text;not null;default:''" json:"url"`
	Title       string    `gorm:"type:text;not null;default:''" json:"title"`
	Description string    `gorm:"type:text;not null;default:''" json:"description"`
	Image       string    `gorm:"type:text;not null;default:''" json:"image"`
	Favicon     string    `gorm:"type:text;not null;default:''" json:"favicon"`
	SiteName    string    `gorm:"column:site_name;type:text;not null;default:''" json:"siteName"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
