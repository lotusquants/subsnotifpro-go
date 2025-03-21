package models

import (
	"time"

	"gorm.io/gorm"
)

type App struct {
	ID string `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	TenantID string `gorm:"type:uuid;not null;uniqueIndex:idx_tenant_platform_identifier,priority:1" json:"tenant_id"`
	Tenant   Tenant `gorm:"foreignKey:TenantID" json:"tenant"`

	Platform string `gorm:"type:varchar(50);not null;uniqueIndex:idx_tenant_platform_identifier,priority:2" json:"platform"` // "play_store", "app_store", etc.

	AppIdentifier string `gorm:"type:varchar(255);not null;uniqueIndex:idx_tenant_platform_identifier,priority:3" json:"app_identifier"` // e.g. com.app.android

	Name string `gorm:"type:varchar(100);not null" json:"name"` // Human-readable name for UI

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
