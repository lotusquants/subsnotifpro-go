package models

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

// OneTimeProductNotification contains one-time purchase details
type OneTimeProductNotification struct {
	Version          string                         `json:"version,omitempty" gorm:"default:null"`
	NotificationType OneTimeProductNotificationType `json:"notificationType,omitempty"` // Now enum
	PurchaseToken    string                         `json:"purchaseToken,omitempty" gorm:"default:null"`
	Sku              string                         `json:"sku,omitempty" gorm:"default:null"`
}
