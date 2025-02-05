package models

import (
	"time"

	"gorm.io/gorm"
)

// GooglePlayServiceAccount represents the stored service account JSON metadata
type GooglePlayServiceAccount struct {
	ID          string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	FileName    string         `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath    string         `gorm:"type:text;not null" json:"file_path"`
	Validated   bool           `gorm:"default:false" json:"validated"`
	LastChecked time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"last_checked"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
