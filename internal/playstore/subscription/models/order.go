package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionOrderIdTransitionHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"` // FK to SubscriptionPurchaseV2

	PreviousOrderID string    `gorm:"type:varchar(100);not null"`
	NewOrderID      string    `gorm:"type:varchar(100);not null"`
	TransitionedAt  time.Time `gorm:"not null;autoCreateTime"` // Time when the transition was recorded

	Reason        string    `gorm:"type:varchar(255);null"` // Optional reason (e.g., "BasePlan change", "Upgrade", "Replace")
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"`
}
