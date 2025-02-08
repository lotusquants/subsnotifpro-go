// internal/google_playstore/rtdn/validator/validator.go
package validator

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"subsnotifpro-go/internal/google_playstore/models"
)

// ValidateWebhookPayload ensures required fields exist in the RTDN payload
func ValidateWebhookPayload(event *models.GooglePlayWebhookEvent) error {
	if event.PackageName == "" {
		return errors.New("package_name is required")
	}
	if event.EventTimeMillis == 0 {
		return errors.New("event_time_millis is required")
	}

	// Validate Subscription Events
	if event.SubscriptionNotification != nil {
		if event.SubscriptionNotification.NotificationType == 0 {
			return errors.New("subscription_notification: notification_type is required")
		}
		if event.SubscriptionNotification.PurchaseToken == "" {
			return errors.New("subscription_notification: purchase_token is required")
		}
		if event.SubscriptionNotification.SubscriptionID == "" {
			return errors.New("subscription_notification: subscription_id is required")
		}
		return nil
	}

	// Validate One-Time Product Events
	if event.OneTimeProductNotification != nil {
		if event.OneTimeProductNotification.NotificationType == 0 {
			return errors.New("one_time_product_notification: notification_type is required")
		}
		if event.OneTimeProductNotification.PurchaseToken == "" {
			return errors.New("one_time_product_notification: purchase_token is required")
		}
		if event.OneTimeProductNotification.Sku == "" {
			return errors.New("one_time_product_notification: sku is required")
		}
		return nil
	}

	// Validate Voided Purchase Events
	if event.VoidedPurchaseNotification != nil {
		if event.VoidedPurchaseNotification.PurchaseToken == "" {
			return errors.New("voided_purchase_notification: purchase_token is required")
		}
		if event.VoidedPurchaseNotification.OrderID == "" {
			return errors.New("voided_purchase_notification: order_id is required")
		}
		if event.VoidedPurchaseNotification.ProductType == 0 {
			return errors.New("voided_purchase_notification: product_type is required")
		}
		return nil
	}

	// Validate Test Notification
	if event.TestNotification != nil {
		if event.TestNotification.Version == "" {
			return errors.New("test_notification: version is required")
		}
		return nil
	}

	return errors.New("invalid RTDN event structure: no valid notification type found")
}

// DecodeRTDNMessage decodes the base64-encoded RTDN message using streaming JSON decoding.
func DecodeRTDNMessage(encodedData string) (models.GooglePlayWebhookEvent, error) {
	// Step 1: Decode base64 payload
	rawData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return models.GooglePlayWebhookEvent{}, errors.New("❌ Failed to decode RTDN message from base64")
	}

	// Step 2: Create a streaming JSON decoder
	decoder := json.NewDecoder(bytes.NewReader(rawData))

	// Step 3: Decode the JSON incrementally
	var event models.GooglePlayWebhookEvent
	if err := decoder.Decode(&event); err != nil && err != io.EOF {
		return models.GooglePlayWebhookEvent{}, errors.New("❌ Failed to parse RTDN JSON payload")
	}

	// Step 4: Ensure valid event
	if event.PackageName == "" && event.SubscriptionNotification == nil &&
		event.OneTimeProductNotification == nil && event.VoidedPurchaseNotification == nil &&
		event.TestNotification == nil {
		return models.GooglePlayWebhookEvent{}, errors.New("⚠️ RTDN event is empty or missing required fields")
	}

	return event, nil
}
