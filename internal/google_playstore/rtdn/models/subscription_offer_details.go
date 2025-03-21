package models

import (
	"time"

	"github.com/google/uuid"
)

// OfferDetails tracks the current offer details for a plan
type OfferDetails struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"` // Tracks which subscription this is related to

	PlanType   string  `gorm:"type:varchar(50);not null;index"` // "AUTO_RENEWING" or "PREPAID"
	BasePlanID string  `gorm:"type:varchar(100);not null;index"`
	OfferID    *string `gorm:"type:varchar(100);null;index"` // Nullable since not all plans have offer ID
	OfferTags  *string `gorm:"type:text;null"`               // Comma-separated tags

	CreatedAt time.Time
	UpdatedAt time.Time
}

type OfferDetailsHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"` // Tracks which subscription this is related to

	PlanType string `gorm:"type:varchar(50);not null;index"` // "AUTO_RENEWING" or "PREPAID"

	// Historical Offer Details
	BasePlanID string  `gorm:"type:varchar(100);not null"`
	OfferID    *string `gorm:"type:varchar(100);null"`
	OfferTags  *string `gorm:"type:text;null"`

	// Metadata about the change
	ChangeType string    `gorm:"type:varchar(50);not null;index"` // ADDED, UPDATED, REMOVED
	ChangedAt  time.Time `gorm:"not null"`

	CreatedAt time.Time
}
