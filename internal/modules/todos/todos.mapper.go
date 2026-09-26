package todos

func ToResponse(t Todo) TodoResponse {
	priority := t.Priority
	if priority == PriorityNone {
		priority = ""
	}
	res := TodoResponse{
		ID:        t.ID,
		NoteID:    t.NoteID,
		Text:      t.Text,
		Checked:   t.Checked,
		Priority:  priority,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	if t.Deadline != nil {
		d := t.Deadline.Format("2006-01-02")
		res.Deadline = &d
	}
	if t.Today != nil {
		d := t.Today.Format("2006-01-02")
		res.Today = &d
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
