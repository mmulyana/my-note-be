package todos

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) FindAll(userID uuid.UUID, labelID *uuid.UUID, page, limit int) ([]Todo, int64, error) {
	var total int64

	q := s.db.Model(&Todo{}).
		Joins("JOIN notes ON notes.id = todos.note_id").
		Where("notes.user_id = ?", userID)

	if labelID != nil {
		q = q.Joins("JOIN note_labels nl ON nl.note_id = todos.note_id").
			Where("nl.label_id = ?", *labelID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	var todos []Todo
	dataQ := s.db.
		Joins("JOIN notes ON notes.id = todos.note_id").
		Where("notes.user_id = ?", userID)

	if labelID != nil {
		dataQ = dataQ.Joins("JOIN note_labels nl ON nl.note_id = todos.note_id").
			Where("nl.label_id = ?", *labelID)
	}

	err := dataQ.
		Order("todos.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&todos).Error

	return todos, total, err
}

func (s *Service) FindOne(id string, userID uuid.UUID) (*Todo, error) {
	var t Todo
	err := s.db.
		Joins("JOIN notes ON notes.id = todos.note_id").
		Where("todos.id = ? AND notes.user_id = ?", id, userID).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Service) Create(userID uuid.UUID, in TodoInput) (*Todo, error) {
	var count int64
	if err := s.db.Table("notes").Where("id = ? AND user_id = ?", in.NoteID, userID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	p := in.Priority
	if p == "" {
		p = PriorityMedium
	}
	t := Todo{
		ID:       in.ID,
		NoteID:   in.NoteID,
		Text:     in.Text,
		Checked:  in.Checked,
		Deadline: parseDate(in.Deadline),
		Today:    parseDate(in.Today),
		Priority: p,
		Tags:     json.RawMessage(`[]`),
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&t).Error; err != nil {
			return err
		}

		var row struct {
			Content string
			Preview string
		}
		if err := tx.Table("notes").Select("content, preview").Where("id = ?", in.NoteID).Scan(&row).Error; err != nil {
			return err
		}

		liHTML := buildTodoLi(t.ID, in.Text, in.Checked, in.Deadline, p)

		return tx.Table("notes").Where("id = ?", in.NoteID).Updates(map[string]any{
			"content":    insertTodoLi(row.Content, in.LastTodoID, liHTML),
			"preview":    insertTodoLi(row.Preview, in.LastTodoID, liHTML),
			"todo_total": gorm.Expr("(SELECT COUNT(*) FROM todos WHERE note_id = ?)", in.NoteID),
			"todo_done":  gorm.Expr("(SELECT COUNT(*) FROM todos WHERE note_id = ? AND checked = true)", in.NoteID),
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// buildTodoLi renders the taskItem <li> for a todo, matching the editor's
// expected shape so the frontend can parse it back into a todo node.
func buildTodoLi(id, text string, checked bool, deadline *string, priority Priority) string {
	checkedAttr := "false"
	inputCheckedAttr := ""
	if checked {
		checkedAttr = "true"
		inputCheckedAttr = ` checked="checked"`
	}
	deadlineAttr := ""
	if deadline != nil && *deadline != "" {
		deadlineAttr = fmt.Sprintf(` data-deadline="%s"`, html.EscapeString(*deadline))
	}
	return fmt.Sprintf(
		`<li data-checked="%s" data-id="%s"%s data-priority="%s" data-type="taskItem"><label><input type="checkbox"%s><span></span></label><div><p>%s</p></div></li>`,
		checkedAttr,
		html.EscapeString(id),
		deadlineAttr,
		html.EscapeString(string(priority)),
		inputCheckedAttr,
		html.EscapeString(text),
	)
}

// insertTodoLi places liHTML into content without touching anything else.
// Priority:
//  1. If lastTodoID is given and found, insert directly after that <li>.
//  2. Else append into the last existing <ul data-type="taskList">.
//  3. Else create a new <ul data-type="taskList"> at the end of content.
func insertTodoLi(content string, lastTodoID *string, liHTML string) string {
	if lastTodoID != nil && *lastTodoID != "" {
		pattern := fmt.Sprintf(`<li[^>]*\sdata-id="%s"[^>]*>.*?</li>`, regexp.QuoteMeta(*lastTodoID))
		re := regexp.MustCompile(pattern)
		if loc := re.FindStringIndex(content); loc != nil {
			return content[:loc[1]] + liHTML + content[loc[1]:]
		}
	}

	taskListRe := regexp.MustCompile(`<ul data-type="taskList"[^>]*>[\s\S]*?</ul>`)
	if matches := taskListRe.FindAllStringIndex(content, -1); len(matches) > 0 {
		last := matches[len(matches)-1]
		closeIdx := last[1] - len("</ul>")
		return content[:closeIdx] + liHTML + content[closeIdx:]
	}

	return content + `<ul data-type="taskList">` + liHTML + `</ul>`
}

func (s *Service) Update(id string, userID uuid.UUID, in TodoUpdateInput) (*Todo, error) {
	t, err := s.FindOne(id, userID)
	if err != nil {
		return nil, err
	}

	fields := map[string]any{}
	if in.Text != nil {
		fields["text"] = *in.Text
	}
	if in.Checked != nil {
		fields["checked"] = *in.Checked
	}
	if in.Deadline.Present {
		fields["deadline"] = parseDate(in.Deadline.Value)
	}
	if in.Today.Present {
		fields["today"] = parseDate(in.Today.Value)
	}
	if in.Priority != nil {
		fields["priority"] = *in.Priority
	}
	if len(fields) == 0 {
		return t, nil
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(t).Updates(fields).Error; err != nil {
			return err
		}

		if in.Text != nil || in.Checked != nil {
			var row struct {
				Content string
				Preview string
			}
			if err := tx.Table("notes").Select("content, preview").Where("id = ?", t.NoteID).Scan(&row).Error; err != nil {
				return err
			}

			noteUpdates := map[string]any{
				"content": patchNoteContent(row.Content, id, in.Checked, in.Text),
				"preview": patchNoteContent(row.Preview, id, in.Checked, in.Text),
			}
			if in.Checked != nil {
				noteUpdates["todo_done"] = gorm.Expr(
					"(SELECT COUNT(*) FROM todos WHERE note_id = ? AND checked = true)", t.NoteID,
				)
			}
			return tx.Table("notes").Where("id = ?", t.NoteID).Updates(noteUpdates).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}

func patchNoteContent(content, todoID string, checked *bool, text *string) string {
	pattern := fmt.Sprintf(`<li[^>]*\sdata-id="%s"[^>]*>.*?</li>`, regexp.QuoteMeta(todoID))
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllStringFunc(content, func(liBlock string) string {
		if checked != nil {
			checkedVal := "false"
			if *checked {
				checkedVal = "true"
			}
			liBlock = regexp.MustCompile(`data-checked="[^"]*"`).
				ReplaceAllString(liBlock, fmt.Sprintf(`data-checked="%s"`, checkedVal))

			inputNew := `<input type="checkbox">`
			if *checked {
				inputNew = `<input type="checkbox" checked="checked">`
			}
			liBlock = regexp.MustCompile(`<input type="checkbox"[^>]*>`).
				ReplaceAllString(liBlock, inputNew)
		}

		if text != nil {
			textRe := regexp.MustCompile(`(<div><p>)(.*?)(</p></div>)`)
			escaped := html.EscapeString(*text)
			liBlock = textRe.ReplaceAllStringFunc(liBlock, func(match string) string {
				subs := textRe.FindStringSubmatch(match)
				if len(subs) < 4 {
					return match
				}
				return subs[1] + escaped + subs[3]
			})
		}

		return liBlock
	})
}

func (s *Service) Remove(id string, userID uuid.UUID) error {
	t, err := s.FindOne(id, userID)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("id = ?", id).Delete(&Todo{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		var row struct {
			Content string
			Preview string
		}
		if err := tx.Table("notes").Select("content, preview").Where("id = ?", t.NoteID).Scan(&row).Error; err != nil {
			return err
		}

		return tx.Table("notes").Where("id = ?", t.NoteID).Updates(map[string]any{
			"content":    removeTodoFromContent(row.Content, id),
			"preview":    removeTodoFromContent(row.Preview, id),
			"todo_total": gorm.Expr("(SELECT COUNT(*) FROM todos WHERE note_id = ?)", t.NoteID),
			"todo_done":  gorm.Expr("(SELECT COUNT(*) FROM todos WHERE note_id = ? AND checked = true)", t.NoteID),
		}).Error
	})
}

func removeTodoFromContent(content, todoID string) string {
	pattern := fmt.Sprintf(`<li[^>]*\sdata-id="%s"[^>]*>.*?</li>`, regexp.QuoteMeta(todoID))
	return regexp.MustCompile(pattern).ReplaceAllString(content, "")
}

func (s *Service) FindGroupByNotes(userID uuid.UUID) ([]NoteGroup, error) {
	type noteRow struct {
		ID    string
		Title string
	}
	var notes []noteRow
	err := s.db.Table("notes").
		Select("notes.id, notes.title").
		Joins("INNER JOIN todos ON todos.note_id = notes.id").
		Where("notes.user_id = ?", userID).
		Group("notes.id, notes.title").
		Order("notes.updated_at DESC").
		Scan(&notes).Error
	if err != nil {
		return nil, err
	}

	out := make([]NoteGroup, len(notes))
	for i, n := range notes {
		var todos []Todo
		if err := s.db.Where("note_id = ?", n.ID).Order("created_at ASC").Find(&todos).Error; err != nil {
			return nil, err
		}
		out[i] = NoteGroup{
			NoteID: n.ID,
			Title:  n.Title,
			Todos:  ToResponses(todos),
		}
	}
	return out, nil
}

func (s *Service) FindGroupByDeadline(userID uuid.UUID) ([]DeadlineGroup, error) {
	var todos []Todo
	err := s.db.
		Joins("JOIN notes ON notes.id = todos.note_id").
		Where("notes.user_id = ?", userID).
		Order("todos.deadline ASC NULLS LAST, todos.created_at ASC").
		Find(&todos).Error
	if err != nil {
		return nil, err
	}

	order := []string{}
	groups := map[string][]TodoResponse{}
	for _, t := range todos {
		key := "none"
		if t.Deadline != nil {
			key = t.Deadline.Format("2006-01-02")
		}
		if _, exists := groups[key]; !exists {
			order = append(order, key)
		}
		groups[key] = append(groups[key], ToResponse(t))
	}

	out := make([]DeadlineGroup, len(order))
	for i, key := range order {
		var dl *string
		if key != "none" {
			dl = &key
		}
		out[i] = DeadlineGroup{Deadline: dl, Todos: groups[key]}
	}
	return out, nil
}

func (s *Service) FindGroupByToday(userID uuid.UUID, date time.Time) (TodayGroups, error) {
	var todayTodos []Todo
	if err := s.db.
		Joins("JOIN notes ON notes.id = todos.note_id").
		Where("notes.user_id = ? AND todos.checked = false AND todos.today = ?", userID, date).
		Order("todos.created_at ASC").
		Find(&todayTodos).Error; err != nil {
		return TodayGroups{}, err
	}

	var overdueTodos []Todo
	if err := s.db.
		Joins("JOIN notes ON notes.id = todos.note_id").
		Where("notes.user_id = ? AND todos.checked = false AND todos.today IS NOT NULL AND todos.today < ?", userID, date).
		Order("todos.today ASC, todos.created_at ASC").
		Find(&overdueTodos).Error; err != nil {
		return TodayGroups{}, err
	}

	var completedTodos []Todo
	if err := s.db.
		Joins("JOIN notes ON notes.id = todos.note_id").
		Where("notes.user_id = ? AND todos.checked = true AND todos.today IS NOT NULL AND todos.today <= ?", userID, date).
		Order("todos.today DESC, todos.created_at ASC").
		Find(&completedTodos).Error; err != nil {
		return TodayGroups{}, err
	}

	return TodayGroups{
		Today:     ToResponses(todayTodos),
		Overdue:   ToResponses(overdueTodos),
		Completed: ToResponses(completedTodos),
	}, nil
}

func parseDate(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}
