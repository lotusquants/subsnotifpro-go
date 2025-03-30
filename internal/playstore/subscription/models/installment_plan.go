package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InstallmentPlan struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID     uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`
	LineItemID         uuid.UUID `gorm:"type:uuid;not null;index"`
	AutoRenewingPlanID uuid.UUID `gorm:"type:uuid;not null;index"`

	InitialCommittedPaymentsCount    int  `gorm:"not null"`
	SubsequentCommittedPaymentsCount *int `gorm:"null"`
	RemainingCommittedPaymentsCount  int  `gorm:"not null"`

	PendingCancellation bool `gorm:"not null"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type InstallmentPlanHistory struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	InstallmentPlanID  uuid.UUID `gorm:"type:uuid;not null;index"`
	SubscriptionID     uuid.UUID `gorm:"type:uuid;not null;index"`
	LineItemID         uuid.UUID `gorm:"type:uuid;not null;index"`
	AutoRenewingPlanID uuid.UUID `gorm:"type:uuid;not null;index"`

	PreviousInitialPaymentsCount *int `gorm:"null"`
	NewInitialPaymentsCount      int  `gorm:"not null"`

	PreviousSubsequentPaymentsCount *int `gorm:"null"`
	NewSubsequentPaymentsCount      *int `gorm:"null"`

	PreviousRemainingPaymentsCount *int `gorm:"null"`
	NewRemainingPaymentsCount      int  `gorm:"not null"`

	PreviousPendingCancellation *bool `gorm:"null"`
	NewPendingCancellation      bool  `gorm:"not null"`

	ChangeType    string    `gorm:"type:varchar(50);not null"` // CREATED, UPDATED, CANCELLED
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}
