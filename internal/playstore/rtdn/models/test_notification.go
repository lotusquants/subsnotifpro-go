package models

// TestNotification represents a test notification from Google Play Console
type TestNotification struct {
	Version string `json:"version,omitempty" gorm:"default:null"`
}
