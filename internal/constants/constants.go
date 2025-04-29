// internal//constants/constants.go
package constants

// One-Time Product Notification Types
const (
	ONE_TIME_PRODUCT_PURCHASED = 1
	ONE_TIME_PRODUCT_CANCELED  = 2
)

// Voided Purchase Notification Types
const (
	PRODUCT_TYPE_SUBSCRIPTION = 1
	PRODUCT_TYPE_ONE_TIME     = 2
)

// Voided Purchase Refund Types
const (
	REFUND_TYPE_FULL    = 1
	REFUND_TYPE_PARTIAL = 2
)

// OneTimeProductNotificationTypes maps Google's integer notification types for one-time purchases
var OneTimeProductNotificationTypes = map[int]string{
	ONE_TIME_PRODUCT_PURCHASED: "ONE_TIME_PRODUCT_PURCHASED",
	ONE_TIME_PRODUCT_CANCELED:  "ONE_TIME_PRODUCT_CANCELED",
}

// VoidedPurchaseNotificationTypes maps voided purchase events
var VoidedPurchaseNotificationTypes = map[int]string{
	PRODUCT_TYPE_SUBSCRIPTION: "PRODUCT_TYPE_SUBSCRIPTION",
	PRODUCT_TYPE_ONE_TIME:     "PRODUCT_TYPE_ONE_TIME",
}

// VoidedPurchaseRefundTypes maps voided purchase event refund types
var VoidedPurchaseRefundTypes = map[int]string{
	REFUND_TYPE_FULL:    "Full refund",
	REFUND_TYPE_PARTIAL: "Partial refund",
}

const (
	// Interval for processing pending events (seconds)
	EventProcessingInterval = 60

	// Max batch size for processing events
	EventBatchSize = 100
)

const DEFAULT_SYNC_BATCH_SIZE = 200
const MAX_ROWS_PER_TRANSACTION = 1000
