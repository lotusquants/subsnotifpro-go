package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/appstore/events"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"subsnotifpro-go/internal/appstore/webhooks/models"
	"subsnotifpro-go/internal/pkg/logger"
	"time"

	"github.com/google/uuid"
)

func (s *appStoreNotificationsService) ProcessNotificationForPublish(ctx context.Context, dto *dto.AppStoreNotification) error {
	// Declare notificationModel outside the transaction
	var notificationModel *models.AppStoreNotification
	// 1. Transactional database operations
	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {

		var err error

		// Convert DTO to domain model
		notificationModel, err = models.FromAppStoreNotificationDTO(dto)
		if err != nil {
			return fmt.Errorf("conversion failed: %w", err)
		}

		// Save notification
		if err := s.repo.Save(txCtx, notificationModel); err != nil {
			return fmt.Errorf("save failed: %w", err)
		}

		// Update initial status
		return s.repo.UpdateStatus(txCtx, notificationModel.ID, models.StatusReceived, "")
	})

	if err != nil {
		logger.Log.Error(ctx, "Transaction failed", "error", err)
		return err
	}

	// Create event with just the ID
	event := &events.AppStorePublishPayload{
		ID: notificationModel.ID,
	}

	// 2. Publish event with timeout and circuit breaker
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = s.cb.Execute(func() (interface{}, error) {
		return nil, s.publisher.PublishAppStoreEvent(publishCtx, event)
	})
	if err != nil {
		// Async status update with retries
		go s.safeUpdateStatusWithRetry(context.Background(), notificationModel.ID, models.StatusFailed, err.Error())
		return fmt.Errorf("queue publish failed: %w", err)
	}

	// 3. Update status to "published" after successful publishing
	err = s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		return s.repo.UpdateStatus(txCtx, notificationModel.ID, models.StatusPublished, "")
	})
	if err != nil {
		logger.Log.Error(ctx, "Failed to update status to published", "error", err)
		return fmt.Errorf("failed to update status to published: %w", err)
	}

	return nil
}

// safeUpdateStatusWithRetry handles async status updates with retries
func (s *appStoreNotificationsService) safeUpdateStatusWithRetry(ctx context.Context, id uuid.UUID, status models.WebhookEventStatus, message string) {
	const maxRetries = 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
			return s.repo.UpdateStatus(txCtx, id, status, message)
		})
		if err == nil {
			return
		}
		lastErr = err
		time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
	}

	logger.Log.Error(ctx, "Failed to update status after retries",
		"id", id,
		"status", status,
		"error", lastErr)
}
