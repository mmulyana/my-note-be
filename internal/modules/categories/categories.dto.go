package categories

type CategoryInput struct {
	Name string `json:"name" binding:"required"`
}

type CategoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
