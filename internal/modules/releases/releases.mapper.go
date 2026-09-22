package releases

func ToResponse(r Release) Response {
	return Response{
		ID:          r.ID.String(),
		Version:     r.Version,
		Title:       r.Title,
		Tag:         r.Tag,
		Summary:     r.Summary,
		Content:     r.Content,
		Image:       r.Image,
		PublishedAt: r.PublishedAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func ToResponses(releases []Release) []Response {
	out := make([]Response, len(releases))
	for i, r := range releases {
		out[i] = ToResponse(r)
	}
	return out
}
