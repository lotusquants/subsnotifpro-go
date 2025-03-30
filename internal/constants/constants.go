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

// RabbitMQ Constants
const (
	MaxRetries           = 3 // Maximum retry attempts before moving to DLQ
	RetryDelay           = 5 // Retry delay in seconds before re-queuing failed messages
	RTDNQueue            = "rtdn_events"
	RTDNDLX              = "rtdn_dead_letter_exchange"
	RTDNDLQ              = "rtdn_dlq"
	RTDNDLQThreshold     = 10 // DLQThreshold defines the warning threshold for the dead-letter queue
	DLQSizeCheckInterval = 300
)

// **Adaptive DLQ Monitoring Thresholds**
const (
	DLQSizeLowThreshold    = 10 // If DLQ has ≤ 10 messages, check every 5 minutes
	DLQSizeMediumThreshold = 50 // If DLQ has ≤ 50 messages, check every 1 minute
	DLQSizeHighThreshold   = 50 // If DLQ has > 50 messages, check every 30 seconds

	DLQCheckIntervalLow    = 5 * 60 // 5 minutes (300s)
	DLQCheckIntervalMedium = 60     // 1 minute (60s)
	DLQCheckIntervalHigh   = 30     // 30 seconds
)

const (
	// Interval for processing pending events (seconds)
	EventProcessingInterval = 60

	// Max batch size for processing events
	EventBatchSize = 100
)

const DEFAULT_SYNC_BATCH_SIZE = 200
const MAX_ROWS_PER_TRANSACTION = 1000
