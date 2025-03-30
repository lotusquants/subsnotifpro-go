package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeferredItemReplacement tracks product replacement (at next renewal)
type DeferredItemReplacement struct {
	ID                uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID    uuid.UUID      `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`
	LineItemID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	PreviousProductID string         `gorm:"type:varchar(100);not null;index"` // Current product
	NewProductID      string         `gorm:"type:varchar(100);not null;index"` // New product replacing it
	CreatedAt         time.Time      `gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `gorm:"autoCreateTime"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

type DeferredItemReplacementHistory struct {
	ID                uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID    uuid.UUID `gorm:"type:uuid;not null;index"`
	LineItemID        uuid.UUID `gorm:"type:uuid;not null;index"`
	PreviousProductID *string   `gorm:"type:varchar(100);not null;index"` // Product being replaced
	NewProductID      string    `gorm:"type:varchar(100);not null;index"` // Replacement product
	ChangeType        string    `gorm:"type:varchar(50);not null;index"`  // ADDED, UPDATED, REMOVED
	ChangeEventID     uuid.UUID `gorm:"type:uuid;not null;index"`         // RTDN/Trigger ID
	CreatedAt         time.Time `gorm:"autoCreateTime"`
}
