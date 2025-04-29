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

// NotificationConverter converts raw App Store notifications to fully decoded DTOs
type NotificationConverter struct{}

// NewNotificationConverter creates a new NotificationConverter
func NewNotificationConverter() *NotificationConverter {
	return &NotificationConverter{}
}

// ConvertToFullDTO converts a raw ResponseBodyV2 to a fully decoded AppStoreNotification
func (c *NotificationConverter) ConvertToFullDTO(raw *dto.ResponseBodyV2) (*dto.AppStoreNotification, error) {
	if raw == nil {
		return nil, errors.New("nil raw notification")
	}

	// Generate UUIDv7 with proper error handling
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	// Create base notification with system fields
	notification := &dto.AppStoreNotification{
		ID:         uuidV7,
		ReceivedAt: time.Now().UTC(),
	}

	// Decode the signed payload if present
	if raw.SignedPayload != "" {
		if err := c.decodeSignedPayload(raw.SignedPayload, notification); err != nil {
			return nil, fmt.Errorf("failed to decode signed payload: %w", err)
		}
	} else {
		return nil, errors.New("missing signed payload in notification")
	}

	return notification, nil
}

// decodeSignedPayload decodes the JWS signed payload into the notification
func (c *NotificationConverter) decodeSignedPayload(signedPayload string, notification *dto.AppStoreNotification) error {
	// Split JWS into parts
	parts := strings.Split(signedPayload, ".")
	if len(parts) != 3 {
		return errors.New("invalid JWS format - expected 3 parts")
	}

	// 1. Decode and parse header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("failed to decode header: %w", err)
	}

	var header dto.JWSDecodedHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return fmt.Errorf("failed to parse header: %w", err)
	}
	notification.JWSDecodedHeader = header

	// 2. Decode and parse payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}

	var payload dto.ResponseBodyV2DecodedPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	// 3. Process the decoded payload
	if err := c.processDecodedPayload(&payload); err != nil {
		return fmt.Errorf("failed to process decoded payload: %w", err)
	}

	notification.ResponseBodyV2DecodedPayload = payload

	return nil
}

// processDecodedPayload handles all nested structures in the decoded payload
func (c *NotificationConverter) processDecodedPayload(payload *dto.ResponseBodyV2DecodedPayload) error {
	// Validate required fields
	if payload.NotificationType == "" {
		return errors.New("missing notificationType in payload")
	}
	if payload.Version == "" {
		return errors.New("missing version in payload")
	}
	if payload.SignedDate == 0 {
		return errors.New("missing signedDate in payload")
	}
	if payload.NotificationUUID == "" {
		return errors.New("missing notificationUUID in payload")
	}

	// Validate that only one of Data, Summary, or ExternalPurchaseToken is present
	count := 0
	if payload.Data != nil {
		count++
	}
	if payload.Summary != nil {
		count++
	}
	if payload.ExternalPurchaseToken != nil {
		count++
	}
	if count != 1 {
		return errors.New("payload must contain exactly one of data, summary, or externalPurchaseToken")
	}

	// Handle Data object if present
	if payload.Data != nil {
		if err := c.processDataPayload(payload.Data); err != nil {
			return fmt.Errorf("failed to process data payload: %w", err)
		}
	}

	// Handle Summary object if present
	if payload.Summary != nil {
		if payload.NotificationType != "RENEWAL_EXTENSION" {
			return fmt.Errorf("summary present for non-renewal extension notification type: %s", payload.NotificationType)
		}
		if payload.Subtype == nil || *payload.Subtype != "SUMMARY" {
			return errors.New("summary present but subtype is not SUMMARY")
		}

		// Validate summary fields
		if payload.Summary.RequestIdentifier == "" {
			return errors.New("summary missing request identifier")
		}
		if payload.Summary.BundleID == "" {
			return errors.New("summary missing bundleId")
		}
		if payload.Summary.ProductId == "" {
			return errors.New("summary missing productId")
		}
		if payload.Summary.AppAppleId == 0 {
			return errors.New("summary missing appAppleId")
		}
		if len(payload.Summary.StorefrontCountryCodes) == 0 {
			return errors.New("summary missing storefrontCountryCodes")
		}
		if payload.Summary.Environment == "" {
			return errors.New("summary missing environment")
		}
	}

	// Handle ExternalPurchaseToken if present
	if payload.ExternalPurchaseToken != nil {
		if payload.NotificationType != "EXTERNAL_PURCHASE_TOKEN" {
			return fmt.Errorf("externalPurchaseToken present for non-external purchase notification type: %s", payload.NotificationType)
		}
		if *payload.ExternalPurchaseToken == "" {
			return errors.New("empty externalPurchaseToken")
		}
	}

	return nil
}

// processDataPayload processes all fields in the Data object
func (c *NotificationConverter) processDataPayload(data *dto.Data) error {
	// Validate required fields
	if data.BundleID == "" {
		return errors.New("missing bundle ID in data")
	}
	if data.BundleVersion == "" {
		return errors.New("missing bundle version in data")
	}
	if data.Environment == "" {
		return errors.New("missing environment in data")
	}

	// Validate environment value
	if data.Environment != dto.EnvironmentSandbox && data.Environment != dto.EnvironmentProduction {
		return fmt.Errorf("invalid environment value: %s", data.Environment)
	}

	// Validate appAppleID if present
	if data.AppAppleID != nil && *data.AppAppleID <= 0 {
		return fmt.Errorf("invalid appAppleId: %d", *data.AppAppleID)
	}

	// Process signed transaction info
	if data.SignedTransactionInfo.JWSTransaction == "" {
		return errors.New("missing signed transaction info in data")
	}

	txPayload, err := c.decodeJWSTransaction(data.SignedTransactionInfo.JWSTransaction)
	if err != nil {
		return fmt.Errorf("failed to decode transaction info: %w", err)
	}
	data.SignedTransactionInfo.JWSTransactionDecodedPayload = *txPayload

	// Validate transaction payload fields
	if err := c.validateTransactionPayload(txPayload); err != nil {
		return fmt.Errorf("invalid transaction payload: %w", err)
	}

	// Process signed renewal info if present
	if data.SignedRenewalInfo != nil && data.SignedRenewalInfo.JWSRenewalInfo != "" {
		renewalPayload, err := c.decodeJWSRenewalInfo(data.SignedRenewalInfo.JWSRenewalInfo)
		if err != nil {
			return fmt.Errorf("failed to decode renewal info: %w", err)
		}
		data.SignedRenewalInfo.JWSRenewalInfoDecodedPayload = *renewalPayload

		// Validate renewal payload fields
		if err := c.validateRenewalPayload(renewalPayload); err != nil {
			return fmt.Errorf("invalid renewal payload: %w", err)
		}
	}

	// Validate status if present
	if data.Status != nil {
		if *data.Status < 1 || *data.Status > 5 {
			return fmt.Errorf("invalid subscription status: %d", *data.Status)
		}
	}

	// Validate consumption request reason if present
	if data.ConsumptionRequestReason != nil {
		switch *data.ConsumptionRequestReason {
		case dto.ConsumptionRequestReasonUnintentedPurchase,
			dto.ConsumptionRequestReasonFullfillmentIssue,
			dto.ConsumptionRequestReasonUnsatisfiedPurchase,
			dto.ConsumptionRequestReasonLegal,
			dto.ConsumptionRequestReasonOther:
			// Valid values
		default:
			return fmt.Errorf("invalid consumption request reason: %s", *data.ConsumptionRequestReason)
		}
	}

	return nil
}
