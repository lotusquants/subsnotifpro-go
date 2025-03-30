package models

import (
	"time"

	"gorm.io/gorm"
)

type AppPlatform string

const (
	PlatformPlayStore AppPlatform = "google_play"
	PlatformAppStore  AppPlatform = "app_store"
	PlatformStripeWeb AppPlatform = "stripe_web"
)

type App struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Slug        string         `gorm:"uniqueIndex;not null" json:"slug"`
	ProjectID   string         `gorm:"not null" json:"project_id"`
	Platform    AppPlatform    `gorm:"not null" json:"platform"`
	PackageName string         `gorm:"not null" json:"package_name"` // bundleId or package
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
