package models

import (
	"time"

	"github.com/google/uuid"
)

type PauseReason string

const (
	PauseReasonUser           PauseReason = "USER_INITIATED"
	PauseReasonSystem         PauseReason = "SYSTEM_PAUSE"
	PauseReasonPaymentFailure PauseReason = "PAYMENT_FAILURE"
)

// SubscriptionPausedContext stores paused state context
type SubscriptionPausedContext struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AutoResumeTime time.Time `gorm:"not null"` // When subscription is expected to auto-resume

	CreatedAt time.Time
	UpdatedAt time.Time
}

// SubscriptionPausedTransitionHistory tracks changes related to subscription pauses.
type SubscriptionPausedTransitionHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index:idx_subscription_paused_history,priority:1"`
	PausedAt       time.Time `gorm:"not null;index:idx_subscription_paused_history,priority:2"`

	PreviousPausedContextID *uuid.UUID `gorm:"type:uuid;null;index"` // Previous paused context (before this event)
	CurrentPausedContextID  *uuid.UUID `gorm:"type:uuid;null;index"` // Current paused context (after this event)

	ResumedAt             *time.Time `gorm:"null"` // When the pause ended (null if still paused)
	AutoResumeTimeAtPause *time.Time `gorm:"null"` // What the planned auto-resume was when it paused

	Reason PauseReason `gorm:"type:varchar(50);null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
