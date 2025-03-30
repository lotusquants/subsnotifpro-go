package models

import "time"

type GoogleAccount struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AppUserID string `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`

	// Required identifier
	ObfuscatedExternalAccountID string `gorm:"type:varchar(100);not null;uniqueIndex"` // Core lookup field

	// Optional fields
	ExternalAccountID           *string `gorm:"type:varchar(100);null"`
	ObfuscatedExternalProfileID *string `gorm:"type:varchar(100);null"`

	ProfileID    *string `gorm:"type:varchar(100);null"`
	ProfileName  *string `gorm:"type:varchar(100);null"`
	EmailAddress *string `gorm:"type:varchar(255);null"`
	GivenName    *string `gorm:"type:varchar(50);null"`
	FamilyName   *string `gorm:"type:varchar(50);null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
