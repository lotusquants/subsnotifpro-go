package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SignupPromotionType string

const (
	SignupPromotionTypeOneTime SignupPromotionType = "ONETIME"
	SignupPromotionTypeVanity  SignupPromotionType = "VANITY"
)

// SignupPromotion tracks initial signup promotion details
type SignupPromotion struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`
	LineItemID     uuid.UUID `gorm:"type:uuid;not null;index"`

	PromotionType SignupPromotionType `gorm:"type:varchar(50);not null"` // ONETIME or VANITY
	PromotionCode *string             `gorm:"type:varchar(100);null"`    // Only for VANITY

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type SignupPromotionHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`
	LineItemID     uuid.UUID `gorm:"type:uuid;not null;index"`

	PromotionType         SignupPromotionType  `gorm:"type:varchar(50);not null"` // ONETIME or VANITY
	PreviousPromotionType *SignupPromotionType `gorm:"type:varchar(50);not null"` // ONETIME or VANITY

	PromotionCode         *string `gorm:"type:varchar(100);null"` // Only for VANITY
	PreviousPromotionCode *string `gorm:"type:varchar(100);null"` // Only for VANITY

	ChangeType string `gorm:"type:varchar(50);not null;index"` // ADDED, UPDATED, REMOVED

	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"` // RTDN/Trigger ID
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}
