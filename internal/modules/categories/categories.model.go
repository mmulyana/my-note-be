package categories

import "github.com/google/uuid"

type Category struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_categories_user_name" json:"userId"`
	Name   string    `gorm:"size:255;not null;uniqueIndex:idx_categories_user_name" json:"name"`
}
