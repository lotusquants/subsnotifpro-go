package mapper

import (
	"strings"

	appStoreModels "subsnotifpro-go/internal/appstore/subscription/models"
	playStoreModels "subsnotifpro-go/internal/playstore/subscription/models"
	"subsnotifpro-go/internal/subscription/models"
)

// StatusMapper handles conversion between platform-specific and unified statuses
type StatusMapper struct{}

func NewStatusMapper() *StatusMapper {
	return &StatusMapper{}
}

// MapAppleStatus converts App Store status to unified status
func (m *StatusMapper) MapAppleStatus(status appStoreModels.SubscriptionStatus) models.SubscriptionStatus {
	switch status {
	case appStoreModels.SubscriptionStatusActive:
		return models.StatusActive
	case appStoreModels.SubscriptionStatusExpired:
		return models.StatusExpired
	case appStoreModels.SubscriptionStatusBillingRetry:
		return models.StatusBillingRetry
	case appStoreModels.SubscriptionStatusGracePeriod:
		return models.StatusGracePeriod
	case appStoreModels.SubscriptionStatusRevoked:
		return models.StatusRevoked
	case appStoreModels.SubscriptionStatusPendingRenewal:
		return models.StatusPendingRenewal
	case appStoreModels.SubscriptionStatusPendingUpgrade:
		return models.StatusPendingUpgrade
	case appStoreModels.SubscriptionStatusPendingDowngrade:
		return models.StatusPendingDowngrade
	default:
		// Fallback for unknown statuses
		return models.SubscriptionStatus(strings.ToUpper(string(status)))
	}
}

// MapGoogleStatus converts Play Store status to unified status
func (m *StatusMapper) MapGoogleStatus(status playStoreModels.SubscriptionState) models.SubscriptionStatus {
	// First normalize by removing prefix and converting to lowercase
	normalized := strings.ToLower(strings.TrimPrefix(string(status), "SUBSCRIPTION_STATE_"))

	switch normalized {
	case "active":
		return models.StatusActive
	case "expired":
		return models.StatusExpired
	case "in_grace_period":
		return models.StatusGracePeriod
	case "paused":
		return models.StatusPaused
	case "on_hold":
		return models.StatusOnHold
	case "canceled":
		return models.StatusCanceled
	case "pending":
		return models.StatusPendingRenewal
	case "pending_purchase_canceled":
		return models.StatusPendingDowngrade
	case "unspecified":
		// Treat unspecified as pending renewal
		return models.StatusPendingRenewal
	default:
		// Fallback for unknown statuses
		return models.SubscriptionStatus(strings.ToUpper(normalized))
	}
}

// ReverseMap converts unified status back to platform-specific status
func (m *StatusMapper) ReverseMap(platform models.PlatformType, status models.SubscriptionStatus) interface{} {
	switch platform {
	case models.PlatformApple:
		return m.reverseMapToApple(status)
	case models.PlatformGoogle:
		return m.reverseMapToGoogle(status)
	default:
		return string(status)
	}
}

func (m *StatusMapper) reverseMapToApple(status models.SubscriptionStatus) appStoreModels.SubscriptionStatus {
	switch status {
	case models.StatusActive:
		return appStoreModels.SubscriptionStatusActive
	case models.StatusExpired:
		return appStoreModels.SubscriptionStatusExpired
	case models.StatusBillingRetry:
		return appStoreModels.SubscriptionStatusBillingRetry
	case models.StatusGracePeriod:
		return appStoreModels.SubscriptionStatusGracePeriod
	case models.StatusRevoked:
		return appStoreModels.SubscriptionStatusRevoked
	case models.StatusPendingRenewal:
		return appStoreModels.SubscriptionStatusPendingRenewal
	case models.StatusPendingUpgrade:
		return appStoreModels.SubscriptionStatusPendingUpgrade
	case models.StatusPendingDowngrade:
		return appStoreModels.SubscriptionStatusPendingDowngrade
	default:
		return appStoreModels.SubscriptionStatus(string(status))
	}
}

func (m *StatusMapper) reverseMapToGoogle(status models.SubscriptionStatus) playStoreModels.SubscriptionState {
	switch status {
	case models.StatusActive:
		return playStoreModels.SubscriptionStateActive
	case models.StatusExpired:
		return playStoreModels.SubscriptionStateExpired
	case models.StatusGracePeriod:
		return playStoreModels.SubscriptionStateInGracePeriod
	case models.StatusPaused:
		return playStoreModels.SubscriptionStatePaused
	case models.StatusOnHold:
		return playStoreModels.SubscriptionStateOnHold
	case models.StatusCanceled:
		return playStoreModels.SubscriptionStateCanceled
	case models.StatusPendingRenewal:
		return playStoreModels.SubscriptionStatePending
	case models.StatusPendingUpgrade, models.StatusPendingDowngrade:
		return playStoreModels.SubscriptionStatePendingPurchaseCanceled
	default:
		return playStoreModels.SubscriptionState(strings.ToUpper("SUBSCRIPTION_STATE_" + string(status)))
	}
}
