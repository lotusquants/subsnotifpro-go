package models

// -------------------------
// ✅ ENUMS for Better Type Safety
// -------------------------

// SubscriptionOfferState Enum
type SubscriptionOfferState string

const (
	OfferStateUnspecified SubscriptionOfferState = "STATE_UNSPECIFIED" // Default value, should not be used.
	OfferStateDraft       SubscriptionOfferState = "DRAFT"             // Offer is not available to users.
	OfferStateActive      SubscriptionOfferState = "ACTIVE"            // Offer is available.
	OfferStateInactive    SubscriptionOfferState = "INACTIVE"          // Offer is unavailable to new users, but existing users retain access.
)

// -------------------------
// ✅ Subscription Offer Model
// -------------------------

type SubscriptionOffer struct {
	PackageName string                 `gorm:"primaryKey;size:40;not null" json:"packageName"`
	ProductID   string                 `gorm:"primaryKey;size:50;not null" json:"productId"`
	BasePlanID  string                 `gorm:"primaryKey;size:50;not null" json:"basePlanId"`
	OfferID     string                 `gorm:"primaryKey;size:50;not null" json:"offerId"`
	State       SubscriptionOfferState `gorm:"type:varchar(20);not null;check:state IN ('STATE_UNSPECIFIED', 'DRAFT', 'ACTIVE', 'INACTIVE')" json:"state"`

	// ✅ Relations with Full Cascade (Delete + Update)
	Phases             []SubscriptionOfferPhase             `gorm:"foreignKey:PackageName,ProductID,BasePlanID,OfferID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"phases"`
	Targeting          *SubscriptionOfferTargeting          `gorm:"foreignKey:PackageName,ProductID,BasePlanID,OfferID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"targeting,omitempty"`
	RegionalConfigs    []RegionalSubscriptionOfferConfig    `gorm:"foreignKey:PackageName,ProductID,BasePlanID,OfferID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"regionalConfigs"`
	OtherRegionsConfig *OtherRegionsSubscriptionOfferConfig `gorm:"foreignKey:PackageName,ProductID,BasePlanID,OfferID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"otherRegionsConfig,omitempty"`
}

// -------------------------
// ✅ Regional Offer Configuration
// -------------------------

type RegionalSubscriptionOfferConfig struct {
	PackageName               string `gorm:"primaryKey;size:40;not null;index:idx_package_name" json:"packageName"`
	ProductID                 string `gorm:"primaryKey;size:50;not null;index:idx_product_id" json:"productId"`
	BasePlanID                string `gorm:"primaryKey;size:50;not null;index:idx_base_plan_id" json:"basePlanId"`
	OfferID                   string `gorm:"primaryKey;size:50;not null;index:idx_offer_id" json:"offerId"`
	RegionCode                string `gorm:"primaryKey;size:3;not null;index:idx_region_code" json:"regionCode"`
	NewSubscriberAvailability bool   `gorm:"not null;default:false" json:"newSubscriberAvailability"`
}

// -------------------------
// ✅ Other Regions Offer Configuration
// -------------------------

type OtherRegionsSubscriptionOfferConfig struct {
	PackageName                           string `gorm:"primaryKey;size:40;not null;index:idx_package_name" json:"packageName"`
	ProductID                             string `gorm:"primaryKey;size:50;not null;index:idx_product_id" json:"productId"`
	BasePlanID                            string `gorm:"primaryKey;size:50;not null;index:idx_base_plan_id" json:"basePlanId"`
	OfferID                               string `gorm:"primaryKey;size:50;not null;index:idx_offer_id" json:"offerId"`
	OtherRegionsNewSubscriberAvailability bool   `gorm:"not null;default:false" json:"otherRegionsNewSubscriberAvailability"`
}

// -------------------------
// ✅ Subscription Offer Phase
// -------------------------

type SubscriptionOfferPhase struct {
	PackageName string `gorm:"primaryKey;size:40;not null" json:"packageName"`
	ProductID   string `gorm:"primaryKey;size:50;not null" json:"productId"`
	BasePlanID  string `gorm:"primaryKey;size:50;not null" json:"basePlanId"`
	OfferID     string `gorm:"primaryKey;size:50;not null" json:"offerId"`

	RecurrenceCount int `gorm:"not null;check:recurrence_count > 0" json:"recurrenceCount"`

	Duration string `gorm:"size:20;not null" json:"duration"` // ISO 8601 Format (e.g., P1M for 1 month)

	// ✅ Relations with Full Cascade (Delete + Update)
	RegionalConfigs    []RegionalSubscriptionOfferPhaseConfig    `gorm:"foreignKey:PackageName,ProductID,BasePlanID,OfferID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"regionalConfigs"`
	OtherRegionsConfig *OtherRegionsSubscriptionOfferPhaseConfig `gorm:"foreignKey:PackageName,ProductID,BasePlanID,OfferID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"otherRegionsConfig,omitempty"`
}

// -------------------------
// ✅ Regional Subscription Offer Phase Config
// -------------------------

type RegionalSubscriptionOfferPhaseConfig struct {
	PackageName string `gorm:"primaryKey;size:40;not null" json:"packageName"`
	ProductID   string `gorm:"primaryKey;size:50;not null" json:"productId"`
	BasePlanID  string `gorm:"primaryKey;size:50;not null" json:"basePlanId"`
	OfferID     string `gorm:"primaryKey;size:50;not null" json:"offerId"`

	RegionCode string `gorm:"primaryKey;size:3;not null" json:"regionCode"` // Standardized region codes (ISO)

	// ✅ Pricing Override Fields (Only one can be set)
	Price            *Money   `gorm:"embedded;default:null" json:"price,omitempty"`
	RelativeDiscount *float64 `gorm:"default:null" json:"relativeDiscount,omitempty"`
	AbsoluteDiscount *Money   `gorm:"embedded;default:null" json:"absoluteDiscount,omitempty"`
	Free             bool     `gorm:"not null;default:false" json:"free"`
}

// -------------------------
// ✅ Other Regions Offer Phase Config (Matches Play API)
// -------------------------

type OtherRegionsSubscriptionOfferPhaseConfig struct {
	PackageName string `gorm:"primaryKey;size:40;not null" json:"packageName"`
	ProductID   string `gorm:"primaryKey;size:50;not null" json:"productId"`
	BasePlanID  string `gorm:"primaryKey;size:50;not null" json:"basePlanId"`
	OfferID     string `gorm:"primaryKey;size:50;not null" json:"offerId"`

	// ✅ Pricing Overrides
	OtherRegionsPrices *OtherRegionsSubscriptionOfferPhasePrices `gorm:"foreignKey:PackageName,ProductID,BasePlanID,OfferID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"otherRegionsPrices,omitempty"`

	RelativeDiscount *float64 `gorm:"default:null" json:"relativeDiscount,omitempty"`

	AbsoluteDiscounts *OtherRegionsSubscriptionOfferPhasePrices `gorm:"foreignKey:PackageName,ProductID,BasePlanID,OfferID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"absoluteDiscounts,omitempty"`

	Free bool `gorm:"not null;default:false" json:"free"`
}

// -------------------------
// ✅ Other Regions Subscription Offer Phase Prices
// -------------------------

type OtherRegionsSubscriptionOfferPhasePrices struct {
	PackageName string `gorm:"primaryKey;size:40;not null" json:"packageName"`
	ProductID   string `gorm:"primaryKey;size:50;not null" json:"productId"`
	BasePlanID  string `gorm:"primaryKey;size:50;not null" json:"basePlanId"`
	OfferID     string `gorm:"primaryKey;size:50;not null" json:"offerId"`

	// ✅ Pricing Fields
	USDPrice Money `gorm:"embedded" json:"usdPrice"`
	EURPrice Money `gorm:"embedded" json:"eurPrice"`
}

// -------------------------
// ✅ Subscription Offer Targeting
// -------------------------

type SubscriptionOfferTargeting struct {
	PackageName string `gorm:"primaryKey;size:40;not null;index:idx_package_name" json:"packageName"`
	ProductID   string `gorm:"primaryKey;size:50;not null;index:idx_product_id" json:"productId"`
	BasePlanID  string `gorm:"primaryKey;size:50;not null;index:idx_base_plan_id" json:"basePlanId"`
	OfferID     string `gorm:"primaryKey;size:50;not null;index:idx_offer_id" json:"offerId"`

	// 🎯 Union Field: ONLY ONE CAN EXIST (Acquisition OR Upgrade)
	AcquisitionRule *AcquisitionTargetingRule `gorm:"foreignKey:PackageName,ProductID,BasePlanID,OfferID;references:PackageName,ProductID,BasePlanID,OfferID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"acquisitionRule,omitempty"`
	UpgradeRule     *UpgradeTargetingRule     `gorm:"foreignKey:PackageName,ProductID,BasePlanID,OfferID;references:PackageName,ProductID,BasePlanID,OfferID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"upgradeRule,omitempty"`
}

// -------------------------
// ✅ Acquisition Targeting Rule
// -------------------------

type AcquisitionTargetingRule struct {
	PackageName string `gorm:"primaryKey;size:40;not null;index:idx_package_name" json:"packageName"`
	ProductID   string `gorm:"primaryKey;size:50;not null;index:idx_product_id" json:"productId"`
	BasePlanID  string `gorm:"primaryKey;size:50;not null;index:idx_base_plan_id" json:"basePlanId"`
	OfferID     string `gorm:"primaryKey;size:50;not null;index:idx_offer_id" json:"offerId"`

	// 🎯 Rule Scope (Defines the eligibility criteria)
	Scope TargetingRuleScope `gorm:"embedded" json:"scope"`
}

// -------------------------
// ✅ Upgrade Targeting Rule
// -------------------------

type UpgradeTargetingRule struct {
	PackageName string `gorm:"primaryKey;size:40;not null;index:idx_package_name" json:"packageName"`
	ProductID   string `gorm:"primaryKey;size:50;not null;index:idx_product_id" json:"productId"`
	BasePlanID  string `gorm:"primaryKey;size:50;not null;index:idx_base_plan_id" json:"basePlanId"`
	OfferID     string `gorm:"primaryKey;size:50;not null;index:idx_offer_id" json:"offerId"`

	BillingPeriodDuration string `gorm:"size:20;not null" json:"billingPeriodDuration"` // ISO 8601 format (e.g., P1M for 1 month)
	OncePerUser           bool   `gorm:"not null;default:false" json:"oncePerUser"`     // Restrict to a one-time upgrade per user

	// 🎯 Rule Scope (Defines the eligibility criteria)
	Scope TargetingRuleScope `gorm:"embedded" json:"scope"`
}

// -------------------------
// ✅ Targeting Rule Scope (Union Field)
// -------------------------

type TargetingRuleScope struct {
	ThisSubscription          *bool   `gorm:"default:null" json:"thisSubscription,omitempty"`                  // Targets only the current subscription
	AnySubscriptionInApp      *bool   `gorm:"default:null" json:"anySubscriptionInApp,omitempty"`              // Targets any subscription in the app
	SpecificSubscriptionInApp *string `gorm:"default:null;size:50" json:"specificSubscriptionInApp,omitempty"` // Specific Subscription ID
}
