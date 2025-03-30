package models

import (
	"subsnotifpro-go/internal/playstore/user/models"
	"time"

	"github.com/google/uuid"
)

type AppUser struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AppUserID      string    `gorm:"type:varchar(100);not null;unique"` // The unique identifier for the user across all platforms
	ActivePlatform string    `gorm:"type:varchar(50);not null"`         // Active platform (Google, Apple, etc.)
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`

	// Platform-specific details (corrected relationships)
	GoogleAccountID *uuid.UUID            `gorm:"type:uuid"`
	GoogleAccount   *models.GoogleAccount `gorm:"foreignKey:GoogleAccountID;constraint:OnDelete:SET NULL"`

	// AppleAccountID  *uuid.UUID       `gorm:"type:uuid"`
	// AppleAccount    *AppleAccount    `gorm:"foreignKey:AppleAccountID;constraint:OnDelete:SET NULL"`

	PlatformChanges []AppUserPlatformChange `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type AppUserPlatformChange struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`  // Link to the User
	User        AppUser   `gorm:"foreignKey:UserID"`         // Explicit relationship
	OldPlatform string    `gorm:"type:varchar(50);not null"` // Previous platform
	NewPlatform string    `gorm:"type:varchar(50);not null"` // New platform
	ChangeDate  time.Time `gorm:"autoCreateTime;index"`      // Timestamp with index
}
