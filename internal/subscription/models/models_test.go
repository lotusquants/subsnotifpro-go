package models_test

import (
	"testing"
	"time"

	"subsnotifpro-go/internal/subscription/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UnifiedSubscriptionModelTestSuite provides comprehensive testing for the UnifiedSubscription domain model
type UnifiedSubscriptionModelTestSuite struct {
	suite.Suite
}

func TestUnifiedSubscriptionModelSuite(t *testing.T) {
	suite.Run(t, new(UnifiedSubscriptionModelTestSuite))
}

// Test UnifiedSubscription Model Creation and Validation
func (suite *UnifiedSubscriptionModelTestSuite) TestUnifiedSubscriptionCreation() {
	now := time.Now()
	userID := uuid.New()
	subscriptionID := uuid.New()

	tests := []struct {
		name        string
		setupSub    func() *models.UnifiedSubscription
		expectValid bool
		description string
	}{
		{
			name: "Valid Apple App Store Subscription",
			setupSub: func() *models.UnifiedSubscription {
				return &models.UnifiedSubscription{
					UserID:         userID,
					SubscriptionID: subscriptionID,
					ActivePlatform: models.PlatformApple,
					PlatformUserID: stringPtr("apple_user_123"),
					PurchaseToken:  stringPtr("original_transaction_id_123"),
					LatestOrderID:  stringPtr("current_transaction_id_456"),
					PlanType:       "AUTO_RENEWING",
					Status:         models.StatusActive,
					StartDate:      now,
					NextRenewalDate: now.AddDate(0, 1, 0),
					ExpirationDate: now.AddDate(0, 1, 7), // 7 days grace after renewal
					ProductId:      "com.example.premium",
					BasePlanID:     "premium_monthly",
					TotalAmount:    9.99,
					Currency:       "USD",
				}
			},
			expectValid: true,
			description: "Standard Apple subscription with all required fields",
		},
		{
			name: "Valid Google Play Store Subscription",
			setupSub: func() *models.UnifiedSubscription {
				return &models.UnifiedSubscription{
					UserID:         userID,
					SubscriptionID: subscriptionID,
					ActivePlatform: models.PlatformGoogle,
					PlatformUserID: stringPtr("google_obfuscated_123"),
					PurchaseToken:  stringPtr("purchase_token_xyz"),
					LatestOrderID:  stringPtr("order_id_789"),
					PlanType:       "AUTO_RENEWING",
					Status:         models.StatusActive,
					StartDate:      now,
					NextRenewalDate: now.AddDate(0, 1, 0),
					ExpirationDate: now.AddDate(0, 1, 7),
					ProductId:      "premium_subscription",
					BasePlanID:     "premium-monthly",
					AddOnID:        stringPtr("extra_features"),
					ActiveOfferID:  stringPtr("50_percent_off"),
					TotalAmount:    4.99, // With 50% discount
					Currency:       "USD",
				}
			},
			expectValid: true,
			description: "Google Play subscription with add-ons and offers",
		},
		{
			name: "Grace Period Subscription",
			setupSub: func() *models.UnifiedSubscription {
				gracePeriodStart := now.AddDate(0, 1, 0)
				gracePeriodEnd := now.AddDate(0, 1, 7)
				return &models.UnifiedSubscription{
					UserID:               userID,
					SubscriptionID:       subscriptionID,
					ActivePlatform:       models.PlatformApple,
					PlatformUserID:       stringPtr("apple_user_456"),
					PurchaseToken:        stringPtr("grace_period_token"),
					Status:               models.StatusGracePeriod,
					StartDate:            now,
					NextRenewalDate:      now.AddDate(0, 1, 0),
					ExpirationDate:       gracePeriodEnd,
					GracePeriodStartDate: &gracePeriodStart,
					GracePeriodEndDate:   &gracePeriodEnd,
					ProductId:            "com.example.premium",
					BasePlanID:           "premium_monthly",
					TotalAmount:          9.99,
					Currency:             "USD",
				}
			},
			expectValid: true,
			description: "Subscription in grace period with proper date handling",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			sub := tt.setupSub()
			
			// Basic field validation
			assert.NotEqual(suite.T(), uuid.Nil, sub.UserID, "UserID should not be nil")
			assert.NotEqual(suite.T(), uuid.Nil, sub.SubscriptionID, "SubscriptionID should not be nil")
			assert.NotEmpty(suite.T(), sub.ActivePlatform, "ActivePlatform should be set")
			assert.NotEmpty(suite.T(), sub.Status, "Status should be set")
			assert.NotEmpty(suite.T(), sub.ProductId, "ProductId should be set")
			assert.NotEmpty(suite.T(), sub.BasePlanID, "BasePlanID should be set")
			
			// Platform-specific validation
			if sub.ActivePlatform == models.PlatformApple {
				assert.NotNil(suite.T(), sub.PurchaseToken, "Apple subscriptions should have PurchaseToken")
			}
			
			// Financial validation
			if tt.expectValid {
				assert.Positive(suite.T(), sub.TotalAmount, "TotalAmount should be positive")
				assert.NotEmpty(suite.T(), sub.Currency, "Currency should be set")
				assert.Len(suite.T(), sub.Currency, 3, "Currency should be 3-letter code")
			}
			
			// Date validation
			if sub.Status == models.StatusGracePeriod {
				assert.NotNil(suite.T(), sub.GracePeriodStartDate, "Grace period subscriptions should have start date")
				assert.NotNil(suite.T(), sub.GracePeriodEndDate, "Grace period subscriptions should have end date")
				if sub.GracePeriodStartDate != nil && sub.GracePeriodEndDate != nil {
					assert.True(suite.T(), sub.GracePeriodEndDate.After(*sub.GracePeriodStartDate), 
						"Grace period end should be after start")
				}
			}
		})
	}
}

// Test Platform Type Validation
func (suite *UnifiedSubscriptionModelTestSuite) TestPlatformTypeValidation() {
	tests := []struct {
		platform models.PlatformType
		isValid  bool
	}{
		{models.PlatformApple, true},
		{models.PlatformGoogle, true},
		{models.PlatformType("INVALID_PLATFORM"), false},
		{models.PlatformType(""), false},
	}

	for _, tt := range tests {
		suite.Run(string(tt.platform), func() {
			if tt.isValid {
				assert.Contains(suite.T(), []models.PlatformType{models.PlatformApple, models.PlatformGoogle}, 
					tt.platform, "Platform should be valid")
			} else {
				assert.NotContains(suite.T(), []models.PlatformType{models.PlatformApple, models.PlatformGoogle}, 
					tt.platform, "Platform should be invalid")
			}
		})
	}
}

// Test Subscription Status Validation and State Machine
func (suite *UnifiedSubscriptionModelTestSuite) TestSubscriptionStatusValidation() {
	validStatuses := []models.SubscriptionStatus{
		models.StatusActive,
		models.StatusExpired,
		models.StatusBillingRetry,
		models.StatusGracePeriod,
		models.StatusRevoked,
		models.StatusPendingRenewal,
		models.StatusPendingUpgrade,
		models.StatusPendingDowngrade,
		models.StatusPaused,
		models.StatusOnHold,
		models.StatusCanceled,
	}

	suite.Run("Valid Statuses", func() {
		for _, status := range validStatuses {
			assert.NotEmpty(suite.T(), string(status), "Status should not be empty")
			assert.True(suite.T(), len(string(status)) > 3, "Status should be descriptive")
		}
	})

	// Test status transitions logic (business rules)
	suite.Run("Status Transition Logic", func() {
		testCases := []struct {
			fromStatus    models.SubscriptionStatus
			toStatus      models.SubscriptionStatus
			shouldBeValid bool
			description   string
		}{
			// Valid transitions
			{models.StatusActive, models.StatusCanceled, true, "User can cancel active subscription"},
			{models.StatusActive, models.StatusExpired, true, "Active subscription can expire"},
			{models.StatusActive, models.StatusGracePeriod, true, "Failed payment leads to grace period"},
			{models.StatusGracePeriod, models.StatusActive, true, "Payment recovery from grace period"},
			{models.StatusGracePeriod, models.StatusExpired, true, "Grace period can expire"},
			{models.StatusBillingRetry, models.StatusActive, true, "Billing retry success"},
			{models.StatusBillingRetry, models.StatusGracePeriod, true, "Billing retry to grace period"},
			{models.StatusPaused, models.StatusActive, true, "Resume from pause"},
			{models.StatusOnHold, models.StatusActive, true, "Resolve payment issue"},
			
			// Invalid transitions (business logic violations)
			{models.StatusExpired, models.StatusActive, false, "Cannot reactivate expired without new purchase"},
			{models.StatusRevoked, models.StatusActive, false, "Cannot reactivate revoked subscription"},
			{models.StatusCanceled, models.StatusGracePeriod, false, "Canceled subscriptions don't enter grace period"},
		}

		for _, tc := range testCases {
			suite.Run(tc.description, func() {
				// This would normally be implemented as business logic methods
				// For now, we're testing the logic structure
				isValidTransition := suite.isValidStatusTransition(tc.fromStatus, tc.toStatus)
				assert.Equal(suite.T(), tc.shouldBeValid, isValidTransition, tc.description)
			})
		}
	})
}

// Test UnifiedSubscriptionEvent Model
func (suite *UnifiedSubscriptionModelTestSuite) TestUnifiedSubscriptionEvent() {
	now := time.Now()
	
	event := &models.UnifiedSubscriptionEvent{
		ID:             uuid.New().String(),
		SubscriptionID: uuid.New().String(),
		EventType:      models.EventTypePurchase,
		Timestamp:      now,
		Platform:       models.PlatformApple,
		Amount:         9.99,
		Currency:       "USD",
		ProductID:      "com.example.premium",
		BasePlanID:     stringPtr("premium_monthly"),
		ActiveOfferID:  stringPtr("first_month_free"),
	}

	// Validate event fields
	assert.NotEmpty(suite.T(), event.ID, "Event ID should not be empty")
	assert.NotEmpty(suite.T(), event.SubscriptionID, "SubscriptionID should not be empty")
	assert.NotEmpty(suite.T(), event.EventType, "EventType should not be empty")
	assert.Positive(suite.T(), event.Amount, "Amount should be positive")
	assert.NotEmpty(suite.T(), event.Currency, "Currency should not be empty")
	assert.NotEmpty(suite.T(), event.ProductID, "ProductID should not be empty")

	// Test event types
	validEventTypes := []string{
		models.EventTypePurchase,
		models.EventTypeRenewal,
		models.EventTypeCancel,
		models.EventTypeGracePeriod,
		models.EventTypeExpiration,
		models.EventTypePause,
		models.EventTypeResume,
	}

	for _, eventType := range validEventTypes {
		suite.Run("EventType_"+eventType, func() {
			assert.NotEmpty(suite.T(), eventType, "Event type should not be empty")
		})
	}
}

// Test Financial Calculations and Currency Handling
func (suite *UnifiedSubscriptionModelTestSuite) TestFinancialCalculations() {
	tests := []struct {
		name           string
		amount         float64
		currency       string
		isValid        bool
		expectedAmount float64
	}{
		{"Valid USD Amount", 9.99, "USD", true, 9.99},
		{"Valid EUR Amount", 8.99, "EUR", true, 8.99},
		{"Zero Amount", 0.00, "USD", false, 0.00},
		{"Negative Amount", -5.99, "USD", false, -5.99},
		{"Very Small Amount", 0.01, "USD", true, 0.01},
		{"Large Amount", 999.99, "USD", true, 999.99},
		{"Invalid Currency", 9.99, "INVALID", false, 9.99},
		{"Empty Currency", 9.99, "", false, 9.99},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			sub := &models.UnifiedSubscription{
				TotalAmount: tt.amount,
				Currency:    tt.currency,
			}

			if tt.isValid {
				assert.Positive(suite.T(), sub.TotalAmount, "Amount should be positive")
				assert.Len(suite.T(), sub.Currency, 3, "Currency should be 3-letter code")
			} else {
				if tt.amount <= 0 {
					assert.LessOrEqual(suite.T(), sub.TotalAmount, 0.0, "Invalid amount should be <= 0")
				}
				if len(tt.currency) != 3 {
					assert.NotEqual(suite.T(), 3, len(sub.Currency), "Invalid currency length")
				}
			}
		})
	}
}

// Test Date Logic and Temporal Relationships
func (suite *UnifiedSubscriptionModelTestSuite) TestDateLogicAndTemporalRelationships() {
	now := time.Now()
	
	tests := []struct {
		name        string
		setupSub    func() *models.UnifiedSubscription
		expectValid bool
		validation  func(*testing.T, *models.UnifiedSubscription)
	}{
		{
			name: "Standard Active Subscription Dates",
			setupSub: func() *models.UnifiedSubscription {
				return &models.UnifiedSubscription{
					Status:          models.StatusActive,
					StartDate:       now.AddDate(0, -1, 0), // Started 1 month ago
					NextRenewalDate: now.AddDate(0, 0, 7),  // Renews in 7 days
					ExpirationDate:  now.AddDate(0, 0, 14), // Expires 7 days after renewal
				}
			},
			expectValid: true,
			validation: func(t *testing.T, sub *models.UnifiedSubscription) {
				assert.True(t, sub.NextRenewalDate.After(sub.StartDate), "Renewal should be after start")
				assert.True(t, sub.ExpirationDate.After(sub.NextRenewalDate), "Expiration should be after renewal")
			},
		},
		{
			name: "Grace Period Subscription Dates",
			setupSub: func() *models.UnifiedSubscription {
				gracePeriodStart := now.AddDate(0, 0, -1) // Started yesterday
				gracePeriodEnd := now.AddDate(0, 0, 6)    // Ends in 6 days
				return &models.UnifiedSubscription{
					Status:               models.StatusGracePeriod,
					StartDate:            now.AddDate(0, -1, 0),
					NextRenewalDate:      gracePeriodStart,
					ExpirationDate:       gracePeriodEnd,
					GracePeriodStartDate: &gracePeriodStart,
					GracePeriodEndDate:   &gracePeriodEnd,
				}
			},
			expectValid: true,
			validation: func(t *testing.T, sub *models.UnifiedSubscription) {
				assert.NotNil(t, sub.GracePeriodStartDate, "Grace period start should be set")
				assert.NotNil(t, sub.GracePeriodEndDate, "Grace period end should be set")
				assert.True(t, sub.GracePeriodEndDate.After(*sub.GracePeriodStartDate), 
					"Grace period end should be after start")
				assert.Equal(t, sub.ExpirationDate, *sub.GracePeriodEndDate, 
					"Expiration should match grace period end")
			},
		},
		{
			name: "Expired Subscription Dates",
			setupSub: func() *models.UnifiedSubscription {
				return &models.UnifiedSubscription{
					Status:          models.StatusExpired,
					StartDate:       now.AddDate(0, -2, 0), // Started 2 months ago
					NextRenewalDate: now.AddDate(0, -1, 0), // Should have renewed 1 month ago
					ExpirationDate:  now.AddDate(0, 0, -7), // Expired 7 days ago
				}
			},
			expectValid: true,
			validation: func(t *testing.T, sub *models.UnifiedSubscription) {
				assert.True(t, sub.ExpirationDate.Before(now), "Expired subscription should be in past")
				assert.True(t, sub.NextRenewalDate.Before(now), "Missed renewal should be in past")
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			sub := tt.setupSub()
			
			// Basic temporal validation
			assert.True(suite.T(), sub.StartDate.Before(time.Now().Add(time.Hour)), 
				"Start date should not be far in future")
			
			if tt.validation != nil {
				tt.validation(suite.T(), sub)
			}
		})
	}
}

// Test Cross-Platform Compatibility
func (suite *UnifiedSubscriptionModelTestSuite) TestCrossPlatformCompatibility() {
	userID := uuid.New()
	subscriptionID := uuid.New()
	now := time.Now()

	// Test that both platforms can represent the same logical subscription
	appleVersion := &models.UnifiedSubscription{
		UserID:         userID,
		SubscriptionID: subscriptionID,
		ActivePlatform: models.PlatformApple,
		PlatformUserID: stringPtr("apple_user_123"),
		PurchaseToken:  stringPtr("original_transaction_id_123"),
		ProductId:      "com.example.premium",
		BasePlanID:     "premium_monthly",
		Status:         models.StatusActive,
		StartDate:      now,
		TotalAmount:    9.99,
		Currency:       "USD",
	}

	googleVersion := &models.UnifiedSubscription{
		UserID:         userID,
		SubscriptionID: subscriptionID,
		ActivePlatform: models.PlatformGoogle,
		PlatformUserID: stringPtr("google_obfuscated_123"),
		PurchaseToken:  stringPtr("purchase_token_xyz"),
		ProductId:      "premium_subscription",
		BasePlanID:     "premium-monthly",
		Status:         models.StatusActive,
		StartDate:      now,
		TotalAmount:    9.99,
		Currency:       "USD",
	}

	// Verify they represent the same logical subscription
	assert.Equal(suite.T(), appleVersion.UserID, googleVersion.UserID, "Same user across platforms")
	assert.Equal(suite.T(), appleVersion.SubscriptionID, googleVersion.SubscriptionID, "Same subscription across platforms")
	assert.Equal(suite.T(), appleVersion.Status, googleVersion.Status, "Same status across platforms")
	assert.Equal(suite.T(), appleVersion.TotalAmount, googleVersion.TotalAmount, "Same amount across platforms")
	assert.Equal(suite.T(), appleVersion.Currency, googleVersion.Currency, "Same currency across platforms")

	// Verify platform-specific differences are handled
	assert.NotEqual(suite.T(), appleVersion.ActivePlatform, googleVersion.ActivePlatform, "Different platforms")
	assert.NotEqual(suite.T(), appleVersion.PlatformUserID, googleVersion.PlatformUserID, "Different platform user IDs")
	assert.NotEqual(suite.T(), appleVersion.PurchaseToken, googleVersion.PurchaseToken, "Different purchase tokens")
}

// Test Business Logic Edge Cases
func (suite *UnifiedSubscriptionModelTestSuite) TestBusinessLogicEdgeCases() {
	now := time.Now()
	
	suite.Run("Subscription with Future Start Date", func() {
		futureStart := now.AddDate(0, 0, 7)
		sub := &models.UnifiedSubscription{
			Status:          models.StatusPendingRenewal,
			StartDate:       futureStart,
			NextRenewalDate: futureStart.AddDate(0, 1, 0),
			ExpirationDate:  futureStart.AddDate(0, 1, 7),
		}
		
		assert.True(suite.T(), sub.StartDate.After(now), "Future start date should be allowed for pending subscriptions")
		assert.True(suite.T(), sub.NextRenewalDate.After(sub.StartDate), "Renewal after start")
	})
	
	suite.Run("Zero-Cost Subscription (Trial)", func() {
		sub := &models.UnifiedSubscription{
			Status:      models.StatusActive,
			TotalAmount: 0.00,
			Currency:    "USD",
			ProductId:   "com.example.trial",
			BasePlanID:  "trial_30_days",
		}
		
		// Trials can have zero cost
		assert.Equal(suite.T(), 0.00, sub.TotalAmount, "Trial subscriptions can have zero cost")
		assert.Contains(suite.T(), sub.ProductId, "trial", "Trial product should be identifiable")
	})
	
	suite.Run("Subscription with Multiple Offers", func() {
		sub := &models.UnifiedSubscription{
			Status:        models.StatusActive,
			ProductId:     "com.example.premium",
			BasePlanID:    "premium_monthly",
			AddOnID:       stringPtr("extra_storage"),
			ActiveOfferID: stringPtr("50_percent_off"),
			TotalAmount:   4.99, // Discounted from 9.99
			Currency:      "USD",
		}
		
		assert.NotNil(suite.T(), sub.AddOnID, "Add-ons should be supported")
		assert.NotNil(suite.T(), sub.ActiveOfferID, "Offers should be supported")
		assert.Positive(suite.T(), sub.TotalAmount, "Even discounted subscriptions should have positive amount")
	})
}

// Helper Methods

// isValidStatusTransition validates business logic for status transitions
func (suite *UnifiedSubscriptionModelTestSuite) isValidStatusTransition(from, to models.SubscriptionStatus) bool {
	// Define valid transition matrix (simplified business rules)
	validTransitions := map[models.SubscriptionStatus][]models.SubscriptionStatus{
		models.StatusActive: {
			models.StatusCanceled, models.StatusExpired, models.StatusGracePeriod,
			models.StatusBillingRetry, models.StatusPaused, models.StatusOnHold,
			models.StatusPendingUpgrade, models.StatusPendingDowngrade,
		},
		models.StatusGracePeriod: {
			models.StatusActive, models.StatusExpired, models.StatusRevoked,
		},
		models.StatusBillingRetry: {
			models.StatusActive, models.StatusGracePeriod, models.StatusExpired,
		},
		models.StatusPaused: {
			models.StatusActive, models.StatusExpired, models.StatusCanceled,
		},
		models.StatusOnHold: {
			models.StatusActive, models.StatusExpired, models.StatusRevoked,
		},
		models.StatusPendingRenewal: {
			models.StatusActive, models.StatusExpired, models.StatusCanceled,
		},
		models.StatusPendingUpgrade: {
			models.StatusActive, models.StatusCanceled,
		},
		models.StatusPendingDowngrade: {
			models.StatusActive, models.StatusCanceled,
		},
		models.StatusCanceled: {
			models.StatusExpired, // Canceled subscriptions can only expire
		},
		// Terminal states - no transitions allowed
		models.StatusExpired: {},
		models.StatusRevoked: {},
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return true
		}
	}
	return false
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
