package converter

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"time"

	"github.com/google/uuid"
)

// decodeJWSRenewalInfo decodes a JWS renewal info string into payload and header
func (c *NotificationConverter) decodeJWSRenewalInfo(jws string) (*dto.JWSRenewalInfoDecodedPayload, error) {
	parts := strings.Split(jws, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWS format - expected 3 parts")
	}

	// Decode and parse header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode renewal info header: %w", err)
	}

	var header dto.JWSDecodedHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to parse renewal info header: %w", err)
	}

	// Decode and parse payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode renewal info payload: %w", err)
	}

	// First unmarshal into a temporary struct that handles the timestamps as numbers
	var tempPayload struct {
		AppAccountToken             *uuid.UUID `json:"appAccountToken"`
		AppTransactionId            string     `json:"appTransactionId"`
		AutoRenewProductId          string     `json:"autoRenewProductId"`
		AutoRenewStatus             int32      `json:"autoRenewStatus"`
		Currency                    string     `json:"currency"`
		EligibleWinBackOfferIds     []string   `json:"eligibleWinBackOfferIds"`
		Environment                 string     `json:"environment"`
		ExpirationIntent            int32      `json:"expirationIntent"`
		GracePeriodExpiresDate      int64      `json:"gracePeriodExpiresDate"`
		IsInBillingRetryPeriod      bool       `json:"isInBillingRetryPeriod"`
		OfferDiscountType           string     `json:"offerDiscountType"`
		OfferIdentifier             string     `json:"offerIdentifier"`
		OfferPeriod                 string     `json:"offerPeriod"`
		OfferType                   int32      `json:"offerType"`
		OriginalTransactionId       string     `json:"originalTransactionId"`
		PriceIncreaseStatus         int32      `json:"priceIncreaseStatus"`
		ProductId                   string     `json:"productId"`
		RecentSubscriptionStartDate int64      `json:"recentSubscriptionStartDate"`
		RenewalDate                 int64      `json:"renewalDate"`
		RenewalPrice                int64      `json:"renewalPrice"`
		SignedDate                  int64      `json:"signedDate"`
	}

	if err := json.Unmarshal(payloadBytes, &tempPayload); err != nil {
		return nil, fmt.Errorf("failed to parse renewal info payload: %w", err)
	}

	// Convert the temporary struct to the final DTO with proper time.Time fields
	payload := dto.JWSRenewalInfoDecodedPayload{
		AppAccountToken:         tempPayload.AppAccountToken,
		AppTransactionId:        tempPayload.AppTransactionId,
		AutoRenewProductId:      tempPayload.AutoRenewProductId,
		AutoRenewStatus:         tempPayload.AutoRenewStatus,
		Currency:                tempPayload.Currency,
		EligibleWinBackOfferIds: tempPayload.EligibleWinBackOfferIds,
		Environment:             dto.Environment(tempPayload.Environment),
		ExpirationIntent:        tempPayload.ExpirationIntent,
		IsInBillingRetryPeriod:  tempPayload.IsInBillingRetryPeriod,
		OfferDiscountType:       tempPayload.OfferDiscountType,
		OfferIdentifier:         tempPayload.OfferIdentifier,
		OfferPeriod:             tempPayload.OfferPeriod,
		OfferType:               tempPayload.OfferType,
		OriginalTransactionId:   tempPayload.OriginalTransactionId,
		PriceIncreaseStatus:     tempPayload.PriceIncreaseStatus,
		ProductId:               tempPayload.ProductId,
		RenewalPrice:            tempPayload.RenewalPrice,
	}

	// Convert timestamp fields from milliseconds to time.Time
	payload.GracePeriodExpiresDate = convertMillisToTime(tempPayload.GracePeriodExpiresDate)
	payload.RecentSubscriptionStartDate = convertMillisToTime(tempPayload.RecentSubscriptionStartDate)
	payload.RenewalDate = convertMillisToTime(tempPayload.RenewalDate)
	payload.SignedDate = convertMillisToTime(tempPayload.SignedDate)

	// Validate payload fields
	if err := c.validateRenewalPayload(&payload); err != nil {
		return nil, fmt.Errorf("invalid renewal payload: %w", err)
	}

	return &payload, nil
}

// validateRenewalPayload validates all fields in the renewal payload
func (c *NotificationConverter) validateRenewalPayload(payload *dto.JWSRenewalInfoDecodedPayload) error {
	// Required fields validation
	if payload.OriginalTransactionId == "" {
		return errors.New("missing originalTransactionId")
	}
	if payload.ProductId == "" {
		return errors.New("missing productId")
	}
	if payload.AutoRenewProductId == "" {
		return errors.New("missing autoRenewProductId")
	}
	if payload.SignedDate.IsZero() {
		return errors.New("missing signedDate")
	}

	// Enum validations
	if payload.Environment != dto.EnvironmentSandbox && payload.Environment != dto.EnvironmentProduction {
		return fmt.Errorf("invalid environment: %s", payload.Environment)
	}

	if payload.AutoRenewStatus != 0 && payload.AutoRenewStatus != 1 {
		return fmt.Errorf("invalid autoRenewStatus: %d", payload.AutoRenewStatus)
	}

	if payload.ExpirationIntent != 0 &&
		(payload.ExpirationIntent < 1 || payload.ExpirationIntent > 5) {
		return fmt.Errorf("invalid expirationIntent: %d", payload.ExpirationIntent)
	}

	if payload.OfferDiscountType != "" &&
		!contains([]string{"FREE_TRIAL", "PAY_AS_YOU_GO", "PAY_UP_FRONT"}, payload.OfferDiscountType) {
		return fmt.Errorf("invalid offerDiscountType: %s", payload.OfferDiscountType)
	}

	if payload.OfferType != 0 &&
		!contains([]int32{1, 2, 3, 4}, payload.OfferType) {
		return fmt.Errorf("invalid offerType: %d", payload.OfferType)
	}

	if payload.PriceIncreaseStatus != 0 && payload.PriceIncreaseStatus != 1 {
		return fmt.Errorf("invalid priceIncreaseStatus: %d", payload.PriceIncreaseStatus)
	}

	// Format validations
	if payload.Currency != "" && len(payload.Currency) != 3 {
		return fmt.Errorf("invalid currency code: %s", payload.Currency)
	}

	if payload.OfferPeriod != "" {
		if _, err := time.ParseDuration(strings.ReplaceAll(payload.OfferPeriod, "P", "")); err != nil {
			return fmt.Errorf("invalid offerPeriod format: %s", payload.OfferPeriod)
		}
	}

	// Value range validations
	if payload.RenewalPrice < 0 {
		return fmt.Errorf("invalid renewalPrice: %d", payload.RenewalPrice)
	}

	// Date consistency checks
	if !payload.GracePeriodExpiresDate.IsZero() && payload.GracePeriodExpiresDate.Before(payload.SignedDate) {
		return fmt.Errorf("gracePeriodExpiresDate (%s) before signedDate (%s)",
			payload.GracePeriodExpiresDate, payload.SignedDate)
	}

	if !payload.RenewalDate.IsZero() && payload.RenewalDate.Before(payload.SignedDate) {
		return fmt.Errorf("renewalDate (%s) before signedDate (%s)",
			payload.RenewalDate, payload.SignedDate)
	}

	if !payload.RecentSubscriptionStartDate.IsZero() {
		if payload.RecentSubscriptionStartDate.After(payload.SignedDate) {
			return fmt.Errorf("recentSubscriptionStartDate (%s) after signedDate (%s)",
				payload.RecentSubscriptionStartDate, payload.SignedDate)
		}
		if !payload.RenewalDate.IsZero() && payload.RecentSubscriptionStartDate.After(payload.RenewalDate) {
			return fmt.Errorf("recentSubscriptionStartDate (%s) after renewalDate (%s)",
				payload.RecentSubscriptionStartDate, payload.RenewalDate)
		}
	}

	return nil
}
