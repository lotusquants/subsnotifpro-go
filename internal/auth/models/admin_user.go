// internal/auth/models/admin_user.go
package models

import (
	"time"

	"subsnotifpro-go/internal/tenant/models"

	"gorm.io/gorm"
)

type AdminUser struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TenantID  string         `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant    models.Tenant  `gorm:"foreignKey:TenantID" json:"-"`
	Email     string         `gorm:"type:varchar(100);unique;not null" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Role      string         `gorm:"type:varchar(50);default:'admin'" json:"role"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
