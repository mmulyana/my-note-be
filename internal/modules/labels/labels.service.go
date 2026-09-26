package labels

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var namePattern = regexp.MustCompile(`^[\p{L}\p{N}_-]{1,32}$`)

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

func normalizeNames(names []string) []string {
	normalized := make([]string, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		normalizedName := strings.ToLower(strings.TrimSpace(name))
		if normalizedName == "" || seen[normalizedName] {
			continue
		}
		seen[normalizedName] = true
		normalized = append(normalized, normalizedName)
	}
	return normalized
}

// note: returns nil label (no error) when not found
func findByName(tx *gorm.DB, userID uuid.UUID, normalizedName string) (*Label, error) {
	var matches []Label
	err := tx.Where("user_id = ? AND lower(name) = ?", userID, normalizedName).
		Limit(1).
		Find(&matches).Error
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, nil
	}
	return &matches[0], nil
}

// note: returns nil label (no error) when the name is invalid
func findOrCreateByName(tx *gorm.DB, userID uuid.UUID, normalizedName string) (*Label, error) {
	existing, err := findByName(tx, userID, normalizedName)
	if err != nil || existing != nil {
		return existing, err
	}
	if !namePattern.MatchString(normalizedName) {
		return nil, nil
	}
	newLabel := &Label{UserID: userID, Name: normalizedName}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(newLabel).Error; err != nil {
		return nil, err
	}
	// note: re-fetch because DoNothing skips insert when a concurrent request created it
	return findByName(tx, userID, normalizedName)
}

func ResolveByNames(tx *gorm.DB, userID uuid.UUID, names []string) ([]Label, error) {
	labels := make([]Label, 0, len(names))
	for _, normalizedName := range normalizeNames(names) {
		label, err := findOrCreateByName(tx, userID, normalizedName)
		if err != nil {
			return nil, err
		}
		if label == nil {
			continue
		}
		labels = append(labels, *label)
	}
	return labels, nil
}

func IDsByNames(tx *gorm.DB, userID uuid.UUID, names []string) ([]string, error) {
	normalizedNames := normalizeNames(names)
	if len(normalizedNames) == 0 {
		return nil, nil
	}
	var ids []string
	err := tx.Model(&Label{}).
		Where("user_id = ? AND lower(name) IN ?", userID, normalizedNames).
		Pluck("id", &ids).Error
	return ids, err
}

func DeleteUnused(tx *gorm.DB, userID uuid.UUID, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return tx.Exec(`
		DELETE FROM labels
		WHERE user_id = ? AND id IN ?
		AND NOT EXISTS (SELECT 1 FROM note_labels nl WHERE nl.label_id = labels.id)
	`, userID, ids).Error
}
