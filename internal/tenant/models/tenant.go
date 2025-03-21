package models

import (
	"time"

	"gorm.io/gorm"
)

type Tenant struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name      string         `gorm:"unique;not null" json:"name"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type TenantSetting struct {
	ID        uint           `gorm:"primaryKey"`
	TenantID  string         `gorm:"type:uuid;index;not null"`
	Key       string         `gorm:"type:varchar(100);index;not null"`
	Value     string         `gorm:"type:text;not null"`
	Tenant    Tenant         `gorm:"foreignKey:TenantID"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
