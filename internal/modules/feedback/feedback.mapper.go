package feedback

type relayHubItemRequest struct {
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Priority      string         `json:"priority"`
	ReporterName  string         `json:"reporter_name"`
	ReporterEmail string         `json:"reporter_email"`
	UserID        string         `json:"user_id"`
	CustomFields  map[string]any `json:"custom_fields,omitempty"`
}

type relayHubItemResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func toRelayHubRequest(in FeedbackInput, reporterName, reporterEmail, userID string) relayHubItemRequest {
	return relayHubItemRequest{
		Type:          in.Type,
		Title:         in.Title,
		Description:   in.Description,
		Priority:      defaultPriority,
		ReporterName:  reporterName,
		ReporterEmail: reporterEmail,
		UserID:        userID,
		CustomFields:  in.CustomFields,
	}
}

func toFeedbackResponse(out relayHubItemResponse) FeedbackResponse {
	return FeedbackResponse{
		ID:        out.ID,
		Type:      out.Type,
		Title:     out.Title,
		Status:    out.Status,
		CreatedAt: out.CreatedAt,
	}
}
