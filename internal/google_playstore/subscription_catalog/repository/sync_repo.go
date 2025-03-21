package repository

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ✅ Upsert Subscription Product (Insert or Update inside Transaction)
func (r *subscriptionCatalogRepository) UpsertSubscriptionProducts(ctx context.Context, subscription *models.ProductSubscription) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "package_name"}, {Name: "product_id"}}, // ✅ Unique constraints
			DoUpdates: clause.AssignmentColumns([]string{"package_name", "product_id"}),
		}).Create(subscription).Error

		if err != nil {
			return fmt.Errorf("failed to upsert subscription product: %w", err)
		}

		return nil
	})
}

func (r *subscriptionCatalogRepository) DeleteAllSubscriptionProducts(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("1 = 1").Delete(&models.ProductSubscription{}).Error; err != nil {
			return fmt.Errorf("failed to delete all subscription products: %w", err)
		}
		return nil
	})
}

// ✅ Upsert Base Plan (Insert or Update inside Transaction)
func (r *subscriptionCatalogRepository) UpsertSubscriptionBasePlans(ctx context.Context, basePlan *models.ProductBasePlan) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "package_name"}, {Name: "product_id"}, {Name: "base_plan_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"package_name", "product_id", "base_plan_id"}),
		}).Create(basePlan).Error

		if err != nil {
			return fmt.Errorf("failed to upsert base plan: %w", err)
		}

		return nil
	})
}

// ✅ Upsert Subscription Offers (Bulk Insert)
func (r *subscriptionCatalogRepository) UpsertSubscriptionOffers(ctx context.Context, offers []models.SubscriptionOffer) error {
	if len(offers) == 0 {
		return nil // No offers to insert
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		batchSize := r.batchSize // Use configured batch size
		for i := 0; i < len(offers); i += batchSize {
			end := i + batchSize
			if end > len(offers) {
				end = len(offers)
			}

			err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "package_name"}, {Name: "product_id"}, {Name: "base_plan_id"}, {Name: "offer_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"state"}), // ✅ Only update the state
			}).CreateInBatches(offers[i:end], batchSize).Error

			if err != nil {
				return fmt.Errorf("failed to upsert offers in batch: %w", err)
			}
		}
		return nil
	})
}

// ✅ Delete All Subscription Offers (Safe and Efficient)
func (r *subscriptionCatalogRepository) DeleteAllSubscriptionOffers(ctx context.Context, packageName string, productID string, basePlanID string) error {
	// ✅ Unscoped delete ensures soft deletes are considered (if enabled)
	return r.db.WithContext(ctx).Where(
		"package_name = ? AND product_id = ? AND base_plan_id = ?", packageName, productID, basePlanID).
		Unscoped().Delete(&models.SubscriptionOffer{}).Error
}
