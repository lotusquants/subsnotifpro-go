package models

import (
	"time"

	"gorm.io/gorm"
)

type SetupStatus string

const (
	SetupPending    SetupStatus = "pending"
	SetupIncomplete SetupStatus = "incomplete"
	SetupComplete   SetupStatus = "complete"
)

type AppSetup struct {
	AppID       string         `gorm:"primaryKey" json:"app_id"`
	Platform    AppPlatform    `gorm:"not null" json:"platform"`
	SetupStatus SetupStatus    `gorm:"not null;default:'pending'" json:"setup_status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type PlaystoreSetup struct {
	AppID                  string         `gorm:"primaryKey" json:"app_id"`
	PackageName            string         `gorm:"not null" json:"package_name"`
	ServiceAccountPath     string         `gorm:"not null" json:"service_account_path"`
	ServiceAccountVerified bool           `gorm:"default:false" json:"service_account_verified"`
	BucketID               *string        `json:"bucket_id"`
	PubSubCreated          bool           `gorm:"default:false" json:"pubsub_created"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}
