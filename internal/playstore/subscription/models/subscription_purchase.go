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
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PackageName    string    `gorm:"type:varchar(50);not null;index"`
	StartTime      time.Time `gorm:"not null;index"`
	LatestOrderID  string    `gorm:"type:varchar(50);not null;index"`
	PurchaseToken  string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_subscription_purchase_token"`
	IsTestPurchase bool      `gorm:"default:false"`

	UserID uuid.UUID           `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`
	User   *userModels.AppUser `gorm:"foreignKey:UserID"`

	RegionCode           string               `gorm:"type:varchar(5);not null;index"`
	SubscriptionState    SubscriptionState    `gorm:"type:varchar(50);not null;index"`
	AcknowledgementState AcknowledgementState `gorm:"type:varchar(50);not null;index"`

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

type SubscriptionEvent struct {
	ID             uuid.UUID               `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID               `gorm:"type:uuid;not null;index:idx_subscription_event"`
	Subscription   *SubscriptionPurchaseV2 `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:CASCADE"`

	EventID   uuid.UUID `gorm:"type:uuid;not null;index:idx_subscription_event"`
	EventType string    `gorm:"type:varchar(50);not null;index"` // e.g., "renewal", "cancellation", "pause", etc.

	CreatedAt time.Time      `gorm:"autoCreateTime;index"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
