package models

import (
	"time"

	"github.com/google/uuid"
)

// Money represents a monetary value in microunits (1 unit = 1,000,000 micros)
type Money struct {
	AmountMicros int64  `gorm:"not null"`                  // Amount in micro-units
	Currency     string `gorm:"type:varchar(10);not null"` // Currency code (USD, INR, etc.)
}

type AutoRenewingPlan struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`

	ProductID         string    `gorm:"type:varchar(100);not null;index"` // Product identifier (SKU)
	CurrentExpiryTime time.Time `gorm:"not null;index"`                   // Current expiry date (most recent)

	AutoRenewEnabled bool  `gorm:"not null;index"`
	RecurringPrice   Money `gorm:"embedded"`

	// Optional: Link to latest price change details (if any)
	PriceChangeDetailsID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`

	// Optional: Link to installment plan (if part of installment plan)
	InstallmentPlanID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`

	OfferDetailsID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`
	// For AutoRenewingPlan
	DeferredItemReplacementID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`
	SignupPromotionID         *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type AutoRenewingPlanChangeType string

const (
	AutoRenewingPlanChangeCreated     AutoRenewingPlanChangeType = "CREATED"       // Initial creation of the plan
	AutoRenewingPlanChangeRenewed     AutoRenewingPlanChangeType = "RENEWED"       // Recurring renewal
	AutoRenewingPlanChangePriceChange AutoRenewingPlanChangeType = "PRICE_CHANGED" // Price update
	AutoRenewingPlanChangeCanceled    AutoRenewingPlanChangeType = "CANCELED"      // Subscription canceled
	AutoRenewingPlanChangePaused      AutoRenewingPlanChangeType = "PAUSED"        // Subscription paused
	AutoRenewingPlanChangeResumed     AutoRenewingPlanChangeType = "RESUMED"       // Subscription resumed from pause
	AutoRenewingPlanChangeReplaced    AutoRenewingPlanChangeType = "REPLACED"      // Replaced by another product
)

type AutoRenewingPlanHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"` // Links to SubscriptionPurchaseV2

	ChangeType AutoRenewingPlanChangeType `gorm:"type:varchar(50);not null;index"` // What type of change (CREATED, RENEWED, etc.)

	ProductID          string     `gorm:"type:varchar(100);not null;index"` // Product identifier at the time of the change
	PreviousExpiryTime *time.Time `gorm:"null"`                             // Expiry time before the change (if any)
	CurrentExpiryTime  *time.Time `gorm:"null"`                             // New expiry time after the change

	PreviousAutoRenewEnabled *bool `gorm:"null"` // Auto-renewal flag before change
	CurrentAutoRenewEnabled  *bool `gorm:"null"` // Auto-renewal flag after change

	PreviousPrice *Money `gorm:"embedded;null"` // Historical price before change
	CurrentPrice  *Money `gorm:"embedded;null"` // Current price after change

	// Optional: Link to price change details if the change was a price update
	PriceChangeDetailsID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`

	// Optional: Link to installment plan if relevant
	InstallmentPlanID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`

	ChangeTime time.Time `gorm:"not null;index"` // When the change occurred
	CreatedAt  time.Time
}

type SubscriptionItemPriceChangeDetails struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`

	NewPrice Money `gorm:"embedded"`

	PriceChangeMode  string `gorm:"type:varchar(50);not null"`
	PriceChangeState string `gorm:"type:varchar(50);not null"`

	ExpectedNewPriceChargeTime *time.Time `gorm:"null"` // When the new price is expected to take effect

	CreatedAt time.Time
	UpdatedAt time.Time
}

type InstallmentPlan struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`

	InitialCommittedPaymentsCount    int  `gorm:"not null"`
	SubsequentCommittedPaymentsCount *int `gorm:"null"` // Optional for fallback to normal plan
	RemainingCommittedPaymentsCount  int  `gorm:"not null"`

	PendingCancellation bool `gorm:"not null"` // Indicates if pending cancellation after commitment

	CreatedAt time.Time
	UpdatedAt time.Time
}

type InstallmentPlanHistory struct {
	ID                          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID              uuid.UUID `gorm:"type:uuid;not null;index"`
	PreviousInitialCount        *int      `gorm:"null"`
	CurrentInitialCount         *int      `gorm:"null"`
	PreviousSubsequentCount     *int      `gorm:"null"`
	CurrentSubsequentCount      *int      `gorm:"null"`
	PreviousRemainingCount      *int      `gorm:"null"`
	CurrentRemainingCount       *int      `gorm:"null"`
	PreviousPendingCancellation *bool     `gorm:"null"`
	CurrentPendingCancellation  *bool     `gorm:"null"`
	ChangeTime                  time.Time `gorm:"not null"`
	CreatedAt                   time.Time
}

type PrepaidPlanChangeType string

const (
	PrepaidPlanChangeCreated         PrepaidPlanChangeType = "CREATED"           // Initial purchase
	PrepaidPlanChangeExtended        PrepaidPlanChangeType = "EXTENDED"          // Extended (top-up)
	PrepaidPlanChangeExpired         PrepaidPlanChangeType = "EXPIRED"           // Expired
	PrepaidPlanChangeConvertedToAuto PrepaidPlanChangeType = "CONVERTED_TO_AUTO" // Converted to auto-renew
)

type PrepaidPlan struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`

	ProductID         string    `gorm:"type:varchar(100);not null;index"` // Product identifier
	CurrentExpiryTime time.Time `gorm:"not null;index"`                   // Current expiry time

	// AllowExtendAfterTime defines when the plan can be extended with a top-up
	AllowExtendAfterTime *time.Time `gorm:"null"`

	OfferDetailsID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`
	// For AutoRenewingPlan
	DeferredItemReplacementID *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`
	SignupPromotionID         *uuid.UUID `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`

	// Recurring price at the time of purchase (although prepaid, this helps track product cost)
	Price Money `gorm:"embedded"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type PrepaidPlanHistory struct {
	ID             uuid.UUID             `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID             `gorm:"type:uuid;not null;index"`
	ChangeType     PrepaidPlanChangeType `gorm:"type:varchar(50);not null;index"` // CREATED, EXTENDED, EXPIRED, CONVERTED_TO_AUTO

	ProductID          string     `gorm:"type:varchar(100);not null;index"` // Historical product ID
	PreviousExpiryTime *time.Time `gorm:"null"`                             // Expiry time before this change
	CurrentExpiryTime  *time.Time `gorm:"null"`                             // Expiry time after this change

	PreviousAllowExtendAfterTime *time.Time `gorm:"null"` // Previous top-up allowed time
	CurrentAllowExtendAfterTime  *time.Time `gorm:"null"` // Current top-up allowed time

	PreviousPrice *Money `gorm:"embedded;null"` // Price before change
	CurrentPrice  *Money `gorm:"embedded;null"` // Price after change

	ChangeTime time.Time `gorm:"not null;index"` // Timestamp when change happened
	CreatedAt  time.Time
}

// DeferredItemReplacement tracks product replacement (at next renewal)
type DeferredItemReplacement struct {
	ID                   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID       uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`
	PlanType             string    `gorm:"type:varchar(50);not null;index"`  // AUTO_RENEWING or PREPAID
	ProductID            string    `gorm:"type:varchar(100);not null;index"` // Current product
	ReplacementProductID string    `gorm:"type:varchar(100);not null;index"` // New product replacing it

	ScheduledReplacementDate *time.Time `gorm:"null"` // Optional date for scheduled replacement

	CreatedAt time.Time
	UpdatedAt time.Time
}

type DeferredItemReplacementHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`
	PlanType       string    `gorm:"type:varchar(50);not null;index"` // AUTO_RENEWING or PREPAID

	PreviousProductID        string     `gorm:"type:varchar(100);not null;index"` // Product being replaced
	NewProductID             string     `gorm:"type:varchar(100);not null;index"` // Replacement product
	ChangeType               string     `gorm:"type:varchar(50);not null;index"`  // ADDED, UPDATED, REMOVED
	ChangeTime               time.Time  `gorm:"not null"`
	ScheduledReplacementDate *time.Time `gorm:"null"`

	CreatedAt time.Time
}

// SignupPromotion tracks initial signup promotion details
type SignupPromotion struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`
	PlanType       string    `gorm:"type:varchar(50);not null;index"` // AUTO_RENEWING or PREPAID

	PromotionType string  `gorm:"type:varchar(50);not null"` // ONETIME or VANITY
	PromotionCode *string `gorm:"type:varchar(100);null"`    // Only for VANITY

	CreatedAt time.Time
	UpdatedAt time.Time
}

type SignupPromotionHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`
	PlanType       string    `gorm:"type:varchar(50);not null;index"` // AUTO_RENEWING or PREPAID

	PromotionType string  `gorm:"type:varchar(50);not null"`
	PromotionCode *string `gorm:"type:varchar(100);null"`

	ChangeType string    `gorm:"type:varchar(50);not null;index"` // ADDED, UPDATED, REMOVED
	ChangeTime time.Time `gorm:"not null"`

	CreatedAt time.Time
}
