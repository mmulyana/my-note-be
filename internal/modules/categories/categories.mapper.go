package categories

import (
	"strings"

	"github.com/google/uuid"
)

func (in CategoryInput) ToModel(userID uuid.UUID) Category {
	return Category{UserID: userID, Name: strings.TrimSpace(in.Name)}
}

func ToResponse(c Category) CategoryResponse {
	res := CategoryResponse{}
	res.ID = c.ID.String()
	res.Name = c.Name
	return res
}

func ToResponses(cats []Category) []CategoryResponse {
	out := make([]CategoryResponse, len(cats))
	for i, c := range cats {
		out[i] = ToResponse(c)
	}
	return out
}

func Names(cats []Category) []string {
	out := make([]string, len(cats))
	for i, c := range cats {
		out[i] = c.Name
	}
	return out
}
