package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AutoRenewingPlanChangeType string

const (
	AutoRenewingPlanChangeCreated     AutoRenewingPlanChangeType = "CREATED"       // Initial creation of the plan
	AutoRenewingPlanChangeRenewed     AutoRenewingPlanChangeType = "RENEWED"       // Recurring renewal
	AutoRenewingPlanChangePriceChange AutoRenewingPlanChangeType = "PRICE_CHANGED" // Price update
	AutoRenewingPlanChangeCanceled    AutoRenewingPlanChangeType = "CANCELED"      // Subscription canceled
	AutoRenewingPlanChangePaused      AutoRenewingPlanChangeType = "PAUSED"        // Subscription paused
	AutoRenewingPlanChangeResumed     AutoRenewingPlanChangeType = "RESUMED"       // Subscription resumed from pause
	AutoRenewingPlanChangeReplaced    AutoRenewingPlanChangeType = "REPLACED"      // Replaced by another product
	AutoRenewingPlanChangeExpired     AutoRenewingPlanChangeType = "EXPIRED"
)

type AutoRenewingPlan struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`
	LineItemID     uuid.UUID `gorm:"type:uuid;not null;index"`

	ProductID  string    `gorm:"type:varchar(100);not null;index"` // Product identifier (SKU)
	ExpiryTime time.Time `gorm:"not null;index"`                   // Current expiry date (most recent)

	AutoRenewEnabled bool  `gorm:"not null;index"`
	RecurringPrice   Money `gorm:"embedded"`

	// Optional: Link to latest price change details (if any)
	PriceChangeDetailsID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`
	// Add proper foreign key references
	PriceChangeDetails *SubscriptionItemPriceChangeDetails `gorm:"foreignKey:PriceChangeDetailsID"`

	// Optional: Link to installment plan (if part of installment plan)
	InstallmentPlanID *uuid.UUID       `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`
	InstallmentPlan   *InstallmentPlan `gorm:"foreignKey:InstallmentPlanID"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type AutoRenewingPlanHistory struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID     uuid.UUID `gorm:"type:uuid;not null;index"` // Links to SubscriptionPurchaseV2
	LineItemID         uuid.UUID `gorm:"type:uuid;not null;index"`
	AutoRenewingPlanID uuid.UUID `gorm:"type:uuid;not null;index"` // Add reference to parent plan

	ChangeType AutoRenewingPlanChangeType `gorm:"type:varchar(50);not null;index"` // What type of change (CREATED, RENEWED, etc.)

	ProductID          string     `gorm:"type:varchar(100);not null;index"` // Product identifier at the time of the change
	PreviousExpiryTime *time.Time `gorm:"null"`                             // Expiry time before the change (if any)
	CurrentExpiryTime  *time.Time `gorm:"null"`                             // New expiry time after the change

	PreviousAutoRenewEnabled *bool `gorm:"null"` // Auto-renewal flag before change
	CurrentAutoRenewEnabled  *bool `gorm:"null"` // Auto-renewal flag after change

	PreviousPrice *Money `gorm:"embedded;null"` // Historical price before change
	CurrentPrice  *Money `gorm:"embedded;null"` // Current price after change

	// Optional: Link to latest price change details (if any)
	PriceChangeDetailsID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`

	// Optional: Link to installment plan (if part of installment plan)
	InstallmentPlanID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`

	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"` // RTDN/Trigger ID
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}
