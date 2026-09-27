package users

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	Email     *string   `gorm:"uniqueIndex"`
	Password  *string
	Username  *string
	Photo     *string
	IsGuest   bool
	CreatedAt *time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *User) EmailOrEmpty() string {
	if u.Email == nil {
		return ""
	}
	return *u.Email
}

func NewUser(email, password string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Email:     &email,
		Password:  &password,
		CreatedAt: &now,
		UpdatedAt: now,
	}
}

func NewGuestUser() *User {
	return &User{
		ID:        uuid.New(),
		IsGuest:   true,
		UpdatedAt: time.Now(),
	}
}

type RefreshToken struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"not null;index"`
	TokenHash string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}
