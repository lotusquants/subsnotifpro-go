package dto

import "fmt"

// SubscriptionStatus represents the status of an auto-renewable subscription at the time
// the App Store signs the notification.
type SubscriptionStatus int32

const (
	// SubscriptionActive indicates the auto-renewable subscription is active.
	// Value: 1
	SubscriptionActive SubscriptionStatus = 1 + iota

	// SubscriptionExpired indicates the auto-renewable subscription is expired.
	// Value: 2
	SubscriptionExpired

	// SubscriptionBillingRetry indicates the auto-renewable subscription is in a billing retry period.
	// Value: 3
	SubscriptionBillingRetry

	// SubscriptionGracePeriod indicates the auto-renewable subscription is in a Billing Grace Period.
	// Value: 4
	SubscriptionGracePeriod

	// SubscriptionRevoked indicates the auto-renewable subscription is revoked.
	// Value: 5
	SubscriptionRevoked
)

// String returns a human-readable description of the subscription status
func (s SubscriptionStatus) String() string {
	switch s {
	case SubscriptionActive:
		return "Active"
	case SubscriptionExpired:
		return "Expired"
	case SubscriptionBillingRetry:
		return "Billing Retry Period"
	case SubscriptionGracePeriod:
		return "Billing Grace Period"
	case SubscriptionRevoked:
		return "Revoked"
	default:
		return fmt.Sprintf("Unknown status (%d)", s)
	}
}

// Valid checks if the status value is within the valid range
func (s SubscriptionStatus) Valid() bool {
	return s >= SubscriptionActive && s <= SubscriptionRevoked
}
