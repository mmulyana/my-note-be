package todos

func ToResponse(t Todo) TodoResponse {
	res := TodoResponse{
		ID:        t.ID,
		NoteID:    t.NoteID,
		Text:      t.Text,
		Checked:   t.Checked,
		Priority:  t.Priority,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	if t.Deadline != nil {
		d := t.Deadline.Format("2006-01-02")
		res.Deadline = &d
	}
	return res
}

func ToResponses(todos []Todo) []TodoResponse {
	out := make([]TodoResponse, len(todos))
	for i, t := range todos {
		out[i] = ToResponse(t)
	}
	return out
}
