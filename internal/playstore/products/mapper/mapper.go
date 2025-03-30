package mapper

import (
	"log"
	"subsnotifpro-go/internal/playstore/products/models"

	"google.golang.org/api/androidpublisher/v3"
)

// -------------------------
// 🔹 Convert API Response to Internal Subscription Model
// -------------------------
func ConvertSubscriptionModel(sub *androidpublisher.Subscription, packageName string) *models.ProductSubscription {
	// ✅ Convert Listings
	var listings []models.ProductSubscriptionListing
	if sub.Listings != nil {
		for _, listing := range sub.Listings {
			listings = append(listings, models.ProductSubscriptionListing{
				PackageName:  packageName,
				ProductID:    sub.ProductId,
				LanguageCode: listing.LanguageCode,
				Title:        listing.Title,
				Description:  listing.Description,
				Benefits:     listing.Benefits,
			})
		}
	}

	// ✅ Convert Tax & Compliance Settings (Prevent Nil Panic)
	var taxAndCompliance models.SubscriptionTaxAndComplianceSettings
	if sub.TaxAndComplianceSettings != nil {
		taxAndCompliance = models.SubscriptionTaxAndComplianceSettings{
			PackageName:             packageName,
			ProductID:               sub.ProductId,
			EeaWithdrawalRightType:  models.EeaWithdrawalRightType(sub.TaxAndComplianceSettings.EeaWithdrawalRightType),
			IsTokenizedDigitalAsset: sub.TaxAndComplianceSettings.IsTokenizedDigitalAsset,
		}
	} else {
		taxAndCompliance = models.SubscriptionTaxAndComplianceSettings{
			PackageName: packageName,
			ProductID:   sub.ProductId,
		}
	}

	// ✅ Convert Restricted Payment Countries (Prevent Nil Panic)
	var restrictedPayment models.RestrictedPaymentCountries
	if sub.RestrictedPaymentCountries != nil {
		restrictedPayment = models.RestrictedPaymentCountries{
			PackageName: packageName,
			ProductID:   sub.ProductId,
			RegionCodes: sub.RestrictedPaymentCountries.RegionCodes,
		}
	} else {
		restrictedPayment = models.RestrictedPaymentCountries{
			PackageName: packageName,
			ProductID:   sub.ProductId,
			RegionCodes: []string{}, // ✅ Return empty slice instead of nil
		}
	}

	// ✅ Return Converted Subscription Model
	return &models.ProductSubscription{
		PackageName:                packageName,
		ProductID:                  sub.ProductId,
		Listings:                   listings,
		BasePlans:                  nil, // Placeholder for base plans
		TaxAndComplianceSettings:   taxAndCompliance,
		RestrictedPaymentCountries: restrictedPayment,
	}
}

// -------------------------
// 🔹 Convert API Response to Internal Base Plan Model
// -------------------------

// ConvertBasePlanModel converts a Google Play API BasePlan into an internal ProductBasePlan model.
func ConvertBasePlanModel(basePlan *androidpublisher.BasePlan, packageName, productID string) *models.ProductBasePlan {
	if basePlan == nil || basePlan.BasePlanId == "" {
		return nil // ✅ Prevent panic if basePlan is nil or BasePlanID is empty
	}

	// Determine the BasePlanType dynamically (only one type can be present)
	var basePlanType models.BasePlanTypeEnum
	switch {
	case basePlan.AutoRenewingBasePlanType != nil:
		basePlanType = models.AutoRenewing
	case basePlan.PrepaidBasePlanType != nil:
		basePlanType = models.Prepaid
	case basePlan.InstallmentsBasePlanType != nil:
		basePlanType = models.Installments
	default:
		log.Println("⚠️ WARNING: BasePlanType is missing for Package:", packageName, "Product:", productID, "BasePlan:", basePlan.BasePlanId)
	}

	// Convert Auto-Renewing Base Plan if exists
	var autoRenewingBasePlan *models.AutoRenewingBasePlanType
	if b := basePlan.AutoRenewingBasePlanType; b != nil {
		autoRenewingBasePlan = &models.AutoRenewingBasePlanType{
			ProductID:                           productID,
			BasePlanID:                          basePlan.BasePlanId,
			BillingPeriodDuration:               b.BillingPeriodDuration,
			GracePeriodDuration:                 b.GracePeriodDuration,
			AccountHoldDuration:                 b.AccountHoldDuration,
			ResubscribeState:                    models.ResubscribeState(b.ResubscribeState),
			ProrationMode:                       models.SubscriptionProrationMode(b.ProrationMode),
			LegacyCompatible:                    b.LegacyCompatible,
			LegacyCompatibleSubscriptionOfferID: b.LegacyCompatibleSubscriptionOfferId,
		}
	}

	// Convert Prepaid Base Plan if exists
	var prepaidBasePlan *models.PrepaidBasePlanType
	if b := basePlan.PrepaidBasePlanType; b != nil {
		prepaidBasePlan = &models.PrepaidBasePlanType{
			ProductID:             productID,
			BasePlanID:            basePlan.BasePlanId,
			BillingPeriodDuration: b.BillingPeriodDuration,
			TimeExtension:         models.TimeExtension(b.TimeExtension),
		}
	}

	// Convert Installments Base Plan if exists
	var installmentsBasePlan *models.InstallmentsBasePlanType
	if b := basePlan.InstallmentsBasePlanType; b != nil {
		installmentsBasePlan = &models.InstallmentsBasePlanType{
			ProductID:              productID,
			BasePlanID:             basePlan.BasePlanId,
			BillingPeriodDuration:  b.BillingPeriodDuration,
			CommittedPaymentsCount: int(b.CommittedPaymentsCount), // ✅ Safe default 0
			RenewalType:            models.RenewalType(b.RenewalType),
			GracePeriodDuration:    b.GracePeriodDuration,
			AccountHoldDuration:    b.AccountHoldDuration,
			ResubscribeState:       models.ResubscribeState(b.ResubscribeState),
			ProrationMode:          models.SubscriptionProrationMode(b.ProrationMode),
		}
	}

	// Convert Regional Configs
	var regionalConfigs []models.BasePlanRegionalConfig
	if basePlan.RegionalConfigs != nil {
		for _, region := range basePlan.RegionalConfigs {
			if region.Price != nil { // ✅ Prevent nil pointer dereference
				regionalConfigs = append(regionalConfigs, models.BasePlanRegionalConfig{
					ProductID:                 productID,
					BasePlanID:                basePlan.BasePlanId,
					RegionCode:                region.RegionCode,
					NewSubscriberAvailability: region.NewSubscriberAvailability,
					Price: models.Money{
						CurrencyCode: region.Price.CurrencyCode,
						Units:        region.Price.Units,
						Nanos:        region.Price.Nanos,
					},
				})
			}
		}
	}

	// Convert Offer Tags
	var offerTags []models.OfferTag
	if basePlan.OfferTags != nil {
		for _, tag := range basePlan.OfferTags {
			offerTags = append(offerTags, models.OfferTag{
				ProductID:  productID,
				BasePlanID: basePlan.BasePlanId,
				Tag:        tag.Tag,
			})
		}
	}

	// Convert Other Regions Config
	var otherRegionsConfig models.OtherRegionsBasePlanConfig
	if b := basePlan.OtherRegionsConfig; b != nil {
		otherRegionsConfig = models.OtherRegionsBasePlanConfig{
			ProductID:                 productID,
			BasePlanID:                basePlan.BasePlanId,
			NewSubscriberAvailability: b.NewSubscriberAvailability,
		}
		// ✅ Ensure USD price exists before accessing it
		if b.UsdPrice != nil {
			otherRegionsConfig.USDPrice = models.Money{
				CurrencyCode: b.UsdPrice.CurrencyCode,
				Units:        b.UsdPrice.Units,
				Nanos:        b.UsdPrice.Nanos,
			}
		}
		// ✅ Ensure EUR price exists before accessing it
		if b.EurPrice != nil {
			otherRegionsConfig.EURPrice = models.Money{
				CurrencyCode: b.EurPrice.CurrencyCode,
				Units:        b.EurPrice.Units,
				Nanos:        b.EurPrice.Nanos,
			}
		}
	}

	// ✅ Return final mapped ProductBasePlan
	return &models.ProductBasePlan{
		PackageName:              packageName,
		ProductID:                productID,
		BasePlanID:               basePlan.BasePlanId,
		State:                    models.BasePlanStateEnum(basePlan.State),
		Type:                     basePlanType, // ✅ Dynamically assigned based on available data
		RegionalConfigs:          regionalConfigs,
		OfferTags:                offerTags,
		OtherRegionsConfig:       otherRegionsConfig,
		AutoRenewingBasePlanType: autoRenewingBasePlan,
		PrepaidBasePlanType:      prepaidBasePlan,
		InstallmentsBasePlanType: installmentsBasePlan,
	}
}
