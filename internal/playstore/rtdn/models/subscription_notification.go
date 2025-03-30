package models

type SubscriptionNotificationType int

const (
	SubscriptionRecovered               SubscriptionNotificationType = 1
	SubscriptionRenewed                 SubscriptionNotificationType = 2
	SubscriptionCanceled                SubscriptionNotificationType = 3
	SubscriptionPurchased               SubscriptionNotificationType = 4
	SubscriptionOnHold                  SubscriptionNotificationType = 5
	SubscriptionInGracePeriod           SubscriptionNotificationType = 6
	SubscriptionRestarted               SubscriptionNotificationType = 7
	SubscriptionPriceChangeConfirmed    SubscriptionNotificationType = 8
	SubscriptionDeferred                SubscriptionNotificationType = 9
	SubscriptionPaused                  SubscriptionNotificationType = 10
	SubscriptionPauseScheduleChanged    SubscriptionNotificationType = 11
	SubscriptionRevoked                 SubscriptionNotificationType = 12
	SubscriptionExpired                 SubscriptionNotificationType = 13
	SubscriptionPendingPurchaseCanceled SubscriptionNotificationType = 20
)

func (t SubscriptionNotificationType) String() string {
	switch t {
	case SubscriptionRecovered:
		return "SUBSCRIPTION_RECOVERED"
	case SubscriptionRenewed:
		return "SUBSCRIPTION_RENEWED"
	case SubscriptionCanceled:
		return "SUBSCRIPTION_CANCELED"
	case SubscriptionPurchased:
		return "SUBSCRIPTION_PURCHASED"
	case SubscriptionOnHold:
		return "SUBSCRIPTION_ON_HOLD"
	case SubscriptionInGracePeriod:
		return "SUBSCRIPTION_IN_GRACE_PERIOD"
	case SubscriptionRestarted:
		return "SUBSCRIPTION_RESTARTED"
	case SubscriptionPriceChangeConfirmed:
		return "SUBSCRIPTION_PRICE_CHANGE_CONFIRMED"
	case SubscriptionDeferred:
		return "SUBSCRIPTION_DEFERRED"
	case SubscriptionPaused:
		return "SUBSCRIPTION_PAUSED"
	case SubscriptionPauseScheduleChanged:
		return "SUBSCRIPTION_PAUSE_SCHEDULE_CHANGED"
	case SubscriptionRevoked:
		return "SUBSCRIPTION_REVOKED"
	case SubscriptionExpired:
		return "SUBSCRIPTION_EXPIRED"
	case SubscriptionPendingPurchaseCanceled:
		return "SUBSCRIPTION_PENDING_PURCHASE_CANCELED"
	default:
		return "UNKNOWN"
	}
}

// SubscriptionNotification contains subscription-specific details
type SubscriptionNotification struct {
	Version          string                       `json:"version,omitempty" gorm:"default:null"`
	NotificationType SubscriptionNotificationType `json:"notificationType,omitempty"`
	PurchaseToken    string                       `json:"purchaseToken,omitempty" gorm:"default:null"`
	SubscriptionID   string                       `json:"subscriptionId,omitempty" gorm:"default:null"`
}
