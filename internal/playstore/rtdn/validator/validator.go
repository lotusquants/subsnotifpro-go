// internal/google_playstore/rtdn/validator/validator.go
package validator

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/playstore/rtdn/models"
)

func ValidateWebhookPayload(event *models.GooglePlayWebhookEvent) error {
	if event.PackageName == "" {
		return errors.New("package_name is required")
	}
	if event.EventTimeMillis == 0 {
		return errors.New("event_time_millis is required")
	}

	// ✅ Ensure only ONE notification type is present
	notificationCount := 0
	if event.SubscriptionNotification != nil {
		notificationCount++
	}
	if event.OneTimeProductNotification != nil {
		notificationCount++
	}
	if event.VoidedPurchaseNotification != nil {
		notificationCount++
	}
	if event.TestNotification != nil {
		notificationCount++
	}

	if notificationCount == 0 {
		return errors.New("invalid RTDN event: no notification type found")
	} else if notificationCount > 1 {
		return errors.New("invalid RTDN event: multiple notification types found in a single payload")
	}

	// ✅ Validate Subscription Events
	if event.SubscriptionNotification != nil {
		if event.SubscriptionNotification.NotificationType == 0 {
			logger.Log.Warnf("⚠️ Unknown subscription event type: %d", event.SubscriptionNotification.NotificationType)
		} else {
			if event.SubscriptionNotification.PurchaseToken == "" {
				return errors.New("subscription_notification: purchase_token is required")
			}
			if event.SubscriptionNotification.SubscriptionID == "" {
				return errors.New("subscription_notification: subscription_id is required")
			}
		}
	}

	// ✅ Validate One-Time Product Events
	if event.OneTimeProductNotification != nil {
		if event.OneTimeProductNotification.NotificationType == 0 {
			logger.Log.Warn("⚠️ Unknown one-time purchase event type:", event.OneTimeProductNotification.NotificationType)
		} else {
			if event.OneTimeProductNotification.PurchaseToken == "" {
				return errors.New("one_time_product_notification: purchase_token is required")
			}
			if event.OneTimeProductNotification.Sku == "" {
				return errors.New("one_time_product_notification: sku is required")
			}
		}
	}

	// ✅ Validate Voided Purchase Events
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
	}

	// ✅ Validate Test Notification
	if event.TestNotification != nil {
		if event.TestNotification.Version == "" {
			return errors.New("test_notification: version is required")
		}
	}

	// ✅ If none of the notifications exist, return an error
	if event.SubscriptionNotification == nil &&
		event.OneTimeProductNotification == nil &&
		event.VoidedPurchaseNotification == nil &&
		event.TestNotification == nil {
		return errors.New("invalid RTDN event: no notification type found")
	}

	return nil
}

// DecodeRTDNMessage decodes the base64-encoded RTDN message
func DecodeRTDNMessage(encodedData string) (models.GooglePlayWebhookEvent, error) {
	rawData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return models.GooglePlayWebhookEvent{}, errors.New("❌ Failed to decode RTDN message from base64")
	}

	var tempPayload map[string]interface{}
	if err := json.Unmarshal(rawData, &tempPayload); err != nil {
		return models.GooglePlayWebhookEvent{}, errors.New("❌ Failed to parse RTDN JSON payload")
	}

	eventTimeInt, err := ParseEventTimeMillis(tempPayload["eventTimeMillis"])
	if err != nil {
		logger.Log.Errorf("❌ eventTimeMillis conversion error: %v", err)
		return models.GooglePlayWebhookEvent{}, errors.New("❌ Invalid eventTimeMillis format")
	}

	event := models.GooglePlayWebhookEvent{
		Version:         SafeString(tempPayload["version"]),
		PackageName:     SafeString(tempPayload["packageName"]),
		EventTimeMillis: eventTimeInt,
		RawPayload:      string(rawData),
		Status:          "pending",
		RetryCount:      0,
	}

	// ✅ Use SafeUnmarshal to avoid panics
	event.SubscriptionNotification = SafeUnmarshal[models.SubscriptionNotification](tempPayload["subscriptionNotification"])
	event.OneTimeProductNotification = SafeUnmarshal[models.OneTimeProductNotification](tempPayload["oneTimeProductNotification"])
	event.VoidedPurchaseNotification = SafeUnmarshal[models.VoidedPurchaseNotification](tempPayload["voidedPurchaseNotification"])
	event.TestNotification = SafeUnmarshal[models.TestNotification](tempPayload["testNotification"])

	return event, nil
}

// ✅ Generic function to safely unmarshal JSON into a struct
func SafeUnmarshal[T any](data interface{}) *T {
	if data == nil {
		return nil
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	var result T
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil
	}
	return &result
}

// ✅ Helper function to safely extract string fields
func SafeString(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// ✅ Safely parse eventTimeMillis into int64
func ParseEventTimeMillis(value interface{}) (int64, error) {
	switch v := value.(type) {
	case string:
		parsedInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, errors.New("❌ Failed to parse eventTimeMillis as string")
		}
		return parsedInt, nil
	case float64:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, errors.New("❌ eventTimeMillis is of unknown type")
	}
}
