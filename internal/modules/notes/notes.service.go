package notes

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var headingRe = regexp.MustCompile(`(?i)<h[1-3][^>]*>(.*?)</h[1-3]>`)
var tagRe = regexp.MustCompile(`<[^>]+>`)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) FindAll(userID uuid.UUID, labelID *uuid.UUID, folderID *uuid.UUID, hasFolder *bool, archived *bool, pinned *bool, search string, page, limit int) ([]Note, int64, error) {
	var total int64

	search = strings.TrimSpace(search)

	filters := func(db *gorm.DB) *gorm.DB {
		if labelID != nil {
			db = db.Joins("JOIN note_labels nl ON nl.note_id = notes.id").
				Where("nl.label_id = ?", *labelID)
		}
		if folderID != nil {
			db = db.Where("notes.folder_id = ?", *folderID)
		}
		if hasFolder != nil {
			if *hasFolder {
				db = db.Where("notes.folder_id IS NOT NULL")
			} else {
				db = db.Where("notes.folder_id IS NULL")
			}
		}
		if archived != nil {
			db = db.Where("notes.archived = ?", *archived)
		} else {
			db = db.Where("notes.archived = false")
		}
		if pinned != nil {
			db = db.Where("notes.pinned = ?", *pinned)
		}
		if search != "" {
			pattern := "%" + search + "%"
			db = db.Where("notes.title ILIKE ? OR notes.text ILIKE ?", pattern, pattern)
		}
		return db
	}

	q := s.db.Model(&Note{}).
		Where("notes.user_id = ?", userID).
		Scopes(filters)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	var notes []Note
	err := s.db.
		Select("notes.id, notes.title, notes.preview, notes.folder_id, notes.pinned, notes.secret, notes.archived, notes.todo_total, notes.todo_done, notes.updated_at").
		Preload("Labels").
		Preload("Folder").
		Where("notes.user_id = ?", userID).
		Scopes(filters).
		Order("notes.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notes).Error

	return notes, total, err
}

func (s *Service) FindOne(id string, userID uuid.UUID) (*Note, error) {
	var note Note
	if err := s.db.
		Preload("Todos").
		Preload("Labels").
		Preload("Folder").
		First(&note, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, err
	}
	return &note, nil
}

func (s *Service) Create(userID uuid.UUID, in CreateNoteInput) (*Note, error) {
	note := Note{
		ID:       in.ID,
		UserID:   userID,
		Content:  in.Content,
		Title:    extractTitle(in.Content),
		FolderID: in.FolderID,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&note).Error; err != nil {
			return err
		}

		for _, labelID := range in.LabelIDs {
			if err := tx.Exec("INSERT INTO note_labels (note_id, label_id) VALUES (?, ?)", note.ID, labelID).Error; err != nil {
				return err
			}
		}

		diff := in.TodoDiff

		if len(diff.Removed) > 0 {
			if err := tx.Where("id IN ? AND note_id = ?", diff.Removed, note.ID).Delete(&Todo{}).Error; err != nil {
				return err
			}
		}

		for _, a := range diff.Added {
			p := a.Priority
			if p == "" {
				p = PriorityMedium
			}
			todo := Todo{
				ID:       a.ID,
				NoteID:   note.ID,
				Text:     a.Text,
				Checked:  a.Checked,
				Deadline: parseDate(a.Deadline),
				Today:    parseDate(a.Today),
				Priority: p,
				Tags:     json.RawMessage(`[]`),
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"text", "checked", "deadline", "today", "priority", "updated_at"}),
			}).Create(&todo).Error; err != nil {
				return err
			}
		}

		for _, u := range diff.Updated {
			fields := buildUpdateMap(u.Fields)
			if len(fields) == 0 {
				continue
			}
			fields["updated_at"] = gorm.Expr("now()")
			if err := tx.Model(&Todo{}).Where("id = ? AND note_id = ?", u.ID, note.ID).Updates(fields).Error; err != nil {
				return err
			}
		}

		if err := applyLinkDiff(tx, note.ID, in.LinkDiff); err != nil {
			return err
		}

		return tx.Exec(`
			UPDATE notes SET
				todo_total = (SELECT COUNT(*) FROM todos WHERE note_id = ?),
				todo_done  = (SELECT COUNT(*) FROM todos WHERE note_id = ? AND checked = true)
			WHERE id = ?
		`, note.ID, note.ID, note.ID).Error
	})
	if err != nil {
		return nil, err
	}
	return s.FindOne(note.ID, userID)
}

func (s *Service) Save(id string, userID uuid.UUID, in SaveNoteInput) (*Note, error) {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var existing Note
		if err := tx.First(&existing, "id = ? AND user_id = ?", id, userID).Error; err != nil {
			return err
		}

		noteUpdates := map[string]any{
			"content":   in.Content,
			"preview":   in.Preview,
			"title":     extractTitle(in.Content),
			"text":      tagRe.ReplaceAllString(in.Preview, ""),
			"folder_id": in.FolderID,
		}
		if in.Pinned != nil {
			noteUpdates["pinned"] = *in.Pinned
		}
		if in.Archived != nil {
			noteUpdates["archived"] = *in.Archived
		}
		if in.Secret != nil {
			noteUpdates["secret"] = *in.Secret
		}
		if err := tx.Model(&Note{}).Where("id = ?", id).Updates(noteUpdates).Error; err != nil {
			return err
		}

		if err := tx.Exec("DELETE FROM note_labels WHERE note_id = ?", id).Error; err != nil {
			return err
		}
		for _, labelID := range in.LabelIDs {
			if err := tx.Exec("INSERT INTO note_labels (note_id, label_id) VALUES (?, ?)", id, labelID).Error; err != nil {
				return err
			}
		}

		diff := in.TodoDiff

		if len(diff.Removed) > 0 {
			if err := tx.Where("id IN ? AND note_id = ?", diff.Removed, id).Delete(&Todo{}).Error; err != nil {
				return err
			}
		}

		for _, a := range diff.Added {
			p := a.Priority
			if p == "" {
				p = PriorityMedium
			}
			todo := Todo{
				ID:       a.ID,
				NoteID:   id,
				Text:     a.Text,
				Checked:  a.Checked,
				Deadline: parseDate(a.Deadline),
				Today:    parseDate(a.Today),
				Priority: p,
				Tags:     json.RawMessage(`[]`),
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"text", "checked", "deadline", "today", "priority", "updated_at"}),
			}).Create(&todo).Error; err != nil {
				return err
			}
		}

		for _, u := range diff.Updated {
			fields := buildUpdateMap(u.Fields)
			if len(fields) == 0 {
				continue
			}
			fields["updated_at"] = gorm.Expr("now()")
			if err := tx.Model(&Todo{}).Where("id = ? AND note_id = ?", u.ID, id).Updates(fields).Error; err != nil {
				return err
			}
		}

		if err := applyLinkDiff(tx, id, in.LinkDiff); err != nil {
			return err
		}

		return tx.Exec(`
			UPDATE notes SET
				todo_total = (SELECT COUNT(*) FROM todos WHERE note_id = ?),
				todo_done  = (SELECT COUNT(*) FROM todos WHERE note_id = ? AND checked = true)
			WHERE id = ?
		`, id, id, id).Error
	})
	if err != nil {
		return nil, err
	}
	return s.FindOne(id, userID)
}

func (s *Service) Remove(id string, userID uuid.UUID) error {
	res := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Note{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// applyLinkDiff nulis perubahan link di dalam transaksi note-nya, biar note dan link-nya nggak pernah beda isi
func applyLinkDiff(tx *gorm.DB, noteID string, diff LinkDiff) error {
	if len(diff.Removed) > 0 {
		if err := tx.Where("id IN ? AND note_id = ?", diff.Removed, noteID).
			Delete(&Link{}).Error; err != nil {
			return err
		}
	}

	for _, a := range diff.Added {
		link := Link{
			ID:          a.ID,
			NoteID:      noteID,
			URL:         a.URL,
			Title:       a.Title,
			Description: a.Description,
			Image:       a.Image,
			Favicon:     a.Favicon,
			SiteName:    a.SiteName,
		}
		// row "added" bisa udah ada: card kesimpen pas masih loading, lalu kesimpen lagi pas metadata-nya dateng
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"url", "title", "description", "image", "favicon", "site_name", "updated_at",
			}),
		}).Create(&link).Error; err != nil {
			return err
		}
	}

	for _, u := range diff.Updated {
		fields := buildLinkUpdateMap(u.Fields)
		if len(fields) == 0 {
			continue
		}
		fields["updated_at"] = gorm.Expr("now()")
		if err := tx.Model(&Link{}).
			Where("id = ? AND note_id = ?", u.ID, noteID).
			Updates(fields).Error; err != nil {
			return err
		}
	}

	return nil
}

var linkColumns = map[string]string{
	"url":         "url",
	"title":       "title",
	"description": "description",
	"image":       "image",
	"favicon":     "favicon",
	"siteName":    "site_name",
}

// cuma field yang dikenal yang lolos, biar key asing nggak bisa nyasar jadi nama kolom di UPDATE
func buildLinkUpdateMap(fields map[string]json.RawMessage) map[string]any {
	out := map[string]any{}
	for key, column := range linkColumns {
		raw, ok := fields[key]
		if !ok {
			continue
		}
		var v string
		if json.Unmarshal(raw, &v) == nil {
			out[column] = v
		}
	}
	return out
}

func buildUpdateMap(fields map[string]json.RawMessage) map[string]any {
	out := map[string]any{}

	if raw, ok := fields["checked"]; ok {
		var v bool
		if json.Unmarshal(raw, &v) == nil {
			out["checked"] = v
		}
	}
	if raw, ok := fields["text"]; ok {
		var v string
		if json.Unmarshal(raw, &v) == nil {
			out["text"] = v
		}
	}
	if raw, ok := fields["deadline"]; ok {
		if string(raw) == "null" {
			out["deadline"] = nil
		} else {
			var ds string
			if json.Unmarshal(raw, &ds) == nil {
				if t, err := time.Parse("2006-01-02", ds); err == nil {
					out["deadline"] = t
				}
			}
		}
	}
	if raw, ok := fields["today"]; ok {
		if string(raw) == "null" {
			out["today"] = nil
		} else {
			var ds string
			if json.Unmarshal(raw, &ds) == nil {
				if t, err := time.Parse("2006-01-02", ds); err == nil {
					out["today"] = t
				}
			}
		}
	}
	if raw, ok := fields["priority"]; ok {
		var v string
		if json.Unmarshal(raw, &v) == nil {
			out["priority"] = v
		}
	}

	return out
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

func extractTitle(content string) string {
	m := headingRe.FindStringSubmatch(content)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(tagRe.ReplaceAllString(m[1], ""))
}
