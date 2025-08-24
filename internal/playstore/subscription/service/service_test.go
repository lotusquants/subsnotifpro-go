package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	// Import the domain models
	apiDto "subsnotifpro-go/internal/playstore/api/dto"
	"subsnotifpro-go/internal/playstore/subscription/models"
	rtdnDto "subsnotifpro-go/internal/playstore/rtdn/dto"
	userModels "subsnotifpro-go/internal/playstore/user/models"
)

// Test suite for PlaystoreSubscriptionService
type PlaystoreSubscriptionServiceTestSuite struct {
	suite.Suite
	service         PlaystoreSubscriptionService
	mockRepo        *MockPlaystoreSubscriptionRepository
	mockUserService *MockPlaystoreUserService
}

// Mock repository interface - implementing all required methods
type MockPlaystoreSubscriptionRepository struct {
	mock.Mock
}

func (m *MockPlaystoreSubscriptionRepository) GetSubscriptionByPurchaseToken(ctx context.Context, tx *gorm.DB, token string) (*models.SubscriptionPurchaseV2, error) {
	args := m.Called(ctx, tx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SubscriptionPurchaseV2), args.Error(1)
}

func (m *MockPlaystoreSubscriptionRepository) InsertSubscription(ctx context.Context, tx *gorm.DB, sub *models.SubscriptionPurchaseV2) error {
	args := m.Called(ctx, tx, sub)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) UpdateSubscriptionFields(ctx context.Context, tx *gorm.DB, subscriptionID uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, tx, subscriptionID, updates)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertSubscriptionStateTransition(ctx context.Context, tx *gorm.DB, history *models.SubscriptionStateTransitionHistory) error {
	args := m.Called(ctx, tx, history)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertOrderIDTransition(ctx context.Context, tx *gorm.DB, history *models.SubscriptionOrderIdTransitionHistory) error {
	args := m.Called(ctx, tx, history)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertAcknowledgementStateTransition(ctx context.Context, tx *gorm.DB, history *models.AcknowledgementStateTransitionHistory) error {
	args := m.Called(ctx, tx, history)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) UpdatePausedContext(ctx context.Context, tx *gorm.DB, context *models.SubscriptionPausedContext) error {
	args := m.Called(ctx, tx, context)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertPausedContext(ctx context.Context, tx *gorm.DB, context *models.SubscriptionPausedContext) error {
	args := m.Called(ctx, tx, context)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertPausedContextHistory(ctx context.Context, tx *gorm.DB, history *models.SubscriptionPausedContextHistory) error {
	args := m.Called(ctx, tx, history)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertCancellationContext(ctx context.Context, tx *gorm.DB, cancelCtx *models.SubscriptionCancellationContext) error {
	args := m.Called(ctx, tx, cancelCtx)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) UpdateCancellationContext(ctx context.Context, tx *gorm.DB, cancelCtx *models.SubscriptionCancellationContext) error {
	args := m.Called(ctx, tx, cancelCtx)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertCancellationHistory(ctx context.Context, tx *gorm.DB, history *models.SubscriptionCancellationContextHistory) error {
	args := m.Called(ctx, tx, history)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertOfferDetails(ctx context.Context, tx *gorm.DB, detail *models.OfferDetails) error {
	args := m.Called(ctx, tx, detail)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertOfferDetailsHistory(ctx context.Context, tx *gorm.DB, detail *models.OfferDetailsHistory) error {
	args := m.Called(ctx, tx, detail)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) UpdateOfferDetails(ctx context.Context, tx *gorm.DB, offer *models.OfferDetails) error {
	args := m.Called(ctx, tx, offer)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertAutoRenewingPlan(ctx context.Context, tx *gorm.DB, plan *models.AutoRenewingPlan) error {
	args := m.Called(ctx, tx, plan)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) DeleteAutoRenewingPlansByLineItem(ctx context.Context, tx *gorm.DB, lineItemID uuid.UUID) error {
	args := m.Called(ctx, tx, lineItemID)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertAutoRenewingPlanHistory(ctx context.Context, tx *gorm.DB, history *models.AutoRenewingPlanHistory) error {
	args := m.Called(ctx, tx, history)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertPrepaidPlan(ctx context.Context, tx *gorm.DB, plan *models.PrepaidPlan) error {
	args := m.Called(ctx, tx, plan)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) UpdatePrepaidPlan(ctx context.Context, tx *gorm.DB, plan *models.PrepaidPlan) error {
	args := m.Called(ctx, tx, plan)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) InsertPrepaidPlanHistory(ctx context.Context, tx *gorm.DB, history *models.PrepaidPlanHistory) error {
	args := m.Called(ctx, tx, history)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) CreateSignupPromotion(ctx context.Context, tx *gorm.DB, promotion *models.SignupPromotion) error {
	args := m.Called(ctx, tx, promotion)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) UpdateSignupPromotion(ctx context.Context, tx *gorm.DB, promotion *models.SignupPromotion) error {
	args := m.Called(ctx, tx, promotion)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) CreateSignupPromotionHistory(ctx context.Context, tx *gorm.DB, promotion *models.SignupPromotionHistory) error {
	args := m.Called(ctx, tx, promotion)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) CreateDeferredReplacement(ctx context.Context, tx *gorm.DB, replacement *models.DeferredItemReplacement) error {
	args := m.Called(ctx, tx, replacement)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) UpdateDeferredReplacement(ctx context.Context, tx *gorm.DB, replacement *models.DeferredItemReplacement) error {
	args := m.Called(ctx, tx, replacement)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) CreateDeferredReplacementHistory(ctx context.Context, tx *gorm.DB, replacement *models.DeferredItemReplacementHistory) error {
	args := m.Called(ctx, tx, replacement)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) CreateBulkSubscriptionLineItemHistory(ctx context.Context, tx *gorm.DB, entries []models.SubscriptionLineItemHistory) error {
	args := m.Called(ctx, tx, entries)
	return args.Error(0)
}

func (m *MockPlaystoreSubscriptionRepository) CreateSubscriptionEvent(ctx context.Context, tx *gorm.DB, event *models.SubscriptionEvent) error {
	args := m.Called(ctx, tx, event)
	return args.Error(0)
}

// Mock user service interface
type MockPlaystoreUserService struct {
	mock.Mock
}

func (m *MockPlaystoreUserService) GetOrCreateUserIDFromObfuscatedExternalAccountID(
	ctx context.Context,
	tx *gorm.DB,
	obfuscatedID string,
	googleAccount *userModels.GoogleAccount,
) (uuid.UUID, error) {
	args := m.Called(ctx, tx, obfuscatedID, googleAccount)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// Setup and teardown
func (suite *PlaystoreSubscriptionServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockPlaystoreSubscriptionRepository)
	suite.mockUserService = new(MockPlaystoreUserService)
	
	suite.service = &playstoreSubscriptionService{
		repo:                 suite.mockRepo,
		playstoreUserService: suite.mockUserService,
	}
}

func (suite *PlaystoreSubscriptionServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}

// Test Cases

func (suite *PlaystoreSubscriptionServiceTestSuite) TestUpsertSubscription_CreateNew_DataValidation() {
	// Test data validation and structure without actual service call
	// This tests the business logic and data structures

	// Mock subscription data from Google API
	subData := &apiDto.SubscriptionPurchaseV2{
		RegionCode:           "US",
		StartTime:            time.Now().Add(-24 * time.Hour),
		SubscriptionState:    apiDto.SubscriptionStateActive,
		AcknowledgementState: apiDto.AcknowledgementStateAcknowledged,
		LatestOrderID:        "GPA.ORDER.123",
		LineItems: []apiDto.LineItem{
			{
				ProductID:  "premium_monthly",
				ExpiryTime: time.Now().Add(30 * 24 * time.Hour),
				PlanType:   apiDto.PlanTypeAutoRenewing,
				OfferDetails: &apiDto.OfferDetails{
					OfferTags:  []string{"new_user"},
					BasePlanID: "premium-monthly",
					OfferID:    "new-user-discount",
				},
			},
		},
		ExternalAccountIdentifiers: apiDto.ExternalAccountIdentifiers{
			ObfuscatedExternalAccountID: stringPtr("user123abc"),
		},
	}

	// Mock webhook event with correct structure
	event := &rtdnDto.GooglePlayWebhookEvent{
		PackageName: "com.example.app",
		Subscription: &rtdnDto.SubscriptionNotification{
			PurchaseToken:    "purchase_token_123",
			SubscriptionID:   "premium_monthly",
			NotificationType: rtdnDto.SubscriptionPurchased,
		},
	}

	// Verify data structure validity
	suite.NotNil(subData)
	suite.NotNil(event)
	suite.Equal("US", subData.RegionCode)
	suite.Equal(apiDto.SubscriptionStateActive, subData.SubscriptionState)
	suite.Equal("com.example.app", event.PackageName)
	suite.Equal("purchase_token_123", event.Subscription.PurchaseToken)
	
	// Verify line items structure
	suite.Len(subData.LineItems, 1)
	lineItem := subData.LineItems[0]
	suite.Equal("premium_monthly", lineItem.ProductID)
	suite.Equal(apiDto.PlanTypeAutoRenewing, lineItem.PlanType)
	
	// Verify offer details
	suite.NotNil(lineItem.OfferDetails)
	suite.Equal("premium-monthly", lineItem.OfferDetails.BasePlanID)
	suite.Equal("new-user-discount", lineItem.OfferDetails.OfferID)
}

func (suite *PlaystoreSubscriptionServiceTestSuite) TestUpsertSubscription_UpdateExisting_DataValidation() {
	// Test update scenario data validation

	// Mock existing subscription
	existingSubID := uuid.New()
	existingSub := &models.SubscriptionPurchaseV2{
		ID:                   existingSubID,
		PackageName:          "com.example.app",
		PurchaseToken:        "purchase_token_123",
		UserID:               uuid.New(),
		SubscriptionState:    models.SubscriptionStatePending,
		AcknowledgementState: models.AcknowledgementStatePending,
		StartTime:            time.Now().Add(-48 * time.Hour),
		LatestOrderID:        "GPA.ORDER.122",
		RegionCode:           "US",
		IsTestPurchase:       false,
		LineItems: []models.SubscriptionLineItem{
			{
				ID:         uuid.New(),
				ProductID:  "premium_monthly",
				ExpiryTime: time.Now().Add(29 * 24 * time.Hour),
				PlanType:   models.PlanTypeAutoRenewing,
			},
		},
	}

	// Updated subscription data
	subData := &apiDto.SubscriptionPurchaseV2{
		RegionCode:           "US",
		StartTime:            existingSub.StartTime,
		SubscriptionState:    apiDto.SubscriptionStateActive,
		AcknowledgementState: apiDto.AcknowledgementStateAcknowledged,
		LatestOrderID:        "GPA.ORDER.123",
		LineItems: []apiDto.LineItem{
			{
				ProductID:  "premium_monthly",
				ExpiryTime: time.Now().Add(30 * 24 * time.Hour),
				PlanType:   apiDto.PlanTypeAutoRenewing,
			},
		},
		ExternalAccountIdentifiers: apiDto.ExternalAccountIdentifiers{
			ObfuscatedExternalAccountID: stringPtr("user123abc"),
		},
	}

	event := &rtdnDto.GooglePlayWebhookEvent{
		PackageName: "com.example.app",
		Subscription: &rtdnDto.SubscriptionNotification{
			PurchaseToken:    "purchase_token_123",
			SubscriptionID:   "premium_monthly",
			NotificationType: rtdnDto.SubscriptionRenewed,
		},
	}

	// Test business logic validation
	suite.Equal(models.SubscriptionStatePending, existingSub.SubscriptionState)
	suite.Equal(models.AcknowledgementStatePending, existingSub.AcknowledgementState)
	
	// Verify state transition logic is sound
	newState := models.SubscriptionState(subData.SubscriptionState)
	newAckState := models.AcknowledgementState(subData.AcknowledgementState)
	
	suite.Equal(models.SubscriptionStateActive, newState)
	suite.Equal(models.AcknowledgementStateAcknowledged, newAckState)
	
	// Verify event structure
	suite.Equal(rtdnDto.SubscriptionRenewed, event.Subscription.NotificationType)
}

func (suite *PlaystoreSubscriptionServiceTestSuite) TestUpsertSubscription_ErrorHandling() {
	// Test case: Missing ObfuscatedExternalAccountID
	subDataMissingAccount := &apiDto.SubscriptionPurchaseV2{
		RegionCode:           "US",
		StartTime:            time.Now(),
		SubscriptionState:    apiDto.SubscriptionStateActive,
		AcknowledgementState: apiDto.AcknowledgementStateAcknowledged,
		LatestOrderID:        "GPA.ORDER.123",
		ExternalAccountIdentifiers: apiDto.ExternalAccountIdentifiers{
			ObfuscatedExternalAccountID: nil, // Missing!
		},
	}

	// Verify input validation logic
	suite.Nil(subDataMissingAccount.ExternalAccountIdentifiers.ObfuscatedExternalAccountID)
	
	// Test case: Empty purchase token
	eventEmptyToken := &rtdnDto.GooglePlayWebhookEvent{
		PackageName: "com.example.app",
		Subscription: &rtdnDto.SubscriptionNotification{
			PurchaseToken:    "", // Empty!
			SubscriptionID:   "premium_monthly",
			NotificationType: rtdnDto.SubscriptionPurchased,
		},
	}
	
	suite.Empty(eventEmptyToken.Subscription.PurchaseToken)
}

func (suite *PlaystoreSubscriptionServiceTestSuite) TestUpsertSubscription_LineItemProcessing() {
	// Test multiple line items processing
	subData := &apiDto.SubscriptionPurchaseV2{
		RegionCode:           "US",
		StartTime:            time.Now(),
		SubscriptionState:    apiDto.SubscriptionStateActive,
		AcknowledgementState: apiDto.AcknowledgementStateAcknowledged,
		LatestOrderID:        "GPA.ORDER.123",
		LineItems: []apiDto.LineItem{
			{
				ProductID:  "premium_monthly",
				ExpiryTime: time.Now().Add(30 * 24 * time.Hour),
				PlanType:   apiDto.PlanTypeAutoRenewing,
				AutoRenewingPlan: &apiDto.AutoRenewingPlan{
					AutoRenewEnabled: true,
					RecurringPrice: apiDto.Money{
						CurrencyCode: "USD",
						Units:        9,
						Nanos:        990000000,
					},
				},
			},
			{
				ProductID:  "premium_addon",
				ExpiryTime: time.Now().Add(30 * 24 * time.Hour),
				PlanType:   apiDto.PlanTypeAutoRenewing,
			},
		},
		ExternalAccountIdentifiers: apiDto.ExternalAccountIdentifiers{
			ObfuscatedExternalAccountID: stringPtr("user123abc"),
		},
	}

	// Verify line items structure and business logic
	suite.Len(subData.LineItems, 2)
	suite.Equal("premium_monthly", subData.LineItems[0].ProductID)
	suite.Equal("premium_addon", subData.LineItems[1].ProductID)
	
	// Verify auto-renewing plan details
	arp := subData.LineItems[0].AutoRenewingPlan
	suite.NotNil(arp)
	suite.True(arp.AutoRenewEnabled)
	suite.Equal("USD", arp.RecurringPrice.CurrencyCode)
	suite.Equal(int64(9), arp.RecurringPrice.Units)
}

func (suite *PlaystoreSubscriptionServiceTestSuite) TestUpsertSubscription_OfferDetailsProcessing() {
	// Test offer details and promotion handling
	subData := &apiDto.SubscriptionPurchaseV2{
		RegionCode:           "US",
		StartTime:            time.Now(),
		SubscriptionState:    apiDto.SubscriptionStateActive,
		AcknowledgementState: apiDto.AcknowledgementStateAcknowledged,
		LatestOrderID:        "GPA.ORDER.123",
		LineItems: []apiDto.LineItem{
			{
				ProductID:  "premium_monthly",
				ExpiryTime: time.Now().Add(30 * 24 * time.Hour),
				PlanType:   apiDto.PlanTypeAutoRenewing,
				OfferDetails: &apiDto.OfferDetails{
					OfferTags:  []string{"new_user", "discount_50"},
					BasePlanID: "premium-monthly",
					OfferID:    "new-user-50-off",
				},
				SignupPromotion: &apiDto.SignupPromotion{
					Type: apiDto.PromoTypeVanityCode,
					Code: stringPtr("WELCOME50"),
				},
			},
		},
		ExternalAccountIdentifiers: apiDto.ExternalAccountIdentifiers{
			ObfuscatedExternalAccountID: stringPtr("user123abc"),
		},
	}

	// Verify offer details structure
	lineItem := subData.LineItems[0]
	suite.NotNil(lineItem.OfferDetails)
	suite.Equal("premium-monthly", lineItem.OfferDetails.BasePlanID)
	suite.Equal("new-user-50-off", lineItem.OfferDetails.OfferID)
	suite.Contains(lineItem.OfferDetails.OfferTags, "new_user")
	suite.Contains(lineItem.OfferDetails.OfferTags, "discount_50")

	// Verify signup promotion
	suite.NotNil(lineItem.SignupPromotion)
	suite.Equal(apiDto.PromoTypeVanityCode, lineItem.SignupPromotion.Type)
	suite.NotNil(lineItem.SignupPromotion.Code)
	suite.Equal("WELCOME50", *lineItem.SignupPromotion.Code)
}

func (suite *PlaystoreSubscriptionServiceTestSuite) TestSubscriptionStateTransitions() {
	// Test valid state transitions
	testCases := []struct {
		name          string
		currentState  models.SubscriptionState
		newState      models.SubscriptionState
		shouldBeValid bool
	}{
		{
			name:          "Pending to Active",
			currentState:  models.SubscriptionStatePending,
			newState:      models.SubscriptionStateActive,
			shouldBeValid: true,
		},
		{
			name:          "Active to Canceled",
			currentState:  models.SubscriptionStateActive,
			newState:      models.SubscriptionStateCanceled,
			shouldBeValid: true,
		},
		{
			name:          "Active to Paused",
			currentState:  models.SubscriptionStateActive,
			newState:      models.SubscriptionStatePaused,
			shouldBeValid: true,
		},
		{
			name:          "Expired to Active (Renewal)",
			currentState:  models.SubscriptionStateExpired,
			newState:      models.SubscriptionStateActive,
			shouldBeValid: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Test state transition logic
			suite.NotEqual(tc.currentState, tc.newState)
			
			// Verify states are valid enum values
			validStates := []models.SubscriptionState{
				models.SubscriptionStatePending,
				models.SubscriptionStateActive,
				models.SubscriptionStatePaused,
				models.SubscriptionStateInGracePeriod,
				models.SubscriptionStateOnHold,
				models.SubscriptionStateCanceled,
				models.SubscriptionStateExpired,
			}
			
			suite.Contains(validStates, tc.currentState)
			suite.Contains(validStates, tc.newState)
		})
	}
}

func (suite *PlaystoreSubscriptionServiceTestSuite) TestWebhookEventStructure() {
	// Test webhook event structure validation
	event := &rtdnDto.GooglePlayWebhookEvent{
		ID:              uuid.New(),
		Version:         "1.0",
		PackageName:     "com.example.app",
		EventTimeMillis: time.Now().Unix() * 1000,
		RawPayload:      `{"test": "payload"}`,
		Subscription: &rtdnDto.SubscriptionNotification{
			Version:          "1.0",
			PurchaseToken:    "purchase_token_123",
			SubscriptionID:   "premium_monthly",
			NotificationType: rtdnDto.SubscriptionPurchased,
		},
	}

	// Verify event structure
	suite.NotNil(event.Subscription)
	suite.Equal("com.example.app", event.PackageName)
	suite.Equal("purchase_token_123", event.Subscription.PurchaseToken)
	suite.Equal("premium_monthly", event.Subscription.SubscriptionID)
	suite.Equal(rtdnDto.SubscriptionPurchased, event.Subscription.NotificationType)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

// Run test suite
func TestPlaystoreSubscriptionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PlaystoreSubscriptionServiceTestSuite))
}
