package migrations

import (
	"gorm.io/gorm"
)

// ApplyCompositeIndexes manually adds composite indexes for performance tuning.
func ApplyCompositeIndexes(db *gorm.DB) error {
	compositeIndexes := []string{
		// SubscriptionPurchaseV2 composite indexes
		`CREATE INDEX IF NOT EXISTS idx_subscription_user_plan_type ON subsnotifpro_subscription_purchase_v2 (user_id, plan_type)`,
		`CREATE INDEX IF NOT EXISTS idx_subscription_plan_start_time ON subsnotifpro_subscription_purchase_v2 (plan_type, start_time)`,

		// AutoRenewingPlanHistory composite indexes
		`CREATE INDEX IF NOT EXISTS idx_auto_renewing_history_subscription_change ON subsnotifpro_auto_renewing_plan_history (subscription_id, change_type)`,
		`CREATE INDEX IF NOT EXISTS idx_auto_renewing_history_change_time ON subsnotifpro_auto_renewing_plan_history (change_time)`,

		// PrepaidPlanHistory composite indexes
		`CREATE INDEX IF NOT EXISTS idx_prepaid_history_subscription_change ON subsnotifpro_prepaid_plan_history (subscription_id, change_type)`,
		`CREATE INDEX IF NOT EXISTS idx_prepaid_history_change_time ON subsnotifpro_prepaid_plan_history (change_time)`,

		// DeferredItemReplacementHistory composite indexes
		`CREATE INDEX IF NOT EXISTS idx_deferred_item_replacement_subscription ON subsnotifpro_deferred_item_replacement_history (subscription_id, plan_type)`,

		// SignupPromotionHistory composite indexes
		`CREATE INDEX IF NOT EXISTS idx_signup_promotion_subscription ON subsnotifpro_signup_promotion_history (subscription_id, plan_type)`,
	}

	for _, query := range compositeIndexes {
		if err := db.Exec(query).Error; err != nil {
			return err
		}
	}
	return nil
}
