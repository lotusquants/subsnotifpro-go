package models

// -------------------------
// 🔹 ENUMS for Base Plan
// -------------------------

type BasePlanStateEnum string

const (
	StateUnspecified BasePlanStateEnum = "STATE_UNSPECIFIED"
	StateDraft       BasePlanStateEnum = "DRAFT"
	StateActive      BasePlanStateEnum = "ACTIVE"
	StateInactive    BasePlanStateEnum = "INACTIVE"
)

type BasePlanTypeEnum string

const (
	AutoRenewing BasePlanTypeEnum = "auto-renewing"
	Prepaid      BasePlanTypeEnum = "prepaid"
	Installments BasePlanTypeEnum = "installments"
)

// -------------------------
// 🔹 PRODUCT BASE PLAN
// -------------------------

type ProductBasePlan struct {
	PackageName string `gorm:"primaryKey;size:40;not null" json:"package_name"`
	ProductID   string `gorm:"primaryKey;size:50;not null" json:"product_id"`
	BasePlanID  string `gorm:"primaryKey;size:50;not null" json:"base_plan_id"`

	State BasePlanStateEnum `json:"state"`
	Type  BasePlanTypeEnum  `json:"type"`

	// ✅ Relationships with Full Cascade on Delete
	RegionalConfigs    []BasePlanRegionalConfig   `gorm:"foreignKey:PackageName,ProductID,BasePlanID;references:PackageName,ProductID,BasePlanID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"regional_configs"`
	OfferTags          []OfferTag                 `gorm:"foreignKey:PackageName,ProductID,BasePlanID;references:PackageName,ProductID,BasePlanID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"offer_tags"`
	OtherRegionsConfig OtherRegionsBasePlanConfig `gorm:"foreignKey:PackageName,ProductID,BasePlanID;references:PackageName,ProductID,BasePlanID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"other_regions_config"`

	AutoRenewingBasePlanType *AutoRenewingBasePlanType `gorm:"foreignKey:PackageName,ProductID,BasePlanID;references:PackageName,ProductID,BasePlanID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"auto_renewing_base_plan"`
	PrepaidBasePlanType      *PrepaidBasePlanType      `gorm:"foreignKey:PackageName,ProductID,BasePlanID;references:PackageName,ProductID,BasePlanID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"prepaid_base_plan"`
	InstallmentsBasePlanType *InstallmentsBasePlanType `gorm:"foreignKey:PackageName,ProductID,BasePlanID;references:PackageName,ProductID,BasePlanID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"installments_base_plan"`
}

// -------------------------
// 🔹 BASE PLAN REGIONAL CONFIG
// -------------------------

type BasePlanRegionalConfig struct {
	PackageName               string `gorm:"primaryKey;size:40;not null" json:"package_name"`
	ProductID                 string `gorm:"primaryKey;size:50;not null" json:"product_id"`
	BasePlanID                string `gorm:"primaryKey;size:50;not null" json:"base_plan_id"`
	RegionCode                string `gorm:"primaryKey;size:3;not null;index" json:"region_code"`
	NewSubscriberAvailability bool   `json:"new_subscriber_availability"`
	Price                     Money  `gorm:"type:jsonb" json:"price"`
}

// -------------------------
// 🔹 OFFER TAG
// -------------------------

type OfferTag struct {
	PackageName string `gorm:"primaryKey;size:40;not null" json:"package_name"`
	ProductID   string `gorm:"primaryKey;size:50;not null" json:"product_id"`
	BasePlanID  string `gorm:"primaryKey;size:50;not null" json:"base_plan_id"`
	Tag         string `gorm:"size:20" json:"tag"`
}

// -------------------------
// 🔹 OTHER REGIONS BASE PLAN CONFIG
// -------------------------

type OtherRegionsBasePlanConfig struct {
	PackageName string `gorm:"primaryKey;size:40;not null" json:"package_name"`
	ProductID   string `gorm:"primaryKey;size:50;not null" json:"product_id"`
	BasePlanID  string `gorm:"primaryKey;size:50;not null" json:"base_plan_id"`

	// ✅ Use Embedded Struct After Implementing Valuer/Scanner
	USDPrice                  Money `gorm:"type:jsonb" json:"usdPrice"`
	EURPrice                  Money `gorm:"type:jsonb" json:"eurPrice"`
	NewSubscriberAvailability bool  `json:"new_subscriber_availability"`
}

// -------------------------
// 🔹 AUTO-RENEWING BASE PLAN
// -------------------------

type AutoRenewingBasePlanType struct {
	PackageName                         string                    `gorm:"primaryKey;size:40;not null" json:"package_name"`
	ProductID                           string                    `gorm:"primaryKey;size:50;not null" json:"product_id"`
	BasePlanID                          string                    `gorm:"primaryKey;size:50;not null" json:"base_plan_id"`
	BillingPeriodDuration               string                    `json:"billing_period_duration"`
	GracePeriodDuration                 string                    `json:"grace_period_duration"`
	AccountHoldDuration                 string                    `json:"account_hold_duration"`
	ResubscribeState                    ResubscribeState          `json:"resubscribe_state"`
	ProrationMode                       SubscriptionProrationMode `json:"proration_mode"`
	LegacyCompatible                    bool                      `json:"legacy_compatible"`
	LegacyCompatibleSubscriptionOfferID string                    `gorm:"index" json:"legacy_compatible_subscription_offer_id"`
}

// -------------------------
// 🔹 ENUMS for AUTO-RENEWING
// -------------------------

type ResubscribeState string

const (
	ResubscribeStateUnspecified ResubscribeState = "RESUBSCRIBE_STATE_UNSPECIFIED"
	ResubscribeStateActive      ResubscribeState = "RESUBSCRIBE_STATE_ACTIVE"
	ResubscribeStateInactive    ResubscribeState = "RESUBSCRIBE_STATE_INACTIVE"
)

type SubscriptionProrationMode string

const (
	SubscriptionProrationModeUnspecified           SubscriptionProrationMode = "SUBSCRIPTION_PRORATION_MODE_UNSPECIFIED"
	SubscriptionProrationModeChargeNextBillingDate SubscriptionProrationMode = "SUBSCRIPTION_PRORATION_MODE_CHARGE_ON_NEXT_BILLING_DATE"
	SubscriptionProrationModeChargeFullPriceNow    SubscriptionProrationMode = "SUBSCRIPTION_PRORATION_MODE_CHARGE_FULL_PRICE_IMMEDIATELY"
)

// -------------------------
// 🔹 PREPAID BASE PLAN
// -------------------------

type PrepaidBasePlanType struct {
	PackageName           string        `gorm:"primaryKey;size:40;not null" json:"package_name"`
	ProductID             string        `gorm:"primaryKey;size:50;not null" json:"product_id"`
	BasePlanID            string        `gorm:"primaryKey;size:50;not null" json:"base_plan_id"`
	BillingPeriodDuration string        `json:"billing_period_duration"`
	TimeExtension         TimeExtension `json:"time_extension"`
}

// -------------------------
// 🔹 ENUMS for PREPAID BASE PLAN
// -------------------------

type TimeExtension string

const (
	TimeExtensionUnspecified TimeExtension = "TIME_EXTENSION_UNSPECIFIED"
	TimeExtensionActive      TimeExtension = "TIME_EXTENSION_ACTIVE"
	TimeExtensionInactive    TimeExtension = "TIME_EXTENSION_INACTIVE"
)

// -------------------------
// 🔹 INSTALLMENTS BASE PLAN
// -------------------------

type InstallmentsBasePlanType struct {
	PackageName            string                    `gorm:"primaryKey;size:40;not null" json:"package_name"`
	ProductID              string                    `gorm:"primaryKey;size:50;not null" json:"product_id"`
	BasePlanID             string                    `gorm:"primaryKey;size:50;not null" json:"base_plan_id"`
	BillingPeriodDuration  string                    `json:"billing_period_duration"`
	CommittedPaymentsCount int                       `json:"committed_payments_count"`
	RenewalType            RenewalType               `json:"renewal_type"`
	GracePeriodDuration    string                    `json:"grace_period_duration"`
	AccountHoldDuration    string                    `json:"account_hold_duration"`
	ResubscribeState       ResubscribeState          `json:"resubscribe_state"`
	ProrationMode          SubscriptionProrationMode `json:"proration_mode"`
}

// -------------------------
// 🔹 ENUMS for INSTALLMENTS BASE PLAN
// -------------------------

type RenewalType string

const (
	RenewalTypeUnspecified             RenewalType = "RENEWAL_TYPE_UNSPECIFIED"
	RenewalTypeRenewsWithoutCommitment RenewalType = "RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT"
	RenewalTypeRenewsWithCommitment    RenewalType = "RENEWAL_TYPE_RENEWS_WITH_COMMITMENT"
)
