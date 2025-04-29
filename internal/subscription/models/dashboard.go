package models

import (
	"time"
)

type SubscriptionDashboard struct {
	PlatformUserID string     `gorm:"column:platform_user_id;primaryKey;index"`
	ProductID      string     `gorm:"column:product_id;primaryKey;index"`
	SubscriptionID string     `gorm:"column:subscription_id;index"`
	BasePlanID     *string    `gorm:"column:base_plan_id;index"` // For filtering
	ActiveOfferID  *string    `gorm:"column:offer_id;index"`     // For filtering
	Platform       string     `gorm:"column:platform;index"`     // "PLAY_STORE" or "APP_STORE"
	Status         string     `gorm:"column:status;index"`       // "ACTIVE", "EXPIRED", etc.
	PlanType       string     `gorm:"column:plan_type;index"`    // "Monthly", "Annual", etc.
	StartDate      *time.Time `gorm:"column:start_date;index"`
	RenewalDate    *time.Time `gorm:"column:renewal_date;index"`
	ExpirationDate *time.Time `gorm:"column:expiration_date;index"`
	LatestOrderID  string     `gorm:"column:latest_order_id;index"` // New unified field
	PurchaseToken  string     `gorm:"column:purchase_token;index"`  // For backward compatibility
	LastModified   time.Time  `gorm:"column:last_modified;index"`   // For sorting
	TotalAmount    *float64   `gorm:"column:total_amount"`          // New unified field
	Currency       *string    `gorm:"column:currency"`              // New unified field

}

func (SubscriptionDashboard) TableName() string {
	return "subsnotifpro_subscription_dashboard_view"
}
