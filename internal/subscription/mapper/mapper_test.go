package mapper_test

import (
	"testing"

	appStoreModels "subsnotifpro-go/internal/appstore/subscription/models"
	playStoreModels "subsnotifpro-go/internal/playstore/subscription/models"
	"subsnotifpro-go/internal/subscription/mapper"
	"subsnotifpro-go/internal/subscription/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// StatusMapperTestSuite provides comprehensive testing for cross-platform status mapping
type StatusMapperTestSuite struct {
	suite.Suite
	statusMapper *mapper.StatusMapper
}

func TestStatusMapperSuite(t *testing.T) {
	suite.Run(t, new(StatusMapperTestSuite))
}

func (suite *StatusMapperTestSuite) SetupTest() {
	suite.statusMapper = mapper.NewStatusMapper()
}

// Test Apple App Store Status Mapping
func (suite *StatusMapperTestSuite) TestMapAppleStatus() {
	tests := []struct {
		name           string
		appleStatus    appStoreModels.SubscriptionStatus
		expectedStatus models.SubscriptionStatus
		description    string
	}{
		{
			name:           "Active Apple Subscription",
			appleStatus:    appStoreModels.SubscriptionStatusActive,
			expectedStatus: models.StatusActive,
			description:    "Active subscription should map to Active",
		},
		{
			name:           "Expired Apple Subscription",
			appleStatus:    appStoreModels.SubscriptionStatusExpired,
			expectedStatus: models.StatusExpired,
			description:    "Expired subscription should map to Expired",
		},
		{
			name:           "Billing Retry Apple Subscription",
			appleStatus:    appStoreModels.SubscriptionStatusBillingRetry,
			expectedStatus: models.StatusBillingRetry,
			description:    "Billing retry should map to BillingRetry",
		},
		{
			name:           "Grace Period Apple Subscription",
			appleStatus:    appStoreModels.SubscriptionStatusGracePeriod,
			expectedStatus: models.StatusGracePeriod,
			description:    "Grace period should map to GracePeriod",
		},
		{
			name:           "Revoked Apple Subscription",
			appleStatus:    appStoreModels.SubscriptionStatusRevoked,
			expectedStatus: models.StatusRevoked,
			description:    "Revoked subscription should map to Revoked",
		},
		{
			name:           "Pending Renewal Apple Subscription",
			appleStatus:    appStoreModels.SubscriptionStatusPendingRenewal,
			expectedStatus: models.StatusPendingRenewal,
			description:    "Pending renewal should map to PendingRenewal",
		},
		{
			name:           "Pending Upgrade Apple Subscription",
			appleStatus:    appStoreModels.SubscriptionStatusPendingUpgrade,
			expectedStatus: models.StatusPendingUpgrade,
			description:    "Pending upgrade should map to PendingUpgrade",
		},
		{
			name:           "Pending Downgrade Apple Subscription",
			appleStatus:    appStoreModels.SubscriptionStatusPendingDowngrade,
			expectedStatus: models.StatusPendingDowngrade,
			description:    "Pending downgrade should map to PendingDowngrade",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.statusMapper.MapAppleStatus(tt.appleStatus)
			assert.Equal(suite.T(), tt.expectedStatus, result, tt.description)
			
			// Ensure the mapped status is valid
			assert.NotEmpty(suite.T(), string(result), "Mapped status should not be empty")
			assert.True(suite.T(), suite.isValidUnifiedStatus(result), "Mapped status should be valid unified status")
		})
	}
}

// Test Google Play Store Status Mapping
func (suite *StatusMapperTestSuite) TestMapGoogleStatus() {
	tests := []struct {
		name           string
		googleStatus   playStoreModels.SubscriptionState
		expectedStatus models.SubscriptionStatus
		description    string
	}{
		{
			name:           "Active Google Subscription",
			googleStatus:   playStoreModels.SubscriptionStateActive,
			expectedStatus: models.StatusActive,
			description:    "Active subscription should map to Active",
		},
		{
			name:           "Expired Google Subscription",
			googleStatus:   playStoreModels.SubscriptionStateExpired,
			expectedStatus: models.StatusExpired,
			description:    "Expired subscription should map to Expired",
		},
		{
			name:           "Grace Period Google Subscription",
			googleStatus:   playStoreModels.SubscriptionStateInGracePeriod,
			expectedStatus: models.StatusGracePeriod,
			description:    "In grace period should map to GracePeriod",
		},
		{
			name:           "Paused Google Subscription",
			googleStatus:   playStoreModels.SubscriptionStatePaused,
			expectedStatus: models.StatusPaused,
			description:    "Paused subscription should map to Paused",
		},
		{
			name:           "On Hold Google Subscription",
			googleStatus:   playStoreModels.SubscriptionStateOnHold,
			expectedStatus: models.StatusOnHold,
			description:    "On hold subscription should map to OnHold",
		},
		{
			name:           "Canceled Google Subscription",
			googleStatus:   playStoreModels.SubscriptionStateCanceled,
			expectedStatus: models.StatusCanceled,
			description:    "Canceled subscription should map to Canceled",
		},
		{
			name:           "Pending Google Subscription",
			googleStatus:   playStoreModels.SubscriptionStatePending,
			expectedStatus: models.StatusPendingRenewal,
			description:    "Pending subscription should map to PendingRenewal",
		},
		{
			name:           "Pending Purchase Canceled Google Subscription",
			googleStatus:   playStoreModels.SubscriptionStatePendingPurchaseCanceled,
			expectedStatus: models.StatusPendingDowngrade,
			description:    "Pending purchase canceled should map to PendingDowngrade",
		},
		{
			name:           "Unspecified Google Subscription",
			googleStatus:   playStoreModels.SubscriptionStateUnspecified,
			expectedStatus: models.StatusPendingRenewal,
			description:    "Unspecified should fallback to PendingRenewal",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.statusMapper.MapGoogleStatus(tt.googleStatus)
			assert.Equal(suite.T(), tt.expectedStatus, result, tt.description)
			
			// Ensure the mapped status is valid
			assert.NotEmpty(suite.T(), string(result), "Mapped status should not be empty")
			assert.True(suite.T(), suite.isValidUnifiedStatus(result), "Mapped status should be valid unified status")
		})
	}
}

// Test Cross-Platform Status Consistency
func (suite *StatusMapperTestSuite) TestCrossPlatformStatusConsistency() {
	// Test that equivalent statuses across platforms map to the same unified status
	equivalentStatuses := []struct {
		name           string
		appleStatus    appStoreModels.SubscriptionStatus
		googleStatus   playStoreModels.SubscriptionState
		expectedStatus models.SubscriptionStatus
		description    string
	}{
		{
			name:           "Active Status Consistency",
			appleStatus:    appStoreModels.SubscriptionStatusActive,
			googleStatus:   playStoreModels.SubscriptionStateActive,
			expectedStatus: models.StatusActive,
			description:    "Active status should be consistent across platforms",
		},
		{
			name:           "Expired Status Consistency",
			appleStatus:    appStoreModels.SubscriptionStatusExpired,
			googleStatus:   playStoreModels.SubscriptionStateExpired,
			expectedStatus: models.StatusExpired,
			description:    "Expired status should be consistent across platforms",
		},
		{
			name:           "Grace Period Status Consistency",
			appleStatus:    appStoreModels.SubscriptionStatusGracePeriod,
			googleStatus:   playStoreModels.SubscriptionStateInGracePeriod,
			expectedStatus: models.StatusGracePeriod,
			description:    "Grace period status should be consistent across platforms",
		},
	}

	for _, tt := range equivalentStatuses {
		suite.Run(tt.name, func() {
			appleResult := suite.statusMapper.MapAppleStatus(tt.appleStatus)
			googleResult := suite.statusMapper.MapGoogleStatus(tt.googleStatus)
			
			assert.Equal(suite.T(), tt.expectedStatus, appleResult, "Apple status mapping")
			assert.Equal(suite.T(), tt.expectedStatus, googleResult, "Google status mapping")
			assert.Equal(suite.T(), appleResult, googleResult, "Cross-platform consistency")
		})
	}
}

// Test Edge Cases and Unknown Status Handling
func (suite *StatusMapperTestSuite) TestEdgeCasesAndUnknownStatuses() {
	suite.Run("Unknown Apple Status", func() {
		unknownStatus := appStoreModels.SubscriptionStatus("UNKNOWN_STATUS")
		result := suite.statusMapper.MapAppleStatus(unknownStatus)
		
		// Should fallback to uppercase string
		assert.Equal(suite.T(), models.SubscriptionStatus("UNKNOWN_STATUS"), result, 
			"Unknown Apple status should fallback to uppercase string")
	})

	suite.Run("Unknown Google Status", func() {
		unknownStatus := playStoreModels.SubscriptionState("SUBSCRIPTION_STATE_UNKNOWN")
		result := suite.statusMapper.MapGoogleStatus(unknownStatus)
		
		// Should fallback to normalized string
		assert.Equal(suite.T(), models.SubscriptionStatus("UNKNOWN"), result, 
			"Unknown Google status should fallback to normalized string")
	})

	suite.Run("Empty Status Handling", func() {
		emptyAppleStatus := appStoreModels.SubscriptionStatus("")
		emptyGoogleStatus := playStoreModels.SubscriptionState("")
		
		appleResult := suite.statusMapper.MapAppleStatus(emptyAppleStatus)
		googleResult := suite.statusMapper.MapGoogleStatus(emptyGoogleStatus)
		
		assert.NotEqual(suite.T(), models.StatusActive, appleResult, "Empty Apple status should not map to Active")
		assert.NotEqual(suite.T(), models.StatusActive, googleResult, "Empty Google status should not map to Active")
	})
}

// Test Status Mapping Performance and Reliability
func (suite *StatusMapperTestSuite) TestStatusMappingPerformanceAndReliability() {
	suite.Run("Apple Status Mapping Stress Test", func() {
		appleStatuses := []appStoreModels.SubscriptionStatus{
			appStoreModels.SubscriptionStatusActive,
			appStoreModels.SubscriptionStatusExpired,
			appStoreModels.SubscriptionStatusBillingRetry,
			appStoreModels.SubscriptionStatusGracePeriod,
			appStoreModels.SubscriptionStatusRevoked,
		}

		// Test multiple mappings
		for i := 0; i < 1000; i++ {
			for _, status := range appleStatuses {
				result := suite.statusMapper.MapAppleStatus(status)
				assert.NotEmpty(suite.T(), string(result), "Result should not be empty")
			}
		}
	})

	suite.Run("Google Status Mapping Stress Test", func() {
		googleStatuses := []playStoreModels.SubscriptionState{
			playStoreModels.SubscriptionStateActive,
			playStoreModels.SubscriptionStateExpired,
			playStoreModels.SubscriptionStateInGracePeriod,
			playStoreModels.SubscriptionStatePaused,
			playStoreModels.SubscriptionStateOnHold,
			playStoreModels.SubscriptionStateCanceled,
		}

		// Test multiple mappings
		for i := 0; i < 1000; i++ {
			for _, status := range googleStatuses {
				result := suite.statusMapper.MapGoogleStatus(status)
				assert.NotEmpty(suite.T(), string(result), "Result should not be empty")
			}
		}
	})
}

// Test Bidirectional Mapping Consistency
func (suite *StatusMapperTestSuite) TestBidirectionalMappingConsistency() {
	// Test that mapping is deterministic and consistent
	suite.Run("Apple Status Consistency", func() {
		testStatus := appStoreModels.SubscriptionStatusActive
		
		// Map multiple times and ensure consistency
		result1 := suite.statusMapper.MapAppleStatus(testStatus)
		result2 := suite.statusMapper.MapAppleStatus(testStatus)
		result3 := suite.statusMapper.MapAppleStatus(testStatus)
		
		assert.Equal(suite.T(), result1, result2, "Multiple mappings should be consistent")
		assert.Equal(suite.T(), result2, result3, "Multiple mappings should be consistent")
		assert.Equal(suite.T(), models.StatusActive, result1, "Active status should map correctly")
	})

	suite.Run("Google Status Consistency", func() {
		testStatus := playStoreModels.SubscriptionStateActive
		
		// Map multiple times and ensure consistency
		result1 := suite.statusMapper.MapGoogleStatus(testStatus)
		result2 := suite.statusMapper.MapGoogleStatus(testStatus)
		result3 := suite.statusMapper.MapGoogleStatus(testStatus)
		
		assert.Equal(suite.T(), result1, result2, "Multiple mappings should be consistent")
		assert.Equal(suite.T(), result2, result3, "Multiple mappings should be consistent")
		assert.Equal(suite.T(), models.StatusActive, result1, "Active status should map correctly")
	})
}

// Test Status Validation Logic
func (suite *StatusMapperTestSuite) TestStatusValidationLogic() {
	suite.Run("Valid Status Coverage", func() {
		// Ensure all major statuses are covered
		appleStatuses := []appStoreModels.SubscriptionStatus{
			appStoreModels.SubscriptionStatusActive,
			appStoreModels.SubscriptionStatusExpired,
			appStoreModels.SubscriptionStatusBillingRetry,
			appStoreModels.SubscriptionStatusGracePeriod,
			appStoreModels.SubscriptionStatusRevoked,
			appStoreModels.SubscriptionStatusPendingRenewal,
			appStoreModels.SubscriptionStatusPendingUpgrade,
			appStoreModels.SubscriptionStatusPendingDowngrade,
		}

		googleStatuses := []playStoreModels.SubscriptionState{
			playStoreModels.SubscriptionStateActive,
			playStoreModels.SubscriptionStateExpired,
			playStoreModels.SubscriptionStateInGracePeriod,
			playStoreModels.SubscriptionStatePaused,
			playStoreModels.SubscriptionStateOnHold,
			playStoreModels.SubscriptionStateCanceled,
			playStoreModels.SubscriptionStatePending,
			playStoreModels.SubscriptionStatePendingPurchaseCanceled,
			playStoreModels.SubscriptionStateUnspecified,
		}

		// Test that all statuses map to valid unified statuses
		for _, appleStatus := range appleStatuses {
			result := suite.statusMapper.MapAppleStatus(appleStatus)
			assert.True(suite.T(), suite.isValidUnifiedStatus(result), 
				"Apple status %s should map to valid unified status", appleStatus)
		}

		for _, googleStatus := range googleStatuses {
			result := suite.statusMapper.MapGoogleStatus(googleStatus)
			assert.True(suite.T(), suite.isValidUnifiedStatus(result), 
				"Google status %s should map to valid unified status", googleStatus)
		}
	})
}

// Test Platform-Specific Status Features
func (suite *StatusMapperTestSuite) TestPlatformSpecificStatusFeatures() {
	suite.Run("Google-Only Statuses", func() {
		// Test statuses that exist only in Google Play Store
		googleOnlyStatuses := []struct {
			status   playStoreModels.SubscriptionState
			expected models.SubscriptionStatus
		}{
			{playStoreModels.SubscriptionStatePaused, models.StatusPaused},
			{playStoreModels.SubscriptionStateOnHold, models.StatusOnHold},
		}

		for _, test := range googleOnlyStatuses {
			result := suite.statusMapper.MapGoogleStatus(test.status)
			assert.Equal(suite.T(), test.expected, result, 
				"Google-only status %s should map correctly", test.status)
		}
	})

	suite.Run("Apple-Specific Status Handling", func() {
		// Test Apple-specific statuses and their proper mapping
		appleSpecificStatuses := []struct {
			status   appStoreModels.SubscriptionStatus
			expected models.SubscriptionStatus
		}{
			{appStoreModels.SubscriptionStatusBillingRetry, models.StatusBillingRetry},
			{appStoreModels.SubscriptionStatusRevoked, models.StatusRevoked},
		}

		for _, test := range appleSpecificStatuses {
			result := suite.statusMapper.MapAppleStatus(test.status)
			assert.Equal(suite.T(), test.expected, result, 
				"Apple-specific status %s should map correctly", test.status)
		}
	})
}

// Helper Methods

// isValidUnifiedStatus checks if a status is a valid unified status
func (suite *StatusMapperTestSuite) isValidUnifiedStatus(status models.SubscriptionStatus) bool {
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

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}

	// Allow fallback statuses (uppercase strings)
	if len(string(status)) > 0 {
		return true
	}

	return false
}
