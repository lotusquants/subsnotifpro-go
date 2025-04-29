package repository

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/pkg/contextutil"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/subscription/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SubscriptionRepository interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
	UpsertSubscription(ctx context.Context, sub *models.UnifiedSubscription) error
	// Add other repository methods here...

}

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) WithTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	tx := r.db.Begin().WithContext(ctx)
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	txCtx := contextutil.WithTx(ctx, tx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Log.Errorf("panic recovered in transaction: %v", r)
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			logger.Log.Errorf("rollback failed: %v", rollbackErr)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

func (r *subscriptionRepository) UpsertSubscription(ctx context.Context, sub *models.UnifiedSubscription) error {
	tx, ok := contextutil.TxFromContext(ctx)
	if !ok {
		return fmt.Errorf("transaction not found in context")
	}

	// Full record upsert
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "subscription_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"user_id",
			"active_platform",
			"platform_user_id",
			"latest_order_id",
			"purchase_token",
			"plan_type",
			"status",
			"start_date",
			"next_renewal_date",
			"expiration_date",
			"grace_period_start_date",
			"grace_period_end_date",
			"product_id",
			"base_plan_id",
			"add_on_id",
			"active_offer_id",
			"currency",
			"total_amount",
			"updated_at",
		}),
	}).Create(sub).Error
}
