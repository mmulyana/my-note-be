package labels

type LabelInput struct {
	Name string `json:"name" binding:"required"`
}

type LabelResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
