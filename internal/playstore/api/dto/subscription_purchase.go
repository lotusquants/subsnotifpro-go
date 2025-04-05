package dto

import (
	"time"
)

// SubscriptionPurchaseV2 represents the complete subscription purchase resource
type SubscriptionPurchaseV2 struct {
	Kind                       string
	RegionCode                 string
	LineItems                  []LineItem
	StartTime                  time.Time
	SubscriptionState          SubscriptionState
	LatestOrderID              string
	LinkedPurchaseToken        *string
	PausedStateContext         *PausedStateContext
	CanceledStateContext       *CanceledStateContext
	IsTestPurchase             bool
	AcknowledgementState       AcknowledgementState
	ExternalAccountIdentifiers ExternalAccountIdentifiers
	SubscribeWithGoogleInfo    *SubscribeWithGoogleInfo
}

// LineItem represents a single purchase line item
type LineItem struct {
	ProductID               string
	ExpiryTime              time.Time
	PlanType                PlanType
	AutoRenewingPlan        *AutoRenewingPlan
	PrepaidPlan             *PrepaidPlan
	OfferDetails            *OfferDetails
	DeferredItemReplacement *DeferredItemReplacement
	SignupPromotion         *SignupPromotion
}

// AutoRenewingPlan details
type AutoRenewingPlan struct {
	AutoRenewEnabled   bool
	RecurringPrice     Money
	PriceChangeDetails *SubscriptionItemPriceChangeDetails
	InstallmentDetails *InstallmentPlan
}

// PrepaidPlan details
type PrepaidPlan struct {
	AllowExtendAfterTime *time.Time
}

// Money represents currency amount
type Money struct {
	CurrencyCode string
	Units        int64
	Nanos        int64
}

// SubscriptionItemPriceChangeDetails for price changes
type SubscriptionItemPriceChangeDetails struct {
	NewPrice                   Money
	PriceChangeMode            PriceChangeMode
	PriceChangeState           PriceChangeState
	ExpectedNewPriceChargeTime time.Time
}

// InstallmentPlan details
type InstallmentPlan struct {
	InitialCommittedPaymentsCount    int
	SubsequentCommittedPaymentsCount int
	RemainingCommittedPaymentsCount  int
	IsPendingCancellation            bool
}

// OfferDetails for promotions
type OfferDetails struct {
	OfferTags  []string
	BasePlanID string
	OfferID    string
}

// DeferredItemReplacement for pending changes
type DeferredItemReplacement struct {
	ProductID string
}

// SignupPromotionType defines the type of promotion applied
type SignupPromotionType string

const (
	PromoTypeOneTimeCode SignupPromotionType = "ONE_TIME_CODE"
	PromoTypeVanityCode  SignupPromotionType = "VANITY_CODE"
)

// SignupPromotion contains promotion details
type SignupPromotion struct {
	Type SignupPromotionType // Type of promotion
	Code *string             // Only for VANITY_CODE
}

// PausedStateContext for paused subscriptions
type PausedStateContext struct {
	AutoResumeTime time.Time
}

const (
	CancellationReasonUserInitiated      CancellationReason = "USER_INITIATED"
	CancellationReasonSystemInitiated    CancellationReason = "SYSTEM_INITIATED"
	CancellationReasonDeveloperInitiated CancellationReason = "DEVELOPER_INITIATED"
	CancellationReasonReplacement        CancellationReason = "REPLACEMENT"
)
const (
	CancelSurveyReasonUnspecified     CancelSurveyReason = "CANCEL_SURVEY_REASON_UNSPECIFIED"
	CancelSurveyReasonNotEnoughUsage  CancelSurveyReason = "CANCEL_SURVEY_REASON_NOT_ENOUGH_USAGE"
	CancelSurveyReasonTechnicalIssues CancelSurveyReason = "CANCEL_SURVEY_REASON_TECHNICAL_ISSUES"
	CancelSurveyReasonCostRelated     CancelSurveyReason = "CANCEL_SURVEY_REASON_COST_RELATED"
	CancelSurveyReasonFoundBetterApp  CancelSurveyReason = "CANCEL_SURVEY_REASON_FOUND_BETTER_APP"
	CancelSurveyReasonOthers          CancelSurveyReason = "CANCEL_SURVEY_REASON_OTHERS"
)

type CanceledStateContext struct {
	Reason             CancellationReason
	CancelTime         *time.Time // Only for user-initiated
	CancelSurveyResult *struct {
		Reason          CancelSurveyReason
		ReasonUserInput *string // Only for SurveyReasonOther
	}
}

// ExternalAccountIdentifiers for third-party accounts
type ExternalAccountIdentifiers struct {
	ExternalAccountID           *string
	ObfuscatedExternalAccountID *string
	ObfuscatedExternalProfileID *string
}

// SubscribeWithGoogleInfo for user profile
type SubscribeWithGoogleInfo struct {
	ProfileID    *string
	ProfileName  *string
	EmailAddress *string
	GivenName    *string
	FamilyName   *string
}

// Enums
type SubscriptionState string
type AcknowledgementState string
type CancelSurveyReason string
type CancellationReason string
type PlanType string
type PriceChangeMode string
type PriceChangeState string

const (
	PriceChangeModeUnspecified PriceChangeMode = "PRICE_CHANGE_MODE_UNSPECIFIED"
	PriceChangeModeOptIn       PriceChangeMode = "OPT_IN"
	PriceChangeModeOptOut      PriceChangeMode = "OPT_OUT"
)

const (
	PriceChangeStateUnspecified PriceChangeState = "PRICE_CHANGE_STATE_UNSPECIFIED"
	PriceChangeStatePending     PriceChangeState = "PENDING"
	PriceChangeStateConfirmed   PriceChangeState = "CONFIRMED"
	PriceChangeStateCancelled   PriceChangeState = "CANCELLED"
)

// Enum values

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

const (
	AcknowledgementStateUnspecified  AcknowledgementState = "ACKNOWLEDGEMENT_STATE_UNSPECIFIED"
	AcknowledgementStatePending      AcknowledgementState = "ACKNOWLEDGEMENT_STATE_PENDING"
	AcknowledgementStateAcknowledged AcknowledgementState = "ACKNOWLEDGEMENT_STATE_ACKNOWLEDGED"
)

const (
	PlanTypeAutoRenewing PlanType = "AUTO_RENEWING"
	PlanTypePrepaid      PlanType = "PREPAID"
)
