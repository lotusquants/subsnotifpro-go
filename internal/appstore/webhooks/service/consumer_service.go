// internal/appstore/webhooks/service/processor.go
package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"subsnotifpro-go/internal/metrics"
	"subsnotifpro-go/internal/pkg/logger"
	"time"

	"github.com/google/uuid"
)

func (s *appStoreNotificationsService) ProcessAppStoreEvent(ctx context.Context, id uuid.UUID) error {
	startTime := time.Now()

	// 1. Get the full notification from DB
	notification, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}
	if notification == nil {
		return fmt.Errorf("notification not found with id: %s", id)
	}

	// 2. Convert to DTO
	notificationDTO := notification.ToDTO()
	if notificationDTO == nil ||
		notificationDTO.ResponseBodyV2DecodedPayload.Data == nil {
		return fmt.Errorf("invalid notification data structure")
	}

	processErr := s.processSubscriptionEvent(ctx, notificationDTO)

	// Record metrics with proper nil checks
	bundleID := "unknown"
	if notificationDTO.ResponseBodyV2DecodedPayload.Data.BundleID != "" {
		bundleID = notificationDTO.ResponseBodyV2DecodedPayload.Data.BundleID
	}

	metrics.EventProcessingTime.WithLabelValues(bundleID).Observe(time.Since(startTime).Seconds())
	if processErr != nil {
		metrics.FailedEvents.WithLabelValues(bundleID).Inc()
		return processErr
	}

	metrics.ProcessedEvents.WithLabelValues(bundleID).Inc()
	return nil
}

func (s *appStoreNotificationsService) processSubscriptionEvent(ctx context.Context, dto *dto.AppStoreNotification) error {
	logger.Log.Info("🔹 Processing App Store Subscription Event")

	// Comprehensive validation
	if dto == nil || dto.ResponseBodyV2DecodedPayload.Data == nil {
		return fmt.Errorf("nil data in subscription event")
	}

	if dto.ResponseBodyV2DecodedPayload.Data.BundleID == "" {
		return fmt.Errorf("invalid subscription event: missing bundle ID")
	}

	// Process the subscription
	err := s.subscriptionService.ProcessSubscriptionNotification(ctx, dto)
	if err != nil {
		return fmt.Errorf("failed to process subscription: %w", err)
	}

	logger.Log.Infof("✅ Subscription processed successfully for transaction %s",
		dto.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload.TransactionId)
	return nil
}
