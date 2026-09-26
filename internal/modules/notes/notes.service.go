package notes

import (
	"database/sql"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"my-note-be/internal/modules/labels"

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

func orderTodosByContent(db *gorm.DB) *gorm.DB {
	return db.
		Order(`NULLIF(strpos((SELECT content FROM notes WHERE notes.id = todos.note_id), 'data-id="' || todos.id || '"'), 0) ASC NULLS LAST`).
		Order("todos.created_at ASC").
		Order("todos.id ASC")
}

func (s *Service) FindAll(userID uuid.UUID, labelID *uuid.UUID, folderID *uuid.UUID, hasFolder *bool, archived *bool, pinned *bool, hasTodo *bool, tf TodoFilter, search string, page, limit int) ([]Note, int64, error) {
	var total int64

	search = strings.TrimSpace(search)

	filters := func(db *gorm.DB) *gorm.DB {
		if labelID != nil {
			db = db.Joins("JOIN note_labels nl ON nl.note_id = notes.id").
				Where("nl.label_id = ?", *labelID)
		}
		if folderID != nil {
			db = db.Where("notes.folder_id = ?", *folderID)
		} else {
			db = db.Where("(notes.folder_id IS NULL OR notes.folder_id NOT IN (SELECT id FROM folders WHERE isolated = true AND deleted_at IS NULL))")
		}
		if hasFolder != nil {
			if *hasFolder {
				db = db.Where("notes.folder_id IS NOT NULL")
			} else {
				db = db.Where("notes.folder_id IS NULL")
			}
		}
		if tf.hasFolderFilter() {
			switch {
			case len(tf.FolderIDs) > 0 && tf.NoFolder:
				db = db.Where("(notes.folder_id IN ? OR notes.folder_id IS NULL)", tf.FolderIDs)
			case len(tf.FolderIDs) > 0:
				db = db.Where("notes.folder_id IN ?", tf.FolderIDs)
			default:
				db = db.Where("notes.folder_id IS NULL")
			}
		}
		if archived != nil {
			db = db.Where("notes.archived = ?", *archived)
		} else if labelID == nil {
			db = db.Where("notes.archived = false")
		}
		if pinned != nil {
			db = db.Where("notes.pinned = ?", *pinned)
		}
		if hasTodo != nil {
			if *hasTodo {
				db = db.Where("notes.todo_total > 0")
			} else {
				db = db.Where("notes.todo_total = 0")
			}
		}
		if hasTodo != nil && *hasTodo {
			if exists := tf.existsSQL(); exists != "" {
				db = db.Where(exists)
			}
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

	dataQ := s.db.
		Select("notes.id, notes.title, notes.preview, notes.folder_id, notes.pinned, notes.secret, notes.archived, notes.todo_total, notes.todo_done, notes.updated_at").
		Preload("Labels").
		Preload("Folder").
		Where("notes.user_id = ?", userID).
		Scopes(filters)

	if hasTodo != nil && *hasTodo {
		dataQ = dataQ.Preload("Todos", tf.preloadScope)
	}

	var notes []Note
	if hasTodo != nil && *hasTodo {
		if order := tf.noteOrderSQL(); order != "" {
			dataQ = dataQ.Order(order)
		}
	}
	err := dataQ.
		Order("notes.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notes).Error

	return notes, total, err
}

func (s *Service) Counts(userID uuid.UUID) (CountsResponse, error) {
	var out CountsResponse
	err := s.db.Raw(`
		SELECT
			(SELECT COUNT(*) FROM notes WHERE user_id = @uid AND archived = false
				AND (folder_id IS NULL OR folder_id NOT IN (SELECT id FROM folders WHERE isolated = true AND deleted_at IS NULL))) AS notes,
			(SELECT COUNT(*) FROM notes WHERE user_id = @uid AND archived = true) AS archive,
			(SELECT COUNT(*) FROM todos t JOIN notes n ON n.id = t.note_id
				WHERE n.user_id = @uid AND n.archived = false AND t.checked = false) AS todos,
			(SELECT COUNT(*) FROM labels WHERE user_id = @uid) AS labels
	`, sql.Named("uid", userID)).Scan(&out).Error
	if err != nil {
		return out, err
	}

	var rows []struct {
		FolderID uuid.UUID `gorm:"column:folder_id"`
		Total    int64     `gorm:"column:total"`
	}
	if err := s.db.Raw(`
		SELECT folder_id, COUNT(*) AS total
		FROM notes
		WHERE user_id = ? AND archived = false AND folder_id IS NOT NULL
		GROUP BY folder_id
	`, userID).Scan(&rows).Error; err != nil {
		return out, err
	}

	out.Folders = make(map[string]int64, len(rows))
	for _, r := range rows {
		out.Folders[r.FolderID.String()] = r.Total
	}
	return out, nil
}

func (s *Service) FindOne(id string, userID uuid.UUID) (*Note, error) {
	var note Note
	if err := s.db.
		Preload("Todos", orderTodosByContent).
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

		if err := syncLabels(tx, note.ID, userID, in.Labels); err != nil {
			return err
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
				p = PriorityNone
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

		if err := syncLabels(tx, id, userID, in.Labels); err != nil {
			return err
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
				p = PriorityNone
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
	return s.db.Transaction(func(tx *gorm.DB) error {
		var labelIDs []string
		if err := tx.Table("note_labels").Where("note_id = ?", id).Pluck("label_id", &labelIDs).Error; err != nil {
			return err
		}

		res := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&Note{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return labels.DeleteUnused(tx, userID, labelIDs)
	})
}

// note: nil names = request did not carry labels, leave the note's labels untouched; empty = clear them all
func syncLabels(tx *gorm.DB, noteID string, userID uuid.UUID, names []string) error {
	if names == nil {
		return nil
	}

	wanted, err := labels.ResolveByNames(tx, userID, names)
	if err != nil {
		return err
	}
	keep := make(map[string]bool, len(wanted))
	for _, label := range wanted {
		keep[label.ID.String()] = true
	}

	var current []string
	if err := tx.Table("note_labels").Where("note_id = ?", noteID).Pluck("label_id", &current).Error; err != nil {
		return err
	}
	var stale []string
	for _, id := range current {
		if !keep[id] {
			stale = append(stale, id)
		}
	}

	if len(stale) > 0 {
		if err := tx.Exec("DELETE FROM note_labels WHERE note_id = ? AND label_id IN ?", noteID, stale).Error; err != nil {
			return err
		}
		if err := labels.DeleteUnused(tx, userID, stale); err != nil {
			return err
		}
	}

	for _, label := range wanted {
		if err := tx.Exec(
			"INSERT INTO note_labels (note_id, label_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			noteID, label.ID,
		).Error; err != nil {
			return err
		}
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
			if v == "" {
				v = string(PriorityNone)
			}
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
