package models

import (
	playstoreRtdnModels "subsnotifpro-go/internal/google_playstore/rtdn/models"
	playstoreSettingsModels "subsnotifpro-go/internal/google_playstore/settings/models"
	playstoreSubscriptionCatalogModels "subsnotifpro-go/internal/google_playstore/subscription_catalog/models"
)

// Collect all models in a list for AutoMigrate
var AllModels = []interface{}{
	&TestModel{},

	// Core Centralized User Tables
	&AppUser{},
	&AppUserPlatformChange{},

	// Google playstore settings tables
	&playstoreSettingsModels.GooglePlayServiceAccount{},
	&playstoreSettingsModels.GooglePlaySettings{},

	// Google playstore webhook notification logs table
	&playstoreRtdnModels.GooglePlayWebhookEvent{},

	// Core Google playstore Subscription Models
	&playstoreRtdnModels.SubscriptionPurchaseV2{},
	&playstoreRtdnModels.SubscriptionStateModel{},
	&playstoreRtdnModels.SubscriptionStateTransitionHistory{},
	&playstoreRtdnModels.AcknowledgementStateModel{},
	&playstoreRtdnModels.AcknowledgementStateTransitionHistory{},
	&playstoreRtdnModels.SubscriptionPausedContext{},
	&playstoreRtdnModels.SubscriptionPausedTransitionHistory{},
	&playstoreRtdnModels.SubscriptionCancellationContext{},
	&playstoreRtdnModels.SubscriptionCancellationHistory{},

	// Google playstore Plan Models (AutoRenewing & Prepaid)
	&playstoreRtdnModels.AutoRenewingPlan{},
	&playstoreRtdnModels.AutoRenewingPlanHistory{},
	&playstoreRtdnModels.PrepaidPlan{},
	&playstoreRtdnModels.PrepaidPlanHistory{},
	&playstoreRtdnModels.InstallmentPlan{},
	&playstoreRtdnModels.InstallmentPlanHistory{},
	&playstoreRtdnModels.SubscriptionItemPriceChangeDetails{},

	// Google playstore Offer, Deferred Replacement & Promotions
	&playstoreRtdnModels.OfferDetails{},
	&playstoreRtdnModels.OfferDetailsHistory{},
	&playstoreRtdnModels.DeferredItemReplacement{},
	&playstoreRtdnModels.DeferredItemReplacementHistory{},
	&playstoreRtdnModels.SignupPromotion{},
	&playstoreRtdnModels.SignupPromotionHistory{},

	// Google playstore Other Supporting Tables
	&playstoreRtdnModels.RegionCode{},
	&playstoreRtdnModels.GoogleAccount{},

	// 🔹 Subscription Sync Models
	&playstoreSubscriptionCatalogModels.ProductSubscription{},
	&playstoreSubscriptionCatalogModels.ProductSubscriptionListing{},
	&playstoreSubscriptionCatalogModels.RestrictedPaymentCountries{},
	&playstoreSubscriptionCatalogModels.SubscriptionTaxAndComplianceSettings{},
	&playstoreSubscriptionCatalogModels.RegionalTaxRateInfo{},

	// 🔹 Base Plan Models
	&playstoreSubscriptionCatalogModels.ProductBasePlan{},
	&playstoreSubscriptionCatalogModels.BasePlanRegionalConfig{},
	&playstoreSubscriptionCatalogModels.OtherRegionsBasePlanConfig{},

	// 🔹 Subscription Offer Models
	&playstoreSubscriptionCatalogModels.SubscriptionOffer{},
	&playstoreSubscriptionCatalogModels.RegionalSubscriptionOfferConfig{},
	&playstoreSubscriptionCatalogModels.OtherRegionsSubscriptionOfferConfig{},
	&playstoreSubscriptionCatalogModels.SubscriptionOfferPhase{},
	&playstoreSubscriptionCatalogModels.RegionalSubscriptionOfferPhaseConfig{},
	&playstoreSubscriptionCatalogModels.OtherRegionsSubscriptionOfferPhaseConfig{},
	&playstoreSubscriptionCatalogModels.OtherRegionsSubscriptionOfferPhasePrices{},

	// 🔹 Subscription Offer Targeting
	&playstoreSubscriptionCatalogModels.SubscriptionOfferTargeting{},
	&playstoreSubscriptionCatalogModels.AcquisitionTargetingRule{},
	&playstoreSubscriptionCatalogModels.UpgradeTargetingRule{},
	&playstoreSubscriptionCatalogModels.TargetingRuleScope{},

	// 🔹 Auto-Renewing Plans
	&playstoreSubscriptionCatalogModels.AutoRenewingBasePlanType{},

	// 🔹 Prepaid Plans
	&playstoreSubscriptionCatalogModels.PrepaidBasePlanType{},

	// 🔹 Installments Plans
	&playstoreSubscriptionCatalogModels.InstallmentsBasePlanType{},

	// 🔹 Shared Models
	&playstoreSubscriptionCatalogModels.Money{},
	&playstoreSubscriptionCatalogModels.OfferTag{},
}
