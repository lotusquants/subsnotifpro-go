package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"subsnotifpro-go/internal/appstore/webhooks/models"
)

type AppStoreWebhookRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo AppstoreNotificationsRepository
	ctx  context.Context
}

func TestAppStoreWebhookRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AppStoreWebhookRepositoryTestSuite))
}

func (suite *AppStoreWebhookRepositoryTestSuite) SetupTest() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)

	// Create a simple test table for SQLite compatibility
	err = db.Exec(`
		CREATE TABLE app_store_notifications (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			received_at DATETIME,
			status TEXT DEFAULT 'RECEIVED',
			retry_count INTEGER DEFAULT 0,
			status_message TEXT
		)
	`).Error
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewAppstoreNotificationsRepository(db)
	suite.ctx = context.Background()
}

func (suite *AppStoreWebhookRepositoryTestSuite) TestBasicCRUDOperations() {
	suite.Run("Save_and_Retrieve_Notification", func() {
		notification := suite.createTestNotification()

		// Test save within transaction
		err := suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			return suite.repo.Save(ctx, notification)
		})
		assert.NoError(suite.T(), err)

		// Test retrieve
		retrievedNotification, err := suite.repo.GetByID(suite.ctx, notification.ID)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), retrievedNotification)
		assert.Equal(suite.T(), notification.ID, retrievedNotification.ID)
		assert.Equal(suite.T(), notification.Status, retrievedNotification.Status)
		assert.Equal(suite.T(), notification.RetryCount, retrievedNotification.RetryCount)
	})

	suite.Run("Check_Notification_Existence", func() {
		notification := suite.createTestNotification()

		// Initially should not exist
		exists, err := suite.repo.Exists(suite.ctx, notification.ID)
		assert.NoError(suite.T(), err)
		assert.False(suite.T(), exists)

		// Save and check again
		err = suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			return suite.repo.Save(ctx, notification)
		})
		assert.NoError(suite.T(), err)

		// Now should exist
		exists, err = suite.repo.Exists(suite.ctx, notification.ID)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), exists)
	})
}

func (suite *AppStoreWebhookRepositoryTestSuite) TestStatusOperations() {
	suite.Run("Update_Notification_Status", func() {
		notification := suite.createTestNotification()

		// Save notification first
		err := suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			return suite.repo.Save(ctx, notification)
		})
		assert.NoError(suite.T(), err)

		// Update status
		err = suite.repo.UpdateStatus(suite.ctx, notification.ID, models.StatusProcessing, "Processing webhook")
		assert.NoError(suite.T(), err)

		// Verify status update
		retrievedNotification, err := suite.repo.GetByID(suite.ctx, notification.ID)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), models.StatusProcessing, retrievedNotification.Status)
	})

	suite.Run("Increment_Retry_Count", func() {
		notification := suite.createTestNotification()

		// Save notification first
		err := suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			return suite.repo.Save(ctx, notification)
		})
		assert.NoError(suite.T(), err)

		// Increment retry count
		err = suite.repo.IncrementRetryCount(suite.ctx, notification.ID)
		assert.NoError(suite.T(), err)

		// Verify retry count increment
		retrievedNotification, err := suite.repo.GetByID(suite.ctx, notification.ID)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), int32(1), retrievedNotification.RetryCount)
	})
}

func (suite *AppStoreWebhookRepositoryTestSuite) TestTransactionBehavior() {
	suite.Run("Successful_Transaction", func() {
		initialCount := suite.getNotificationCount()

		notification := suite.createTestNotification()
		err := suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			return suite.repo.Save(ctx, notification)
		})

		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), initialCount+1, suite.getNotificationCount())
	})

	suite.Run("Transaction_Rollback_on_Error", func() {
		initialCount := suite.getNotificationCount()

		err := suite.repo.WithTransaction(suite.ctx, func(ctx context.Context) error {
			// This should fail due to invalid context manipulation
			return assert.AnError
		})

		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), initialCount, suite.getNotificationCount())
	})
}

func (suite *AppStoreWebhookRepositoryTestSuite) TestErrorCases() {
	suite.Run("Get_Non-existent_Notification", func() {
		nonExistentID := uuid.New()
		notification, err := suite.repo.GetByID(suite.ctx, nonExistentID)

		// Should return error for non-existent notification
		assert.Error(suite.T(), err)
		assert.Nil(suite.T(), notification)
	})

	suite.Run("Update_Status_for_Non-existent_Notification", func() {
		nonExistentID := uuid.New()
		err := suite.repo.UpdateStatus(suite.ctx, nonExistentID, models.StatusProcessing, "test message")
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "notification not found")
	})

	suite.Run("Increment_Retry_Count_for_Non-existent_Notification", func() {
		nonExistentID := uuid.New()
		err := suite.repo.IncrementRetryCount(suite.ctx, nonExistentID)
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "notification not found")
	})
}

// Helper methods
func (suite *AppStoreWebhookRepositoryTestSuite) createTestNotification() *models.AppStoreNotification {
	id := uuid.New()
	return &models.AppStoreNotification{
		ID:         id,
		Status:     models.StatusReceived,
		RetryCount: 0,
	}
}

func (suite *AppStoreWebhookRepositoryTestSuite) getNotificationCount() int64 {
	var count int64
	result := suite.db.Raw("SELECT COUNT(*) FROM app_store_notifications").Scan(&count)
	if result.Error != nil {
		return 0
	}
	return count
}
