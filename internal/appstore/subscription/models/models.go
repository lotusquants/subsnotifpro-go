package models

import (
	"fmt"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"subsnotifpro-go/internal/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Environment string

const (
	EnvironmentSandbox    Environment = "SANDBOX"
	EnvironmentProduction Environment = "PRODUCTION"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive           SubscriptionStatus = "ACTIVE"
	SubscriptionStatusExpired          SubscriptionStatus = "EXPIRED"
	SubscriptionStatusBillingRetry     SubscriptionStatus = "BILLING_RETRY"
	SubscriptionStatusGracePeriod      SubscriptionStatus = "GRACE_PERIOD"
	SubscriptionStatusRevoked          SubscriptionStatus = "REVOKED"
	SubscriptionStatusPendingRenewal   SubscriptionStatus = "PENDING_RENEWAL"
	SubscriptionStatusPendingUpgrade   SubscriptionStatus = "PENDING_UPGRADE"
	SubscriptionStatusPendingDowngrade SubscriptionStatus = "PENDING_DOWNGRADE"
)

type AutoRenewStatus string

const (
	AutoRenewOn  AutoRenewStatus = "ON"
	AutoRenewOff AutoRenewStatus = "OFF"
)

func MapAutoRenewStatus(autoRenewStatus int32) AutoRenewStatus {
	switch autoRenewStatus {
	case 1:
		return AutoRenewOn
	case 0:
		return AutoRenewOff
	default:
		// Handle unexpected values - you might want to log this
		return AutoRenewOff // Default to OFF for safety
	}
}

type InAppOwnershipType string

const (
	OwnershipPurchased    InAppOwnershipType = "PURCHASED"
	OwnershipFamilyShared InAppOwnershipType = "FAMILY_SHARED"
)

// Subscription represents the current state of a user's subscription
type AppStoreSubscription struct {
	gorm.Model
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OriginalTransactionID string    `gorm:"size:128;uniqueIndex:idx_original_transaction"`
	CurrentTransactionID  string    `gorm:"size:128;index"`
	WebOrderLineItemID    string    `gorm:"size:128;index"`

	BundleID string `gorm:"size:128;index"`

	// Relationships
	UserID              uuid.UUID `gorm:"type:uuid;index"` // Your system's user ID
	ProductID           string    `gorm:"size:128;index"`
	SubscriptionGroupID string    `gorm:"size:128;index"`

	// Core state
	Status          SubscriptionStatus `gorm:"size:32;index"`
	AutoRenewStatus AutoRenewStatus    `gorm:"size:16;index"`
	Environment     Environment        `gorm:"size:16;index"` // SANDBOX/PRODUCTION

	// Timing information
	OriginalPurchaseDate   time.Time  `gorm:"index"`
	PurchaseDate           time.Time  `gorm:"index"`
	ExpiresDate            time.Time  `gorm:"index"`
	GracePeriodExpiresDate *time.Time `gorm:"index"`
	RevocationDate         *time.Time `gorm:"index"`

	// Ownership and access
	InAppOwnershipType InAppOwnershipType `gorm:"size:32"`
	IsUpgraded         bool
	AppAccountToken    *string `gorm:"type:uuid;index"`

	// Pricing information
	Currency    string `gorm:"size:3"`
	Price       int64  // In milliunits
	CountryCode string `gorm:"size:3"`

	// Offer details
	OfferType       *int32
	OfferIdentifier *string `gorm:"size:128"`
	OfferDuration   *string `gorm:"size:32"`

	ConsumptionRequestReason *dto.ConsumptionRequestReason `gorm:"size:32"`

	// System fields
	LastVerifiedAt *time.Time `gorm:"index"`

	// Stores the complete decoded payload as JSON
	LatestRawData *dto.ResponseBodyV2DecodedPayload `gorm:"type:jsonb"`

	// Relationships
	Events []AppStoreSubscriptionEvent `gorm:"foreignKey:SubscriptionID"`
}

type SubscriptionEventType string

func (s SubscriptionEventType) String() {
	panic("unimplemented")
}

const (
	// Core lifecycle events
	EventTypeInitialPurchase SubscriptionEventType = "INITIAL_PURCHASE"
	EventTypeResubscribe     SubscriptionEventType = "RESUSCRIBE"

	EventTypeInteractiveRenew SubscriptionEventType = "INTERACTIVE_RENEW" // User-initiated renewal
	EventTypeAutomaticRenew   SubscriptionEventType = "AUTOMATIC_RENEW"
	EventTypeExpiration       SubscriptionEventType = "EXPIRATION"

	// Billing events
	EventTypeBillingRecovery  SubscriptionEventType = "BILLING_RECOVERY" // After failed renewal
	EventTypeBillingRetry     SubscriptionEventType = "BILLING_RETRY"
	EventTypeGracePeriodEnter SubscriptionEventType = "GRACE_PERIOD_ENTER"
	EventTypeGracePeriodExit  SubscriptionEventType = "GRACE_PERIOD_EXIT"

	// User-initiated changes
	EventTypeCancelled        SubscriptionEventType = "CANCELLED"
	EventTypeUpgrade          SubscriptionEventType = "UPGRADE"
	EventTypeDowngrade        SubscriptionEventType = "DOWNGRADE"
	EventTypeCrossgrade       SubscriptionEventType = "CROSSGRADE" // Same price tier change
	EventTypeAutoRenewEnable  SubscriptionEventType = "AUTO_RENEW_ENABLE"
	EventTypeAutoRenewDisable SubscriptionEventType = "AUTO_RENEW_DISABLE"

	// Admin/System events
	EventTypePriceChange         SubscriptionEventType = "PRICE_CHANGE"
	EventTypePriceIncrease       SubscriptionEventType = "PRICE_INCREASE"
	EventTypePriceDecrease       SubscriptionEventType = "PRICE_DECREASE"
	EventTypeRefund              SubscriptionEventType = "REFUND"
	EventTypeRefundDeclined      SubscriptionEventType = "REFUND_DECLINED"
	EventTypeRefundReversed      SubscriptionEventType = "REFUND_REVERSED"
	EventTypeRevocation          SubscriptionEventType = "REVOCATION"
	EventTypeBillingIssueResolve SubscriptionEventType = "BILLING_ISSUE_RESOLVE"

	// Special cases
	EventTypeFamilyShareChange      SubscriptionEventType = "FAMILY_SHARE_CHANGE"
	EventTypeServiceExtension       SubscriptionEventType = "SERVICE_EXTENSION" // Apple-approved extensions
	EventTypeServiceExtensionFailed SubscriptionEventType = "SERVICE_EXTENSION_FAILED"
	EventTypeRenewalExtension       SubscriptionEventType = "RENEWAL_EXTENSION"
	EventTypeWinbackOfferRedeem     SubscriptionEventType = "WINBACK_OFFER_REDEEM"
	EventTypeOfferCodeRedeem        SubscriptionEventType = "OFFER_CODE_REDEEM"

	// Diagnostic events
	EventTypeConsumptionRequest SubscriptionEventType = "CONSUMPTION_REQUEST"
	EventTypeTestNotification   SubscriptionEventType = "TEST_NOTIFICATION"

	EventTypeMetadataUpdate SubscriptionEventType = "METADATA_UPDATE"
	EventTypeMigration      SubscriptionEventType = "MIGRATION"
	EventTypeOneTimeCharge  SubscriptionEventType = "ONE_TIME_CHARGE"
)

// AppStoreSubscriptionEvent captures the complete state of a subscription at the time of an event
type AppStoreSubscriptionEvent struct {
	gorm.Model
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Reference to the subscription
	SubscriptionID uuid.UUID `gorm:"index;not null"`

	BundleID *string `gorm:"size:128;index"`

	// Event metadata
	Type             SubscriptionEventType `gorm:"size:32;index;not null"`
	Subtype          *string               `gorm:"size:32;index"`
	NotificationType *string               `gorm:"size:64;index"`
	EventDate        time.Time             `gorm:"index;not null"`

	// Full snapshot of subscription fields (all nullable except EventDate)
	OriginalTransactionID *string             `gorm:"size:128;index"`
	CurrentTransactionID  *string             `gorm:"size:128;index"`
	WebOrderLineItemID    *string             `gorm:"size:128;index"`
	UserID                *uuid.UUID          `gorm:"type:uuid;index"`
	ProductID             *string             `gorm:"size:128;index"`
	SubscriptionGroupID   *string             `gorm:"size:128;index"`
	Status                *SubscriptionStatus `gorm:"size:32;index"`
	AutoRenewStatus       *AutoRenewStatus    `gorm:"size:16;index"`
	Environment           *Environment        `gorm:"size:16;index"`

	// Timing information
	OriginalPurchaseDate   *time.Time `gorm:"index"`
	PurchaseDate           *time.Time `gorm:"index"`
	ExpiresDate            *time.Time `gorm:"index"`
	GracePeriodExpiresDate *time.Time `gorm:"index"`
	RevocationDate         *time.Time `gorm:"index"`

	// Ownership and access
	InAppOwnershipType *InAppOwnershipType `gorm:"size:32"`
	IsUpgraded         *bool
	AppAccountToken    *string `gorm:"type:uuid;index"`

	// Pricing information
	Currency    *string `gorm:"size:3"`
	Price       *int64
	CountryCode *string `gorm:"size:3"`

	// Offer details
	OfferType       *int32
	OfferIdentifier *string `gorm:"size:128"`
	OfferDuration   *string `gorm:"size:32"`

	// Additional context
	Reason                   *string                       `gorm:"type:text"`
	ConsumptionRequestReason *dto.ConsumptionRequestReason `gorm:"size:32"`
	// Stores the complete decoded payload as JSON
	RawData *dto.ResponseBodyV2DecodedPayload `gorm:"type:jsonb"`
}

// CreateEventFromSubscription creates a comprehensive event from the current subscription state
// Added String() methods for better logging/debugging
func (s SubscriptionStatus) String() string {
	return string(s)
}

func (a AutoRenewStatus) String() string {
	return string(a)
}

func (i InAppOwnershipType) String() string {
	return string(i)
}

func (e Environment) String() string {
	return string(e)
}

// Enhanced AppStoreSubscription with validation
func (s *AppStoreSubscription) BeforeCreate(tx *gorm.DB) error {
	if s.OriginalTransactionID == "" {
		return fmt.Errorf("original transaction ID is required")
	}
	if s.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	return nil
}

// Enhanced AppStoreSubscriptionEvent with validation
func (e *AppStoreSubscriptionEvent) BeforeCreate(tx *gorm.DB) error {
	if e.SubscriptionID == uuid.Nil {
		return fmt.Errorf("subscription ID is required")
	}
	if e.EventDate.IsZero() {
		e.EventDate = time.Now()
	}
	return nil
}

// Improved CreateEventFromSubscription with error handling
func (s *AppStoreSubscription) CreateEventFromSubscription(
	eventType SubscriptionEventType,
	notification *dto.AppStoreNotification,
) (*AppStoreSubscriptionEvent, error) {
	if s == nil {
		return nil, fmt.Errorf("subscription cannot be nil")
	}

	event := &AppStoreSubscriptionEvent{
		ID: func() uuid.UUID {
			if notification != nil {
				return notification.ID
			}
			return uuid.New()
		}(),
		SubscriptionID:        s.ID,
		Type:                  eventType,
		EventDate:             time.Now().UTC(), // Use UTC for consistency
		OriginalTransactionID: &s.OriginalTransactionID,
		CurrentTransactionID:  &s.CurrentTransactionID,
		WebOrderLineItemID:    &s.WebOrderLineItemID,
		UserID:                &s.UserID,
		ProductID:             &s.ProductID,
		SubscriptionGroupID:   &s.SubscriptionGroupID,
		Status:                &s.Status,
		AutoRenewStatus:       &s.AutoRenewStatus,
		Environment:           &s.Environment,
		OriginalPurchaseDate:  &s.OriginalPurchaseDate,
		PurchaseDate:          &s.PurchaseDate,
		ExpiresDate:           &s.ExpiresDate,
		InAppOwnershipType:    &s.InAppOwnershipType,
		IsUpgraded:            &s.IsUpgraded,
		Currency:              &s.Currency,
		Price:                 &s.Price,
		CountryCode:           &s.CountryCode,
		BundleID:              &s.BundleID,
	}

	// Handle nullable fields safely
	if s.GracePeriodExpiresDate != nil {
		event.GracePeriodExpiresDate = s.GracePeriodExpiresDate
	}
	if s.RevocationDate != nil {
		event.RevocationDate = s.RevocationDate
	}
	if s.AppAccountToken != nil {
		event.AppAccountToken = s.AppAccountToken
	}
	if s.OfferType != nil {
		event.OfferType = s.OfferType
	}
	if s.OfferIdentifier != nil {
		event.OfferIdentifier = s.OfferIdentifier
	}
	if s.OfferDuration != nil {
		event.OfferDuration = s.OfferDuration
	}

	if notification != nil {
		// Deep copy the payload to prevent mutation
		payloadCopy := notification.ResponseBodyV2DecodedPayload
		event.RawData = &payloadCopy

		if notification.ResponseBodyV2DecodedPayload.Subtype != nil {
			subtype := string(*notification.ResponseBodyV2DecodedPayload.Subtype)
			event.Subtype = &subtype
		}

		notifType := string(notification.ResponseBodyV2DecodedPayload.NotificationType)
		event.NotificationType = &notifType
	}

	return event, nil
}

// Enhanced AddEvent with transaction handling
func (s *AppStoreSubscription) AddEvent(
	tx *gorm.DB,
	eventType SubscriptionEventType,
	notification *dto.AppStoreNotification,
	reason string,
) error {
	event, err := s.CreateEventFromSubscription(eventType, notification)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	if reason != "" {
		event.Reason = &reason
	}

	if err := tx.Create(event).Error; err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	// Safely append to events slice
	if s.Events == nil {
		s.Events = make([]AppStoreSubscriptionEvent, 0)
	}
	s.Events = append(s.Events, *event)

	// Update the latest raw data if this is a state-changing event
	if notification != nil && shouldUpdateLatestData(

	// eventType

	) {
		payloadCopy := notification.ResponseBodyV2DecodedPayload
		s.LatestRawData = &payloadCopy
		if err := tx.Model(s).Update("latest_raw_data", s.LatestRawData).Error; err != nil {
			logger.Log.Warnf("Failed to update latest raw data: %v", err)
		}
	}

	return nil
}

func shouldUpdateLatestData(

// eventType SubscriptionEventType

) bool {
	// List of events that should update the latest raw data
	// switch eventType {
	// case EventTypeInitialPurchase, EventTypeResubscribe,
	// 	EventTypeAutomaticRenew, EventTypeUpgrade,
	// 	EventTypeDowngrade, EventTypeRefund,
	// 	EventTypeRevocation, EventTypeBillingRecovery:
	return true
	// default:
	// 	return false
	// }
}
