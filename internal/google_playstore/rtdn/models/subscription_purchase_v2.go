package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlanType string

const (
	PlanTypeAutoRenewing PlanType = "AUTO_RENEWING"
	PlanTypePrepaid      PlanType = "PREPAID"
)

// SubscriptionPurchaseV2 represents a subscription purchase object in Google Play
// SubscriptionPurchaseV2 (Stores only the latest state)
type SubscriptionPurchaseV2 struct {
	ID                          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Kind                        string    `gorm:"type:varchar(50);not null"`
	UserID                      uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`
	RegionCodeID                uuid.UUID `gorm:"type:uuid;not null;index"`
	SubscriptionStateModelID    uuid.UUID `gorm:"type:uuid;not null;index"`
	AcknowledgementStateModelID uuid.UUID `gorm:"type:uuid;not null;index"`

	SubscriptionPausedContextID       *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`
	SubscriptionCancellationContextID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`

	// 🔹 A subscription can either be Auto-Renewing or Prepaid (Mutually Exclusive)
	PlanType           PlanType   `gorm:"type:varchar(20);not null"`
	AutoRenewingPlanID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`
	PrepaidPlanID      *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`

	StartTime           time.Time `gorm:"not null;index"`
	LatestOrderID       string    `gorm:"type:varchar(50);not null;unique"`
	LinkedPurchaseToken string    `gorm:"type:varchar(255);not null;index"`

	CreatedAt time.Time      `gorm:"autoCreateTime;index"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
