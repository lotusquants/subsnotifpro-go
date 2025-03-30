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

type ResumeReason string

const (
	ResumeReasonAuto   ResumeReason = "AUTO"
	ResumeReasonManual ResumeReason = "MANUAL"
)

type PauseContextStatus string

const (
	PauseContextStatusScheduled PauseContextStatus = "SCHEDULED"
	PauseContextStatusPaused    PauseContextStatus = "PAUSED"
	PauseContextStatusResumed   PauseContextStatus = "RESUMED"
	PauseContextStatusCancelled PauseContextStatus = "CANCELLED"
	PauseContextStatusExpired   PauseContextStatus = "EXPIRED"
)

type PauseChangeReason string

const (
	PauseChangeCreated        PauseChangeReason = "CREATED"
	PauseChangeCancelled      PauseChangeReason = "CANCELLED"
	PauseChangePausedAuto     PauseChangeReason = "PAUSED_AUTO"
	PauseChangePausedManual   PauseChangeReason = "PAUSED_MANUAL"
	PauseChangeResumedAuto    PauseChangeReason = "RESUMED_AUTO"
	PauseChangeResumedManual  PauseChangeReason = "RESUMED_MANUAL"
	PauseChangeExpired        PauseChangeReason = "EXPIRED"
	PauseChangeScheduleCreate PauseChangeReason = "SCHEDULE_CREATED"
	PauseChangeScheduleEdit   PauseChangeReason = "SCHEDULE_UPDATED"
)

// SubscriptionPausedContext stores paused state context
type SubscriptionPausedContext struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`

	ScheduledAt time.Time  `gorm:"not null"` // When the user scheduled it
	CancelledAt *time.Time // If user cancels before pause begins

	AutoPauseTime  time.Time  `gorm:"not null;index"` // When subscription is expected to pause
	PausedAt       *time.Time `gorm:"null;index"`     // Because it's a pointer// When actually paused(can be paused manually before autoPauseTime)
	AutoResumeTime time.Time  `gorm:"not null;index"` // When subscription is expected to auto-resume
	ResumedAt      *time.Time `gorm:"null;index"`     // When actually resumed(can be resumed manually before autoPauseTime)
	ExpiredAt      *time.Time `gorm:"null;index"`     // When this paused Context got expired.( in case of a subscription expiring)

	Status PauseContextStatus `gorm:"type:varchar(20);not null;default:'SCHEDULED';index"` // Current Status of the pause

	// 🧠 Reason metadata
	PauseReason  PauseReason   `gorm:"type:varchar(50);null"`
	ResumeReason *ResumeReason `gorm:"type:varchar(20);null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoCreateTime"`
}

type SubscriptionPausedContextHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PauseContextID uuid.UUID `gorm:"type:uuid;not null;index"` // FK to current context (optional but good for joins)
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Snapshot of values at this point in time
	Status         PauseContextStatus `gorm:"type:varchar(20);not null;index"` // SCHEDULED, PAUSED, etc.
	ScheduledAt    time.Time          `gorm:"not null"`
	CancelledAt    *time.Time         `gorm:"null"`
	AutoPauseTime  time.Time          `gorm:"not null"`
	PausedAt       *time.Time         `gorm:"null"`
	AutoResumeTime time.Time          `gorm:"not null"`
	ResumedAt      *time.Time         `gorm:"null"`
	ExpiredAt      *time.Time         `gorm:"null;index"`

	SubscriptionState SubscriptionState `gorm:"type:uuid;not null;index"`

	PauseReason  PauseReason   `gorm:"type:varchar(50);null"`
	ResumeReason *ResumeReason `gorm:"type:varchar(20);null"`

	// What triggered this change
	PauseChangeReason PauseChangeReason `gorm:"type:varchar(100);not null;index"`
	ChangeEventID     uuid.UUID         `gorm:"type:uuid;not null;index"` // RTDN/Event ID triggering this (for traceability)

	ChangedAt time.Time `gorm:"autoCreateTime;not null"`
}
