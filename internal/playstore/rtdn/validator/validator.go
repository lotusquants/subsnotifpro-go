// internal/google_playstore/rtdn/validator/validator.go
package validator

import (
	"errors"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/playstore/rtdn/dto"
)

func ValidateWebhookPayload(event *dto.GooglePlayWebhookEvent) error {
	if event.PackageName == "" {
		return errors.New("package_name is required")
	}
	if event.EventTimeMillis == 0 {
		return errors.New("event_time_millis is required")
	}

	// ✅ Ensure only ONE notification type is present
	notificationCount := 0
	if event.Subscription != nil {
		notificationCount++
	}
	if event.OneTimeProduct != nil {
		notificationCount++
	}
	if event.VoidedPurchase != nil {
		notificationCount++
	}
	if event.Test != nil {
		notificationCount++
	}

	if notificationCount == 0 {
		return errors.New("invalid RTDN event: no notification type found")
	} else if notificationCount > 1 {
		return errors.New("invalid RTDN event: multiple notification types found in a single payload")
	}

	// ✅ Validate Subscription Events
	if event.Subscription != nil {
		if event.Subscription.NotificationType == 0 {
			logger.Log.Warnf("⚠️ Unknown subscription event type: %d", event.Subscription.NotificationType)
		} else {
			if event.Subscription.PurchaseToken == "" {
				return errors.New("subscription_notification: purchase_token is required")
			}
			if event.Subscription.SubscriptionID == "" {
				return errors.New("subscription_notification: subscription_id is required")
			}
		}
	}

	// ✅ Validate One-Time Product Events
	if event.OneTimeProduct != nil {
		if event.OneTimeProduct.NotificationType == 0 {
			logger.Log.Warn("⚠️ Unknown one-time purchase event type:", event.OneTimeProduct.NotificationType)
		} else {
			if event.OneTimeProduct.PurchaseToken == "" {
				return errors.New("one_time_product_notification: purchase_token is required")
			}
			if event.OneTimeProduct.Sku == "" {
				return errors.New("one_time_product_notification: sku is required")
			}
		}
	}

	// ✅ Validate Voided Purchase Events
	if event.VoidedPurchase != nil {
		if event.VoidedPurchase.PurchaseToken == "" {
			return errors.New("voided_purchase_notification: purchase_token is required")
		}
		if event.VoidedPurchase.OrderID == "" {
			return errors.New("voided_purchase_notification: order_id is required")
		}
		if event.VoidedPurchase.ProductType == 0 {
			return errors.New("voided_purchase_notification: product_type is required")
		}
	}

	// ✅ Validate Test Notification
	if event.Test != nil {
		if event.Test.Version == "" {
			return errors.New("test_notification: version is required")
		}
	}

	// ✅ If none of the notifications exist, return an error
	if event.Subscription == nil &&
		event.OneTimeProduct == nil &&
		event.VoidedPurchase == nil &&
		event.Test == nil {
		return errors.New("invalid RTDN event: no notification type found")
	}

	return nil
}
