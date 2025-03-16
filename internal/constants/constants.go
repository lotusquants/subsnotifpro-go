// internal//constants/constants.go
package constants

// Subscription Notification Types (Explicit Constants)
const (
	SUBSCRIPTION_RECOVERED                 = 1
	SUBSCRIPTION_RENEWED                   = 2
	SUBSCRIPTION_CANCELED                  = 3
	SUBSCRIPTION_PURCHASED                 = 4
	SUBSCRIPTION_ON_HOLD                   = 5
	SUBSCRIPTION_IN_GRACE_PERIOD           = 6
	SUBSCRIPTION_RESTARTED                 = 7
	SUBSCRIPTION_PRICE_CHANGE_CONFIRMED    = 8
	SUBSCRIPTION_DEFERRED                  = 9
	SUBSCRIPTION_PAUSED                    = 10
	SUBSCRIPTION_PAUSE_SCHEDULE_CHANGED    = 11
	SUBSCRIPTION_REVOKED                   = 12
	SUBSCRIPTION_EXPIRED                   = 13
	SUBSCRIPTION_PENDING_PURCHASE_CANCELED = 20
)

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

// SubscriptionNotificationTypes maps Google's integer notification types to readable event names
var SubscriptionNotificationTypes = map[int]string{
	SUBSCRIPTION_RECOVERED:                 "SUBSCRIPTION_RECOVERED",
	SUBSCRIPTION_RENEWED:                   "SUBSCRIPTION_RENEWED",
	SUBSCRIPTION_CANCELED:                  "SUBSCRIPTION_CANCELED",
	SUBSCRIPTION_PURCHASED:                 "SUBSCRIPTION_PURCHASED",
	SUBSCRIPTION_ON_HOLD:                   "SUBSCRIPTION_ON_HOLD",
	SUBSCRIPTION_IN_GRACE_PERIOD:           "SUBSCRIPTION_IN_GRACE_PERIOD",
	SUBSCRIPTION_RESTARTED:                 "SUBSCRIPTION_RESTARTED",
	SUBSCRIPTION_PRICE_CHANGE_CONFIRMED:    "SUBSCRIPTION_PRICE_CHANGE_CONFIRMED",
	SUBSCRIPTION_DEFERRED:                  "SUBSCRIPTION_DEFERRED",
	SUBSCRIPTION_PAUSED:                    "SUBSCRIPTION_PAUSED",
	SUBSCRIPTION_PAUSE_SCHEDULE_CHANGED:    "SUBSCRIPTION_PAUSE_SCHEDULE_CHANGED",
	SUBSCRIPTION_REVOKED:                   "SUBSCRIPTION_REVOKED",
	SUBSCRIPTION_EXPIRED:                   "SUBSCRIPTION_EXPIRED",
	SUBSCRIPTION_PENDING_PURCHASE_CANCELED: "SUBSCRIPTION_PENDING_PURCHASE_CANCELED",
}

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
