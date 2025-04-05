package dto

// OneTimeProductNotificationType represents one-time purchase notification types
type OneTimeProductNotificationType int

const (
	OneTimeProductPurchased OneTimeProductNotificationType = 1
	OneTimeProductCanceled  OneTimeProductNotificationType = 2
)

func (t OneTimeProductNotificationType) String() string {
	switch t {
	case OneTimeProductPurchased:
		return "ONE_TIME_PRODUCT_PURCHASED"
	case OneTimeProductCanceled:
		return "ONE_TIME_PRODUCT_CANCELED"
	default:
		return "UNKNOWN"
	}
}

type OneTimeProductNotification struct {
	Version          string                         `json:"version,omitempty"`
	NotificationType OneTimeProductNotificationType `json:"notificationType,omitempty"`
	PurchaseToken    string                         `json:"purchaseToken,omitempty"`
	Sku              string                         `json:"sku,omitempty"`
}
