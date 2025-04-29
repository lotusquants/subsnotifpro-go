package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func main() {
	// Configuration - update these values as needed
	targetURL := "http://localhost:8080/api/app-store/webhooks" // Updated to match your endpoint
	bundleID := "com.your.app"
	productID := "com.your.app.basic"
	originalTransactionID := "1000000123456009"
	transactionID := "1000000123456795"
	renewalTransactionId := "1000000123456793"
	appAppleID := int64(123456789)
	appAccountToken := "79a64a4a-0b69-47df-97f7-9a9a735f5bc5"

	// Test 1: Initial Purchase Notification
	fmt.Println("Sending INITIAL_BUY notification...")

	initialBuyPayload := createInitialBuyNotification(bundleID, productID, originalTransactionID, transactionID, appAppleID, appAccountToken)
	sendNotification(targetURL, initialBuyPayload)

	// Test 2: Renewal Notification
	fmt.Println("\nSending RENEWAL notification...")
	renewalPayload := createRenewalNotification(bundleID, productID, originalTransactionID, renewalTransactionId, appAppleID, appAccountToken)
	sendNotification(targetURL, renewalPayload)
}

// createInitialBuyNotification creates a properly structured INITIAL_BUY notification
func createInitialBuyNotification(bundleID, productID, originalTxnID, txnID string, appAppleID int64, appAccountToken string) map[string]interface{} {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	purchaseDate := now - 86400000  // 1 day ago
	expiresDate := now + 2592000000 // 30 days from now

	// Create the transaction info payload
	transactionInfo := map[string]interface{}{
		"appAccountToken":             appAccountToken,
		"transactionId":               txnID,
		"originalTransactionId":       originalTxnID,
		"bundleId":                    bundleID,
		"productId":                   productID,
		"purchaseDate":                purchaseDate,
		"expiresDate":                 expiresDate,
		"quantity":                    1,
		"type":                        "Auto-Renewable Subscription",
		"inAppOwnershipType":          "PURCHASED",
		"signedDate":                  now,
		"environment":                 "Sandbox",
		"storefront":                  "USA",
		"storefrontId":                "143441",
		"currency":                    "USD",
		"price":                       999,
		"subscriptionGroupIdentifier": "group_123",
	}

	// Create the main notification payload
	notification := map[string]interface{}{
		"notificationType": "SUBSCRIBED",
		"subtype":          "INITIAL_BUY",
		"version":          "2.0",
		"signedDate":       now,
		"notificationUUID": uuid.New().String(),
		"data": map[string]interface{}{

			"appAppleId":    appAppleID,
			"bundleId":      bundleID,
			"bundleVersion": "1.0",
			"environment":   "Sandbox",
			"signedTransactionInfo": map[string]interface{}{
				"JWSTransaction": createJWSPayload(transactionInfo),
			},
		},
	}

	return map[string]interface{}{
		"signedPayload": createJWSPayload(notification),
	}
}

// createRenewalNotification creates a properly structured RENEWAL notification
func createRenewalNotification(bundleID, productID, originalTxnID, txnID string, appAppleID int64, appAccountToken string) map[string]interface{} {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	purchaseDate := now - 86400000           // 1 day ago
	expiresDate := now + 3*2592000000        // 30 days from now
	originalPurchaseDate := now - 2592000000 // 30 days ago

	// Create the transaction info payload
	transactionInfo := map[string]interface{}{
		"appAccountToken":             appAccountToken,
		"transactionId":               txnID,
		"originalTransactionId":       originalTxnID,
		"bundleId":                    bundleID,
		"productId":                   productID,
		"purchaseDate":                purchaseDate,
		"expiresDate":                 expiresDate,
		"originalPurchaseDate":        originalPurchaseDate,
		"quantity":                    1,
		"type":                        "Auto-Renewable Subscription",
		"inAppOwnershipType":          "PURCHASED",
		"signedDate":                  now,
		"environment":                 "Sandbox",
		"storefront":                  "USA",
		"storefrontId":                "143441",
		"currency":                    "USD",
		"price":                       999,
		"subscriptionGroupIdentifier": "group_123",
		"transactionReason":           "RENEWAL",
	}

	// Create the renewal info payload
	renewalInfo := map[string]interface{}{
		"autoRenewProductId":    productID,
		"autoRenewStatus":       1,
		"productId":             productID,
		"originalTransactionId": originalTxnID,
		"expirationIntent":      0,
		"signedDate":            now,
		"environment":           "Sandbox",
		"renewalDate":           expiresDate,
		"renewalPrice":          999,
	}

	// Create the main notification payload
	notification := map[string]interface{}{
		"notificationType": "DID_RENEW",
		"version":          "2.0",
		"signedDate":       now,
		"notificationUUID": uuid.New().String(),
		"data": map[string]interface{}{
			"appAppleId":    appAppleID,
			"bundleId":      bundleID,
			"bundleVersion": "1.0",
			"environment":   "Sandbox",
			"signedTransactionInfo": map[string]interface{}{
				"JWSTransaction": createJWSPayload(transactionInfo),
			},

			"signedRenewalInfo": map[string]interface{}{
				"JWSRenewalInfo": createJWSPayload(renewalInfo),
			},
		},
	}

	return map[string]interface{}{
		"signedPayload": createJWSPayload(notification),
	}
}

// createJWSPayload creates a mock JWS payload (header.payload.signature)
func createJWSPayload(payload map[string]interface{}) string {
	header := map[string]interface{}{
		"alg": "ES256",
		"x5c": []string{"MIIE...", "MIIC..."},
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signature := "mock-signature"

	return fmt.Sprintf("%s.%s.%s", headerB64, payloadB64, signature)
}

// sendNotification sends the notification to the target endpoint
func sendNotification(targetURL string, payload map[string]interface{}) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	resp, err := http.Post(targetURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Response Status: %s\n", resp.Status)

	// Print response body if available
	if resp.Body != nil {
		var body bytes.Buffer
		body.ReadFrom(resp.Body)
		if body.Len() > 0 {
			fmt.Printf("Response Body: %s\n", body.String())
		}
	}
}
