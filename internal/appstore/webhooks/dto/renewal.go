package dto

import (
	"time"

	"github.com/google/uuid"
)

// JWSRenewalInfo
// Subscription renewal information signed by the App Store, in JSON Web Signature (JWS) format.
// The JWSRenewalInfo type is a string of three Base64 URL-encoded components, separated by a period. The string contains the JWS representation of the subscription renewal information, signed by the App Store according to the JSON Web Signature (JWS) IETF RFC 7515 specification.

// The three components in the string are a header, a payload, and a signature, in that order.

// To read the subscription renewal information, Base64 URL-decode the payload. Use a JWSRenewalInfoDecodedPayload object to read the payload information.

// To read the header, Base64 URL-decode it and use a JWSDecodedHeader object to access the information. Use the information in the header to verify the signature.

type JWSRenewalInfo struct {
	JWSRenewalInfo               string `json:"JWSRenewalInfo"`
	JWSRenewalInfoDecodedPayload JWSRenewalInfoDecodedPayload
}

// A decoded payload containing subscription renewal information for an auto-renewable subscription.
type JWSRenewalInfoDecodedPayload struct {
	AppAccountToken *uuid.UUID `json:"appAccountToken"`
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

	AutoRenewProductId string `json:"autoRenewProductId"`
	// The identifier of the product that renews at the next billing period.

	AutoRenewStatus int32 `json:"autoRenewStatus"`
	// The renewal status for an auto-renewable subscription.
	// Possible Values
	// 0 - Automatic renewal is off. The customer has turned off automatic renewal for the subscription,
	// and it won’t renew at the end of the current subscription period.
	// 1 - Automatic renewal is on. The subscription renews at the end of the current subscription period.

	Currency string `json:"currency"`
	// The three-letter ISO 4217 currency code for the price of the product.
	// The currency property contains an ISO 4217 alpha-3 string that represents the currency of the price of the product.
	// Don’t use the currency value to infer the storefront. Use the storefront value in the transaction instead.

	EligibleWinBackOfferIds []string `json:"eligibleWinBackOfferIds"`
	// An array of win-back offer identifiers that a customer is eligible to redeem,
	// which sorts the identifiers to present the better offers first.

	Environment Environment `json:"environment"`
	// The server environment, either sandbox or production.
	// Possible Values
	// Sandbox - Indicates that the notification applies to testing in the sandbox environment.
	// Production - Indicates that the notification applies to the production environment.
	// You receive notifications in the sandbox environment when you opt in to receive notifications
	// in the sandbox environment and test your app in the sandbox environment.
	// TestFlight also uses the sandbox environment to send notifications.

	ExpirationIntent int32 `json:"expirationIntent"`
	// The reason a subscription expired.
	// 	Possible Values
	// 1 - The customer canceled their subscription.
	// 2 - Billing error; for example, the customer’s payment information is no longer valid.
	// 3 - The customer didn’t consent to an auto-renewable subscription price increase that requires customer consent, allowing the subscription to expire.
	// 4 - The product wasn’t available for purchase at the time of renewal.
	// 5 - The subscription expired for some other reason.

	GracePeriodExpiresDate time.Time `json:"gracePeriodExpiresDate"`
	// The time when the billing grace period for a subscription renewal expires.

	IsInBillingRetryPeriod bool `json:"isInBillingRetryPeriod"`
	// A Boolean value that indicates whether the App Store is attempting to automatically renew a subscription that expired due to a billing issue.

	OfferDiscountType string `json:"offerDiscountType"`
	// The payment mode for a subscription offer for an auto-renewable subscription.
	// 	Possible Values
	// FREE_TRIAL - A payment mode of a product discount that indicates a free trial.
	// PAY_AS_YOU_GO - A payment mode of a product discount that customers pay over a single or multiple billing periods.
	// PAY_UP_FRONT - A payment mode of a product discount that customers pay up front.

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

	OriginalTransactionId string `json:"originalTransactionId"`
	// The original transaction identifier of a purchase.
	// This value is identical to the transaction identifier (transactionId)
	// except when the user restores or renews a subscription.

	PriceIncreaseStatus int32 `json:"priceIncreaseStatus"`
	// The status that indicates whether an auto-renewable subscription is subject to a price increase.
	// 	Possible Values
	// 0 - The customer hasn’t yet responded to an auto-renewable subscription price increase that requires customer consent.
	// 1 - The customer consented to an auto-renewable subscription price increase that requires customer consent,
	// or the App Store has notified the customer of an auto-renewable subscription price increase that doesn’t require consent.

	ProductId string `json:"productId"`
	// The product identifier of the In-App Purchase. You create product identifiers for in-app purchases in App Store Connect.

	RecentSubscriptionStartDate time.Time `json:"recentSubscriptionStartDate"`
	// The earliest start date of a subscription in a series of auto-renewable
	// subscription purchases that ignores all lapses of paid service shorter than 60 days.

	RenewalDate time.Time `json:"renewalDate"`
	// The UNIX time, in milliseconds, when the most recent auto-renewable subscription purchase expires.
	// The renewalDate is a value that’s always present in the payload for auto-renewable subscriptions,
	//  even for expired subscriptions. This date indicates the expiration date of the most
	//  recent auto-renewable subscription purchase, including renewals, and may be in the past.
	// For subscriptions that renew successfully, the renewalDate is the date when the subscription renews.

	RenewalPrice int64 `json:"renewalPrice"`
	// The renewal price, in milliunits, of the auto-renewable subscription that renews at the next billing period.
	// 	This value represents the renewal price, in milliunits of the currency,
	// of the auto-renewable subscription. One unit of the currency equals 1000 milliunits.
	// If the next billing period includes an offer specified by the offerIdentifier,
	// the renewalPrice value reflects the discount.

	SignedDate time.Time `json:"signedDate"`
	// The UNIX time, in milliseconds, that the App Store signed the JSON Web Signature data.

	// AdvancedCommerceRenewalInfo
	// later

}
