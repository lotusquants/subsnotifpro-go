package models

import (
	"subsnotifpro-go/internal/playstore/user/models"
	"time"
)

type AppUser struct {
	ID             string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AppUserID      string    `gorm:"type:varchar(100);not null;unique"` // The unique identifier for the user across all platforms
	ActivePlatform string    `gorm:"type:varchar(50);not null"`         // Active platform (Google, Apple, etc.)
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`

	// Platform-specific details
	GoogleAccount *models.GoogleAccount `gorm:"foreignKey:UserID"`
	// AppleAccount  *AppleAccount  `gorm:"foreignKey:UserID"`

	// You can extend it for other platforms (e.g., Stripe, Facebook, etc.)
}

type AppUserPlatformChange struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      string    `gorm:"type:uuid;index"`           // Link to the User
	OldPlatform string    `gorm:"type:varchar(50);not null"` // Previous platform (e.g., "Google", "Apple")
	NewPlatform string    `gorm:"type:varchar(50);not null"` // New platform (e.g., "Google", "Apple")
	ChangeDate  time.Time `gorm:"autoCreateTime"`            // Timestamp of when the change happened
}
