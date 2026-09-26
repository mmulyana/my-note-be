package labels

func ToResponse(l Label) LabelResponse {
	res := LabelResponse{}
	res.ID = l.ID.String()
	res.Name = l.Name
	return res
}

func ToResponses(labels []Label) []LabelResponse {
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
