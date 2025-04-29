package dto

import (
	"time"
)

// JWSTransaction represents the decoded transaction info
// Transaction information signed by the App Store, in JSON Web Signature (JWS) Compact Serialization format.
// The JWSTransaction type is a string of three Base64URL-encoded components separated by a period. The string contains the JWS Compact Serialization of the transaction information, signed by the App Store according to the JSON Web Signature (JWS) IETF RFC 7515 specification.
// The three components of the string are a header, a payload, and a signature, in that order.
// To read the transaction information, Base64URL-decode the payload. Use a JWSTransactionDecodedPayload object to read the payload information.
// To read the header, decode it and use a JWSDecodedHeader object to access the information. Use the information in the header to verify the signature.
type JWSTransaction struct {
	JWSTransaction               string `json:"JWSTransaction"`
	JWSTransactionDecodedPayload JWSTransactionDecodedPayload
}

// A decoded payload that contains transaction information.
type JWSTransactionDecodedPayload struct {
	AppAccountToken *string `json:"appAccountToken"`
	// A UUID you create at the time of purchase that associates the transaction with a customer on your own service.
	// If your app doesn’t provide an appAccountToken, this field is omitted.
	// When a customer initiates an in-app purchase, your app may create an appAccountToken(_:) and
	// send it to the App Store. The App Store returns the same value in
	// appAccountToken in the transaction information after the customer completes the purchase.
	// If you’re using the Original API for In-App Purchase and provide a UUID in the applicationUsername property,
	// then the appAccountToken field contains that value.

	AppTransactionId string `json:"appTransactionId"`
	// The unique identifier of the app download transaction.
	// The App Store generates a single, globally unique appTransactionID for each Apple Account that downloads
	// your app and for each family group member for apps that support Family Sharing.
	// This value remains the same for the same Apple Account and app if the customer
	// redownloads the app on any device, receives a refund, repurchases the app,
	// or changes the storefront. For apps that support Family Sharing, the appTransactionID is unique
	// for each family group member.
	// The appTransactionID is available even if a customer makes no in-app purchases.

	BundleId string `json:"bundleId"` // The bundle identifier of the app.

	Currency string `json:"currency"`
	// The three-letter ISO 4217 currency code for the price of the product.
	// The currency property contains an ISO 4217 alpha-3 string that represents the currency of the price of the product.
	// Don’t use the currency value to infer the storefront. Use the storefront value in the transaction instead.

	Environment Environment `json:"environment"`
	// The server environment, either sandbox or production.
	// Possible Values
	// Sandbox - Indicates that the notification applies to testing in the sandbox environment.
	// Production - Indicates that the notification applies to the production environment.
	// You receive notifications in the sandbox environment when you opt in to receive notifications
	// in the sandbox environment and test your app in the sandbox environment.
	// TestFlight also uses the sandbox environment to send notifications.

	ExpiresDate time.Time `json:"expiresDate"`
	// The UNIX time, in milliseconds, an auto-renewable subscription purchase expires or renews.
	// ( converted to time.time in converter)
	// The expiresDate is a static value that applies for each transaction.
	// When the auto-renewable subscription renews, the App Store creates a new transaction with a new expiresDate.

	InAppOwnershipType string `json:"inAppOwnershipType"`
	// A string that describes whether the transaction was purchased by the customer,
	// or is available to them through Family Sharing.
	// Possible Values
	// FAMILY_SHARED - The transaction belongs to a family member who benefits from the service.
	// PURCHASED - The transaction belongs to the purchaser.

	IsUpgraded bool `json:"isUpgraded"`
	// A Boolean value that indicates whether the customer upgraded to another subscription.
	// If isUpgraded is true, the customer has upgraded the subscription represented by this
	// transaction to another subscription. This value appears in the transaction only when the value is true.
	// To determine the service that the customer is entitled to, look for another transaction that has a
	// subscription with a higher level of service.

	OfferDiscountType string `json:"offerDiscountType"`
	// The payment mode for a subscription offer for an auto-renewable subscription.
	// Possible Values
	// FREE_TRIAL - A payment mode of a product discount that indicates a free trial.
	// PAY_AS_YOU_GO - A payment mode of a product discount that customers pay over a single or multiple billing periods.
	// PAY_UP_FRONT - A payment mode of a product discount that customers pay up front.
	// You set up subscription offers and determine the payment mode when you configure subscriptions in App Store Connect.

	OfferIdentifier string `json:"offerIdentifier"`
	// The string identifier of a subscription offer that you create in App Store Connect.
	// The offerIdentifier is a string that you provide in App Store Connect when you set up a subscription offer.
	// All offer types (offerType) have offer identifiers, except for introductory offers.

	OfferPeriod string `json:"offerPeriod"` // The duration of the offer. This field is in ISO 8601 duration format.

	OfferType int32 `json:"offerType"`
	// The type of subscription offer.
	// Possible Values
	// 1 - An introductory offer.
	// 2 - A promotional offer.
	// 3 - An offer with a subscription offer code.
	// 4 - A win-back offer.

	OriginalPurchaseDate time.Time `json:"originalPurchaseDate"`
	// The purchase date of the transaction associated with the original transaction identifier.
	// The original purchase date is in UNIX time, in milliseconds.( converted to time.time)

	OriginalTransactionId string `json:"originalTransactionId"`
	// The original transaction identifier of a purchase.
	// This value is identical to the transaction identifier (transactionId)
	// except when the user restores or renews a subscription.

	PreviousOriginalTransactionId *string `json:"previousOriginalTransactionId"`
	// The original transaction identifer of a subscription before migration.

	Price int64 `json:"price"`
	// The price, in milliunits, of the In-App Purchase that the system records in the transaction.
	// This value represents the price, in milliunits of the currency, of the In-App Purchase that the system
	// records in the transaction. One unit of the currency equals 1000 milliunits.
	// The price value reflects all of the following:
	// The price you configured in App Store Connect, which the system records on the purchase date (purchaseDate).
	// The discount from a subscription offer in the offerIdentifier, if the transaction includes an offer.
	// The quantity of a consumable in-app purchase. The price value shows the total amount of the transaction
	//  for the quantity the customer purchased.

	ProductId string `json:"productId"`
	// The product identifier of the In-App Purchase. You create product identifiers for in-app purchases in App Store Connect.

	PurchaseDate time.Time `json:"purchaseDate"`
	// The time that the App Store charged the customer’s account for a purchase,
	//  a restored product, a subscription, or a subscription renewal after a lapse.
	// The purchase date is in UNIX time, in milliseconds.( converted to time.Time in converter)

	Quantity int32 `json:"quantity"`
	// The number of purchased consumable products.

	RevocationDate *time.Time `json:"revocationDate"`
	// The UNIX time, in milliseconds, that the App Store refunded the transaction or revoked it from Family Sharing.

	RevocationReason *int32 `json:"revocationReason"`
	// The reason for a refunded transaction.
	// Possible Values
	// 0 - The App Store refunded the transaction on behalf of the customer for other reasons,
	// for example, an accidental purchase.
	// 1 - The App Store refunded the transaction on behalf of the customer due to an actual
	// or perceived issue within your app.

	SignedDate time.Time `json:"signedDate"`
	// The UNIX time, in milliseconds, that the App Store signed the JSON Web Signature data.
	// converted to time.Time in converter

	Storefront string `json:"storefront"`
	// The three-letter code that represents the country or region associated with the App Store storefront for the purchase.
	// This property uses the ISO 3166-1 alpha-3 country code representation. This property is the same as the countryCode in StoreKit.

	StorefrontId string `json:"storefrontId"`
	// An Apple-defined value that uniquely identifies an App Store storefront.This value is the same as the id value in StoreKit.

	SubscriptionGroupIdentifier string `json:"subscriptionGroupIdentifier"`
	// The identifier of the subscription group that the subscription belongs to.
	// Auto-renewable subscriptions always belong to a subscription group. You create the subscription group identifiers
	// in App Store Connect before you create and add an auto-renewable subscription.

	TransactionId string `json:"transactionId"`
	// The unique identifier for a transaction, such as an In-App Purchase, restored purchase, or subscription renewal.
	// The App Store generates a new value for transaction identifier every time
	// the subscription automatically renews or the user restores it on a new device.
	// When a user first purchases a subscription, the transaction identifier always matches
	// the original transaction identifier (originalTransactionId).
	// For a restore or renewal, the transaction identifier doesn’t match the original transaction identifier.
	// If a user restores or renews the same subscription multiple times,
	// each restore or renewal has a unique transaction identifier.

	TransactionReason string `json:"transactionReason"`
	// The cause of a purchase transaction, which indicates whether it’s a
	// customer’s purchase or a renewal for an auto-renewable subscription that the system initiates.
	// Possible Values
	// PURCHASE - The customer initiated the purchase, which may be for any in-app purchase type: consumable, non-consumable, non-renewing subscription, or auto-renewable subscription.
	// RENEWAL - The App Store server initiated the purchase transaction to renew an auto-renewable subscription.
	// If a customer upgrades an auto-renewable subscription, the upgrade is effective immediately
	// and the transactionReason is PURCHASE.
	// If a customer downgrades an auto-renewable subscription, the product change occurs on the
	// subscription renewal date. The resulting transactionReason is RENEWAL.

	Type string `json:"type"`
	// The product type of the In-App Purchase.
	// Possible Values
	// Auto-Renewable Subscription - An auto-renewable subscription.
	// Non-Consumable - A non-consumable In-App Purchase.
	// Consumable - A consumable In-App Purchase.
	// Non-Renewing Subscription - A non-renewing subscription.

	WebOrderLineItemId string `json:"webOrderLineItemId"`
	// The unique identifier of subscription purchase events across devices, including subscription renewals.
	// This value applies only to auto-renewable subscriptions.

	// AdvancedCommerceTransactionInfo AdvancedCommerceTransactionInfo
	// will do later

}
