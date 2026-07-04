package labels

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrDuplicateName is returned when a label name already exists for the user.
var ErrDuplicateName = errors.New("label already exists")

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) FindAll(userID uuid.UUID) ([]Label, error) {
	var labels []Label
	err := s.db.Where("user_id = ?", userID).Order("name").Find(&labels).Error
	return labels, err
}

func (s *Service) FindOne(id uuid.UUID, userID uuid.UUID) (*Label, error) {
	var l Label
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *Service) Create(userID uuid.UUID, in LabelInput) (*Label, error) {
	l := in.ToModel(userID)
	if err := s.db.Create(&l).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateName
		}
		return nil, err
	}
	return &l, nil
}

func (s *Service) Update(id uuid.UUID, userID uuid.UUID, in LabelInput) (*Label, error) {
	var l Label
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&l).Error; err != nil {
		return nil, err
	}
	l.Name = strings.TrimSpace(in.Name)
	if err := s.db.Save(&l).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateName
		}
		return nil, err
	}
	return &l, nil
}

func (s *Service) Remove(id uuid.UUID, userID uuid.UUID) error {
	res := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Label{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func ResolveByNames(tx *gorm.DB, userID uuid.UUID, names []string) ([]Label, error) {
	labels := make([]Label, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true

		var l Label
		if err := tx.Where(Label{UserID: userID, Name: name}).
			FirstOrCreate(&l).Error; err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, nil
}
