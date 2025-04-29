package models

import (
	"reflect"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"time"

	"github.com/google/uuid"
)

// FromAppStoreNotificationDTO converts a dto.AppStoreNotification to models.AppStoreNotification
func FromAppStoreNotificationDTO(dtoNotification *dto.AppStoreNotification) (*AppStoreNotification, error) {
	notification := &AppStoreNotification{
		ID:         uuid.New(),
		ReceivedAt: dtoNotification.ReceivedAt,
	}

	// Convert JWS header if present
	if !reflect.DeepEqual(dtoNotification.JWSDecodedHeader, dto.JWSDecodedHeader{}) {
		jwsHeader := FromJWSDecodedHeaderDTO(dtoNotification.JWSDecodedHeader)
		notification.JWSDecodedHeader = jwsHeader
	}

	// Convert response body payload
	payload, err := FromResponseBodyV2DecodedPayloadDTO(dtoNotification.ResponseBodyV2DecodedPayload)
	if err != nil {
		return nil, err
	}
	notification.ResponseBodyV2DecodedPayload = payload

	return notification, nil
}

// FromJWSDecodedHeaderDTO converts a dto.JWSDecodedHeader to models.JWSDecodedHeader
func FromJWSDecodedHeaderDTO(dtoHeader dto.JWSDecodedHeader) *JWSDecodedHeader {
	return &JWSDecodedHeader{
		Algorithm: dtoHeader.Alg,
		X5c:       dtoHeader.X5c,
	}
}

// FromResponseBodyV2DecodedPayloadDTO converts a dto.ResponseBodyV2DecodedPayload to models.ResponseBodyV2DecodedPayload
func FromResponseBodyV2DecodedPayloadDTO(dtoPayload dto.ResponseBodyV2DecodedPayload) (*ResponseBodyV2DecodedPayload, error) {
	payload := &ResponseBodyV2DecodedPayload{}

	// Convert environment
	switch dtoPayload.Data.Environment {
	case dto.EnvironmentSandbox:
		payload.Environment = EnvironmentSandbox
	case dto.EnvironmentProduction:
		payload.Environment = EnvironmentProduction
	default:
		payload.Environment = EnvironmentSandbox
	}

	// Convert data
	data, err := FromDataDTO(*dtoPayload.Data)
	if err != nil {
		return nil, err
	}
	payload.Data = data

	// Convert summary if present
	if dtoPayload.Summary != nil {
		summary, err := FromSummaryDTO(*dtoPayload.Summary)
		if err != nil {
			return nil, err
		}
		payload.Summary = summary
	}

	// Set external purchase token if present
	if dtoPayload.ExternalPurchaseToken != nil {
		payload.ExternalPurchaseToken = dtoPayload.ExternalPurchaseToken
	}

	// Set notification metadata
	payload.Version = dtoPayload.Version
	payload.SignedDate = time.Unix(dtoPayload.SignedDate/1000, 0)
	payload.NotificationUUID = dtoPayload.NotificationUUID
	payload.Subtype = (*NotificationSubtype)(dtoPayload.Subtype)
	payload.NotificationType = NotificationType(dtoPayload.NotificationType)

	return payload, nil
}

// FromDataDTO converts a dto.Data to models.Data
func FromDataDTO(dtoData dto.Data) (*AppStoreNotificationData, error) {
	data := &AppStoreNotificationData{
		BundleID:      dtoData.BundleID,
		BundleVersion: dtoData.BundleVersion,
	}

	if dtoData.AppAppleID != nil {
		data.AppAppleID = dtoData.AppAppleID
	}

	// Convert environment
	switch dtoData.Environment {
	case dto.EnvironmentSandbox:
		data.Environment = EnvironmentSandbox
	case dto.EnvironmentProduction:
		data.Environment = EnvironmentProduction
	default:
		data.Environment = EnvironmentSandbox
	}

	// Convert transaction info
	transaction, err := FromJWSTransactionDTO(dtoData.SignedTransactionInfo)
	if err != nil {
		return nil, err
	}
	data.SignedTransactionInfo = transaction

	// Convert renewal info if present
	if dtoData.SignedRenewalInfo != nil {
		renewalInfo, err := FromJWSRenewalInfoDTO(*dtoData.SignedRenewalInfo)
		if err != nil {
			return nil, err
		}
		data.SignedRenewalInfo = renewalInfo
	}

	// Set status if present
	if dtoData.Status != nil {
		status := int32(*dtoData.Status)
		data.Status = &status
	}

	// Set consumption request reason if present
	if dtoData.ConsumptionRequestReason != nil {
		reason := string(*dtoData.ConsumptionRequestReason)
		data.ConsumptionRequestReason = &reason
	}

	return data, nil
}

// FromSummaryDTO converts a dto.Summary to models.Summary
func FromSummaryDTO(dtoSummary dto.Summary) (*AppStoreNotificationSummary, error) {
	return &AppStoreNotificationSummary{
		RequestIdentifier:      dtoSummary.RequestIdentifier,
		AppAppleID:             dtoSummary.AppAppleId,
		BundleID:               dtoSummary.BundleID,
		ProductId:              dtoSummary.ProductId,
		StorefrontCountryCodes: dtoSummary.StorefrontCountryCodes,
		FailedCount:            dtoSummary.FailedCount,
		SucceededCount:         dtoSummary.SucceededCount,
		Environment:            Environment(dtoSummary.Environment),
	}, nil
}

// FromJWSTransactionDTO converts a dto.JWSTransaction to models.JWSTransaction
func FromJWSTransactionDTO(dtoTransaction dto.JWSTransaction) (*JWSTransaction, error) {
	transaction := &JWSTransaction{
		TransactionString: dtoTransaction.JWSTransaction,
	}

	// Convert decoded payload
	payload, err := FromJWSTransactionDecodedPayloadDTO(dtoTransaction.JWSTransactionDecodedPayload)
	if err != nil {
		return nil, err
	}
	transaction.DecodedPayload = payload

	return transaction, nil
}

// FromJWSTransactionDecodedPayloadDTO converts a dto.JWSTransactionDecodedPayload to models.JWSTransactionDecodedPayload
func FromJWSTransactionDecodedPayloadDTO(dtoPayload dto.JWSTransactionDecodedPayload) (*JWSTransactionDecodedPayload, error) {
	payload := &JWSTransactionDecodedPayload{
		AppTransactionId:            dtoPayload.AppTransactionId,
		BundleId:                    dtoPayload.BundleId,
		Currency:                    dtoPayload.Currency,
		InAppOwnershipType:          dtoPayload.InAppOwnershipType,
		IsUpgraded:                  dtoPayload.IsUpgraded,
		OfferDiscountType:           dtoPayload.OfferDiscountType,
		OfferIdentifier:             dtoPayload.OfferIdentifier,
		OfferPeriod:                 dtoPayload.OfferPeriod,
		OfferType:                   dtoPayload.OfferType,
		OriginalTransactionId:       dtoPayload.OriginalTransactionId,
		Price:                       dtoPayload.Price,
		ProductId:                   dtoPayload.ProductId,
		Quantity:                    dtoPayload.Quantity,
		Storefront:                  dtoPayload.Storefront,
		StorefrontId:                dtoPayload.StorefrontId,
		SubscriptionGroupIdentifier: dtoPayload.SubscriptionGroupIdentifier,
		TransactionId:               dtoPayload.TransactionId,
		TransactionReason:           dtoPayload.TransactionReason,
		Type:                        dtoPayload.Type,
		WebOrderLineItemId:          dtoPayload.WebOrderLineItemId,
		Environment:                 string(dtoPayload.Environment),
		ExpiresDate:                 dtoPayload.ExpiresDate,
		OriginalPurchaseDate:        dtoPayload.OriginalPurchaseDate,
		PurchaseDate:                dtoPayload.PurchaseDate,
		SignedDate:                  dtoPayload.SignedDate,
	}

	// Set optional fields
	if dtoPayload.AppAccountToken != nil {
		payload.AppAccountToken = dtoPayload.AppAccountToken
	}
	if dtoPayload.PreviousOriginalTransactionId != nil {
		payload.PreviousOriginalTransactionId = dtoPayload.PreviousOriginalTransactionId
	}
	if dtoPayload.RevocationDate != nil {
		payload.RevocationDate = dtoPayload.RevocationDate
	}
	if dtoPayload.RevocationReason != nil {
		payload.RevocationReason = dtoPayload.RevocationReason
	}

	return payload, nil
}

// FromJWSRenewalInfoDTO converts a dto.JWSRenewalInfo to models.JWSRenewalInfo
func FromJWSRenewalInfoDTO(dtoRenewalInfo dto.JWSRenewalInfo) (*JWSRenewalInfo, error) {
	renewalInfo := &JWSRenewalInfo{
		RenewalInfoString: dtoRenewalInfo.JWSRenewalInfo,
	}

	// Convert decoded payload
	payload, err := FromJWSRenewalInfoDecodedPayloadDTO(dtoRenewalInfo.JWSRenewalInfoDecodedPayload)
	if err != nil {
		return nil, err
	}
	renewalInfo.DecodedPayload = payload

	return renewalInfo, nil
}

// FromJWSRenewalInfoDecodedPayloadDTO converts a dto.JWSRenewalInfoDecodedPayload to models.JWSRenewalInfoDecodedPayload
func FromJWSRenewalInfoDecodedPayloadDTO(dtoPayload dto.JWSRenewalInfoDecodedPayload) (*JWSRenewalInfoDecodedPayload, error) {
	payload := &JWSRenewalInfoDecodedPayload{
		AppTransactionId:            dtoPayload.AppTransactionId,
		AutoRenewProductId:          dtoPayload.AutoRenewProductId,
		AutoRenewStatus:             dtoPayload.AutoRenewStatus,
		Currency:                    dtoPayload.Currency,
		EligibleWinBackOfferIds:     dtoPayload.EligibleWinBackOfferIds,
		ExpirationIntent:            dtoPayload.ExpirationIntent,
		IsInBillingRetryPeriod:      dtoPayload.IsInBillingRetryPeriod,
		OfferDiscountType:           dtoPayload.OfferDiscountType,
		OfferIdentifier:             dtoPayload.OfferIdentifier,
		OfferPeriod:                 dtoPayload.OfferPeriod,
		OfferType:                   dtoPayload.OfferType,
		OriginalTransactionId:       dtoPayload.OriginalTransactionId,
		PriceIncreaseStatus:         dtoPayload.PriceIncreaseStatus,
		ProductId:                   dtoPayload.ProductId,
		RenewalPrice:                dtoPayload.RenewalPrice,
		Environment:                 string(dtoPayload.Environment),
		GracePeriodExpiresDate:      dtoPayload.GracePeriodExpiresDate,
		RecentSubscriptionStartDate: dtoPayload.RecentSubscriptionStartDate,
		RenewalDate:                 dtoPayload.RenewalDate,
		SignedDate:                  dtoPayload.SignedDate,
	}

	// Set optional fields
	if dtoPayload.AppAccountToken != nil {
		payload.AppAccountToken = dtoPayload.AppAccountToken
	}

	return payload, nil
}

// ToDTO converts models.AppStoreNotification to dto.AppStoreNotification
func (n *AppStoreNotification) ToDTO() *dto.AppStoreNotification {
	dtoNotification := &dto.AppStoreNotification{
		ReceivedAt: n.ReceivedAt,
	}

	// Convert JWS header if present
	if n.JWSDecodedHeader != nil {
		dtoNotification.JWSDecodedHeader = n.JWSDecodedHeader.ToDTO()
	}

	// Convert response body payload
	if n.ResponseBodyV2DecodedPayload != nil {
		dtoNotification.ResponseBodyV2DecodedPayload = n.ResponseBodyV2DecodedPayload.ToDTO()
	}

	return dtoNotification
}

// ToDTO converts models.JWSDecodedHeader to dto.JWSDecodedHeader
func (h *JWSDecodedHeader) ToDTO() dto.JWSDecodedHeader {
	return dto.JWSDecodedHeader{
		Alg: h.Algorithm,
		X5c: h.X5c,
	}
}

// ToDTO converts models.ResponseBodyV2DecodedPayload to dto.ResponseBodyV2DecodedPayload
func (p *ResponseBodyV2DecodedPayload) ToDTO() dto.ResponseBodyV2DecodedPayload {
	dtoPayload := dto.ResponseBodyV2DecodedPayload{
		Version:          p.Version,
		NotificationUUID: p.NotificationUUID,
		SignedDate:       p.SignedDate.UnixNano() / int64(time.Millisecond),
	}

	// Convert notification type
	dtoPayload.NotificationType = dto.NotificationType(p.NotificationType)

	// Convert subtype if present

	dtoPayload.Subtype = (*dto.NotificationSubtype)(p.Subtype)

	// Convert data if present
	if p.Data != nil {
		data := p.Data.ToDTO()
		dtoPayload.Data = &data
	}

	// Convert summary if present
	if p.Summary != nil {
		summary := p.Summary.ToDTO()
		dtoPayload.Summary = &summary
	}

	// Set external purchase token if present
	if p.ExternalPurchaseToken != nil {
		dtoPayload.ExternalPurchaseToken = p.ExternalPurchaseToken
	}

	return dtoPayload
}

// ToDTO converts models.Data to dto.Data
func (d *AppStoreNotificationData) ToDTO() dto.Data {
	dtoData := dto.Data{
		BundleID:      d.BundleID,
		BundleVersion: d.BundleVersion,
	}

	if d.AppAppleID != nil {
		dtoData.AppAppleID = d.AppAppleID
	}

	// Convert environment
	switch d.Environment {
	case EnvironmentSandbox:
		dtoData.Environment = dto.EnvironmentSandbox
	case EnvironmentProduction:
		dtoData.Environment = dto.EnvironmentProduction
	default:
		dtoData.Environment = dto.EnvironmentSandbox
	}

	// Convert transaction info
	if d.SignedTransactionInfo != nil {
		dtoData.SignedTransactionInfo = d.SignedTransactionInfo.ToDTO()
	}

	// Convert renewal info if present
	if d.SignedRenewalInfo != nil {
		renewalInfo := d.SignedRenewalInfo.ToDTO()
		dtoData.SignedRenewalInfo = &renewalInfo
	}

	// Set status if present
	if d.Status != nil {
		status := dto.SubscriptionStatus(*d.Status)
		dtoData.Status = &status
	}

	// Set consumption request reason if present
	if d.ConsumptionRequestReason != nil {
		reason := dto.ConsumptionRequestReason(*d.ConsumptionRequestReason)
		dtoData.ConsumptionRequestReason = &reason
	}

	return dtoData
}

// ToDTO converts models.Summary to dto.Summary
func (s *AppStoreNotificationSummary) ToDTO() dto.Summary {
	return dto.Summary{
		RequestIdentifier:      s.RequestIdentifier,
		AppAppleId:             s.AppAppleID,
		BundleID:               s.BundleID,
		ProductId:              s.ProductId,
		StorefrontCountryCodes: s.StorefrontCountryCodes,
		FailedCount:            s.FailedCount,
		SucceededCount:         s.SucceededCount,
		Environment:            dto.Environment(s.Environment),
	}
}

// ToDTO converts models.JWSTransaction to dto.JWSTransaction
func (t *JWSTransaction) ToDTO() dto.JWSTransaction {
	dtoTransaction := dto.JWSTransaction{
		JWSTransaction: t.TransactionString,
	}

	// Convert decoded payload if present
	if t.DecodedPayload != nil {
		dtoTransaction.JWSTransactionDecodedPayload = t.DecodedPayload.ToDTO()
	}

	return dtoTransaction
}

// ToDTO converts models.JWSTransactionDecodedPayload to dto.JWSTransactionDecodedPayload
func (p *JWSTransactionDecodedPayload) ToDTO() dto.JWSTransactionDecodedPayload {
	dtoPayload := dto.JWSTransactionDecodedPayload{
		AppTransactionId:            p.AppTransactionId,
		BundleId:                    p.BundleId,
		Currency:                    p.Currency,
		InAppOwnershipType:          p.InAppOwnershipType,
		IsUpgraded:                  p.IsUpgraded,
		OfferDiscountType:           p.OfferDiscountType,
		OfferIdentifier:             p.OfferIdentifier,
		OfferPeriod:                 p.OfferPeriod,
		OfferType:                   p.OfferType,
		OriginalTransactionId:       p.OriginalTransactionId,
		Price:                       p.Price,
		ProductId:                   p.ProductId,
		Quantity:                    p.Quantity,
		Storefront:                  p.Storefront,
		StorefrontId:                p.StorefrontId,
		SubscriptionGroupIdentifier: p.SubscriptionGroupIdentifier,
		TransactionId:               p.TransactionId,
		TransactionReason:           p.TransactionReason,
		Type:                        p.Type,
		WebOrderLineItemId:          p.WebOrderLineItemId,
		ExpiresDate:                 p.ExpiresDate,
		OriginalPurchaseDate:        p.OriginalPurchaseDate,
		PurchaseDate:                p.PurchaseDate,
		SignedDate:                  p.SignedDate,
	}

	// Convert environment
	switch p.Environment {
	case "Sandbox":
		dtoPayload.Environment = dto.EnvironmentSandbox
	case "Production":
		dtoPayload.Environment = dto.EnvironmentProduction
	default:
		dtoPayload.Environment = dto.EnvironmentSandbox
	}

	// Set optional fields
	if p.AppAccountToken != nil {
		dtoPayload.AppAccountToken = p.AppAccountToken
	}
	if p.PreviousOriginalTransactionId != nil {
		dtoPayload.PreviousOriginalTransactionId = p.PreviousOriginalTransactionId
	}
	if p.RevocationDate != nil {
		dtoPayload.RevocationDate = p.RevocationDate
	}
	if p.RevocationReason != nil {
		dtoPayload.RevocationReason = p.RevocationReason
	}

	return dtoPayload
}

// ToDTO converts models.JWSRenewalInfo to dto.JWSRenewalInfo
func (r *JWSRenewalInfo) ToDTO() dto.JWSRenewalInfo {
	dtoRenewalInfo := dto.JWSRenewalInfo{
		JWSRenewalInfo: r.RenewalInfoString,
	}

	// Convert decoded payload if present
	if r.DecodedPayload != nil {
		dtoRenewalInfo.JWSRenewalInfoDecodedPayload = r.DecodedPayload.ToDTO()
	}

	return dtoRenewalInfo
}

// ToDTO converts models.JWSRenewalInfoDecodedPayload to dto.JWSRenewalInfoDecodedPayload
func (p *JWSRenewalInfoDecodedPayload) ToDTO() dto.JWSRenewalInfoDecodedPayload {
	dtoPayload := dto.JWSRenewalInfoDecodedPayload{
		AppTransactionId:            p.AppTransactionId,
		AutoRenewProductId:          p.AutoRenewProductId,
		AutoRenewStatus:             p.AutoRenewStatus,
		Currency:                    p.Currency,
		EligibleWinBackOfferIds:     p.EligibleWinBackOfferIds,
		ExpirationIntent:            p.ExpirationIntent,
		IsInBillingRetryPeriod:      p.IsInBillingRetryPeriod,
		OfferDiscountType:           p.OfferDiscountType,
		OfferIdentifier:             p.OfferIdentifier,
		OfferPeriod:                 p.OfferPeriod,
		OfferType:                   p.OfferType,
		OriginalTransactionId:       p.OriginalTransactionId,
		PriceIncreaseStatus:         p.PriceIncreaseStatus,
		ProductId:                   p.ProductId,
		RenewalPrice:                p.RenewalPrice,
		GracePeriodExpiresDate:      p.GracePeriodExpiresDate,
		RecentSubscriptionStartDate: p.RecentSubscriptionStartDate,
		RenewalDate:                 p.RenewalDate,
		SignedDate:                  p.SignedDate,
	}

	// Convert environment
	switch p.Environment {
	case "Sandbox":
		dtoPayload.Environment = dto.EnvironmentSandbox
	case "Production":
		dtoPayload.Environment = dto.EnvironmentProduction
	default:
		dtoPayload.Environment = dto.EnvironmentSandbox
	}

	// Set optional fields
	if p.AppAccountToken != nil {
		dtoPayload.AppAccountToken = p.AppAccountToken
	}

	return dtoPayload
}
