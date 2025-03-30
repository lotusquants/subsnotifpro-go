package models

import (
	orgModels "subsnotifpro-go/internal/organization/models"
	playstoreSubscriptionCatalogModels "subsnotifpro-go/internal/playstore/products/models"
	playstoreRtdnModels "subsnotifpro-go/internal/playstore/rtdn/models"
	playstoreSettingsModels "subsnotifpro-go/internal/playstore/settings/models"
	playstoreSubscriptionModels "subsnotifpro-go/internal/playstore/subscription/models"
	playstoreUserModels "subsnotifpro-go/internal/playstore/user/models"
	usersModels "subsnotifpro-go/internal/users/models"
)

// Collect all models in a list for AutoMigrate
var AllModels = []interface{}{
	&TestModel{},

	&orgModels.Organization{},

	// User Tables
	&usersModels.AppUser{},
	&usersModels.AppUserPlatformChange{},
	// playstore user
	&playstoreUserModels.GoogleAccount{},

	// Google playstore settings tables
	&playstoreSettingsModels.GooglePlayServiceAccount{},
	&playstoreSettingsModels.GooglePlaySettings{},

	// Google playstore webhook notification logs table
	&playstoreRtdnModels.GooglePlayWebhookEvent{},

	// Core Google playstore Subscription Models
	&playstoreSubscriptionModels.AcknowledgementStateModel{},
	&playstoreSubscriptionModels.AcknowledgementStateTransitionHistory{},
	&playstoreSubscriptionModels.AutoRenewingPlan{},
	&playstoreSubscriptionModels.AutoRenewingPlanHistory{},
	&playstoreSubscriptionModels.SubscriptionCancellationContext{},
	&playstoreSubscriptionModels.SubscriptionCancellationContextHistory{},
	&playstoreSubscriptionModels.SubscriptionChangeEvent{},
	&playstoreSubscriptionModels.DeferredItemReplacement{},
	&playstoreSubscriptionModels.DeferredItemReplacementHistory{},
	&playstoreSubscriptionModels.InstallmentPlan{},
	&playstoreSubscriptionModels.InstallmentPlanHistory{},
	&playstoreSubscriptionModels.SubscriptionLineItem{},
	&playstoreSubscriptionModels.SubscriptionLineItemHistory{},
	&playstoreSubscriptionModels.OfferDetails{},
	&playstoreSubscriptionModels.OfferDetailsHistory{},
	&playstoreSubscriptionModels.SubscriptionPausedContext{},
	&playstoreSubscriptionModels.SubscriptionPausedContextHistory{},
	&playstoreSubscriptionModels.PrepaidPlan{},
	&playstoreSubscriptionModels.PrepaidPlanHistory{},
	&playstoreSubscriptionModels.SubscriptionItemPriceChangeDetails{},
	&playstoreSubscriptionModels.SubscriptionItemPriceChangeDetailsHistory{},
	&playstoreSubscriptionModels.RegionCode{},
	&playstoreSubscriptionModels.SignupPromotion{},
	&playstoreSubscriptionModels.SignupPromotionHistory{},
	&playstoreSubscriptionModels.SubscriptionPurchaseV2{},
	&playstoreSubscriptionModels.SubscriptionOrderIdTransitionHistory{},
	&playstoreSubscriptionModels.SubscriptionStateModel{},
	&playstoreSubscriptionModels.SubscriptionStateTransitionHistory{},

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
