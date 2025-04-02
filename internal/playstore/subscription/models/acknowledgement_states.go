package models

import (
	"time"

	"github.com/google/uuid"
)

// AcknowledgementState represents the possible acknowledgement states from Google Play
type AcknowledgementState string

const (
	AcknowledgementStateUnspecified  AcknowledgementState = "ACKNOWLEDGEMENT_STATE_UNSPECIFIED"
	AcknowledgementStatePending      AcknowledgementState = "ACKNOWLEDGEMENT_STATE_PENDING"
	AcknowledgementStateAcknowledged AcknowledgementState = "ACKNOWLEDGEMENT_STATE_ACKNOWLEDGED"
)

// IsValid checks if the ack state is valid
func (s AcknowledgementState) IsValid() bool {
	switch s {
	case
		AcknowledgementStateUnspecified, AcknowledgementStatePending, AcknowledgementStateAcknowledged:
		return true
	default:
		return false
	}
}

// String returns the string representation (implements fmt.Stringer)
func (s AcknowledgementState) String() string {
	return string(s)
}

// AcknowledgementStateTransitionHistory used for tracking acknowledgement state changes over time
type AcknowledgementStateTransitionHistory struct {
	ID             string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;not null;index"` // Foreign key to SubscriptionPurchaseV2

	PreviousState *AcknowledgementState `gorm:"type:varchar(50);index"`
	CurrentState  AcknowledgementState  `gorm:"type:varchar(50);not null;index"`

	Reason        string    `gorm:"type:varchar(255);null"` // Optional: Reason for state change (e.g., user action, system action)
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index"`
	ChangedAt     time.Time `gorm:"not null;autoCreateTime"` // Timestamp when state changed
}

// AcknowledgementStateDetails holds additional information about each acknowledgement state
type AcknowledgementStateDetails struct {
	RequiresAcknowledgement bool   // Whether the subscription needs to be acknowledged
	Description             string // Full description from Google Play
	DeveloperMessage        string // Message for developers
}

// AcknowledgementStateDetailsMap is a map storing the details for each acknowledgement state.
var AcknowledgementStateDetailsMap = map[AcknowledgementState]AcknowledgementStateDetails{
	AcknowledgementStateUnspecified: {
		RequiresAcknowledgement: false,
		Description:             "Unspecified acknowledgement state. This should not occur in production.",
		DeveloperMessage:        "Investigate this case. This state should not appear normally.",
	},
	AcknowledgementStatePending: {
		RequiresAcknowledgement: true,
		Description:             "The subscription is not acknowledged yet.",
		DeveloperMessage:        "Call the Google Play API to acknowledge the subscription.",
	},
	AcknowledgementStateAcknowledged: {
		RequiresAcknowledgement: false,
		Description:             "The subscription has been acknowledged.",
		DeveloperMessage:        "No action required. Subscription is properly acknowledged.",
	},
}

// GetAcknowledgementDetails returns detailed information about an acknowledgement state
func GetAcknowledgementDetails(state AcknowledgementState) AcknowledgementStateDetails {
	if details, exists := AcknowledgementStateDetailsMap[state]; exists {
		return details
	}
	// If state is not found, return default details (Unknown state)
	return AcknowledgementStateDetails{
		RequiresAcknowledgement: false,
		Description:             "Unknown acknowledgement state.",
		DeveloperMessage:        "Unexpected state. Investigate this case.",
	}
}
