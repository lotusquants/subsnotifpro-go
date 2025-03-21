package models

import "time"

type GoogleAccount struct {
	ID                          string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AppUserID                   string `gorm:"type:uuid;index"`            // Foreign key to the User model
	ExternalAccountID           string `gorm:"type:varchar(100);not null"` // Google External Account ID
	ObfuscatedExternalAccountID string `gorm:"type:varchar(100);not null"` // Obfuscated Google External Account ID
	ObfuscatedExternalProfileID string `gorm:"type:varchar(100);not null"` // Obfuscated Profile ID

	ProfileID    string `gorm:"type:varchar(100);not null"` // Google Profile ID
	ProfileName  string `gorm:"type:varchar(100);not null"` // Google Profile Name
	EmailAddress string `gorm:"type:varchar(255);not null"` // Email Address
	GivenName    string `gorm:"type:varchar(50);not null"`  // Given Name
	FamilyName   string `gorm:"type:varchar(50);not null"`  // Family Name

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
