package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"subsnotifpro-go/internal/subscription/models"
	"subsnotifpro-go/internal/subscription/repository"
	"subsnotifpro-go/internal/pkg/contextutil"
)

type SubscriptionRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.SubscriptionRepository
	ctx  context.Context
}

func TestSubscriptionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionRepositoryTestSuite))
}

func (suite *SubscriptionRepositoryTestSuite) SetupSuite() {
	// Setup in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(suite.T(), err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.UnifiedSubscription{})
	require.NoError(suite.T(), err)

	suite.db = db
	suite.repo = repository.NewSubscriptionRepository(db)
	suite.ctx = context.Background()
}

func (suite *SubscriptionRepositoryTestSuite) SetupTest() {
	// Clean the database before each test - try both possible table names
	suite.db.Exec("DELETE FROM unified_subscriptions")
	suite.db.Exec("DELETE FROM unified_subscription")
}

func (suite *SubscriptionRepositoryTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func (suite *SubscriptionRepositoryTestSuite) TestBasicCRUDOperations() {
	suite.Run("Create and Retrieve Subscription", func() {
		testSub := suite.createTestSubscription()

		// Test upsert within transaction
		err := suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			return suite.repo.UpsertSubscription(ctx, testSub)
		})
		assert.NoError(suite.T(), err)

		// Test retrieval by subscription ID
		retrievedSub, err := suite.repo.GetBySubscriptionID(suite.ctx, testSub.SubscriptionID.String())
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), retrievedSub)
		assert.Equal(suite.T(), testSub.SubscriptionID, retrievedSub.SubscriptionID)
		assert.Equal(suite.T(), testSub.UserID, retrievedSub.UserID)
		assert.Equal(suite.T(), testSub.Status, retrievedSub.Status)

		// Test platform retrieval
		platform, err := suite.repo.GetPlatformBySubscriptionID(suite.ctx, testSub.SubscriptionID.String())
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), testSub.ActivePlatform, platform)
	})
}

func (suite *SubscriptionRepositoryTestSuite) TestTransactionBehavior() {
	suite.Run("Transaction Rollback", func() {
		initialCount := suite.getSubscriptionCount()

		err := suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			testSub := suite.createTestSubscription()
			err := suite.repo.UpsertSubscription(ctx, testSub)
			if err != nil {
				return err
			}
			
			// Simulate an error to trigger rollback
			return assert.AnError
		})

		assert.Error(suite.T(), err)
		finalCount := suite.getSubscriptionCount()
		assert.Equal(suite.T(), initialCount, finalCount, "Transaction should have rolled back")
	})

	suite.Run("Successful Transaction", func() {
		var executed bool
		err := suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			// Verify transaction context is available
			tx, ok := contextutil.TxFromContext(ctx)
			assert.True(suite.T(), ok)
			assert.NotNil(suite.T(), tx)
			executed = true
			return nil
		})

		assert.NoError(suite.T(), err)
		assert.True(suite.T(), executed)
	})
}

func (suite *SubscriptionRepositoryTestSuite) TestSubscriptionUpsert() {
	suite.Run("Update Existing Subscription", func() {
		testSub := suite.createTestSubscription()

		// Create initial subscription
		err := suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			return suite.repo.UpsertSubscription(ctx, testSub)
		})
		assert.NoError(suite.T(), err)

		// Update the subscription
		testSub.Status = models.StatusExpired
		testSub.TotalAmount = 19.99
		testSub.Currency = "EUR"

		err = suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			return suite.repo.UpsertSubscription(ctx, testSub)
		})
		assert.NoError(suite.T(), err)

		// Verify the update
		retrievedSub, err := suite.repo.GetBySubscriptionID(suite.ctx, testSub.SubscriptionID.String())
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), models.StatusExpired, retrievedSub.Status)
		assert.Equal(suite.T(), 19.99, retrievedSub.TotalAmount)
		assert.Equal(suite.T(), "EUR", retrievedSub.Currency)
	})
}

func (suite *SubscriptionRepositoryTestSuite) TestErrorCases() {
	suite.Run("Upsert Without Transaction Context", func() {
		testSub := suite.createTestSubscription()
		err := suite.repo.UpsertSubscription(suite.ctx, testSub)
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "transaction not found in context")
	})

	suite.Run("Get Non-existent Subscription", func() {
		sub, err := suite.repo.GetBySubscriptionID(suite.ctx, uuid.New().String())
		assert.Error(suite.T(), err)
		assert.Nil(suite.T(), sub)
		assert.Contains(suite.T(), err.Error(), "subscription not found")
	})

	suite.Run("Get Platform for Non-existent Subscription", func() {
		platform, err := suite.repo.GetPlatformBySubscriptionID(suite.ctx, uuid.New().String())
		assert.Error(suite.T(), err)
		assert.Empty(suite.T(), platform)
		assert.Contains(suite.T(), err.Error(), "subscription not found")
	})
}

// Helper methods
func (suite *SubscriptionRepositoryTestSuite) createTestSubscription() *models.UnifiedSubscription {
	now := time.Now()
	platformUserID := "platform-user-123"
	latestOrderID := "order-123"
	purchaseToken := "purchase-token-123"
	addOnID := ""
	activeOfferID := ""
	
	return &models.UnifiedSubscription{
		SubscriptionID:         uuid.New(),
		UserID:                 uuid.New(),
		ActivePlatform:         models.PlatformGoogle,
		PlatformUserID:         &platformUserID,
		LatestOrderID:          &latestOrderID,
		PurchaseToken:          &purchaseToken,
		PlanType:               "premium",
		Status:                 models.StatusActive,
		StartDate:              now,
		NextRenewalDate:        now.Add(30 * 24 * time.Hour),
		ExpirationDate:         now.Add(30 * 24 * time.Hour),
		GracePeriodStartDate:   nil,
		GracePeriodEndDate:     nil,
		ProductId:              "product-123",
		BasePlanID:             "base-plan-123",
		AddOnID:                &addOnID,
		ActiveOfferID:          &activeOfferID,
		Currency:               "USD",
		TotalAmount:            9.99,
	}
}

func (suite *SubscriptionRepositoryTestSuite) getSubscriptionCount() int64 {
	var count int64
	// Try both possible table names
	result := suite.db.Model(&models.UnifiedSubscription{}).Count(&count)
	if result.Error != nil {
		// Table might not exist yet, return 0
		return 0
	}
	return count
}
