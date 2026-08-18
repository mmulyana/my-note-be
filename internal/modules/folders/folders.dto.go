package folders

import (
	"time"

	"github.com/google/uuid"
)

type FolderInput struct {
	Name   string `json:"name" binding:"required"`
	Color  string `json:"color"`
	Secret bool   `json:"secret"`
}

type FolderResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Secret    bool      `json:"secret"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (in FolderInput) ToModel(userID uuid.UUID) Folder {
	color := in.Color
	if color == "" {
		color = "default"
	}
	return Folder{UserID: userID, Name: in.Name, Color: color, Secret: in.Secret}
}

func ToResponse(f Folder) FolderResponse {
	return FolderResponse{
		ID:        f.ID.String(),
		Name:      f.Name,
		Color:     f.Color,
		Secret:    f.Secret,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

func ToResponses(folders []Folder) []FolderResponse {
	out := make([]FolderResponse, len(folders))
	for i, f := range folders {
		out[i] = ToResponse(f)
	}
	return out
}

type FolderNoteResponse struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type FolderWithNotesResponse struct {
	FolderResponse
	Notes []FolderNoteResponse `json:"notes"`
}

func ToWithNotesResponse(f Folder, notes []FolderNote) FolderWithNotesResponse {
	out := make([]FolderNoteResponse, len(notes))
	for i, n := range notes {
		out[i] = FolderNoteResponse{Title: n.Title, Text: n.Text}
	}
	return FolderWithNotesResponse{FolderResponse: ToResponse(f), Notes: out}
}

func ToWithNotesResponses(folders []Folder, notesByFolder map[uuid.UUID][]FolderNote) []FolderWithNotesResponse {
	out := make([]FolderWithNotesResponse, len(folders))
	for i, f := range folders {
		out[i] = ToWithNotesResponse(f, notesByFolder[f.ID])
	}
	return out
}
