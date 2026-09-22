package releases

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrInvalidTag = errors.New("tag must be one of new, improved, fixed")

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) FindAll(page, limit int, includeDrafts bool) ([]Release, int64, error) {
	base := func(db *gorm.DB) *gorm.DB {
		q := db.Model(&Release{})
		if !includeDrafts {
			q = q.Where("published_at IS NOT NULL")
		}
		return q
	}

	var total int64
	if err := base(s.db).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []Release
	err := base(s.db).
		Order("COALESCE(published_at, created_at) DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&list).Error

	return list, total, err
}

func (s *Service) FindOne(id uuid.UUID, includeDrafts bool) (*Release, error) {
	q := s.db.Where("id = ?", id)
	if !includeDrafts {
		q = q.Where("published_at IS NOT NULL")
	}
	var r Release
	if err := q.First(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Service) Create(in Input) (*Release, error) {
	tag, err := normalizeTag(in.Tag)
	if err != nil {
		return nil, err
	}
	r := Release{
		Version:     in.Version,
		Title:       in.Title,
		Tag:         tag,
		Summary:     in.Summary,
		Content:     in.Content,
		Image:       in.Image,
		PublishedAt: in.PublishedAt,
	}
	if err := s.db.Create(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

// Update replaces the entry wholesale, the way the folders module does: a
// payload without publishedAt turns a published entry back into a draft.
func (s *Service) Update(id uuid.UUID, in Input) (*Release, error) {
	tag, err := normalizeTag(in.Tag)
	if err != nil {
		return nil, err
	}

	var r Release
	if err := s.db.Where("id = ?", id).First(&r).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&r).Updates(map[string]any{
		"version":      in.Version,
		"title":        in.Title,
		"tag":          tag,
		"summary":      in.Summary,
		"content":      in.Content,
		"image":        in.Image,
		"published_at": in.PublishedAt,
	}).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Service) Remove(id uuid.UUID) error {
	res := s.db.Where("id = ?", id).Delete(&Release{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Unread counts published entries the user has not opened the panel for. A
// user who has never opened it has everything unread.
func (s *Service) Unread(userID uuid.UUID) (int64, *time.Time, error) {
	seenAt, err := s.seenAt(userID)
	if err != nil {
		return 0, nil, err
	}

	q := s.db.Model(&Release{}).Where("published_at IS NOT NULL")
	if seenAt != nil {
		// note: compare against whichever is later, because the two ways an
		// entry becomes visible are both "new" to a reader — publishing it now,
		// and publishing it with a backdated date so it sorts into place. A
		// plain published_at check would silently skip the backdated one.
		q = q.Where("GREATEST(published_at, created_at) > ?", *seenAt)
	}

	var count int64
	if err := q.Count(&count).Error; err != nil {
		return 0, nil, err
	}
	return count, seenAt, nil
}

func (s *Service) MarkSeen(userID uuid.UUID) (time.Time, error) {
	now := time.Now()
	res := s.db.Table("users").Where("id = ?", userID).Update("releases_seen_at", now)
	if res.Error != nil {
		return time.Time{}, res.Error
	}
	if res.RowsAffected == 0 {
		return time.Time{}, gorm.ErrRecordNotFound
	}
	return now, nil
}

func (s *Service) seenAt(userID uuid.UUID) (*time.Time, error) {
	var row struct {
		ReleasesSeenAt *time.Time
	}
	err := s.db.Table("users").
		Select("releases_seen_at").
		Where("id = ?", userID).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return row.ReleasesSeenAt, nil
}

func normalizeTag(t Tag) (Tag, error) {
	if t == "" {
		return TagNew, nil
	}
	if !t.valid() {
		return "", ErrInvalidTag
	}
	return t, nil
}
