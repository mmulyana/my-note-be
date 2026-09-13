package links

func toResponse(r row) Response {
	return Response{
		ID:          r.ID,
		NoteID:      r.NoteID,
		NoteTitle:   r.NoteTitle,
		URL:         r.URL,
		Title:       r.Title,
		Description: r.Description,
		Image:       r.Image,
		Favicon:     r.Favicon,
		SiteName:    r.SiteName,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func toResponses(rows []row) []Response {
	out := make([]Response, 0, len(rows))
	for _, r := range rows {
		out = append(out, toResponse(r))
	}
	return out
}
