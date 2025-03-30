package models

import (
	"time"

	"gorm.io/gorm"
)

type Organization struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Slug      string         `gorm:"uniqueIndex;not null" json:"slug"`
	OwnerID   string         `gorm:"not null" json:"owner_id"` // foreign key to users.id
	Plan      string         `gorm:"default:'free'" json:"plan"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
