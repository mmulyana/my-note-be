package users

import (
	"errors"
	"time"

	"my-note-be/internal/middleware"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken           = errors.New("email already exists")
	ErrNotGuest             = errors.New("account is not a guest")
	ErrIsGuest              = errors.New("guest accounts have no password; save your account first")
	ErrCurrentPasswordWrong = errors.New("current password is incorrect")
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
		return nil, ErrEmailTaken
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

func (s *Service) CreateGuest() (*User, error) {
	user := NewGuestUser()
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpgradeGuest(id, email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var taken int64
		if err := tx.Model(&User{}).Where("email = ?", email).Count(&taken).Error; err != nil {
			return err
		}
		if taken > 0 {
			return ErrEmailTaken
		}

		now := time.Now()
		res := tx.Model(&User{}).Where("id = ? AND is_guest = ?", id, true).Updates(map[string]any{
			"email":      email,
			"password":   string(hashedPassword),
			"is_guest":   false,
			"created_at": now,
			"updated_at": now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotGuest
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.FindByID(id)
}

func (s *Service) Login(email, password string) (*User, error) {
	var user User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if user.Password == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)); err != nil {
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

func (s *Service) ChangePassword(id, currentPassword, newPassword string) error {
	user, err := s.FindByID(id)
	if err != nil {
		return err
	}

	if user.IsGuest || user.Password == nil {
		return ErrIsGuest
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(currentPassword)); err != nil {
		return ErrCurrentPasswordWrong
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Model(&User{}).Where("id = ?", id).Update("password", string(hashedPassword)).Error
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
