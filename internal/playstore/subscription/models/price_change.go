package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PriceChangeMode defines how a price change will be applied
type PriceChangeMode string

const (
	PriceChangeModeUnspecified PriceChangeMode = "PRICE_CHANGE_MODE_UNSPECIFIED"
	PriceChangeModeOptIn       PriceChangeMode = "OPT_IN"
	PriceChangeModeOptOut      PriceChangeMode = "OPT_OUT"
)

// Scan implements the Scanner interface for database deserialization
func (m *PriceChangeMode) Scan(value interface{}) error {
	if v, ok := value.(string); ok {
		*m = PriceChangeMode(v)
		return nil
	}
	return errors.New("failed to scan PriceChangeMode")
}

// Value implements the driver Valuer interface for database serialization
func (m PriceChangeMode) Value() (driver.Value, error) {
	return string(m), nil
}

// PriceChangeState defines the state of a price change
type PriceChangeState string

const (
	PriceChangeStateUnspecified PriceChangeState = "PRICE_CHANGE_STATE_UNSPECIFIED"
	PriceChangeStatePending     PriceChangeState = "PENDING"
	PriceChangeStateConfirmed   PriceChangeState = "CONFIRMED"
	PriceChangeStateCancelled   PriceChangeState = "CANCELLED"
)

// Scan implements the Scanner interface for database deserialization
func (s *PriceChangeState) Scan(value interface{}) error {
	if v, ok := value.(string); ok {
		*s = PriceChangeState(v)
		return nil
	}
	return errors.New("failed to scan PriceChangeState")
}

// Value implements the driver Valuer interface for database serialization
func (s PriceChangeState) Value() (driver.Value, error) {
	return string(s), nil
}

// SubscriptionItemPriceChangeDetails tracks price change information for subscription items
type SubscriptionItemPriceChangeDetails struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID     uuid.UUID `gorm:"type:uuid;not null;index"`
	LineItemID         uuid.UUID `gorm:"type:uuid;not null;index"`
	AutoRenewingPlanID uuid.UUID `gorm:"type:uuid;not null;index"`

	NewPrice Money `gorm:"embedded;embeddedPrefix:new_price_"`

	PriceChangeMode  PriceChangeMode  `gorm:"type:varchar(50);not null"`
	PriceChangeState PriceChangeState `gorm:"type:varchar(50);not null"`

	ExpectedNewPriceChangeTime *time.Time `gorm:"null"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// PriceChangeDetailsHistory tracks historical changes to price change details
type SubscriptionItemPriceChangeDetailsHistory struct {
	ID                   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PriceChangeDetailsID uuid.UUID `gorm:"type:uuid;not null;index"`
	SubscriptionID       uuid.UUID `gorm:"type:uuid;not null;index"`
	LineItemID           uuid.UUID `gorm:"type:uuid;not null;index"`
	AutoRenewingPlanID   uuid.UUID `gorm:"type:uuid;not null;index"`

	PreviousPrice *Money `gorm:"embedded;embeddedPrefix:previous_price_;null"`
	NewPrice      Money  `gorm:"embedded;embeddedPrefix:new_price_"`

	PreviousPriceChangeMode *PriceChangeMode `gorm:"type:varchar(50)"`
	NewPriceChangeMode      PriceChangeMode  `gorm:"type:varchar(50);not null"`

	PreviousPriceChangeState *PriceChangeState `gorm:"type:varchar(50)"`
	NewPriceChangeState      PriceChangeState  `gorm:"type:varchar(50);not null"`

	PreviousExpectedChangeTime *time.Time `gorm:"null"`
	NewExpectedChangeTime      *time.Time `gorm:"null"`

	ChangeType    string    `gorm:"type:varchar(50);not null"` // CREATED, UPDATED, EXPIRED
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}
