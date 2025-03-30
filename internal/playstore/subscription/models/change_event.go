package models

import (
	"time"

	rtdnModels "subsnotifpro-go/internal/playstore/rtdn/models"

	"github.com/google/uuid"
)

// SubscriptionChangeEvent represents a logical grouping of changes made to a subscription
type SubscriptionChangeEvent struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE"`

	MetadataJSON *string `gorm:"type:jsonb;null"` // Optional structured metadata (diff, source IP, etc.)

	NotificationType rtdnModels.SubscriptionNotificationType `gorm:"type:varchar(50);not null;index"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}
