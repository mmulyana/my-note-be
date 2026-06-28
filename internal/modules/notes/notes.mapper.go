package notes

import (
	"time"

	"github.com/google/uuid"
)

type CategoryResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type NoteListItemResponse struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Preview     string             `json:"preview"`
	TodoSummary TodoSummary        `json:"todoSummary"`
	Categories  []CategoryResponse `json:"categories"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

type TodoSummary struct {
	Total int `json:"total"`
	Done  int `json:"done"`
}

func ToListItemResponse(n Note) NoteListItemResponse {
	cats := make([]CategoryResponse, len(n.Categories))
	for i, c := range n.Categories {
		cats[i] = CategoryResponse{ID: c.ID, Name: c.Name}
	}
	return NoteListItemResponse{
		ID:      n.ID,
		Title:   n.Title,
		Preview: n.Preview,
		TodoSummary: TodoSummary{
			Total: n.TodoTotal,
			Done:  n.TodoDone,
		},
		Categories: cats,
		UpdatedAt:  n.UpdatedAt,
	}
}

func ToListItemResponses(notes []Note) []NoteListItemResponse {
	out := make([]NoteListItemResponse, len(notes))
	for i, n := range notes {
		out[i] = ToListItemResponse(n)
	}
	return out
}

type NoteDetailResponse struct {
	ID         string             `json:"id"`
	Title      string             `json:"title"`
	Content    string             `json:"content"`
	FolderID   *uuid.UUID         `json:"folderId"`
	Todos      []TodoResponse     `json:"todos"`
	Categories []CategoryResponse `json:"categories"`
	CreatedAt  time.Time          `json:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt"`
}

type TodoResponse struct {
	ID       string       `json:"id"`
	Checked  bool         `json:"checked"`
	Text     string       `json:"text"`
	Deadline *string      `json:"deadline"`
	Priority TodoPriority `json:"priority"`
}

func ToDetailResponse(n Note) NoteDetailResponse {
	todos := make([]TodoResponse, len(n.Todos))
	for i, t := range n.Todos {
		todos[i] = toTodoResponse(t)
	}
	cats := make([]CategoryResponse, len(n.Categories))
	for i, c := range n.Categories {
		cats[i] = CategoryResponse{ID: c.ID, Name: c.Name}
	}
	return NoteDetailResponse{
		ID:         n.ID,
		Title:      n.Title,
		Content:    n.Content,
		FolderID:   n.FolderID,
		Todos:      todos,
		Categories: cats,
		CreatedAt:  n.CreatedAt,
		UpdatedAt:  n.UpdatedAt,
	}
}

func toTodoResponse(t Todo) TodoResponse {
	res := TodoResponse{
		ID:       t.ID,
		Checked:  t.Checked,
		Text:     t.Text,
		Priority: t.Priority,
	}
	if t.Deadline != nil {
		d := t.Deadline.Format("2006-01-02")
		res.Deadline = &d
	}
	return res
}
