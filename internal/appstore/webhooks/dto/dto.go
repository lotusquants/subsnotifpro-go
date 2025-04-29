package dto

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AppStoreNotification represents the fully decoded notification
type AppStoreNotification struct {

	// System-generated fields
	ID         uuid.UUID `json:"Id" validate:"required"` // Primary identifier
	ReceivedAt time.Time `json:"receivedAt" validate:"required"`

	ResponseBodyV2DecodedPayload ResponseBodyV2DecodedPayload

	JWSDecodedHeader JWSDecodedHeader
}

// ResponseBodyV2 represents the top-level response from App Store Server Notifications V2
// The signedPayload object is a JWS representation. To get the transaction and subscription
// renewal details from the notification payload, process the signedPayload as follows:
// Parse signedPayload to identify the JWS header, payload, and signature representations.
// Base64URL-decode the payload to get the responseBodyV2DecodedPayload.
// The decoded payload contains details of the notification such as the notification type and data.
// The data object contains a signedTransactionInfo (JWSTransaction) and based
// on the notification type, a signedRenewalInfo (JWSRenewalInfo).
// Parse and Base64URL-decode these signed JWS representations to get transaction and subscription renewal details.
// Each of the signed JWS representations, signedPayload, signedTransactionInfo, and signedRenewalInfo,
//
//	have a JWS signature that you can validate on your server. Use the algorithm specified in the header’s
//	alg parameter to validate the signature. For more information about validating signatures,
//	see the JSON Web Signature (JWS) IETF RFC 7515 specification.
type ResponseBodyV2 struct {
	SignedPayload string `json:"signedPayload"`
	// A cryptographically signed payload, in JSON Web Signature (JWS) format, that contains the response body for a version 2 notification.
	// The signedPayload is a string of three Base64URL-encoded components, separated by a period.
	// The string contains a JWS representation of the notification response body,
	// signed by the App Store according to the JSON Web Signature (JWS) IETF RFC 7515 specification.
	// The three components of the string are a header, a payload, and a signature, in that order.
	// To read the notification response body, Base64URL-decode the payload.
	//  Use a responseBodyV2DecodedPayload object to read the payload information.
	// To read the header, Base64URL-decode it and use a JWSDecodedHeader object to access the information.
	// Use the information in the decoded header to verify the signature.

}

// ResponseBodyV2DecodedPayload represents the decoded payload from the JWS
type ResponseBodyV2DecodedPayload struct {
	NotificationType NotificationType `json:"notificationType"`
	// The type that describes the in-app purchase or external purchase event
	// for which the App Store sends the version 2 notification.

	Subtype *NotificationSubtype `json:"subtype,omitempty"`
	// Additional information that identifies the notification event.
	// The subtype field is present only for specific version 2 notifications.

	// Only one of these fields will be present
	Data *Data `json:"data,omitempty"`
	//The object that contains the app metadata and signed renewal and transaction information.
	// The data, summary, and externalPurchaseToken fields are mutually exclusive.
	// The payload contains only one of these fields.
	// The data object is part of the responseBodyV2DecodedPayload.
	// It’s present in the payload for notificationType values related to in-app purchases,
	// except for the RENEWAL_EXTENSION notification type with a SUMMARY subtype,
	//  and the EXTERNAL_PURCHASE_TOKEN notification type.
	// Use the notification type along with the transaction and subscription renewal information
	// in the data object to update a user’s service or present promotional offers according to your business logic.

	Summary *Summary `json:"summary,omitempty"`
	// The payload data for a subscription-renewal-date extension notification.
	// The summary object appears in the responseBodyV2DecodedPayload when the notificationType is
	// RENEWAL_EXTENSION and the subtype is SUMMARY.
	// This notification occurs when the App Store completes your request to extend the subscription
	// renewal date for eligible subscribers.

	ExternalPurchaseToken *string `json:"externalPurchaseToken,omitempty"`

	Version string `json:"version"`
	// A string that indicates the notification’s App Store Server Notifications version number.

	SignedDate int64 `json:"signedDate"`
	// The UNIX time, in milliseconds, that the App Store signed the JSON Web Signature data.

	NotificationUUID string `json:"notificationUUID"`
	// A unique identifier for the notification.
	//The App Store server assigns a unique identifer to each notification it sends. Use this value to identify, and ignore, duplicate notifications.

}

// In your dto package (where ResponseBodyV2DecodedPayload is defined)

// Scan implements the sql.Scanner interface
func (r *ResponseBodyV2DecodedPayload) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}
	return json.Unmarshal(b, r)
}

// Value implements the driver.Valuer interface
func (r ResponseBodyV2DecodedPayload) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// JWSDecodedHeader represents the decoded JWS header
// All JWS representations, including the signedPayload, contain a JWS header.
// When you Base64 URL-decode the header, use the JWSDecodedHeader object to read its contents.
//
//	Use the information in the JWSDecodedHeader to validate the JWS signature.
//	For more information about validating signatures, see the JSON Web Signature (JWS) IETF RFC 7515 specification.
//
// The App Store signs transaction and renewal information that you receive in App Store Server Notifications V2
// and in the App Store Server API. It uses the following x5c certificate chain, in order:
// A certificate that contains the public key that corresponds to the key the App Store uses to digitally sign the JWS.
//
//	Section 4.11.10 Mac App Store Receipt Signing Certificates of the Apple Inc.
//
// Certificate Practice Statement Worldwide Developer Relations document defines the policy for this certificate.
// An Apple intermediate certificate from the Apple PKI site that starts with Worldwide Developer Relations.
// An Apple root certificate.
type JWSDecodedHeader struct {
	Alg string `json:"alg"`
	// alg - The JSON Web Signature (JWS) header parameter that identifies the cryptographic algorithm used to secure the JWS.

	X5c []string `json:"x5c"`
	// The JSON Web Signature (JWS) header parameter that contains the certificate chain that corresponds to the key used to digitally sign the JWS.
}

// Data represents the data object in the decoded payload
type Data struct {
	AppAppleID    *int64      `json:"appAppleId,omitempty"` // The unique identifier of an app in the App Store.
	BundleID      string      `json:"bundleId"`             // The bundle identifier of an app.
	BundleVersion string      `json:"bundleVersion"`        // The version of the build that identifies an iteration of the bundle.
	Environment   Environment `json:"environment"`          // The server environment that the notification applies to, either sandbox or production.

	SignedTransactionInfo JWSTransaction `json:"signedTransactionInfo"`
	// Transaction information signed by the App Store, in JSON Web Signature (JWS) format.

	SignedRenewalInfo *JWSRenewalInfo `json:"signedRenewalInfo,omitempty"`
	// The JWSRenewalInfo type is a string of three Base64 URL-encoded components, separated by a period.
	// The string contains the JWS representation of the subscription renewal information,
	// signed by the App Store according to the JSON Web Signature (JWS) IETF RFC 7515 specification.
	// The three components in the string are a header, a payload, and a signature, in that order.
	// To read the subscription renewal information, Base64 URL-decode the payload.
	// Use a JWSRenewalInfoDecodedPayload object to read the payload information.
	// To read the header, Base64 URL-decode it and use a JWSDecodedHeader object to access the information.
	//  Use the information in the header to verify the signature.

	Status *SubscriptionStatus `json:"status,omitempty"`
	// the status of an auto-renewable subscription at the time the App Store signs the notification.
	// Possible Values
	// 1 : The auto-renewable subscription is active.
	// 2 : The auto-renewable subscription is expired.
	// 3 : The auto-renewable subscription is in a billing retry period.
	// 4 : The auto-renewable subscription is in a Billing Grace Period.
	// 5 : The auto-renewable subscription is revoked.
	// This status value is current as of the signedDate in the decoded payload, responseBodyV2DecodedPayload.

	ConsumptionRequestReason *ConsumptionRequestReason `json:"consumptionRequestReason,omitempty"`
	// Possible Values
	// UNINTENDED_PURCHASE : The customer didn’t intend to make the in-app purchase.
	// FULFILLMENT_ISSUE : The customer had issues with receiving or using the in-app purchase.
	// UNSATISFIED_WITH_PURCHASE : The customer wasn’t satisfied with the in-app purchase.
	// LEGAL : The customer requested a refund based on a legal reason.
	// OTHER : The customer requested a refund for other reasons.
	// When a customer initiates a refund request for a consumable in-app purchase or auto-renewable subscription,
	// the App Store sends a CONSUMPTION_REQUEST notificationType to your server.
	// The notification includes the consumptionRequestReason in the data object.
}

// Summary represents the summary object for RENEWAL_EXTENSION notifications
type Summary struct {
	RequestIdentifier string `json:"requestIdentifier"`
	// A string that contains a unique identifier for a subscription-renewal-date extension request.
	Environment Environment `json:"environment"`

	AppAppleId int64 `json:"appAppleId"` // The unique identifier of an app in the App Store.

	BundleID string `json:"bundleId"` // The bundle identifier of an app.

	ProductId string `json:"productId"`
	// The product identifier of the In-App Purchase. You create product identifiers for in-app
	// purchases in App Store Connect.

	StorefrontCountryCodes []string `json:"storefrontCountryCodes"`
	// A list of storefront country codes for limiting the storefronts for a subscription-renewal-date extension.
	// The three-letter code that represents the country or region associated with the App Store storefront.

	FailedCount int64 `json:"failedCount"`
	// The final count of subscriptions that fail to receive a subscription-renewal-date extension.

	SucceededCount int64 `json:"succeededCount"`
	// The count of subscriptions that successfully receive a subscription-renewal-date extension.
}

type ConsumptionRequestReason string

const (
	ConsumptionRequestReasonUnintentedPurchase  ConsumptionRequestReason = "UNINTENDED_PURCHASE"
	ConsumptionRequestReasonFullfillmentIssue   ConsumptionRequestReason = "FULFILLMENT_ISSUE"
	ConsumptionRequestReasonUnsatisfiedPurchase ConsumptionRequestReason = "UNSATISFIED_WITH_PURCHASE"
	ConsumptionRequestReasonLegal               ConsumptionRequestReason = "LEGAL"
	ConsumptionRequestReasonOther               ConsumptionRequestReason = "OTHER"
)

type Environment string

const (
	EnvironmentSandbox    Environment = "Sandbox"
	EnvironmentProduction Environment = "Production"
)
