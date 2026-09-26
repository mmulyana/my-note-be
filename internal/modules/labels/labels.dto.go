package labels

type LabelResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	NoteCount int64  `json:"noteCount"`
}
