package todos

import "time"

type TodoInput struct {
	ID         string   `json:"id" binding:"required"`
	NoteID     string   `json:"noteId" binding:"required"`
	Text       string   `json:"text"`
	Checked    bool     `json:"checked"`
	Deadline   *string  `json:"deadline"`
	Today      *string  `json:"today"`
	Priority   Priority `json:"priority"`
	LastTodoID *string  `json:"lastTodoId"`
}

type TodoUpdateInput struct {
	Text     *string   `json:"text"`
	Checked  *bool     `json:"checked"`
	Deadline *string   `json:"deadline"`
	Today    *string   `json:"today"`
	Priority *Priority `json:"priority"`
}

type TodoResponse struct {
	ID        string    `json:"id"`
	NoteID    string    `json:"noteId"`
	Text      string    `json:"text"`
	Checked   bool      `json:"checked"`
	Deadline  *string   `json:"deadline"`
	Today     *string   `json:"today"`
	Priority  Priority  `json:"priority"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type NoteGroup struct {
	NoteID string         `json:"noteId"`
	Title  string         `json:"title"`
	Todos  []TodoResponse `json:"todos"`
}

type DeadlineGroup struct {
	Deadline *string        `json:"deadline"`
	Todos    []TodoResponse `json:"todos"`
}
