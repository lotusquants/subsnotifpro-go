package models

import (
	"time"
)

// AppStoreSettings represents the configuration for App Store Connect API
type AppStoreSettings struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	BundleID  string    `json:"bundle_id" gorm:"not null;uniqueIndex:idx_bundle_settings"`
	IssuerID  string    `json:"issuer_id" gorm:"not null"`
	KeyID     string    `json:"key_id" gorm:"not null"` // Also known as KID
	P8Key     string    `json:"p8_key" gorm:"not null"` // Content of the .p8 file
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AppStoreSettingsRequest represents the API request for creating/updating settings
type AppStoreSettingsRequest struct {
	BundleID string `json:"bundle_id" validate:"required"`
	IssuerID string `json:"issuer_id" validate:"required"`
	KeyID    string `json:"key_id" validate:"required"`
	P8Key    string `json:"p8_key" validate:"required"`
}

// AppStoreSettingsResponse represents the API response for settings
type AppStoreSettingsResponse struct {
	ID       uint   `json:"id"`
	BundleID string `json:"bundle_id"`
	IssuerID string `json:"issuer_id"`
	KeyID    string `json:"key_id"`
	P8Key    string `json:"p8_key"`
}

type TestNotificationResponse struct {
	TestNotificationToken string `json:"testNotificationToken"`
}
