package feedback

type FeedbackInput struct {
	Type         string         `json:"type" binding:"required"`
	Title        string         `json:"title" binding:"required"`
	Description  string         `json:"description"`
	CustomFields map[string]any `json:"customFields"`
}

type FeedbackResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}
