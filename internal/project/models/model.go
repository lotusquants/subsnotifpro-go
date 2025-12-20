package models

import (
	"gorm.io/gorm"
	"time"
)

type Project struct {
	ID             string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name           string         `gorm:"not null" json:"name"`
	Slug           string         `gorm:"uniqueIndex;not null" json:"slug"`
	OrganizationID string         `gorm:"not null" json:"organization_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
