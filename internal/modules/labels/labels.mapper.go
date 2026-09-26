package labels

func ToResponse(l LabelWithCount) LabelResponse {
	res := LabelResponse{}
	res.ID = l.ID.String()
	res.Name = l.Name
	res.NoteCount = l.NoteCount
	return res
}

func ToResponses(labels []LabelWithCount) []LabelResponse {
	out := make([]LabelResponse, len(labels))
	for i, l := range labels {
		out[i] = ToResponse(l)
	}
	return out
}

func Names(labels []Label) []string {
	out := make([]string, len(labels))
	for i, l := range labels {
		out[i] = l.Name
	}
	return out
}
