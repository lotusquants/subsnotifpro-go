package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"subsnotifpro-go/internal/appstore/subscription/models"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"subsnotifpro-go/internal/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *appstoreSubscriptionService) handleConsumptionRequest(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload
	consumptionRequestReason := notification.ResponseBodyV2DecodedPayload.Data.ConsumptionRequestReason

	// Find existing subscription
	subscription, err := s.repo.FindByOriginalTransactionID(ctx, tx, payload.OriginalTransactionId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subscription not found for original transaction ID: %s",
				payload.OriginalTransactionId)
		}
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	// Verify user ownership
	if userID != nil && subscription.UserID != uuid.Nil && subscription.UserID != *userID {
		return nil, fmt.Errorf("user ID mismatch: subscription belongs to %s but request is for %s",
			subscription.UserID, *userID)
	}

	// Update subscription with consumption reason
	subscription.ConsumptionRequestReason = consumptionRequestReason
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	// Create consumption request record
	consumptionReq := &models.ConsumptionRequest{
		SubscriptionID:        subscription.ID,
		OriginalTransactionID: payload.OriginalTransactionId,
		AccountTenure:         s.getAccountTenure(subscription.UserID),
		AppAccountToken:       subscription.AppAccountToken,
		CustomerConsented:     true, // Required by Apple
		DeliveryStatus:        0,    // 0 = delivered
		// ... set other fields ...
	}

	// Save consumption request
	if err := s.consumptionRepo.Create(ctx, tx, consumptionReq); err != nil {
		return nil, fmt.Errorf("failed to create consumption request: %w", err)
	}

	// Create event
	reason := fmt.Sprintf("Apple requested consumption info (reason: %s)", *consumptionRequestReason)
	if err := subscription.AddEvent(tx, models.EventTypeConsumptionRequest, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add consumption request event: %w", err)
	}

	logger.Log.Infof("Created consumption request %s for subscription %s",
		consumptionReq.ID, subscription.ID)

	// // Get customer consent (pseudo-code)
	// if consented, err := s.requestConsent(ctx, subscription.UserID); err != nil {
	// 	return fmt.Errorf("failed to get customer consent: %w", err)
	// } else if !consented {
	// 	logger.Log.Warnf("Customer declined to provide consumption data")
	// 	return nil
	// }

	// Prepare and send consumption info
	if err := s.sendConsumptionData(ctx, tx, consumptionReq); err != nil {
		// Update request status
		consumptionReq.RequestStatus = "failed"
		if updateErr := s.consumptionRepo.Update(ctx, tx, consumptionReq); updateErr != nil {
			logger.Log.Errorf("Failed to update failed consumption request: %v", updateErr)
		}
		return nil, fmt.Errorf("failed to send consumption data: %w", err)
	}

	// Mark request as completed
	consumptionReq.RequestStatus = "completed"
	now := time.Now()
	consumptionReq.ResponseReceived = &now
	if err := s.consumptionRepo.Update(ctx, tx, consumptionReq); err != nil {
		logger.Log.Errorf("Failed to mark consumption request as completed: %v", err)
	}

	return subscription, nil
}

func (s *appstoreSubscriptionService) sendConsumptionData(
	ctx context.Context,
	tx *gorm.DB,
	req *models.ConsumptionRequest,
) error {
	// // Convert model to Apple's API format
	// appleReq := dto.ConsumptionRequest{
	// 	AccountTenure:            req.AccountTenure,
	// 	AppAccountToken:          req.AppAccountToken,
	// 	ConsumptionStatus:        dto.ConsumptionStatus(req.ConsumptionStatus),
	// 	CustomerConsented:        req.CustomerConsented,
	// 	DeliveryStatus:           dto.DeliveryStatus(req.DeliveryStatus),
	// 	LifetimeDollarsPurchased: req.LifetimeDollarsPurchased,
	// 	LifetimeDollarsRefunded:  req.LifetimeDollarsRefunded,
	// 	Platform:                 dto.Platform(req.Platform),
	// 	PlayTime:                 req.PlayTime,
	// 	SampleContentProvided:    req.SampleContentProvided,
	// 	UserStatus:               dto.UserStatus(req.UserStatus),
	// }

	// // Send to Apple (your existing sendConsumptionInfo implementation)
	// if err := s.sendConsumptionInfo(ctx, req.OriginalTransactionID, appleReq); err != nil {
	// 	return err
	// }

	return nil
}

// Helper methods for gathering consumption data
func (s *appstoreSubscriptionService) getAccountTenure(userID uuid.UUID) int32 {
	// Implement logic to get account tenure in days
	// Return 0 if undeclared
	return 0
}

func (s *appstoreSubscriptionService) getLifetimePurchases(userID uuid.UUID) string {
	// Implement logic to get lifetime purchases in USD
	// Format as "0.00" if undeclared
	return "0.00"
}

func (s *appstoreSubscriptionService) getLifetimeRefunds(userID uuid.UUID) string {
	// Implement logic to get lifetime refunds in USD
	// Format as "0.00" if undeclared
	return "0.00"
}

func (s *appstoreSubscriptionService) getPlayTime(userID uuid.UUID, productID string) int32 {
	// Implement logic to get playtime in minutes
	// Return 0 if undeclared
	return 0
}

func (s *appstoreSubscriptionService) sendConsumptionInfo(
	ctx context.Context,
	transactionID string,
	consumptionInfo dto.ConsumptionRequest,
) error {
	// // Determine API endpoint based on environment
	// endpoint := "https://api.storekit.itunes.apple.com/inApps/v1/transactions/consumption/" + transactionID
	// if s.isSandboxEnvironment() {
	// 	endpoint = "https://api.storekit-sandbox.itunes.apple.com/inApps/v1/transactions/consumption/" + transactionID
	// }

	// // Create HTTP request
	// req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, nil)
	// if err != nil {
	// 	return fmt.Errorf("failed to create request: %w", err)
	// }

	// // Set headers
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Authorization", "Bearer "+s.generateJWTToken())

	// // Marshal consumption info to JSON
	// body, err := json.Marshal(consumptionInfo)
	// if err != nil {
	// 	return fmt.Errorf("failed to marshal consumption info: %w", err)
	// }
	// req.Body = io.NopCloser(bytes.NewReader(body))

	// // Send request
	// resp, err := s.httpClient.Do(req)
	// if err != nil {
	// 	return fmt.Errorf("failed to send consumption info: %w", err)
	// }
	// defer resp.Body.Close()

	// // Handle response
	// switch resp.StatusCode {
	// case http.StatusAccepted:
	// 	return nil // Success
	// case http.StatusBadRequest:
	// 	return parseConsumptionError(resp.Body)
	// case http.StatusUnauthorized:
	// 	return fmt.Errorf("unauthorized - invalid JWT token")
	// case http.StatusNotFound:
	// 	return fmt.Errorf("transaction ID not found")
	// case http.StatusTooManyRequests:
	// 	return fmt.Errorf("rate limit exceeded")
	// case http.StatusInternalServerError:
	// 	return fmt.Errorf("app store server error")
	// default:
	// 	return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	// }
	return nil
}

func parseConsumptionError(body io.Reader) error {
	// Parse Apple's error response
	var errorResponse struct {
		ErrorCode string `json:"errorCode"`
	}
	if err := json.NewDecoder(body).Decode(&errorResponse); err != nil {
		return fmt.Errorf("bad request with unparseable error")
	}
	return fmt.Errorf("bad request - error code: %s", errorResponse.ErrorCode)
}
