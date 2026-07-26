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
