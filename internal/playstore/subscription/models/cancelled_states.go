package models

import (
	"time"

	"github.com/google/uuid"
)

type CancellationReason string

const (
	CancellationReasonUserInitiated      CancellationReason = "USER_INITIATED"
	CancellationReasonSystemInitiated    CancellationReason = "SYSTEM_INITIATED"
	CancellationReasonDeveloperInitiated CancellationReason = "DEVELOPER_INITIATED"
	CancellationReasonReplacement        CancellationReason = "REPLACEMENT"
)

type CancelSurveyReason string

const (
	CancelSurveyReasonUnspecified     CancelSurveyReason = "CANCEL_SURVEY_REASON_UNSPECIFIED"
	CancelSurveyReasonNotEnoughUsage  CancelSurveyReason = "CANCEL_SURVEY_REASON_NOT_ENOUGH_USAGE"
	CancelSurveyReasonTechnicalIssues CancelSurveyReason = "CANCEL_SURVEY_REASON_TECHNICAL_ISSUES"
	CancelSurveyReasonCostRelated     CancelSurveyReason = "CANCEL_SURVEY_REASON_COST_RELATED"
	CancelSurveyReasonFoundBetterApp  CancelSurveyReason = "CANCEL_SURVEY_REASON_FOUND_BETTER_APP"
	CancelSurveyReasonOthers          CancelSurveyReason = "CANCEL_SURVEY_REASON_OTHERS"
)

type SubscriptionCancellationContext struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"`

	CancellationReason CancellationReason `gorm:"type:varchar(50);not null"` // User, System, Developer, Replacement
	CancelTime         time.Time          `gorm:"not null"`                  // When the subscription was canceled
	ResubscribeTime    *time.Time         `gorm:"null"`

	CancelSurveyReason    *CancelSurveyReason `gorm:"type:varchar(50);null"`  // Optional user-provided survey reason
	CancelSurveyUserInput *string             `gorm:"type:varchar(255);null"` // Optional freeform input (if reason is "Others")

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoCreateTime"`
}

type SubscriptionCancellationContextHistory struct {
	ID                    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CancellationContextID uuid.UUID `gorm:"type:uuid;not null;index"`
	SubscriptionID        uuid.UUID `gorm:"type:uuid;not null;index"`

	CancellationReason CancellationReason `gorm:"type:varchar(50);not null"`
	CancelTime         time.Time          `gorm:"not null"`

	CancelSurveyReason    *CancelSurveyReason `gorm:"type:varchar(50);null"`
	CancelSurveyUserInput *string             `gorm:"type:varchar(255);null"`

	ResubscribedTime *time.Time `gorm:"null"`

	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"` // RTDN/Trigger ID
	ChangedAt     time.Time `gorm:"autoCreateTime;not null"`
}
