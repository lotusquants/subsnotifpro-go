// internal/subscription/repository/dashboard_repository.go
package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	RefreshView(ctx context.Context) error
	GetDashboardData(ctx context.Context, filter DashboardFilter) ([]DashboardRecord, int64, error)
}

type DashboardRecord struct {
	SubscriptionID string     `gorm:"column:id"`
	PlatformUserID string     `gorm:"column:platform_user_id"`
	ProductID      string     `gorm:"column:product_id"`
	BasePlanID     *string    `gorm:"column:base_plan_id"`
	ActiveOfferID  *string    `gorm:"column:offer_id"`
	Platform       string     `gorm:"column:platform"`
	Status         string     `gorm:"column:status"`
	PlanType       string     `gorm:"column:plan_type"`
	StartDate      *time.Time `gorm:"column:start_date"`
	RenewalDate    *time.Time `gorm:"column:renewal_date"`
	ExpirationDate *time.Time `gorm:"column:expiration_date"`
	LatestOrderID  string     `gorm:"column:latest_order_id"`
	PurchaseToken  string     `gorm:"column:purchase_token"`
	TotalAmount    *float64   `gorm:"column:total_amount"`
	Currency       *string    `gorm:"column:currency"`
	LastModified   time.Time  `gorm:"column:last_modified"`
}

type DashboardFilter struct {
	PlatformUserIDs []string
	Statuses        []string
	Platforms       []string
	PlanTypes       []string
	DateFrom        time.Time
	DateTo          time.Time
	Page            int
	PageSize        int
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) RefreshView(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec(`
		REFRESH MATERIALIZED VIEW CONCURRENTLY subsnotifpro_subscription_dashboard_view
	`).Error
}

func (r *dashboardRepository) GetDashboardData(
	ctx context.Context,
	filter DashboardFilter,
) ([]DashboardRecord, int64, error) {
	var records []DashboardRecord
	var total int64

	query := r.db.WithContext(ctx).Table("subsnotifpro_subscription_dashboard_view")

	if len(filter.PlatformUserIDs) > 0 {
		query = query.Where("platform_user_id IN ?", filter.PlatformUserIDs)
	}

	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}

	if len(filter.Platforms) > 0 {
		query = query.Where("platform IN ?", filter.Platforms)
	}

	if len(filter.PlanTypes) > 0 {
		query = query.Where("plan_type IN ?", filter.PlanTypes)
	}

	if !filter.DateFrom.IsZero() {
		query = query.Where("last_modified >= ?", filter.DateFrom)
	}

	if !filter.DateTo.IsZero() {
		query = query.Where("last_modified <= ?", filter.DateTo)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	// Apply pagination
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Execute query
	err := query.
		Order("last_modified DESC").
		Find(&records).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get dashboard data: %w", err)
	}

	return records, total, nil
}
