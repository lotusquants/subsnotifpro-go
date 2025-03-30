package parser

import (
	"encoding/json"
	rtdnModels "subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/rtdn/validator"
)

// parseUnwrappedRTDN parses an unwrapped RTDN event directly from JSON body.
func ParseUnwrappedRTDN(rawBody []byte) (rtdnModels.GooglePlayWebhookEvent, error) {
	var tempPayload map[string]interface{}
	if err := json.Unmarshal(rawBody, &tempPayload); err != nil {
		return rtdnModels.GooglePlayWebhookEvent{}, err
	}

	eventTimeMillis, err := validator.ParseEventTimeMillis(tempPayload["eventTimeMillis"])
	if err != nil {
		return rtdnModels.GooglePlayWebhookEvent{}, err
	}

	rawJSON, _ := json.Marshal(tempPayload)
	event := rtdnModels.GooglePlayWebhookEvent{
		Version:         validator.SafeString(tempPayload["version"]),
		PackageName:     validator.SafeString(tempPayload["packageName"]),
		EventTimeMillis: eventTimeMillis,
		RawPayload:      string(rawJSON),
		Status:          "pending",
		RetryCount:      0,
	}

	// Extract the correct notification type
	if subNotif, exists := tempPayload["subscriptionNotification"]; exists {
		event.SubscriptionNotification = validator.SafeUnmarshal[rtdnModels.SubscriptionNotification](subNotif)
	}
	if oneTimeNotif, exists := tempPayload["oneTimeProductNotification"]; exists {
		event.OneTimeProductNotification = validator.SafeUnmarshal[rtdnModels.OneTimeProductNotification](oneTimeNotif)
	}
	if voidedNotif, exists := tempPayload["voidedPurchaseNotification"]; exists {
		event.VoidedPurchaseNotification = validator.SafeUnmarshal[rtdnModels.VoidedPurchaseNotification](voidedNotif)
	}
	if testNotif, exists := tempPayload["testNotification"]; exists {
		event.TestNotification = validator.SafeUnmarshal[rtdnModels.TestNotification](testNotif)
	}

	return event, nil
}
