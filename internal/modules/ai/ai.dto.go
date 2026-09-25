package ai

type StreamInput struct {
	Action    string `json:"action" binding:"required"`
	Prompt    string `json:"prompt"`
	Selection string `json:"selection"`
}
