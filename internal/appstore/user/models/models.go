package models

import (
	"time"

	"github.com/google/uuid"
)

type AppleAccount struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AppUserID uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`

	// Primary identifiers from Apple
	AppAccountToken string  `gorm:"type:uuid;not null;uniqueIndex"`
	BundleID        *string `gorm:"type:varchar(255);null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
