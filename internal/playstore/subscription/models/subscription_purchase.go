package models

import (
	"time"

	userModels "subsnotifpro-go/internal/users/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionPurchaseV2 represents a subscription purchase object in Google Play
// SubscriptionPurchaseV2 (Stores only the latest state)
type SubscriptionPurchaseV2 struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PackageName   string    `gorm:"type:varchar(50);not null;index"`
	StartTime     time.Time `gorm:"not null;index"`
	LatestOrderId string    `gorm:"type:varchar(50);not null;index"`
	PurchaseToken string    `gorm:"type:varchar(255);not null;uniqueIndex"`

	UserID uuid.UUID           `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`
	User   *userModels.AppUser `gorm:"foreignKey:UserID"`

	RegionCodeID uuid.UUID  `gorm:"type:uuid;not null;index"`
	RegionCode   RegionCode `gorm:"foreignKey:RegionCodeID"`

	SubscriptionStateModelID uuid.UUID               `gorm:"type:uuid;not null;index"`
	SubscriptionStateModel   *SubscriptionStateModel `gorm:"foreignKey:SubscriptionStateModelID"`

	AcknowledgementStateModelID uuid.UUID                  `gorm:"type:uuid;not null;index"`
	AcknowledgementStateModel   *AcknowledgementStateModel `gorm:"foreignKey:AcknowledgementStateModelID"`

	SubscriptionPausedContextID *uuid.UUID                 `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`
	SubscriptionPausedContext   *SubscriptionPausedContext `gorm:"foreignKey:SubscriptionPausedContextID"`

	SubscriptionCancellationContextID *uuid.UUID                       `gorm:"type:uuid;null;index;constraint:OnDelete:SET NULL"`
	SubscriptionCancellationContext   *SubscriptionCancellationContext `gorm:"foreignKey:SubscriptionCancellationContextID"`

	// 🔁 One subscription -> many line items (e.g., base + add-ons)
	LineItems                []SubscriptionLineItem `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:CASCADE"`
	LinkedPurchaseToken      *string                `gorm:"type:varchar(255);"`
	LinkedFromSubscriptionID *uuid.UUID             `gorm:"type:uuid;index;constraint:OnDelete:SET NULL"`

	CreatedAt time.Time      `gorm:"autoCreateTime;index"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type SubscriptionOrderIdTransitionHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"` // FK to SubscriptionPurchaseV2

	PreviousOrderID string    `gorm:"type:varchar(100);not null"`
	NewOrderID      string    `gorm:"type:varchar(100);not null"`
	TransitionedAt  time.Time `gorm:"not null;autoCreateTime"` // Time when the transition was recorded

	Reason        string    `gorm:"type:varchar(255);null"` // Optional reason (e.g., "BasePlan change", "Upgrade", "Replace")
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"`
}
