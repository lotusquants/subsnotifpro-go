package converter

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"time"
)

// decodeJWSTransaction decodes a JWS transaction string into payload and header
// decodeJWSTransaction decodes a JWS transaction string into payload and header
func (c *NotificationConverter) decodeJWSTransaction(jws string) (*dto.JWSTransactionDecodedPayload, error) {
	parts := strings.Split(jws, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWS format - expected 3 parts")
	}

	// Decode and parse header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	var header dto.JWSDecodedHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	// Decode and parse payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	// First unmarshal into a temporary struct that handles the timestamps as numbers
	var tempPayload struct {
		AppAccountToken               *string `json:"appAccountToken"`
		AppTransactionId              string  `json:"appTransactionId"`
		BundleId                      string  `json:"bundleId"`
		Currency                      string  `json:"currency"`
		Environment                   string  `json:"environment"`
		ExpiresDate                   int64   `json:"expiresDate"`
		InAppOwnershipType            string  `json:"inAppOwnershipType"`
		IsUpgraded                    bool    `json:"isUpgraded"`
		OfferDiscountType             string  `json:"offerDiscountType"`
		OfferIdentifier               string  `json:"offerIdentifier"`
		OfferPeriod                   string  `json:"offerPeriod"`
		OfferType                     int32   `json:"offerType"`
		OriginalPurchaseDate          int64   `json:"originalPurchaseDate"`
		OriginalTransactionId         string  `json:"originalTransactionId"`
		PreviousOriginalTransactionId *string `json:"previousOriginalTransactionId"`
		Price                         int64   `json:"price"`
		ProductId                     string  `json:"productId"`
		PurchaseDate                  int64   `json:"purchaseDate"`
		Quantity                      int32   `json:"quantity"`
		RevocationDate                *int64  `json:"revocationDate"`
		RevocationReason              *int32  `json:"revocationReason"`
		SignedDate                    int64   `json:"signedDate"`
		Storefront                    string  `json:"storefront"`
		StorefrontId                  string  `json:"storefrontId"`
		SubscriptionGroupIdentifier   string  `json:"subscriptionGroupIdentifier"`
		TransactionId                 string  `json:"transactionId"`
		TransactionReason             string  `json:"transactionReason"`
		Type                          string  `json:"type"`
		WebOrderLineItemId            string  `json:"webOrderLineItemId"`
	}

	if err := json.Unmarshal(payloadBytes, &tempPayload); err != nil {
		return nil, fmt.Errorf("failed to parse payload: %w", err)
	}

	// Convert the temporary struct to the final DTO with proper time.Time fields
	payload := dto.JWSTransactionDecodedPayload{
		AppAccountToken:               tempPayload.AppAccountToken,
		AppTransactionId:              tempPayload.AppTransactionId,
		BundleId:                      tempPayload.BundleId,
		Currency:                      tempPayload.Currency,
		Environment:                   dto.Environment(tempPayload.Environment),
		ExpiresDate:                   convertMillisToTime(tempPayload.ExpiresDate),
		InAppOwnershipType:            tempPayload.InAppOwnershipType,
		IsUpgraded:                    tempPayload.IsUpgraded,
		OfferDiscountType:             tempPayload.OfferDiscountType,
		OfferIdentifier:               tempPayload.OfferIdentifier,
		OfferPeriod:                   tempPayload.OfferPeriod,
		OfferType:                     tempPayload.OfferType,
		OriginalPurchaseDate:          convertMillisToTime(tempPayload.OriginalPurchaseDate),
		OriginalTransactionId:         tempPayload.OriginalTransactionId,
		PreviousOriginalTransactionId: tempPayload.PreviousOriginalTransactionId,
		Price:                         tempPayload.Price,
		ProductId:                     tempPayload.ProductId,
		PurchaseDate:                  convertMillisToTime(tempPayload.PurchaseDate),
		Quantity:                      tempPayload.Quantity,
		RevocationReason:              tempPayload.RevocationReason,
		SignedDate:                    convertMillisToTime(tempPayload.SignedDate),
		Storefront:                    tempPayload.Storefront,
		StorefrontId:                  tempPayload.StorefrontId,
		SubscriptionGroupIdentifier:   tempPayload.SubscriptionGroupIdentifier,
		TransactionId:                 tempPayload.TransactionId,
		TransactionReason:             tempPayload.TransactionReason,
		Type:                          tempPayload.Type,
		WebOrderLineItemId:            tempPayload.WebOrderLineItemId,
	}

	// Handle optional revocation date
	if tempPayload.RevocationDate != nil {
		revocationDate := convertMillisToTime(*tempPayload.RevocationDate)
		payload.RevocationDate = &revocationDate
	}

	// Validate required fields
	if err := c.validateTransactionPayload(&payload); err != nil {
		return nil, fmt.Errorf("invalid transaction payload: %w", err)
	}

	return &payload, nil
}

// convertMillisToTime converts a Unix timestamp in milliseconds to time.Time
func convertMillisToTime(millis int64) time.Time {
	if millis == 0 {
		return time.Time{}
	}
	return time.Unix(0, millis*int64(time.Millisecond))
}

// validateTransactionPayload validates all fields in the transaction payload
func (c *NotificationConverter) validateTransactionPayload(payload *dto.JWSTransactionDecodedPayload) error {
	// Required fields validation
	if payload.TransactionId == "" {
		return errors.New("missing transactionId")
	}
	if payload.OriginalTransactionId == "" {
		return errors.New("missing originalTransactionId")
	}
	if payload.BundleId == "" {
		return errors.New("missing bundleId")
	}
	if payload.ProductId == "" {
		return errors.New("missing productId")
	}
	if payload.PurchaseDate.IsZero() {
		return errors.New("missing purchaseDate")
	}
	if payload.SignedDate.IsZero() {
		return errors.New("missing signedDate")
	}

	// Enum validations
	if payload.Environment != dto.EnvironmentSandbox && payload.Environment != dto.EnvironmentProduction {
		return fmt.Errorf("invalid environment: %s", payload.Environment)
	}

	if payload.InAppOwnershipType != "" &&
		payload.InAppOwnershipType != "FAMILY_SHARED" &&
		payload.InAppOwnershipType != "PURCHASED" {
		return fmt.Errorf("invalid inAppOwnershipType: %s", payload.InAppOwnershipType)
	}

	if payload.OfferDiscountType != "" &&
		!contains([]string{"FREE_TRIAL", "PAY_AS_YOU_GO", "PAY_UP_FRONT"}, payload.OfferDiscountType) {
		return fmt.Errorf("invalid offerDiscountType: %s", payload.OfferDiscountType)
	}

	if payload.OfferType != 0 &&
		!contains([]int32{1, 2, 3, 4}, payload.OfferType) {
		return fmt.Errorf("invalid offerType: %d", payload.OfferType)
	}

	if payload.TransactionReason != "" &&
		!contains([]string{"PURCHASE", "RENEWAL"}, payload.TransactionReason) {
		return fmt.Errorf("invalid transactionReason: %s", payload.TransactionReason)
	}

	if payload.Type != "" &&
		!contains([]string{
			"Auto-Renewable Subscription",
			"Non-Consumable",
			"Consumable",
			"Non-Renewing Subscription",
		}, payload.Type) {
		return fmt.Errorf("invalid type: %s", payload.Type)
	}

	if payload.RevocationReason != nil &&
		!contains([]int32{0, 1}, *payload.RevocationReason) {
		return fmt.Errorf("invalid revocationReason: %d", *payload.RevocationReason)
	}

	// Value range validations
	if payload.Quantity <= 0 {
		return fmt.Errorf("invalid quantity: %d", payload.Quantity)
	}
	if payload.Price < 0 {
		return fmt.Errorf("invalid price: %d", payload.Price)
	}

	// Date consistency checks
	if !payload.ExpiresDate.IsZero() && payload.ExpiresDate.Before(payload.PurchaseDate) {
		return fmt.Errorf("expiresDate (%s) before purchaseDate (%s)",
			payload.ExpiresDate, payload.PurchaseDate)
	}

	if payload.RevocationDate != nil && payload.RevocationDate.Before(payload.PurchaseDate) {
		return fmt.Errorf("revocationDate (%s) before purchaseDate (%s)",
			*payload.RevocationDate, payload.PurchaseDate)
	}

	return nil
}

// contains is a helper function for enum validation
func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
