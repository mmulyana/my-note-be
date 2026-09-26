package notes

import (
	"time"

	"my-note-be/internal/modules/folders"

	"github.com/google/uuid"
)

type LabelResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type FolderResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Color  string    `json:"color"`
	Secret bool      `json:"secret"`
}

func toFolderResponse(f *folders.Folder) *FolderResponse {
	if f == nil {
		return nil
	}
	return &FolderResponse{ID: f.ID, Name: f.Name, Color: f.Color, Secret: f.Secret}
}

type NoteListItemResponse struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Preview     string          `json:"preview"`
	TodoSummary TodoSummary     `json:"todoSummary"`
	Todos       []TodoResponse  `json:"todos,omitempty"`
	Labels      []LabelResponse `json:"labels"`
	Folder      *FolderResponse `json:"folder"`
	Pinned      bool            `json:"pinned"`
	Secret      bool            `json:"secret"`
	Archived    bool            `json:"archived"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type TodoSummary struct {
	Total int `json:"total"`
	Done  int `json:"done"`
}

func ToListItemResponse(n Note) NoteListItemResponse {
	lbls := make([]LabelResponse, len(n.Labels))
	for i, l := range n.Labels {
		lbls[i] = LabelResponse{ID: l.ID, Name: l.Name}
	}
	var todos []TodoResponse
	if len(n.Todos) > 0 {
		todos = make([]TodoResponse, len(n.Todos))
		for i, t := range n.Todos {
			todos[i] = toTodoResponse(t)
		}
	}
	return NoteListItemResponse{
		ID:      n.ID,
		Title:   n.Title,
		Preview: n.Preview,
		TodoSummary: TodoSummary{
			Total: n.TodoTotal,
			Done:  n.TodoDone,
		},
		Todos:     todos,
		Labels:    lbls,
		Folder:    toFolderResponse(n.Folder),
		Pinned:    n.Pinned,
		Secret:    n.Secret,
		Archived:  n.Archived,
		UpdatedAt: n.UpdatedAt,
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
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Content   string          `json:"content"`
	FolderID  *uuid.UUID      `json:"folderId"`
	Folder    *FolderResponse `json:"folder"`
	Todos     []TodoResponse  `json:"todos"`
	Labels    []LabelResponse `json:"labels"`
	Pinned    bool            `json:"pinned"`
	Secret    bool            `json:"secret"`
	Archived  bool            `json:"archived"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type TodoResponse struct {
	ID       string       `json:"id"`
	Checked  bool         `json:"checked"`
	Text     string       `json:"text"`
	Deadline *string      `json:"deadline"`
	Today    *string      `json:"today"`
	Priority TodoPriority `json:"priority"`
}

func ToDetailResponse(n Note) NoteDetailResponse {
	todos := make([]TodoResponse, len(n.Todos))
	for i, t := range n.Todos {
		todos[i] = toTodoResponse(t)
	}
	lbls := make([]LabelResponse, len(n.Labels))
	for i, l := range n.Labels {
		lbls[i] = LabelResponse{ID: l.ID, Name: l.Name}
	}
	return NoteDetailResponse{
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		FolderID:  n.FolderID,
		Folder:    toFolderResponse(n.Folder),
		Todos:     todos,
		Labels:    lbls,
		Pinned:    n.Pinned,
		Secret:    n.Secret,
		Archived:  n.Archived,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

func toTodoResponse(t Todo) TodoResponse {
	priority := t.Priority
	if priority == PriorityNone {
		priority = ""
	}
	res := TodoResponse{
		ID:       t.ID,
		Checked:  t.Checked,
		Text:     t.Text,
		Priority: priority,
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
