package releases

import "time"

// Input is the authoring payload. Every field but the title is optional, so a
// draft can be created with a title alone and filled in later.
type Input struct {
	Version string `json:"version"`
	Title   string `json:"title" binding:"required"`
	Tag     Tag    `json:"tag"`
	Summary string `json:"summary"`
	Content     string     `json:"content"`
	Image       string     `json:"image"`
	PublishedAt *time.Time `json:"publishedAt"`
}

type Response struct {
	ID          string     `json:"id"`
	Version     string     `json:"version"`
	Title       string     `json:"title"`
	Tag         Tag        `json:"tag"`
	Summary     string     `json:"summary"`
	Content     string     `json:"content"`
	Image       string     `json:"image"`
	PublishedAt *time.Time `json:"publishedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type UnreadResponse struct {
	Count      int64      `json:"count"`
	LastSeenAt *time.Time `json:"lastSeenAt"`
}
