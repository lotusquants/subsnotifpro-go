package models

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionLinkage tracks historical relationships between subscriptions
type SubscriptionLinkage struct {
	ID                  uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	NewSubscriptionID   uuid.UUID `gorm:"type:uuid;not null;index"`
	OldSubscriptionID   uuid.UUID `gorm:"type:uuid;not null;index"`
	LinkedPurchaseToken string    `gorm:"type:varchar(255);not null"`
	NewPurchaseToken    string    `gorm:"type:varchar(255);not null"`

	// Financial context (nullable)
	ProratedChargeAmount *float64 `gorm:"type:decimal(10,2)"`
	ProratedRefundAmount *float64 `gorm:"type:decimal(10,2)"`
	Currency             *string  `gorm:"type:char(3)"`

	CreatedAt     time.Time `gorm:"autoCreateTime;index"`
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"` // RTDN/Trigger ID

	// Relationships
	NewSubscription SubscriptionPurchaseV2 `gorm:"foreignKey:NewSubscriptionID"`
	OldSubscription SubscriptionPurchaseV2 `gorm:"foreignKey:OldSubscriptionID"`
}
