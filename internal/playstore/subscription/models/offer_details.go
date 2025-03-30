package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OfferDetails struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`
	LineItemID     uuid.UUID `gorm:"type:uuid;not null;index"`

	BasePlanID string    `gorm:"type:varchar(100);not null;index"`
	OfferID    *string   `gorm:"type:varchar(100);null;index"`
	OfferTags  *[]string `gorm:"type:text;null"`

	// Pricing information resolved from products catalog
	BasePlanPrice          Money `gorm:"type:jsonb"`
	CurrentOfferPhaseIndex *int  `gorm:"type:int;null"`
	CurrentPhasePrice      Money `gorm:"type:jsonb"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type OfferDetailsHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OfferDetailsID uuid.UUID `gorm:"type:uuid;not null;index"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`
	LineItemID     uuid.UUID `gorm:"type:uuid;not null;index"`

	BasePlanID string  `gorm:"type:varchar(100);not null"`
	OfferID    *string `gorm:"type:varchar(100);null"`

	OfferTags         *[]string `gorm:"type:text;null"`
	PreviousOfferTags *[]string `gorm:"type:text;null"`

	PreviousBasePlanID *string `gorm:"type:varchar(100);not null"`
	PreviousOfferID    *string `gorm:"type:varchar(100);null"`

	PreviousOfferPhaseIndex *int `gorm:"type:int;null"`
	CurrentOfferPhaseIndex  *int `gorm:"type:int;null"`

	// Pricing history
	PreviousBasePlanPrice *Money `gorm:"type:jsonb"`
	CurrentBasePlanPrice  Money  `gorm:"type:jsonb"`

	PreviousPhasePrice *Money    `gorm:"type:jsonb"`
	CurrentPhasePrice  Money     `gorm:"type:jsonb"`
	ChangeType         string    `gorm:"type:varchar(50);not null;index"` // ADDED, UPDATED, REMOVED
	ChangeEventID      uuid.UUID `gorm:"type:uuid;not null;index"`
	ChangedAt          time.Time `gorm:"autoCreateTime"`
}
