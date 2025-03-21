package models

import (
	"time"

	"subsnotifpro-go/internal/tenant/models"

	"gorm.io/gorm"
)

// GooglePlaySettings stores package name and reference to the latest service account

type GooglePlaySettings struct {
	ID    uint       `gorm:"primaryKey" json:"id"`
	AppID string     `gorm:"type:uuid;not null;index" json:"app_id"`
	App   models.App `gorm:"foreignKey:AppID" json:"app"`

	LatestServiceAccountID *string                   `gorm:"type:uuid;index" json:"latest_account_id"`
	LatestServiceAccount   *GooglePlayServiceAccount `gorm:"foreignKey:LatestServiceAccountID;references:ID" json:"latest_account"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete support (optional in JSON)
}

// GooglePlayServiceAccount represents the stored service account JSON metadata
type GooglePlayServiceAccount struct {
	ID          string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	FileName    string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"file_name"`
	FilePath    string     `gorm:"type:text;not null" json:"file_path"`
	Validated   bool       `gorm:"default:false" json:"validated"`
	LastChecked *time.Time `gorm:"default:null" json:"last_checked"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
