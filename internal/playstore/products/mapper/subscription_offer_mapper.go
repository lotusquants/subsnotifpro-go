package mapper

import (
	"log"
	"subsnotifpro-go/internal/playstore/products/models"

	"google.golang.org/api/androidpublisher/v3"
)

// ConvertSubscriptionOfferModel converts a Google Play API SubscriptionOffer into an internal SubscriptionOffer model.
func ConvertSubscriptionOfferModel(offer *androidpublisher.SubscriptionOffer, packageName string, productID string, basePlanID string) models.SubscriptionOffer {
	if offer == nil || offer.OfferId == "" {
		log.Printf("⚠️ WARNING: Invalid Offer for Package: %s, Product: %s, BasePlan: %s", packageName, productID, basePlanID)
		return models.SubscriptionOffer{}
	}

	// ✅ Convert Phases (Optimized slice allocation)
	phases := make([]models.SubscriptionOfferPhase, 0, len(offer.Phases))
	for phaseIndex, phase := range offer.Phases {
		phases = append(phases, convertSubscriptionOfferPhase(phase, packageName, productID, basePlanID, offer.OfferId, phaseIndex))
	}

	// ✅ Convert Targeting
	targeting := convertSubscriptionOfferTargeting(offer.Targeting, packageName, productID, basePlanID, offer.OfferId)

	// ✅ Convert Regional Subscription Offer Config
	regionalConfigs := convertRegionalSubscriptionOfferConfig(offer.RegionalConfigs, packageName, productID, basePlanID, offer.OfferId)

	// ✅ Convert Other Regions Subscription Offer Config
	otherRegionsConfig := convertOtherRegionsSubscriptionOfferConfig(offer.OtherRegionsConfig, packageName, productID, basePlanID, offer.OfferId)

	// ✅ Return the mapped SubscriptionOffer model
	return models.SubscriptionOffer{
		PackageName:        packageName,
		ProductID:          productID,
		BasePlanID:         basePlanID,
		OfferID:            offer.OfferId,
		State:              models.SubscriptionOfferState(offer.State),
		Phases:             phases,
		Targeting:          targeting,
		RegionalConfigs:    regionalConfigs,
		OtherRegionsConfig: otherRegionsConfig,
	}
}

// ✅ Convert Regional Subscription Offer Config
func convertRegionalSubscriptionOfferConfig(regionalConfigs []*androidpublisher.RegionalSubscriptionOfferConfig, packageName string, productID string, basePlanID string, offerID string) []models.RegionalSubscriptionOfferConfig {
	if regionalConfigs == nil {
		return nil
	}

	convertedConfigs := make([]models.RegionalSubscriptionOfferConfig, 0, len(regionalConfigs))
	for _, region := range regionalConfigs {
		convertedConfigs = append(convertedConfigs, models.RegionalSubscriptionOfferConfig{
			PackageName:               packageName,
			ProductID:                 productID,
			BasePlanID:                basePlanID,
			OfferID:                   offerID,
			RegionCode:                region.RegionCode,
			NewSubscriberAvailability: region.NewSubscriberAvailability,
		})
	}
	return convertedConfigs
}

// ✅ Convert Other Regions Subscription Offer Config
func convertOtherRegionsSubscriptionOfferConfig(config *androidpublisher.OtherRegionsSubscriptionOfferConfig, packageName, productID, basePlanID, offerID string) *models.OtherRegionsSubscriptionOfferConfig {
	if config == nil {
		return nil
	}

	return &models.OtherRegionsSubscriptionOfferConfig{
		PackageName:                           packageName,
		ProductID:                             productID,
		BasePlanID:                            basePlanID,
		OfferID:                               offerID,
		OtherRegionsNewSubscriberAvailability: config.OtherRegionsNewSubscriberAvailability,
	}
}

// Convert Subscription Offer Phase
func convertSubscriptionOfferPhase(phase *androidpublisher.SubscriptionOfferPhase, packageName, productID, basePlanID, offerID string, phaseIndex int) models.SubscriptionOfferPhase {
	if phase == nil {
		log.Printf("⚠️ WARNING: Null Phase for Offer %s", offerID)
		return models.SubscriptionOfferPhase{}
	}

	// ✅ Convert Regional Configs
	regionalConfigs := convertRegionalSubscriptionOfferPhaseConfig(phase.RegionalConfigs, packageName, productID, basePlanID, offerID, phaseIndex)

	// ✅ Convert Other Regions Config
	otherRegionsConfig := convertOtherRegionsSubscriptionOfferPhaseConfig(phase.OtherRegionsConfig, packageName, productID, basePlanID, offerID, phaseIndex)

	// ✅ Return mapped phase model
	return models.SubscriptionOfferPhase{
		PackageName:        packageName,
		ProductID:          productID,
		BasePlanID:         basePlanID,
		OfferID:            offerID,
		PhaseIndex:         phaseIndex,
		RecurrenceCount:    int(phase.RecurrenceCount),
		Duration:           phase.Duration,
		RegionalConfigs:    regionalConfigs,
		OtherRegionsConfig: otherRegionsConfig,
	}
}

// ✅ Convert Regional Subscription Offer Phase Config
func convertRegionalSubscriptionOfferPhaseConfig(regionalConfigs []*androidpublisher.RegionalSubscriptionOfferPhaseConfig, packageName string, productID string, basePlanID string, offerID string, phaseIndex int) []models.RegionalSubscriptionOfferPhaseConfig {
	if regionalConfigs == nil {
		return nil
	}

	convertedConfigs := make([]models.RegionalSubscriptionOfferPhaseConfig, 0, len(regionalConfigs))
	for _, region := range regionalConfigs {
		if region == nil {
			continue
		}

		convertedConfigs = append(convertedConfigs, models.RegionalSubscriptionOfferPhaseConfig{
			PackageName:      packageName,
			ProductID:        productID,
			BasePlanID:       basePlanID,
			OfferID:          offerID,
			PhaseIndex:       phaseIndex,
			RegionCode:       region.RegionCode,
			Price:            convertMoney(region.Price),
			RelativeDiscount: safeRelativeDiscount(region.RelativeDiscount),
			AbsoluteDiscount: convertMoney(region.AbsoluteDiscount),
			Free:             region.Free != nil,
		})
	}
	return convertedConfigs
}

// ✅ Convert Other Regions Subscription Offer Phase Config
func convertOtherRegionsSubscriptionOfferPhaseConfig(config *androidpublisher.OtherRegionsSubscriptionOfferPhaseConfig, packageName string, productID string, basePlanID string, offerID string, phaseIndex int) *models.OtherRegionsSubscriptionOfferPhaseConfig {
	if config == nil {
		return nil
	}

	return &models.OtherRegionsSubscriptionOfferPhaseConfig{
		PackageName:        packageName,
		ProductID:          productID,
		BasePlanID:         basePlanID,
		OfferID:            offerID,
		PhaseIndex:         phaseIndex,
		RelativeDiscount:   safeRelativeDiscount(config.RelativeDiscount),
		Free:               config.Free != nil,
		OtherRegionsPrices: convertOtherRegionsPrices(config.OtherRegionsPrices, packageName, productID, basePlanID, offerID, phaseIndex),
		AbsoluteDiscounts:  convertOtherRegionsPrices(config.AbsoluteDiscounts, packageName, productID, basePlanID, offerID, phaseIndex),
	}
}

// ✅ Convert Other Regions Prices
func convertOtherRegionsPrices(prices *androidpublisher.OtherRegionsSubscriptionOfferPhasePrices, packageName string, productID string, basePlanID string, offerID string, phaseIndex int) *models.OtherRegionsSubscriptionOfferPhasePrices {
	if prices == nil {
		return nil
	}

	convertedPrices := &models.OtherRegionsSubscriptionOfferPhasePrices{
		PackageName: packageName,
		ProductID:   productID,
		BasePlanID:  basePlanID,
		OfferID:     offerID,
		PhaseIndex:  phaseIndex,
	}

	// ✅ Map USD Price if available
	if prices.UsdPrice != nil {
		convertedPrices.USDPrice = *convertMoney(prices.UsdPrice)
	}

	// ✅ Map EUR Price if available
	if prices.EurPrice != nil {
		convertedPrices.EURPrice = *convertMoney(prices.EurPrice)
	}

	return convertedPrices
}

// ✅ Convert Money (Handles nil safely)
func convertMoney(money *androidpublisher.Money) *models.Money {
	if money == nil {
		return nil
	}

	return &models.Money{
		CurrencyCode: money.CurrencyCode,
		Units:        money.Units,
		Nanos:        money.Nanos,
	}
}

// ✅ Safe Relative Discount (Avoids nil pointer errors)
func safeRelativeDiscount(discount float64) *float64 {
	if discount > 0 {
		return &discount
	}
	return nil
}

// ✅ Convert Subscription Offer Targeting (Handles nil cases safely)
func convertSubscriptionOfferTargeting(targeting *androidpublisher.SubscriptionOfferTargeting, packageName, productID, basePlanID, offerID string) *models.SubscriptionOfferTargeting {
	if targeting == nil {
		return nil
	}

	return &models.SubscriptionOfferTargeting{
		PackageName:     packageName,
		ProductID:       productID,
		BasePlanID:      basePlanID,
		OfferID:         offerID,
		AcquisitionRule: convertAcquisitionTargetingRule(targeting.AcquisitionRule, packageName, productID, basePlanID, offerID),
		UpgradeRule:     convertUpgradeTargetingRule(targeting.UpgradeRule, packageName, productID, basePlanID, offerID),
	}
}

// ✅ Convert Acquisition Targeting Rule
func convertAcquisitionTargetingRule(rule *androidpublisher.AcquisitionTargetingRule, packageName, productID, basePlanID, offerID string) *models.AcquisitionTargetingRule {
	if rule == nil {
		return nil
	}

	return &models.AcquisitionTargetingRule{
		PackageName: packageName,
		ProductID:   productID,
		BasePlanID:  basePlanID,
		OfferID:     offerID,
		Scope:       convertTargetingRuleScope(rule.Scope),
	}
}

// ✅ Convert Upgrade Targeting Rule
func convertUpgradeTargetingRule(rule *androidpublisher.UpgradeTargetingRule, packageName, productID, basePlanID, offerID string) *models.UpgradeTargetingRule {
	if rule == nil {
		return nil
	}

	return &models.UpgradeTargetingRule{
		PackageName:           packageName,
		ProductID:             productID,
		BasePlanID:            basePlanID,
		OfferID:               offerID,
		BillingPeriodDuration: rule.BillingPeriodDuration,
		OncePerUser:           rule.OncePerUser,
		Scope:                 convertTargetingRuleScope(rule.Scope),
	}
}

// ✅ Convert Targeting Rule Scope (Handles nil safely)
func convertTargetingRuleScope(scope *androidpublisher.TargetingRuleScope) models.TargetingRuleScope {
	if scope == nil {
		return models.TargetingRuleScope{}
	}

	return models.TargetingRuleScope{
		ThisSubscription:          boolPtr(scope.ThisSubscription != nil),
		AnySubscriptionInApp:      boolPtr(scope.AnySubscriptionInApp != nil),
		SpecificSubscriptionInApp: stringPtr(scope.SpecificSubscriptionInApp),
	}
}

// ✅ Safe Pointer Helpers
func boolPtr(value bool) *bool {
	if !value {
		return nil
	}
	return &value
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
