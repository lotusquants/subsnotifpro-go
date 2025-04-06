package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// enums

type PlanType string

// enum values
const (
	PlanTypeAutoRenewing PlanType = "AUTO_RENEWING"
	PlanTypePrepaid      PlanType = "PREPAID"
)

// SubscriptionLineItem represents an individual component (base plan or add-on) of a subscription purchase.
type SubscriptionLineItem struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	// 🔗 FK to SubscriptionPurchaseV2
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`

	// 📦 Product/Plan info
	ProductID string `gorm:"type:varchar(100);not null;index"`

	ExpiryTime time.Time `gorm:"not null;index"`
	PlanType   PlanType  `gorm:"type:varchar(20);not null;index"` // AUTO_RENEWING or PREPAID

	// 🧩 Optional Linked Entities
	AutoRenewingPlanID *uuid.UUID        `gorm:"type:uuid;index;constraint:OnDelete:SET NULL"`
	AutoRenewingPlan   *AutoRenewingPlan `gorm:"foreignKey:AutoRenewingPlanID"`

	PrepaidPlanID *uuid.UUID   `gorm:"type:uuid;index;constraint:OnDelete:SET NULL"`
	PrepaidPlan   *PrepaidPlan `gorm:"foreignKey:PrepaidPlanID"`

	OfferDetailsID *uuid.UUID    `gorm:"type:uuid;index;constraint:OnDelete:SET NULL"`
	OfferDetails   *OfferDetails `gorm:"foreignKey:OfferDetailsID"`

	DeferredItemReplacementID *uuid.UUID               `gorm:"type:uuid;index;constraint:OnDelete:SET NULL"`
	DeferredItemReplacement   *DeferredItemReplacement `gorm:"foreignKey:DeferredItemReplacementID"`

	SignupPromotionID *uuid.UUID       `gorm:"type:uuid;index;constraint:OnDelete:SET NULL"`
	SignupPromotion   *SignupPromotion `gorm:"foreignKey:SignupPromotionID"`

	// 🔄 Historical changes can be logged in SubscriptionLineItemHistory
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type SubscriptionLineItemHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`

	LineItemID uuid.UUID `gorm:"type:uuid;not null;index"`
	PlanType   PlanType  `gorm:"type:varchar(20);not null"`

	PreviousExpiryTime *time.Time `gorm:"null"`
	CurrentExpiryTime  time.Time  `gorm:"not null"`

	ChangedAt     time.Time `gorm:"autoCreateTime"`
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"` // RTDN/Trigger ID
	Reason        string    `gorm:"type:varchar(255);null"`   // e.g. "Renewed", "Extended", "Replaced", "Expired", "New"
}
