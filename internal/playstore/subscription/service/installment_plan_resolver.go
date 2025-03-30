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

func (s *playstoreSubscriptionService) createInstallmentPlan(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	lineItem *models.SubscriptionLineItem,
	planData *androidpublisher.SubscriptionPurchaseLineItem,
	changeEventID uuid.UUID,
) (*models.InstallmentPlan, error) {
	if planData.AutoRenewingPlan == nil || planData.AutoRenewingPlan.InstallmentDetails == nil {
		return nil, nil
	}

	details := planData.AutoRenewingPlan.InstallmentDetails
	installmentPlan := &models.InstallmentPlan{
		SubscriptionID:                  subscriptionID,
		LineItemID:                      lineItem.ID,
		AutoRenewingPlanID:              lineItem.AutoRenewingPlan.ID,
		InitialCommittedPaymentsCount:   int(details.InitialCommittedPaymentsCount),
		RemainingCommittedPaymentsCount: int(details.InitialCommittedPaymentsCount),
		PendingCancellation:             details.PendingCancellation != nil,
	}

	// Handle optional subsequent payments
	if details.SubsequentCommittedPaymentsCount > 0 {
		subsequentCount := int(details.SubsequentCommittedPaymentsCount)
		installmentPlan.SubsequentCommittedPaymentsCount = &subsequentCount
	}

	if err := tx.Create(installmentPlan).Error; err != nil {
		return nil, fmt.Errorf("failed to create installment plan: %w", err)
	}

	// Create history entry
	history := models.InstallmentPlanHistory{
		InstallmentPlanID:          installmentPlan.ID,
		SubscriptionID:             subscriptionID,
		LineItemID:                 lineItem.ID,
		AutoRenewingPlanID:         lineItem.AutoRenewingPlan.ID,
		NewInitialPaymentsCount:    installmentPlan.InitialCommittedPaymentsCount,
		NewSubsequentPaymentsCount: installmentPlan.SubsequentCommittedPaymentsCount,
		NewRemainingPaymentsCount:  installmentPlan.RemainingCommittedPaymentsCount,
		NewPendingCancellation:     installmentPlan.PendingCancellation,
		ChangeType:                 "CREATED",
		ChangeEventID:              changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create installment plan history: %w", err)
	}

	return installmentPlan, nil
}

func (s *playstoreSubscriptionService) updateInstallmentPlan(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existingLineItem *models.SubscriptionLineItem,
	planData *androidpublisher.SubscriptionPurchaseLineItem,
	changeEventID uuid.UUID,
) (*models.InstallmentPlan, error) {
	if planData.AutoRenewingPlan == nil || planData.AutoRenewingPlan.InstallmentDetails == nil {
		return nil, nil
	}

	if existingLineItem.AutoRenewingPlan == nil || existingLineItem.AutoRenewingPlan.InstallmentPlan == nil {
		return nil, nil
	}

	details := planData.AutoRenewingPlan.InstallmentDetails
	newSubsequentCount := int(details.SubsequentCommittedPaymentsCount)
	var subsequentCountPtr *int
	if newSubsequentCount > 0 {
		subsequentCountPtr = &newSubsequentCount
	}

	existing := existingLineItem.AutoRenewingPlan.InstallmentPlan

	// Prepare updates
	updates := models.InstallmentPlan{
		InitialCommittedPaymentsCount:    int(details.InitialCommittedPaymentsCount),
		RemainingCommittedPaymentsCount:  int(details.RemainingCommittedPaymentsCount),
		SubsequentCommittedPaymentsCount: subsequentCountPtr,
		PendingCancellation:              details.PendingCancellation != nil,
	}

	// Create history entry before updating
	history := models.InstallmentPlanHistory{
		InstallmentPlanID:               existing.ID,
		SubscriptionID:                  subscriptionID,
		LineItemID:                      existingLineItem.ID,
		PreviousInitialPaymentsCount:    &existing.InitialCommittedPaymentsCount,
		PreviousSubsequentPaymentsCount: existing.SubsequentCommittedPaymentsCount,
		PreviousRemainingPaymentsCount:  &existing.RemainingCommittedPaymentsCount,
		PreviousPendingCancellation:     &existing.PendingCancellation,
		NewInitialPaymentsCount:         updates.InitialCommittedPaymentsCount,
		NewSubsequentPaymentsCount:      updates.SubsequentCommittedPaymentsCount,
		NewRemainingPaymentsCount:       updates.RemainingCommittedPaymentsCount,
		NewPendingCancellation:          updates.PendingCancellation,
		ChangeType:                      "UPDATED",
		ChangeEventID:                   changeEventID,
	}

	// Update the existing record
	if err := tx.Model(&existing).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update installment plan: %w", err)
	}

	// Create history entry
	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create installment plan history: %w", err)
	}

	return existing, nil
}

func (s *playstoreSubscriptionService) expireInstallmentPlan(
	ctx context.Context,
	tx *gorm.DB,
	expiredLineItem models.SubscriptionLineItem,
	changeEventID uuid.UUID,
) error {
	// Verify we have an installment plan to expire
	if expiredLineItem.AutoRenewingPlan == nil || expiredLineItem.AutoRenewingPlan.InstallmentPlanID == nil {
		return nil // Nothing to expire
	}

	// Get the full installment plan record
	var plan models.InstallmentPlan
	if err := tx.First(&plan, "id = ?", *expiredLineItem.AutoRenewingPlan.InstallmentPlanID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Already expired or doesn't exist
		}
		return fmt.Errorf("failed to find installment plan: %w", err)
	}

	// Create history before soft deleting
	history := models.InstallmentPlanHistory{
		InstallmentPlanID:               plan.ID,
		SubscriptionID:                  plan.SubscriptionID,
		LineItemID:                      plan.LineItemID,
		AutoRenewingPlanID:              plan.AutoRenewingPlanID,
		PreviousInitialPaymentsCount:    &plan.InitialCommittedPaymentsCount,
		PreviousSubsequentPaymentsCount: plan.SubsequentCommittedPaymentsCount,
		PreviousRemainingPaymentsCount:  &plan.RemainingCommittedPaymentsCount,
		PreviousPendingCancellation:     &plan.PendingCancellation,
		ChangeType:                      "EXPIRED",
		ChangeEventID:                   changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return fmt.Errorf("failed to create installment plan history: %w", err)
	}

	// Soft delete the installment plan
	if err := tx.Delete(&plan).Error; err != nil {
		return fmt.Errorf("failed to soft delete installment plan: %w", err)
	}

	return nil
}
