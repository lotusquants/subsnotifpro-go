package service

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/playstore/api/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) ResolveLinkedPurchaseToken(
	ctx context.Context,
	tx *gorm.DB,
	newSubData *dto.SubscriptionPurchaseV2,
	newSubModelId uuid.UUID,
	newPurchaseToken string,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newSubData.LinkedPurchaseToken == nil {
		return nil, nil
	}

	linkedToken := *newSubData.LinkedPurchaseToken
	if linkedToken == "" {
		return nil, nil
	}

	// 1. Lookup old subscription
	oldSub, err := s.repo.GetSubscriptionByPurchaseToken(ctx, tx, linkedToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return nil, nil // Non-critical
		}
		return nil, fmt.Errorf("failed to find linked subscription: %w", err)
	}

	return &oldSub.ID, nil
}
