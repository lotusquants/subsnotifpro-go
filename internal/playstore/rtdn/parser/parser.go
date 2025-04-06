package parser

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/playstore/rtdn/dto"
)

// ParseWrappedRTDN handles Pub/Sub wrapped RTDN messages
func ParseWrappedRTDN(encodedData string) (*dto.GooglePlayWebhookEvent, error) {
	rawData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		logger.Log.Error("Failed to decode base64 RTDN message")
		return nil, errors.New("invalid base64 encoding")
	}

	return ParseUnwrappedRTDN(rawData) // Reuse unwrapped parser after decoding
}
func ParseUnwrappedRTDN(rawBody []byte) (*dto.GooglePlayWebhookEvent, error) {
	var tempPayload map[string]interface{}
	if err := json.Unmarshal(rawBody, &tempPayload); err != nil {
		return nil, err
	}

	eventTimeMillis, err := ParseEventTimeMillis(tempPayload["eventTimeMillis"])
	if err != nil {
		return nil, err
	}

	rawJSON, _ := json.Marshal(tempPayload)
	event := &dto.GooglePlayWebhookEvent{
		Version:         SafeString(tempPayload["version"]),
		PackageName:     SafeString(tempPayload["packageName"]),
		EventTimeMillis: eventTimeMillis,
		RawPayload:      string(rawJSON),
	}

	// Extract notification types
	if subNotif, exists := tempPayload["subscriptionNotification"]; exists {
		event.Subscription = SafeUnmarshal[dto.SubscriptionNotification](subNotif)
	}
	if oneTimeNotif, exists := tempPayload["oneTimeProductNotification"]; exists {
		event.OneTimeProduct = SafeUnmarshal[dto.OneTimeProductNotification](oneTimeNotif)
	}
	if voidedNotif, exists := tempPayload["voidedPurchaseNotification"]; exists {
		event.VoidedPurchase = SafeUnmarshal[dto.VoidedPurchaseNotification](voidedNotif)
	}
	if testNotif, exists := tempPayload["testNotification"]; exists {
		event.Test = SafeUnmarshal[dto.TestNotification](testNotif)
	}

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
