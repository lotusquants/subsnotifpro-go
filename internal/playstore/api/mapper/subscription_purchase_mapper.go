package mapper

import (
	"log"
	"subsnotifpro-go/internal/playstore/api/dto"
	"time"

	"google.golang.org/api/androidpublisher/v3"
)

func ToSubscriptionPurchaseV2(googleSub *androidpublisher.SubscriptionPurchaseV2) (*dto.SubscriptionPurchaseV2, error) {
	sub := &dto.SubscriptionPurchaseV2{
		Kind:                 googleSub.Kind,
		RegionCode:           googleSub.RegionCode,
		LatestOrderID:        googleSub.LatestOrderId,
		SubscriptionState:    dto.SubscriptionState(googleSub.SubscriptionState),
		AcknowledgementState: dto.AcknowledgementState(googleSub.AcknowledgementState),
	}

	if googleSub.LinkedPurchaseToken != "" {
		sub.LinkedPurchaseToken = &googleSub.LinkedPurchaseToken
	}

	// Parse timestamps
	if startTime, err := time.Parse(time.RFC3339, googleSub.StartTime); err == nil {
		sub.StartTime = startTime
	}

	if googleSub.ExternalAccountIdentifiers != nil {
		sub.ExternalAccountIdentifiers = convertExternalAccountIdentifiers(googleSub.ExternalAccountIdentifiers)
	}

	if googleSub.SubscribeWithGoogleInfo != nil {
		sub.SubscribeWithGoogleInfo = convertSubscribeWithGoogleInfo(googleSub.SubscribeWithGoogleInfo)
	}

	// Convert other nested objects
	if googleSub.TestPurchase != nil {
		sub.IsTestPurchase = true
	} else {
		sub.IsTestPurchase = false
	}

	// Convert line items
	for _, item := range googleSub.LineItems {
		if item == nil {
			continue
		}
		lineItem := convertLineItem(item)
		sub.LineItems = append(sub.LineItems, lineItem)
	}

	// Convert nested contexts
	if googleSub.PausedStateContext != nil {
		sub.PausedStateContext = convertPausedStateContext(googleSub.PausedStateContext)
	}

	if googleSub.CanceledStateContext != nil {
		sub.CanceledStateContext = convertCanceledStateContext(googleSub.CanceledStateContext)
	}

	return sub, nil
}

func convertLineItem(item *androidpublisher.SubscriptionPurchaseLineItem) dto.LineItem {
	li := dto.LineItem{
		ProductID: item.ProductId,
	}

	if item.ExpiryTime != "" {
		if et, err := time.Parse(time.RFC3339, item.ExpiryTime); err == nil {
			li.ExpiryTime = et
		}
	}

	// Handle plan_type union
	if item.AutoRenewingPlan != nil {
		li.AutoRenewingPlan = convertAutoRenewingPlan(item.AutoRenewingPlan)
		li.PlanType = dto.PlanTypeAutoRenewing
	} else if item.PrepaidPlan != nil {
		li.PrepaidPlan = convertPrepaidPlan(item.PrepaidPlan)
		li.PlanType = dto.PlanTypePrepaid
	}

	// Handle other fields
	if item.OfferDetails != nil {
		li.OfferDetails = convertOfferDetails(item.OfferDetails)
	}

	if item.DeferredItemReplacement != nil {
		li.DeferredItemReplacement = &dto.DeferredItemReplacement{
			ProductID: item.DeferredItemReplacement.ProductId,
		}
	}

	if item.SignupPromotion != nil {
		li.SignupPromotion = convertSignupPromotion(item.SignupPromotion)
	}

	return li
}

func convertAutoRenewingPlan(plan *androidpublisher.AutoRenewingPlan) *dto.AutoRenewingPlan {
	if plan == nil {
		return nil
	}

	arp := &dto.AutoRenewingPlan{
		AutoRenewEnabled: plan.AutoRenewEnabled,
		RecurringPrice: dto.Money{
			CurrencyCode: plan.RecurringPrice.CurrencyCode,
			Units:        plan.RecurringPrice.Units,
			Nanos:        plan.RecurringPrice.Nanos,
		},
	}

	if plan.PriceChangeDetails != nil {
		arp.PriceChangeDetails = &dto.SubscriptionItemPriceChangeDetails{
			NewPrice: dto.Money{
				CurrencyCode: plan.PriceChangeDetails.NewPrice.CurrencyCode,
				Units:        plan.PriceChangeDetails.NewPrice.Units,
				Nanos:        plan.PriceChangeDetails.NewPrice.Nanos,
			},
			PriceChangeMode:  dto.PriceChangeMode(plan.PriceChangeDetails.PriceChangeMode),
			PriceChangeState: dto.PriceChangeState(plan.PriceChangeDetails.PriceChangeState),
		}
		if plan.PriceChangeDetails.ExpectedNewPriceChargeTime != "" {
			if t, err := time.Parse(time.RFC3339, plan.PriceChangeDetails.ExpectedNewPriceChargeTime); err == nil {
				arp.PriceChangeDetails.ExpectedNewPriceChargeTime = t
			}
		}
	}

	if plan.InstallmentDetails != nil {
		arp.InstallmentDetails = &dto.InstallmentPlan{
			InitialCommittedPaymentsCount:    int(plan.InstallmentDetails.InitialCommittedPaymentsCount),
			SubsequentCommittedPaymentsCount: int(plan.InstallmentDetails.SubsequentCommittedPaymentsCount),
			RemainingCommittedPaymentsCount:  int(plan.InstallmentDetails.RemainingCommittedPaymentsCount),
		}
		if plan.InstallmentDetails.PendingCancellation != nil {
			arp.InstallmentDetails.IsPendingCancellation = true
		} else {
			arp.InstallmentDetails.IsPendingCancellation = false
		}
	}

	return arp
}
func convertPrepaidPlan(plan *androidpublisher.PrepaidPlan) *dto.PrepaidPlan {
	if plan == nil {
		return nil
	}

	pp := &dto.PrepaidPlan{}

	// Only parse and set AllowExtendAfterTime if it's present and non-empty
	if plan.AllowExtendAfterTime != "" {
		if t, err := time.Parse(time.RFC3339, plan.AllowExtendAfterTime); err == nil {
			pp.AllowExtendAfterTime = &t // Store as pointer to time.Time
		} else {
			// Log parsing error but continue with nil AllowExtendAfterTime
			log.Printf("WARN: failed to parse AllowExtendAfterTime: %v", err)
		}
	}

	return pp
}

func convertOfferDetails(details *androidpublisher.OfferDetails) *dto.OfferDetails {
	if details == nil {
		return nil
	}

	return &dto.OfferDetails{
		OfferTags:  details.OfferTags,
		BasePlanID: details.BasePlanId,
		OfferID:    details.OfferId,
	}
}

func convertSignupPromotion(promo *androidpublisher.SignupPromotion) *dto.SignupPromotion {
	if promo == nil {
		return nil
	}

	result := &dto.SignupPromotion{}

	switch {
	case promo.OneTimeCode != nil:
		result.Type = dto.PromoTypeOneTimeCode
	case promo.VanityCode != nil && promo.VanityCode.PromotionCode != "":
		result.Type = dto.PromoTypeVanityCode
		code := promo.VanityCode.PromotionCode
		result.Code = &code
	default:
		return nil // No valid promotion found
	}

	return result
}

// PausedStateContext
func convertPausedStateContext(ctx *androidpublisher.PausedStateContext) *dto.PausedStateContext {
	if ctx == nil {
		return nil
	}

	psc := &dto.PausedStateContext{}
	if ctx.AutoResumeTime != "" {
		if t, err := time.Parse(time.RFC3339, ctx.AutoResumeTime); err == nil {
			psc.AutoResumeTime = t
		}
	}
	return psc
}

// Cancellation Context
func convertCanceledStateContext(ctx *androidpublisher.CanceledStateContext) *dto.CanceledStateContext {
	if ctx == nil {
		return nil
	}

	result := &dto.CanceledStateContext{}

	switch {
	case ctx.UserInitiatedCancellation != nil:
		result.Reason = dto.CancellationReasonUserInitiated
		if cancelTime, err := time.Parse(time.RFC3339, ctx.UserInitiatedCancellation.CancelTime); err == nil {
			result.CancelTime = &cancelTime
		}
		if survey := ctx.UserInitiatedCancellation.CancelSurveyResult; survey != nil {
			result.CancelSurveyResult = &struct {
				Reason          dto.CancelSurveyReason
				ReasonUserInput *string
			}{
				Reason:          convertCancelSurveyReason(survey.Reason),
				ReasonUserInput: nilIfEmpty(survey.ReasonUserInput),
			}
		}

	case ctx.SystemInitiatedCancellation != nil:
		result.Reason = dto.CancellationReasonSystemInitiated

	case ctx.DeveloperInitiatedCancellation != nil:
		result.Reason = dto.CancellationReasonDeveloperInitiated

	case ctx.ReplacementCancellation != nil:
		result.Reason = dto.CancellationReasonReplacement
	}

	return result
}

func convertCancelSurveyReason(reason string) dto.CancelSurveyReason {
	switch reason {
	case "CANCEL_SURVEY_REASON_NOT_ENOUGH_USAGE":
		return dto.CancelSurveyReasonNotEnoughUsage
	case "CANCEL_SURVEY_REASON_TECHNICAL_ISSUES":
		return dto.CancelSurveyReasonTechnicalIssues
	case "CANCEL_SURVEY_REASON_COST_RELATED":
		return dto.CancelSurveyReasonCostRelated
	case "CANCEL_SURVEY_REASON_FOUND_BETTER_APP":
		return dto.CancelSurveyReasonFoundBetterApp
	case "CANCEL_SURVEY_REASON_OTHERS":
		return dto.CancelSurveyReasonOthers
	default:
		return dto.CancelSurveyReasonUnspecified
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Account
// ExternalAccountIdentifiers
func convertExternalAccountIdentifiers(ids *androidpublisher.ExternalAccountIdentifiers) dto.ExternalAccountIdentifiers {
	result := dto.ExternalAccountIdentifiers{}

	if ids != nil {
		// All fields are optional, only set if they exist
		if ids.ExternalAccountId != "" {
			result.ExternalAccountID = &ids.ExternalAccountId
		}
		if ids.ObfuscatedExternalAccountId != "" {
			result.ObfuscatedExternalAccountID = &ids.ObfuscatedExternalAccountId
		}
		if ids.ObfuscatedExternalProfileId != "" {
			result.ObfuscatedExternalProfileID = &ids.ObfuscatedExternalProfileId
		}
	}

	return result
}

func convertSubscribeWithGoogleInfo(info *androidpublisher.SubscribeWithGoogleInfo) *dto.SubscribeWithGoogleInfo {
	if info == nil {
		return nil
	}

	result := &dto.SubscribeWithGoogleInfo{}

	// Only set fields that exist in the source
	if info.ProfileId != "" {
		result.ProfileID = &info.ProfileId
	}
	if info.ProfileName != "" {
		result.ProfileName = &info.ProfileName
	}
	if info.EmailAddress != "" {
		result.EmailAddress = &info.EmailAddress
	}
	if info.GivenName != "" {
		result.GivenName = &info.GivenName
	}
	if info.FamilyName != "" {
		result.FamilyName = &info.FamilyName
	}

	// Return nil if no fields were set (empty struct)
	if result.ProfileID == nil &&
		result.ProfileName == nil &&
		result.EmailAddress == nil &&
		result.GivenName == nil &&
		result.FamilyName == nil {
		return nil
	}

	return result
}
