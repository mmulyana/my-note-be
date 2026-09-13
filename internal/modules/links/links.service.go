package links

import (
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Preview di-fetch di server, bukan di browser, karena browser nggak boleh baca HTML lintas origin
func (s *Service) Preview(rawURL string) (*Preview, error) {
	return fetchPreview(rawURL)
}

// FindAll: semua link user, terbaru dulu. Row-nya ditulis note save yang bawa linkDiff, modul ini nggak pernah insert.
func (s *Service) FindAll(userID uuid.UUID, search string, page, limit int) ([]row, int64, error) {
	base := func(db *gorm.DB) *gorm.DB {
		db = db.Table("links").
			Joins("JOIN notes ON notes.id = links.note_id").
			Where("notes.user_id = ?", userID)
		if search != "" {
			pattern := "%" + search + "%"
			db = db.Where("links.title ILIKE ? OR links.url ILIKE ?", pattern, pattern)
		}
		return db
	}

	var total int64
	if err := base(s.db).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []row
	err := base(s.db).
		Select("links.*, notes.title AS note_title").
		Order("links.created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&rows).Error

	return rows, total, err
}
