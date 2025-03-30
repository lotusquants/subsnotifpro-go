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

type LineItemStatus string

const (
	LineItemStatusActive   LineItemStatus = "ACTIVE"
	LineItemStatusExpired  LineItemStatus = "EXPIRED"
	LineItemStatusCanceled LineItemStatus = "CANCELED"
)

type LineItemType string

const (
	LineItemTypeBase  LineItemType = "BASE"
	LineItemTypeAddOn LineItemType = "ADD_ON"
)

// SubscriptionLineItem represents an individual component (base plan or add-on) of a subscription purchase.
type SubscriptionLineItem struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	// 🔗 FK to SubscriptionPurchaseV2
	SubscriptionID uuid.UUID               `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`
	Subscription   *SubscriptionPurchaseV2 `gorm:"foreignKey:SubscriptionID"`

	// 📦 Product/Plan info
	ProductID string `gorm:"type:varchar(100);not null;index"`

	ExpiryTime time.Time      `gorm:"not null;index"`
	PlanType   PlanType       `gorm:"type:varchar(20);not null;index"` // AUTO_RENEWING or PREPAID
	Status     LineItemStatus `gorm:"type:varchar(20);not null;index;default:ACTIVE"`
	ItemType   LineItemType   `gorm:"type:varchar(20);not null;index"` // BASE or ADD_ON

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
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type SubscriptionLineItemHistory struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	LineItemID uuid.UUID `gorm:"type:uuid;not null;index"`
	PlanType   PlanType  `gorm:"type:varchar(20);not null"`

	PreviousExpiryTime *time.Time `gorm:"null"`
	CurrentExpiryTime  time.Time  `gorm:"not null"`

	PreviousStatus *LineItemStatus `gorm:"type:varchar(20);null"`
	CurrentStatus  LineItemStatus  `gorm:"type:varchar(20);not null"`

	ChangedAt     time.Time
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"` // RTDN/Trigger ID
	Reason        string    `gorm:"type:varchar(255);null"`   // e.g. "Renewed", "Extended", "Replaced", "Expired", "New"
}
