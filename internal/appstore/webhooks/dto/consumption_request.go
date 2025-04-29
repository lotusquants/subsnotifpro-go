// internal/appstore/webhooks/dto/consumption_request.go
package dto

type ConsumptionRequest struct {
	AccountTenure            int32   `json:"accountTenure"`
	AppAccountToken          *string `json:"appAccountToken,omitempty"`
	ConsumptionStatus        int32   `json:"consumptionStatus"`
	CustomerConsented        bool    `json:"customerConsented"`
	DeliveryStatus           int32   `json:"deliveryStatus"`
	LifetimeDollarsPurchased string  `json:"lifetimeDollarsPurchased"`
	LifetimeDollarsRefunded  string  `json:"lifetimeDollarsRefunded"`
	Platform                 int32   `json:"platform"`
	PlayTime                 int32   `json:"playTime"`
	SampleContentProvided    bool    `json:"sampleContentProvided"`
	UserStatus               int32   `json:"userStatus"`
	RefundPreference         *int32  `json:"refundPreference,omitempty"`
}

// Consumption status values
const (
	ConsumptionStatusUndeclared        = 0
	ConsumptionStatusNotConsumed       = 1
	ConsumptionStatusPartiallyConsumed = 2
	ConsumptionStatusFullyConsumed     = 3
)

// Delivery status values
const (
	DeliveryStatusUndeclared   = 0
	DeliveryStatusDelivered    = 1
	DeliveryStatusNotDelivered = 2
)

// Platform values
const (
	PlatformUndeclared = 0
	PlatformApple      = 1
	PlatformNonApple   = 2
)

// User status values
const (
	UserStatusUndeclared    = 0
	UserStatusActive        = 1
	UserStatusSuspended     = 2
	UserStatusTerminated    = 3
	UserStatusLimitedAccess = 4
)

// Refund preference values
const (
	RefundPreferenceUndeclared   = 0
	RefundPreferenceRequested    = 1
	RefundPreferenceNotRequested = 2
)
