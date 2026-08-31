package users

import (
	"errors"
	"time"

	"my-note-be/internal/middleware"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Register(email, password string) (*User, error) {
	var existing User
	if err := s.db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := NewUser(email, string(hashedPassword))
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(email, password string) (*User, error) {
	var user User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

func (s *Service) FindByID(id string) (*User, error) {
	var user User
	if err := s.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) UpdateProfile(id string, in UpdateProfileInput) (*User, error) {
	updates := map[string]any{}
	if in.Username != nil {
		updates["username"] = *in.Username
	}
	if in.Photo != nil {
		updates["photo"] = *in.Photo
	}
	if len(updates) == 0 {
		return s.FindByID(id)
	}

	if err := s.db.Model(&User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.FindByID(id)
}

func (s *Service) CreateRefreshToken(userID uuid.UUID, ttl time.Duration) (string, error) {
	rawToken, err := middleware.GenerateRandomToken()
	if err != nil {
		return "", err
	}

	tokenHash := middleware.HashToken(rawToken)
	rt := RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&rt).Error; err != nil {
		return "", err
	}

	return rawToken, nil
}

func (s *Service) RotateRefreshToken(rawToken string, ttl time.Duration) (*User, string, error) {
	tokenHash := middleware.HashToken(rawToken)

	var rt RefreshToken
	if err := s.db.Where("token_hash = ?", tokenHash).First(&rt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("invalid refresh token")
		}
		return nil, "", err
	}

	if time.Now().After(rt.ExpiresAt) {
		s.db.Delete(&rt)
		return nil, "", errors.New("refresh token expired")
	}

	var user User
	if err := s.db.Where("id = ?", rt.UserID).First(&user).Error; err != nil {
		return nil, "", errors.New("user not found")
	}

	newRawToken, err := middleware.GenerateRandomToken()
	if err != nil {
		return nil, "", err
	}
	newTokenHash := middleware.HashToken(newRawToken)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&rt).Error; err != nil {
			return err
		}
		newRT := RefreshToken{
			ID:        uuid.New(),
			UserID:    rt.UserID,
			TokenHash: newTokenHash,
			ExpiresAt: time.Now().Add(ttl),
			CreatedAt: time.Now(),
		}
		return tx.Create(&newRT).Error
	})

	if err != nil {
		return nil, "", err
	}

	return &user, newRawToken, nil
}

func (s *Service) RevokeRefreshToken(rawToken string) error {
	tokenHash := middleware.HashToken(rawToken)
	return s.db.Where("token_hash = ?", tokenHash).Delete(&RefreshToken{}).Error
}

