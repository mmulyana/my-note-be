package folders

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrDuplicateName is returned when a folder name already exists for the user.
var ErrDuplicateName = errors.New("folder already exists")

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) FindAll(userID uuid.UUID) ([]Folder, error) {
	var folders []Folder
	err := s.db.Where("user_id = ?", userID).Order("name").Find(&folders).Error
	return folders, err
}

func (s *Service) FindOne(id uuid.UUID, userID uuid.UUID) (*Folder, error) {
	var f Folder
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *Service) Create(userID uuid.UUID, in FolderInput) (*Folder, error) {
	f := in.ToModel(userID)
	if err := s.db.Create(&f).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateName
		}
		return nil, err
	}
	return &f, nil
}

func (s *Service) Update(id uuid.UUID, userID uuid.UUID, in FolderInput) (*Folder, error) {
	var f Folder
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&f).Error; err != nil {
		return nil, err
	}
	color := in.Color
	if color == "" {
		color = "default"
	}
	if err := s.db.Model(&f).Updates(map[string]any{
		"name":   in.Name,
		"color":  color,
		"secret": in.Secret,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateName
		}
		return nil, err
	}
	return &f, nil
}

func (s *Service) Remove(id uuid.UUID, userID uuid.UUID) error {
	res := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Folder{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

const maxNotesPerFolder = 3

type FolderNote struct {
	FolderID uuid.UUID `gorm:"column:folder_id"`
	Title    string    `gorm:"column:title"`
	Text     string    `gorm:"column:text"`
}

func (s *Service) FindAllWithNotes(userID uuid.UUID, page, limit int) ([]Folder, int64, map[uuid.UUID][]FolderNote, error) {
	var total int64
	if err := s.db.Model(&Folder{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, nil, err
	}

	var folders []Folder
	offset := (page - 1) * limit
	if err := s.db.
		Where("user_id = ?", userID).
		Order("name").
		Offset(offset).
		Limit(limit).
		Find(&folders).Error; err != nil {
		return nil, 0, nil, err
	}

	grouped := make(map[uuid.UUID][]FolderNote, len(folders))
	if len(folders) == 0 {
		return folders, total, grouped, nil
	}

	ids := make([]uuid.UUID, len(folders))
	for i, f := range folders {
		ids[i] = f.ID
	}

	var rows []FolderNote
	if err := s.db.Raw(`
		SELECT f.id AS folder_id, n.title, n.text
		FROM folders f
		CROSS JOIN LATERAL (
			SELECT title, text
			FROM notes
			WHERE notes.folder_id = f.id AND notes.user_id = ? AND notes.archived = false
			LIMIT ?
		) n
		WHERE f.id IN ?
	`, userID, maxNotesPerFolder, ids).Scan(&rows).Error; err != nil {
		return nil, 0, nil, err
	}

	for _, r := range rows {
		grouped[r.FolderID] = append(grouped[r.FolderID], r)
	}
	return folders, total, grouped, nil
}
