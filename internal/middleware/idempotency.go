package middleware

import (
	"context"
	"fmt"
	"time"

	"subsnotifpro-go/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IdempotencyRecord represents a processed notification for idempotency checking
type IdempotencyRecord struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventID     string    `gorm:"uniqueIndex;not null"`
	EventType   string    `gorm:"not null"`
	ProcessedAt time.Time `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
	ExpiresAt   time.Time `gorm:"not null;index"`
}

// IdempotencyManager handles idempotency checking for webhook notifications
type IdempotencyManager struct {
	db *gorm.DB
}

// NewIdempotencyManager creates a new idempotency manager
func NewIdempotencyManager(db *gorm.DB) *IdempotencyManager {
	return &IdempotencyManager{db: db}
}

// CheckIdempotency checks if a notification has already been processed
func (im *IdempotencyManager) CheckIdempotency(ctx context.Context, eventID, eventType string) (bool, error) {
	if eventID == "" {
		return false, fmt.Errorf("eventID cannot be empty")
	}

	var count int64
	err := im.db.WithContext(ctx).Model(&IdempotencyRecord{}).
		Where("event_id = ? AND event_type = ?", eventID, eventType).
		Count(&count).Error

	if err != nil {
		logger.Log.Errorf("Failed to check idempotency: %v", err)
		return false, err
	}

	alreadyProcessed := count > 0
	if alreadyProcessed {
		logger.Log.Infof("Duplicate notification detected: %s (type: %s)", eventID, eventType)
	}

	return alreadyProcessed, nil
}

// RecordProcessed records that a notification has been processed
func (im *IdempotencyManager) RecordProcessed(ctx context.Context, eventID, eventType string) error {
	if eventID == "" {
		return fmt.Errorf("eventID cannot be empty")
	}

	now := time.Now()
	record := &IdempotencyRecord{
		EventID:     eventID,
		EventType:   eventType,
		ProcessedAt: now,
		CreatedAt:   now,
		ExpiresAt:   now.Add(7 * 24 * time.Hour), // Keep records for 7 days
	}

	err := im.db.WithContext(ctx).Create(record).Error
	if err != nil {
		logger.Log.Errorf("Failed to record idempotency: %v", err)
		return err
	}

	logger.Log.Infof("Recorded processed notification: %s (type: %s)", eventID, eventType)
	return nil
}

// CleanupExpiredRecords removes expired idempotency records
func (im *IdempotencyManager) CleanupExpiredRecords(ctx context.Context) error {
	result := im.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&IdempotencyRecord{})

	if result.Error != nil {
		logger.Log.Errorf("Failed to cleanup expired idempotency records: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected > 0 {
		logger.Log.Infof("Cleaned up %d expired idempotency records", result.RowsAffected)
	}

	return nil
}

// IdempotencyMiddleware creates a Gin middleware for idempotency checking
func IdempotencyMiddleware(im *IdempotencyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract event ID from context (set by webhook handlers)
		eventID, exists := c.Get("eventID")
		if !exists {
			// If no event ID is set, skip idempotency check
			c.Next()
			return
		}

		eventType, exists := c.Get("eventType")
		if !exists {
			eventType = "unknown"
		}

		// Check if already processed
		processed, err := im.CheckIdempotency(c.Request.Context(), eventID.(string), eventType.(string))
		if err != nil {
			logger.Log.Errorf("Idempotency check failed: %v", err)
			c.JSON(500, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		if processed {
			logger.Log.Infof("Duplicate notification ignored: %s", eventID)
			c.JSON(200, gin.H{"status": "already processed"})
			c.Abort()
			return
		}

		// Continue processing
		c.Next()

		// Record as processed if the request was successful
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			err := im.RecordProcessed(c.Request.Context(), eventID.(string), eventType.(string))
			if err != nil {
				logger.Log.Errorf("Failed to record processed notification: %v", err)
			}
		}
	}
}

// AutoMigrate creates the idempotency table
func (im *IdempotencyManager) AutoMigrate() error {
	return im.db.AutoMigrate(&IdempotencyRecord{})
}
