package service

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"google.golang.org/api/androidpublisher/v3"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) ResolveLinkedPurchaseToken(
	ctx context.Context,
	tx *gorm.DB,
	newSubData *androidpublisher.SubscriptionPurchaseV2,
	newSubModelId uuid.UUID,
	newPurchaseToken string,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newSubData.LinkedPurchaseToken == "" {
		return nil, nil
	}

	linkedToken := newSubData.LinkedPurchaseToken

	// 1. Lookup old subscription
	oldSub, err := s.repo.GetSubscriptionByPurchaseToken(ctx, tx, linkedToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return nil, nil // Non-critical
		}
		return nil, fmt.Errorf("failed to find linked subscription: %w", err)
	}

	if err := s.createLinkageRecord(
		ctx,
		tx,
		oldSub.ID,
		changeEventID,
		linkedToken,
		&newSubModelId,
		newPurchaseToken,
	); err != nil {
		return nil, fmt.Errorf("failed to create subscription linking record: %w", err)
	}

	return &oldSub.ID, nil
}

func (s *playstoreSubscriptionService) createLinkageRecord(
	ctx context.Context,
	tx *gorm.DB,
	oldSubID uuid.UUID,
	changeEventID uuid.UUID,
	linkedToken string,
	newSubModelID *uuid.UUID,
	newPurchaseToken string,
) error {
	linkage := &models.SubscriptionLinkage{
		OldSubscriptionID:   oldSubID,
		NewSubscriptionID:   *newSubModelID,
		ChangeEventID:       changeEventID,
		LinkedPurchaseToken: linkedToken,
		NewPurchaseToken:    newPurchaseToken,
	}

	return tx.WithContext(ctx).Create(linkage).Error
}
