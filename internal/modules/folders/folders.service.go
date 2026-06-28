package folders

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
		"name":  in.Name,
		"color": color,
	}).Error; err != nil {
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
