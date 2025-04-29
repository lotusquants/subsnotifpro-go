// // internal/migrations/views.go
// package migrations

// import "gorm.io/gorm"

// func CreateMaterializedViews(db *gorm.DB) error {
// 	views := []string{
// 		`CREATE MATERIALIZED VIEW IF NOT EXISTS subsnotifpro_subscription_dashboard_view AS
//         WITH ranked_subscriptions AS (
//             SELECT
//                 s.platform_user_id,
//                 s.product_id,
//                 CASE
//                     WHEN s.active_platform = 'GOOGLE_PLAYSTORE' THEN 'PLAY_STORE'
//                     WHEN s.active_platform = 'APPLE_APPSTORE' THEN 'APP_STORE'
//                     ELSE s.active_platform
//                 END AS platform,
//                 s.status,
//                 s.plan_type,
//                 s.start_date,
//                 s.next_renewal_date AS renewal_date,
//                 s.updated_at AS last_modified,
//                 ROW_NUMBER() OVER (
//                     PARTITION BY s.platform_user_id, s.product_id,
//                     CASE
//                         WHEN s.active_platform = 'GOOGLE_PLAYSTORE' THEN 'PLAY_STORE'
//                         WHEN s.active_platform = 'APPLE_APPSTORE' THEN 'APP_STORE'
//                         ELSE s.active_platform
//                     END
//                     ORDER BY s.updated_at DESC
//                 ) as row_num
//             FROM
//                 subsnotifpro_unified_subscription s
//         )
//         SELECT
//             platform_user_id,
//             product_id,
//             platform,
//             status,
//             plan_type,
//             start_date,
//             renewal_date,
//             last_modified
//         FROM
//             ranked_subscriptions
//         WHERE
//             row_num = 1`,
// 	}

// 	for _, view := range views {
// 		if err := db.Exec(view).Error; err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// // CreateViewIndexes adds indexes to materialized views
// func CreateViewIndexes(db *gorm.DB) error {
// 	indexes := []string{
// 		`CREATE UNIQUE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_pk
// 		 ON subsnotifpro_subscription_dashboard_view (platform_user_id, product_id, platform)`,
// 		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_user
// 		 ON subsnotifpro_subscription_dashboard_view (platform_user_id)`,
// 		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_status
// 		 ON subsnotifpro_subscription_dashboard_view (status)`,
// 		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_platform
// 		 ON subsnotifpro_subscription_dashboard_view (platform)`,
// 		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_plan_type
// 		 ON subsnotifpro_subscription_dashboard_view (plan_type)`,
// 		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_renewal
// 		 ON subsnotifpro_subscription_dashboard_view (renewal_date)`,
// 		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_modified
// 		 ON subsnotifpro_subscription_dashboard_view (last_modified)`,
// 	}

// 	for _, index := range indexes {
// 		if err := db.Exec(index).Error; err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// internal/migrations/views.go
package migrations

import "gorm.io/gorm"

func CreateMaterializedViews(db *gorm.DB) error {
	views := []string{
		`DROP MATERIALIZED VIEW IF EXISTS subsnotifpro_subscription_dashboard_view`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS subsnotifpro_subscription_dashboard_view AS
        WITH ranked_subscriptions AS (
            SELECT 
                s.id,
                s.platform_user_id,
                s.product_id,
                s.base_plan_id,
                s.active_offer_id,
                CASE 
                    WHEN s.active_platform = 'GOOGLE_PLAYSTORE' THEN 'PLAY_STORE'
                    WHEN s.active_platform = 'APPLE_APPSTORE' THEN 'APP_STORE'
                    ELSE s.active_platform
                END AS platform,
                s.status,
                s.plan_type,
                s.start_date,
                s.next_renewal_date AS renewal_date,
                s.expiration_date,
                s.latest_order_id,
                s.purchase_token,
                s.updated_at AS last_modified,
                s.currency,
                s.total_amount,
                ROW_NUMBER() OVER (
                    PARTITION BY s.platform_user_id, s.product_id, 
                    CASE 
                        WHEN s.active_platform = 'GOOGLE_PLAYSTORE' THEN 'PLAY_STORE'
                        WHEN s.active_platform = 'APPLE_APPSTORE' THEN 'APP_STORE'
                        ELSE s.active_platform
                    END
                    ORDER BY s.updated_at DESC
                ) as row_num
            FROM 
                subsnotifpro_unified_subscription s
        )
        SELECT 
            id,
            platform_user_id,
            product_id,
            base_plan_id,
            active_offer_id,
            platform,
            status,
            plan_type,
            start_date,
            renewal_date,
            expiration_date,
            latest_order_id,
            purchase_token,
            last_modified,
            currency,
            total_amount
        FROM 
            ranked_subscriptions
        WHERE 
            row_num = 1`,
	}

	for _, view := range views {
		if err := db.Exec(view).Error; err != nil {
			return err
		}
	}
	return nil
}

// CreateViewIndexes adds indexes to materialized views
func CreateViewIndexes(db *gorm.DB) error {
	indexes := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_pk 
		 ON subsnotifpro_subscription_dashboard_view (id)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_user 
		 ON subsnotifpro_subscription_dashboard_view (platform_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_product 
		 ON subsnotifpro_subscription_dashboard_view (product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_status 
		 ON subsnotifpro_subscription_dashboard_view (status)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_platform 
		 ON subsnotifpro_subscription_dashboard_view (platform)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_plan_type 
		 ON subsnotifpro_subscription_dashboard_view (plan_type)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_renewal 
		 ON subsnotifpro_subscription_dashboard_view (renewal_date)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_expiration 
		 ON subsnotifpro_subscription_dashboard_view (expiration_date)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_modified 
		 ON subsnotifpro_subscription_dashboard_view (last_modified)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_order_id 
		 ON subsnotifpro_subscription_dashboard_view (latest_order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_purchase_token 
		 ON subsnotifpro_subscription_dashboard_view (purchase_token)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_base_plan 
		 ON subsnotifpro_subscription_dashboard_view (base_plan_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subsnotifpro_subscription_dashboard_view_offer 
		 ON subsnotifpro_subscription_dashboard_view (active_offer_id)`,
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			return err
		}
	}

	return nil
}
