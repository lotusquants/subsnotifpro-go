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

// AcknowledgementStateModel represents a single acknowledgement state
type AcknowledgementStateModel struct {
	ID    uuid.UUID            `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	State AcknowledgementState `gorm:"type:varchar(50);not null;unique"`
}

// AcknowledgementStateTransitionHistory used for tracking acknowledgement state changes over time
type AcknowledgementStateTransitionHistory struct {
	ID              string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID  uuid.UUID `gorm:"type:uuid;not null;index"` // Foreign key to SubscriptionPurchaseV2
	PreviousStateID uuid.UUID `gorm:"type:uuid;not null;index"` // Foreign key to AcknowledgementState (previous state)
	CurrentStateID  uuid.UUID `gorm:"type:uuid;not null;index"` // Foreign key to AcknowledgementState (current state)
	ChangedAt       time.Time `gorm:"not null;autoCreateTime"`  // Timestamp when state changed
	Reason          string    `gorm:"type:varchar(255);null"`   // Optional: Reason for state change (e.g., user action, system action)
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
