package notes

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TodoStatusOverdue = "overdue"
	TodoStatusOpen    = "open"
	TodoStatusDone    = "done"

	TodoSortPriority = "priority"
	TodoSortOverdue  = "overdue"
	TodoSortUpdated  = "updated"
)

const todoPriorityRank = `(CASE todos.priority WHEN 'high' THEN 3 WHEN 'medium' THEN 2 WHEN 'low' THEN 1 ELSE 0 END)`

type TodoFilter struct {
	FolderIDs  []uuid.UUID
	NoFolder   bool
	Priorities []TodoPriority
	Status     string
	Sort       string
	Desc       bool
	Today      time.Time
}

func (f TodoFilter) hasFolderFilter() bool {
	return len(f.FolderIDs) > 0 || f.NoFolder
}

func (f TodoFilter) overdueSQL() string {
	return "(todos.checked = false AND todos.deadline < '" + f.Today.Format("2006-01-02") + "'::date)"
}

func (f TodoFilter) todoCondSQL() string {
	var parts []string

	if len(f.Priorities) > 0 {
		quoted := make([]string, len(f.Priorities))
		for i, p := range f.Priorities {
			quoted[i] = "'" + string(p) + "'"
		}
		parts = append(parts, "todos.priority IN ("+strings.Join(quoted, ",")+")")
	}

	switch f.Status {
	case TodoStatusOverdue:
		parts = append(parts, f.overdueSQL())
	case TodoStatusOpen:
		parts = append(parts, "todos.checked = false")
	case TodoStatusDone:
		parts = append(parts, "todos.checked = true")
	}

	return strings.Join(parts, " AND ")
}

func (f TodoFilter) dir() string {
	if f.Desc {
		return "DESC"
	}
	return "ASC"
}

func (f TodoFilter) todoOrderSQL() string {
	switch f.Sort {
	case TodoSortPriority:
		return todoPriorityRank + " " + f.dir()
	case TodoSortOverdue:
		return "(CASE WHEN " + f.overdueSQL() + " THEN 1 ELSE 0 END) " + f.dir() + ", todos.deadline ASC NULLS LAST"
	case TodoSortUpdated:
		return "todos.updated_at " + f.dir()
	}
	return ""
}

func (f TodoFilter) noteScope(cond string) string {
	scope := "SELECT %s FROM todos WHERE todos.note_id = notes.id"
	if cond != "" {
		scope += " AND " + cond
	}
	return scope
}

func (f TodoFilter) noteOrderSQL() string {
	cond := f.todoCondSQL()
	switch f.Sort {
	case TodoSortPriority:
		q := strings.Replace(f.noteScope(cond), "%s", "COALESCE(MAX"+todoPriorityRank+", -1)", 1)
		return "(" + q + ") " + f.dir() + ", notes.updated_at DESC"
	case TodoSortOverdue:
		overdue := f.overdueSQL()
		hasOverdue := strings.Replace(f.noteScope(cond+andIf(cond)+overdue), "%s", "1", 1)
		firstOverdue := strings.Replace(f.noteScope(cond+andIf(cond)+overdue), "%s", "MIN(todos.deadline)", 1)
		return "EXISTS (" + hasOverdue + ") " + f.dir() + ", (" + firstOverdue + ") ASC NULLS LAST, notes.updated_at DESC"
	case TodoSortUpdated:
		return "notes.updated_at " + f.dir()
	}
	return ""
}

func andIf(cond string) string {
	if cond == "" {
		return ""
	}
	return " AND "
}

func (f TodoFilter) preloadScope(db *gorm.DB) *gorm.DB {
	if cond := f.todoCondSQL(); cond != "" {
		db = db.Where(cond)
	}
	if order := f.todoOrderSQL(); order != "" {
		db = db.Order(order)
	}
	return orderTodosByContent(db)
}

func (f TodoFilter) existsSQL() string {
	cond := f.todoCondSQL()
	if cond == "" {
		return ""
	}
	return "EXISTS (" + strings.Replace(f.noteScope(cond), "%s", "1", 1) + ")"
}
