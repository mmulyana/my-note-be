package categories

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) FindAll(userID uuid.UUID) ([]Category, error) {
	var cats []Category
	err := s.db.Where("user_id = ?", userID).Order("name").Find(&cats).Error
	return cats, err
}

func (s *Service) FindOne(id uuid.UUID, userID uuid.UUID) (*Category, error) {
	var c Category
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) Create(userID uuid.UUID, in CategoryInput) (*Category, error) {
	c := in.ToModel(userID)
	if err := s.db.Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) Update(id uuid.UUID, userID uuid.UUID, in CategoryInput) (*Category, error) {
	var c Category
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&c).Error; err != nil {
		return nil, err
	}
	c.Name = strings.TrimSpace(in.Name)
	if err := s.db.Save(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) Remove(id uuid.UUID, userID uuid.UUID) error {
	res := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Category{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func ResolveByNames(tx *gorm.DB, userID uuid.UUID, names []string) ([]Category, error) {
	cats := make([]Category, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true

		var c Category
		if err := tx.Where(Category{UserID: userID, Name: name}).
			FirstOrCreate(&c).Error; err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}
