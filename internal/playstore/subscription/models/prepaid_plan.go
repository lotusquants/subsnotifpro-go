package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	LineItemID     uuid.UUID `gorm:"type:uuid;not null;index"`

	ProductID string `gorm:"type:varchar(100);not null;index"`

	ExpiryTime time.Time `gorm:"not null;index"`

	// AllowExtendAfterTime defines when the plan can be extended with a top-up
	AllowExtendAfterTime *time.Time     `gorm:"null"`
	CreatedAt            time.Time      `gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `gorm:"autoCreateTime"`
	DeletedAt            gorm.DeletedAt `gorm:"index"`
}

type PrepaidPlanHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`
	LineItemID     uuid.UUID `gorm:"type:uuid;not null;index"`

	ChangeType PrepaidPlanChangeType `gorm:"type:varchar(50);not null;index"`

	ProductID string `gorm:"type:varchar(100);not null;index"`

	PreviousAllowExtendAfterTime *time.Time `gorm:"null"` // Previous top-up allowed time
	CurrentAllowExtendAfterTime  *time.Time `gorm:"null"` // Current top-up allowed time

	PreviousExpiryTime *time.Time `gorm:"null"` // Expiry time before the change (if any)
	CurrentExpiryTime  *time.Time `gorm:"null"` // New expiry time after the change

	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"` // RTDN/Trigger ID
	ChangedAt     time.Time `gorm:"autoCreateTime"`
}
