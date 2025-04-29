package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebhookEventStatus string

const (
	StatusReceived   WebhookEventStatus = "RECEIVED"
	StatusPublished  WebhookEventStatus = "PUBLISHED"
	StatusProcessing WebhookEventStatus = "PROCESSING"
	StatusProcessed  WebhookEventStatus = "PROCESSED"
	StatusFailed     WebhookEventStatus = "FAILED"
	StatusRetrying   WebhookEventStatus = "RETRYING"
	StatusDeadLetter WebhookEventStatus = "DEAD_LETTER"
)

// AppStoreNotification represents the main notification entity stored in the database
type AppStoreNotification struct {
	gorm.Model
	ID         uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ReceivedAt time.Time          `gorm:"index"`
	Status     WebhookEventStatus `gorm:"type:varchar(20);default:'RECEIVED';index"`
	RetryCount int32              `gorm:"default:0"`

	// Relationships
	ResponseBodyV2DecodedPayloadID *uint
	ResponseBodyV2DecodedPayload   *ResponseBodyV2DecodedPayload `gorm:"foreignKey:ResponseBodyV2DecodedPayloadID"`

	JWSDecodedHeaderID *uint
	JWSDecodedHeader   *JWSDecodedHeader `gorm:"foreignKey:JWSDecodedHeaderID"`
}

// JWSDecodedHeader represents the decoded JWS header
type JWSDecodedHeader struct {
	gorm.Model
	Algorithm string   `gorm:"size:50"` // alg field
	X5c       []string `gorm:"serializer:json"`
}

// ResponseBodyV2DecodedPayload represents the decoded payload from the JWS
type ResponseBodyV2DecodedPayload struct {
	gorm.Model
	Environment Environment `gorm:"size:20;index"`

	// Relationships - only one of these will be populated
	DataID *uint
	Data   *AppStoreNotificationData `gorm:"foreignKey:DataID"`

	SummaryID *uint
	Summary   *AppStoreNotificationSummary `gorm:"foreignKey:SummaryID"`

	ExternalPurchaseToken *string `gorm:"size:255"`

	// Notification metadata
	NotificationType NotificationType     `gorm:"size:50;index"`
	Subtype          *NotificationSubtype `gorm:"size:50;index"`
	NotificationUUID string               `gorm:"size:64;uniqueIndex"`
	Version          string               `gorm:"size:10"`
	SignedDate       time.Time            `gorm:"index"`
}

// Data represents the data object in the decoded payload
type AppStoreNotificationData struct {
	gorm.Model
	Environment   Environment `gorm:"size:20;index"`
	BundleID      string      `gorm:"size:255;index"`
	BundleVersion string      `gorm:"size:50"`
	AppAppleID    *int64      `gorm:"index"`

	// Relationships
	SignedTransactionInfoID *uint
	SignedTransactionInfo   *JWSTransaction `gorm:"foreignKey:SignedTransactionInfoID"`

	SignedRenewalInfoID *uint
	SignedRenewalInfo   *JWSRenewalInfo `gorm:"foreignKey:SignedRenewalInfoID"`

	Status                   *int32  `gorm:"index"` // SubscriptionStatus
	ConsumptionRequestReason *string `gorm:"size:50"`
}

// Summary represents the summary object for RENEWAL_EXTENSION notifications
type AppStoreNotificationSummary struct {
	gorm.Model
	RequestIdentifier      string      `gorm:"size:255;index"`
	Environment            Environment `gorm:"size:20;index"`
	AppAppleID             int64       `gorm:"index"`
	BundleID               string      `gorm:"size:255;index"`
	ProductId              string      `gorm:"size:255;index"`
	StorefrontCountryCodes []string    `gorm:"serializer:json"`
	FailedCount            int64
	SucceededCount         int64
}

// JWSTransaction represents the decoded transaction info
type JWSTransaction struct {
	gorm.Model
	TransactionString string `gorm:"type:text"` // Original JWS string

	// Relationships
	DecodedPayloadID *uint
	DecodedPayload   *JWSTransactionDecodedPayload `gorm:"foreignKey:DecodedPayloadID"`
}

// JWSTransactionDecodedPayload represents the decoded transaction payload
type JWSTransactionDecodedPayload struct {
	gorm.Model
	AppAccountToken               *string   `gorm:"type:uuid;index"`
	AppTransactionId              string    `gorm:"size:255;index"`
	BundleId                      string    `gorm:"size:255;index"`
	Currency                      string    `gorm:"size:3"`
	Environment                   string    `gorm:"size:20;index"`
	ExpiresDate                   time.Time `gorm:"index"`
	InAppOwnershipType            string    `gorm:"size:50"`
	IsUpgraded                    bool
	OfferDiscountType             string `gorm:"size:50"`
	OfferIdentifier               string `gorm:"size:255"`
	OfferPeriod                   string `gorm:"size:50"`
	OfferType                     int32
	OriginalPurchaseDate          time.Time `gorm:"index"`
	OriginalTransactionId         string    `gorm:"size:255;index"`
	PreviousOriginalTransactionId *string   `gorm:"size:255"`
	Price                         int64
	ProductId                     string    `gorm:"size:255;index"`
	PurchaseDate                  time.Time `gorm:"index"`
	Quantity                      int32
	RevocationDate                *time.Time `gorm:"index"`
	RevocationReason              *int32
	SignedDate                    time.Time `gorm:"index"`
	Storefront                    string    `gorm:"size:3"`
	StorefrontId                  string    `gorm:"size:255"`
	SubscriptionGroupIdentifier   string    `gorm:"size:255"`
	TransactionId                 string    `gorm:"size:255;index"`
	TransactionReason             string    `gorm:"size:50"`
	Type                          string    `gorm:"size:50"`
	WebOrderLineItemId            string    `gorm:"size:255"`
}

// JWSRenewalInfo represents the decoded renewal info
type JWSRenewalInfo struct {
	gorm.Model
	RenewalInfoString string `gorm:"type:text"` // Original JWS string

	// Relationships
	DecodedPayloadID *uint
	DecodedPayload   *JWSRenewalInfoDecodedPayload `gorm:"foreignKey:DecodedPayloadID"`
}

// JWSRenewalInfoDecodedPayload represents the decoded renewal info payload
type JWSRenewalInfoDecodedPayload struct {
	gorm.Model
	AppAccountToken             *uuid.UUID `gorm:"type:uuid;index"`
	AppTransactionId            string     `gorm:"size:255;index"`
	AutoRenewProductId          string     `gorm:"size:255;index"`
	AutoRenewStatus             int32      `gorm:"index"`
	Currency                    string     `gorm:"size:3"`
	EligibleWinBackOfferIds     []string   `gorm:"serializer:json"`
	Environment                 string     `gorm:"size:20;index"`
	ExpirationIntent            int32
	GracePeriodExpiresDate      time.Time `gorm:"index"`
	IsInBillingRetryPeriod      bool
	OfferDiscountType           string `gorm:"size:50"`
	OfferIdentifier             string `gorm:"size:255"`
	OfferPeriod                 string `gorm:"size:50"`
	OfferType                   int32
	OriginalTransactionId       string `gorm:"size:255;index"`
	PriceIncreaseStatus         int32
	ProductId                   string    `gorm:"size:255;index"`
	RecentSubscriptionStartDate time.Time `gorm:"index"`
	RenewalDate                 time.Time `gorm:"index"`
	RenewalPrice                int64
	SignedDate                  time.Time `gorm:"index"`
}

// Environment represents the server environment
type Environment string

const (
	EnvironmentSandbox    Environment = "Sandbox"
	EnvironmentProduction Environment = "Production"
)

// ConsumptionRequestReason represents possible consumption request reasons
type ConsumptionRequestReason string

const (
	ConsumptionRequestReasonUnintentedPurchase  ConsumptionRequestReason = "UNINTENDED_PURCHASE"
	ConsumptionRequestReasonFullfillmentIssue   ConsumptionRequestReason = "FULFILLMENT_ISSUE"
	ConsumptionRequestReasonUnsatisfiedPurchase ConsumptionRequestReason = "UNSATISFIED_WITH_PURCHASE"
	ConsumptionRequestReasonLegal               ConsumptionRequestReason = "LEGAL"
	ConsumptionRequestReasonOther               ConsumptionRequestReason = "OTHER"
)

// SubscriptionStatus represents the status of an auto-renewable subscription
type SubscriptionStatus int32

const (
	SubscriptionActive SubscriptionStatus = 1 + iota
	SubscriptionExpired
	SubscriptionBillingRetry
	SubscriptionGracePeriod
	SubscriptionRevoked
)

// NotificationType represents the type of App Store Server Notification
type NotificationType string

const (
	// CONSUMPTION_REQUEST indicates that the customer initiated a refund request for a consumable
	// in-app purchase or auto-renewable subscription, and the App Store is requesting consumption data.
	//
	// This requires the app to provide information about the consumption status of the item.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/consumptionrequest
	CONSUMPTION_REQUEST NotificationType = "CONSUMPTION_REQUEST"

	// DID_CHANGE_RENEWAL_PREF indicates the customer changed their subscription plan.
	//
	// Subtypes:
	// - UPGRADE: User upgraded subscription (effective immediately)
	// - DOWNGRADE: User downgraded subscription (effective at next renewal)
	// - Empty: User reverted to current subscription (canceled downgrade)
	//
	// Upgrades start new billing period with prorated refund for unused portion.
	// Downgrades take effect at next renewal date.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/did_changerenewalpref
	DID_CHANGE_RENEWAL_PREF NotificationType = "DID_CHANGE_RENEWAL_PREF"

	// DID_CHANGE_RENEWAL_STATUS indicates the customer changed subscription auto-renewal status.
	//
	// Subtypes:
	// - AUTO_RENEW_ENABLED: Customer re-enabled auto-renewal
	// - AUTO_RENEW_DISABLED: Customer disabled auto-renewal or App Store disabled after refund
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/did_changerenewalstatus
	DID_CHANGE_RENEWAL_STATUS NotificationType = "DID_CHANGE_RENEWAL_STATUS"

	// DID_FAIL_TO_RENEW indicates subscription failed to renew due to billing issue.
	//
	// Subtypes:
	// - GRACE_PERIOD: Subscription in grace period (continue service)
	// - Empty: Not in grace period (can stop service)
	//
	// App Store retries billing for 60 days or until resolved.
	// Notify customer of billing issue.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/did_failtorenew
	DID_FAIL_TO_RENEW NotificationType = "DID_FAIL_TO_RENEW"

	// DID_RENEW indicates subscription successfully renewed.
	//
	// Subtypes:
	// - BILLING_RECOVERY: Expired subscription that previously failed has renewed
	// - Empty: Active subscription auto-renewed normally
	//
	// Provide customer access to content/service.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/did_renew
	DID_RENEW NotificationType = "DID_RENEW"

	// EXPIRED indicates a subscription expired.
	//
	// Subtypes:
	// - VOLUNTARY: User disabled renewal
	// - BILLING_RETRY: Billing retry period ended without success
	// - PRICE_INCREASE: Customer didn't consent to price increase
	// - PRODUCT_NOT_FOR_SALE: Product unavailable at renewal time
	// - Empty: Expired for other reasons
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/expired
	EXPIRED NotificationType = "EXPIRED"

	// EXTERNAL_PURCHASE_TOKEN indicates Apple created an external purchase token but didn't receive report.
	//
	// Subtypes:
	// - UNREPORTED: Token was not reported
	//
	// Only for apps using External Purchase for alternative payments.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/external_purchasetoken
	EXTERNAL_PURCHASE_TOKEN NotificationType = "EXTERNAL_PURCHASE_TOKEN"

	// GRACE_PERIOD_EXPIRED indicates billing grace period ended without renewal.
	//
	// Turn off access and notify customer of possible billing issue.
	// App Store continues retry billing for 60 days or until resolved.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/grace_periodexpired
	GRACE_PERIOD_EXPIRED NotificationType = "GRACE_PERIOD_EXPIRED"

	// METADATA_UPDATE indicates subscription metadata was changed via Change Subscription Metadata endpoint.
	//
	// Only applies to apps using Advanced Commerce API.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/metadata_update
	METADATA_UPDATE NotificationType = "METADATA_UPDATE"

	// MIGRATION indicates subscription was migrated to Advanced Commerce API.
	//
	// Only applies to apps using Advanced Commerce API.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/migration
	MIGRATION NotificationType = "MIGRATION"

	// OFFER_REDEEMED indicates customer redeemed a subscription offer.
	//
	// Subtypes:
	// - UPGRADE: Offer to upgrade (effective immediately)
	// - DOWNGRADE: Offer to downgrade (effective at next renewal)
	// - Empty: Offer for current subscription
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/offer_redeemed
	OFFER_REDEEMED NotificationType = "OFFER_REDEEMED"

	// ONE_TIME_CHARGE indicates purchase of consumable, non-consumable, or non-renewing subscription.
	//
	// Also sent when customer receives access to non-consumable via Family Sharing.
	// Currently sandbox-only for testing.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/one_time_charge
	ONE_TIME_CHARGE NotificationType = "ONE_TIME_CHARGE"

	// PRICE_CHANGE indicates subscription price was changed via Change Subscription Price endpoint.
	//
	// Only applies to apps using Advanced Commerce API.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/price_change
	PRICE_CHANGE NotificationType = "PRICE_CHANGE"

	// PRICE_INCREASE indicates system informed customer of subscription price increase.
	//
	// Subtypes:
	// - PENDING: Customer hasn't responded (consent required)
	// - ACCEPTED: Customer consented or consent not required
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/price_increase
	PRICE_INCREASE NotificationType = "PRICE_INCREASE"

	// REFUND indicates App Store refunded a transaction.
	//
	// Applies to all purchase types. Includes:
	// - revocationDate: When refund occurred
	// - originalTransactionId: Original transaction
	// - productId: Product refunded
	// - revocationReason: Reason for refund
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/refund
	REFUND NotificationType = "REFUND"

	// REFUND_DECLINED indicates App Store declined a refund request.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/refund_declined
	REFUND_DECLINED NotificationType = "REFUND_DECLINED"

	// REFUND_REVERSED indicates App Store reversed a previously granted refund due to dispute.
	//
	// Applies to all purchase types. For subscriptions, renewal date remains unchanged.
	// If content was revoked due to refund, it should be reinstated.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/refund_reversed
	REFUND_REVERSED NotificationType = "REFUND_REVERSED"

	// RENEWAL_EXTENDED indicates App Store extended subscription renewal date.
	//
	// Result of calling Extend a Subscription Renewal Date or Extend Subscription Renewal Dates for All Active Subscribers.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/renewal_extended
	RENEWAL_EXTENDED NotificationType = "RENEWAL_EXTENDED"

	// RENEWAL_EXTENSION indicates App Store is attempting to extend renewal dates.
	//
	// Subtypes:
	// - SUMMARY: Completed for all eligible subscribers
	// - FAILURE: Failed for specific subscription
	//
	// See details in responseBodyV2DecodedPayload.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/renewal_extension
	RENEWAL_EXTENSION NotificationType = "RENEWAL_EXTENSION"

	// REVOKE indicates Family Sharing entitlement is no longer available.
	//
	// Sent when:
	// - Purchaser disables Family Sharing
	// - Purchaser/family member leaves family group
	// - Purchaser receives refund
	//
	// Applies to non-consumables and auto-renewable subscriptions.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/revoke
	REVOKE NotificationType = "REVOKE"

	// SUBSCRIBED indicates customer subscribed to auto-renewable subscription.
	//
	// Subtypes:
	// - INITIAL_BUY: First-time purchase or Family Sharing access
	// - RESUBSCRIBE: Resubscribed or access to same/different subscription in group
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subscribed
	SUBSCRIBED NotificationType = "SUBSCRIBED"

	// TEST indicates test notification requested via Request a Test Notification endpoint.
	//
	// Only received when explicitly requested for server testing.
	// See: https://developer.apple.com/documentation/appstoreservernotifications/test
	TEST NotificationType = "TEST"
)

// HasSubtype checks if a notification type can have subtypes
func (n NotificationType) HasSubtype() bool {
	switch n {
	case DID_CHANGE_RENEWAL_PREF, DID_CHANGE_RENEWAL_STATUS, DID_FAIL_TO_RENEW,
		DID_RENEW, EXPIRED, OFFER_REDEEMED, PRICE_INCREASE, RENEWAL_EXTENSION,
		SUBSCRIBED:
		return true
	default:
		return false
	}
}

// ValidSubtypes returns valid subtypes for a notification type
func (n NotificationType) ValidSubtypes() []NotificationSubtype {
	switch n {
	case DID_CHANGE_RENEWAL_PREF:
		return []NotificationSubtype{UPGRADE, DOWNGRADE}
	case DID_CHANGE_RENEWAL_STATUS:
		return []NotificationSubtype{AUTO_RENEW_ENABLED, AUTO_RENEW_DISABLED}
	case DID_FAIL_TO_RENEW:
		return []NotificationSubtype{GRACE_PERIOD}
	case DID_RENEW:
		return []NotificationSubtype{BILLING_RECOVERY}
	case EXPIRED:
		return []NotificationSubtype{VOLUNTARY, BILLING_RETRY, PRICE_INCREASE_EXPIRED, PRODUCT_NOT_FOR_SALE}
	case OFFER_REDEEMED:
		return []NotificationSubtype{UPGRADE, DOWNGRADE}
	case PRICE_INCREASE:
		return []NotificationSubtype{PENDING, ACCEPTED}
	case RENEWAL_EXTENSION:
		return []NotificationSubtype{SUMMARY, FAILURE}
	case SUBSCRIBED:
		return []NotificationSubtype{INITIAL_BUY, RESUBSCRIBE}
	default:
		return nil
	}
}

// NotificationSubtype represents possible subtypes for App Store Server Notifications
type NotificationSubtype string

const (
	// ACCEPTED applies to PRICE_INCREASE notifications.
	//
	// Indicates:
	// - Customer consented to price increase (when consent required)
	// - System notified customer (when consent not required)
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	ACCEPTED NotificationSubtype = "ACCEPTED"

	// AUTO_RENEW_DISABLED applies to DID_CHANGE_RENEWAL_STATUS notifications.
	//
	// Indicates:
	// - User disabled auto-renewal
	// - App Store disabled auto-renewal after refund request
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	AUTO_RENEW_DISABLED NotificationSubtype = "AUTO_RENEW_DISABLED"

	// AUTO_RENEW_ENABLED applies to DID_CHANGE_RENEWAL_STATUS notifications.
	//
	// Indicates user re-enabled subscription auto-renewal.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	AUTO_RENEW_ENABLED NotificationSubtype = "AUTO_RENEW_ENABLED"

	// BILLING_RECOVERY applies to DID_RENEW notifications.
	//
	// Indicates an expired subscription that previously failed to renew
	// has now successfully renewed.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	BILLING_RECOVERY NotificationSubtype = "BILLING_RECOVERY"

	// BILLING_RETRY applies to EXPIRED notifications.
	//
	// Indicates subscription expired because billing retry period
	// ended without successful renewal.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	BILLING_RETRY NotificationSubtype = "BILLING_RETRY"

	// DOWNGRADE applies to DID_CHANGE_RENEWAL_PREF and OFFER_REDEEMED notifications.
	//
	// Indicates:
	// - User downgraded subscription
	// - Cross-grade to subscription with different duration
	//
	// Takes effect at next renewal date.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	DOWNGRADE NotificationSubtype = "DOWNGRADE"

	// FAILURE applies to RENEWAL_EXTENSION notifications.
	//
	// Indicates subscription renewal date extension failed for
	// an individual subscription.
	//
	// Check responseBodyV2DecodedPayload for details.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	FAILURE NotificationSubtype = "FAILURE"

	// GRACE_PERIOD applies to DID_FAIL_TO_RENEW notifications.
	//
	// Indicates subscription failed to renew due to billing issue.
	// Continue providing service during grace period.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	GRACE_PERIOD NotificationSubtype = "GRACE_PERIOD"

	// INITIAL_BUY applies to SUBSCRIBED notifications.
	//
	// Indicates:
	// - First-time subscription purchase
	// - First-time access via Family Sharing
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	INITIAL_BUY NotificationSubtype = "INITIAL_BUY"

	// PENDING applies to PRICE_INCREASE notifications.
	//
	// Indicates system informed user of price increase but
	// user hasn't yet responded (when consent required).
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	PENDING NotificationSubtype = "PENDING"

	// PRICE_INCREASE applies to EXPIRED notifications.
	//
	// Indicates subscription expired because user didn't
	// consent to required price increase.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	PRICE_INCREASE_EXPIRED NotificationSubtype = "PRICE_INCREASE"

	// PRODUCT_NOT_FOR_SALE applies to EXPIRED notifications.
	//
	// Indicates subscription expired because product was
	// unavailable at renewal time.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	PRODUCT_NOT_FOR_SALE NotificationSubtype = "PRODUCT_NOT_FOR_SALE"

	// RESUBSCRIBE applies to SUBSCRIBED notifications.
	//
	// Indicates:
	// - User resubscribed
	// - Access via Family Sharing to same/other subscription in group
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	RESUBSCRIBE NotificationSubtype = "RESUBSCRIBE"

	// SUMMARY applies to RENEWAL_EXTENSION notifications.
	//
	// Indicates completion of renewal date extension for
	// all eligible subscribers.
	//
	// Check summary object in responseBodyV2DecodedPayload for details.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	SUMMARY NotificationSubtype = "SUMMARY"

	// UPGRADE applies to DID_CHANGE_RENEWAL_PREF and OFFER_REDEEMED notifications.
	//
	// Indicates:
	// - User upgraded subscription
	// - Cross-grade to subscription with same duration
	//
	// Takes effect immediately.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	UPGRADE NotificationSubtype = "UPGRADE"

	// UNREPORTED applies to EXTERNAL_PURCHASE_TOKEN notifications.
	//
	// Indicates Apple created token but didn't receive report.
	// For apps using External Purchase system.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	UNREPORTED NotificationSubtype = "UNREPORTED"

	// VOLUNTARY applies to EXPIRED notifications.
	//
	// Indicates subscription expired after user disabled auto-renewal.
	//
	// See: https://developer.apple.com/documentation/appstoreservernotifications/subtype
	VOLUNTARY NotificationSubtype = "VOLUNTARY"
)

// ValidForNotificationType checks if a subtype is valid for a given notification type
func (s NotificationSubtype) ValidForNotificationType(nt NotificationType) bool {
	switch nt {
	case DID_CHANGE_RENEWAL_PREF:
		return s == UPGRADE || s == DOWNGRADE
	case DID_CHANGE_RENEWAL_STATUS:
		return s == AUTO_RENEW_ENABLED || s == AUTO_RENEW_DISABLED
	case DID_FAIL_TO_RENEW:
		return s == GRACE_PERIOD
	case DID_RENEW:
		return s == BILLING_RECOVERY
	case EXPIRED:
		return s == VOLUNTARY || s == BILLING_RETRY || s == PRICE_INCREASE_EXPIRED || s == PRODUCT_NOT_FOR_SALE
	case OFFER_REDEEMED:
		return s == UPGRADE || s == DOWNGRADE
	case PRICE_INCREASE:
		return s == PENDING || s == ACCEPTED
	case RENEWAL_EXTENSION:
		return s == SUMMARY || s == FAILURE
	case SUBSCRIBED:
		return s == INITIAL_BUY || s == RESUBSCRIBE
	case EXTERNAL_PURCHASE_TOKEN:
		return s == UNREPORTED
	default:
		return false
	}
}

// Description returns a human-readable description of the subtype
func (s NotificationSubtype) Description() string {
	switch s {
	case ACCEPTED:
		return "Customer consented to price increase or was notified"
	case AUTO_RENEW_DISABLED:
		return "User disabled auto-renewal or App Store disabled after refund"
	case AUTO_RENEW_ENABLED:
		return "User re-enabled subscription auto-renewal"
	case BILLING_RECOVERY:
		return "Expired subscription that failed to renew has now renewed"
	case BILLING_RETRY:
		return "Subscription expired after billing retry period ended"
	case DOWNGRADE:
		return "User downgraded or cross-graded subscription (effective next renewal)"
	case FAILURE:
		return "Renewal date extension failed for individual subscription"
	case GRACE_PERIOD:
		return "Subscription in billing grace period"
	case INITIAL_BUY:
		return "First-time purchase or Family Sharing access"
	case PENDING:
		return "Price increase pending customer consent"
	case PRICE_INCREASE_EXPIRED:
		return "Subscription expired due to unapproved price increase"
	case PRODUCT_NOT_FOR_SALE:
		return "Subscription expired because product was unavailable"
	case RESUBSCRIBE:
		return "User resubscribed or accessed via Family Sharing"
	case SUMMARY:
		return "Renewal date extension completed for eligible subscribers"
	case UPGRADE:
		return "User upgraded or cross-graded subscription (effective immediately)"
	case UNREPORTED:
		return "External purchase token created but not reported"
	case VOLUNTARY:
		return "Subscription expired after user disabled auto-renewal"
	default:
		return "Unknown subtype"
	}
}
