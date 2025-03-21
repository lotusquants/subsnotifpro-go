package models

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionState represents the possible subscription states from Google Play
type SubscriptionState string

const (
	SubscriptionStateUnspecified             SubscriptionState = "SUBSCRIPTION_STATE_UNSPECIFIED"
	SubscriptionStatePending                 SubscriptionState = "SUBSCRIPTION_STATE_PENDING"
	SubscriptionStateActive                  SubscriptionState = "SUBSCRIPTION_STATE_ACTIVE"
	SubscriptionStatePaused                  SubscriptionState = "SUBSCRIPTION_STATE_PAUSED"
	SubscriptionStateInGracePeriod           SubscriptionState = "SUBSCRIPTION_STATE_IN_GRACE_PERIOD"
	SubscriptionStateOnHold                  SubscriptionState = "SUBSCRIPTION_STATE_ON_HOLD"
	SubscriptionStateCanceled                SubscriptionState = "SUBSCRIPTION_STATE_CANCELED"
	SubscriptionStateExpired                 SubscriptionState = "SUBSCRIPTION_STATE_EXPIRED"
	SubscriptionStatePendingPurchaseCanceled SubscriptionState = "SUBSCRIPTION_STATE_PENDING_PURCHASE_CANCELED"
)

// SubscriptionStateModel represents a unique subscription state
type SubscriptionStateModel struct {
	ID    uint              `gorm:"primaryKey;autoIncrement"`               // 🔥 Use Integer for Faster Queries
	State SubscriptionState `gorm:"type:varchar(50);not null;unique;index"` // Ensure Unique & Indexed
}

// SubscriptionStateTransitionHistory used for tracking subscription state changes over time
type SubscriptionStateTransitionHistory struct {
	ID              string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriptionID  uuid.UUID `gorm:"type:uuid;not null;index"` // Foreign key to SubscriptionPurchaseV2
	PreviousStateID uuid.UUID `gorm:"type:uuid;not null;index"` // Foreign key to SubscriptionState (previous state)
	CurrentStateID  uuid.UUID `gorm:"type:uuid;not null;index"` // Foreign key to SubscriptionState (current state)
	ChangedAt       time.Time `gorm:"not null;autoCreateTime"`  // Timestamp when state changed
	Reason          string    `gorm:"type:varchar(255);null"`   // Optional: Reason for state change (e.g., user action, system action)
}

// SubscriptionStateDetails holds additional information about each state
type SubscriptionStateDetails struct {
	HasAccess        bool   // Whether the user has access
	ShortDescription string // Short readable description
	FullDescription  string // Full official description from Google Play
	RequiresAction   bool   // Whether the user needs to take action
	DeveloperMessage string // Message for developers
}

// SubscriptionStateDetailsMap is a map storing the details for each subscription state.
var SubscriptionStateDetailsMap = map[SubscriptionState]SubscriptionStateDetails{
	SubscriptionStateUnspecified: {
		HasAccess:        false,
		ShortDescription: "Unknown subscription state.",
		FullDescription:  "Unspecified subscription state. This should not occur in a production environment.",
		RequiresAction:   false,
		DeveloperMessage: "Investigate this case. This state should not appear normally.",
	},
	SubscriptionStatePending: {
		HasAccess:        false,
		ShortDescription: "Pending payment.",
		FullDescription:  "Subscription was created but awaiting payment during signup. In this state, all items are awaiting payment.",
		RequiresAction:   true,
		DeveloperMessage: "User initiated a purchase but has not completed payment.",
	},
	SubscriptionStateActive: {
		HasAccess:        true,
		ShortDescription: "Subscription is active.",
		FullDescription:  "Subscription is active. (1) If auto-renewing, at least one item is autoRenewEnabled and not expired. (2) If prepaid, at least one item is not expired.",
		RequiresAction:   false,
		DeveloperMessage: "No action required. The user has an active subscription.",
	},
	SubscriptionStatePaused: {
		HasAccess:        false,
		ShortDescription: "Subscription is paused.",
		FullDescription:  "Subscription is paused. This state is only available when the subscription is an auto-renewing plan. In this state, all items are in paused state.",
		RequiresAction:   true,
		DeveloperMessage: "User needs to manually resume subscription before regaining access.",
	},
	SubscriptionStateInGracePeriod: {
		HasAccess:        true,
		ShortDescription: "Payment failed, grace period active.",
		FullDescription:  "Subscription is in a grace period. This state is only available when the subscription is an auto-renewing plan. In this state, all items are in the grace period.",
		RequiresAction:   true,
		DeveloperMessage: "User should update payment method before grace period ends.",
	},
	SubscriptionStateOnHold: {
		HasAccess:        false,
		ShortDescription: "Subscription on hold (payment issue).",
		FullDescription:  "Subscription is on hold (suspended). This state is only available when the subscription is an auto-renewing plan. In this state, all items are on hold.",
		RequiresAction:   true,
		DeveloperMessage: "User needs to update payment details to resume subscription.",
	},
	SubscriptionStateCanceled: {
		HasAccess:        false,
		ShortDescription: "Subscription canceled but still active until expiry.",
		FullDescription:  "Subscription is canceled but not expired yet. This state is only available when the subscription is an auto-renewing plan. All items have autoRenewEnabled set to false.",
		RequiresAction:   false,
		DeveloperMessage: "User canceled the subscription but still has access until expiry.",
	},
	SubscriptionStateExpired: {
		HasAccess:        false,
		ShortDescription: "Subscription has expired.",
		FullDescription:  "Subscription is expired. All items have expiryTime in the past.",
		RequiresAction:   true,
		DeveloperMessage: "User needs to re-purchase the subscription.",
	},
	SubscriptionStatePendingPurchaseCanceled: {
		HasAccess:        false,
		ShortDescription: "Pending purchase canceled.",
		FullDescription:  "Pending transaction for the subscription is canceled. If this pending purchase was for an existing subscription, use linkedPurchaseToken to get the current state of that subscription.",
		RequiresAction:   false,
		DeveloperMessage: "No action required, but user may retry purchase.",
	},
}
