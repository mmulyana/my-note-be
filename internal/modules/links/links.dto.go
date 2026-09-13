package links

import "time"

type Preview struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Favicon     string `json:"favicon"`
	SiteName    string `json:"siteName"`
}

type Response struct {
	ID          string    `json:"id"`
	NoteID      string    `json:"noteId"`
	NoteTitle   string    `json:"noteTitle"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Favicon     string    `json:"favicon"`
	SiteName    string    `json:"siteName"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type row struct {
	Link      `gorm:"embedded"`
	NoteTitle string
}
