package labels

import "github.com/google/uuid"

type Label struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_labels_user_name" json:"userId"`
	Name   string    `gorm:"size:255;not null;uniqueIndex:idx_labels_user_name" json:"name"`
}
