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

func (s *Service) FindAll(userID uuid.UUID) ([]Todo, error) {
	var todos []Todo
	err := s.db.
		Joins("JOIN notes ON notes.id = todos.note_id").
		Where("notes.user_id = ?", userID).
		Order("todos.created_at DESC").
		Find(&todos).Error
	return todos, err
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
		Deadline: parseDeadline(in.Deadline),
		Priority: p,
		Tags:     json.RawMessage(`[]`),
	}
	if err := s.db.Create(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
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
	if in.Deadline != nil {
		fields["deadline"] = parseDeadline(in.Deadline)
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
	res := s.db.
		Where("id = ? AND note_id IN (SELECT id FROM notes WHERE user_id = ?)", id, userID).
		Delete(&Todo{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

func parseDeadline(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}
