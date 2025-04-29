// internal/subscription/service/dashboard_service_impl.go
package service

import (
	"context"
	"sync"
	"time"

	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/subscription/dto"
	"subsnotifpro-go/internal/subscription/repository"
)

type DashboardService interface {
	RefreshDashboard(ctx context.Context) error
	SchedulePeriodicRefresh(ctx context.Context, interval time.Duration)
	GetDashboardData(ctx context.Context, filter dto.DashboardFilter) ([]dto.DashboardRecord, int64, error)
}

type dashboardService struct {
	repo            repository.DashboardRepository
	mu              sync.Mutex
	lastRefresh     time.Time
	refreshInterval time.Duration
}

func NewDashboardService(
	repo repository.DashboardRepository,
	refreshInterval time.Duration,
) DashboardService {
	return &dashboardService{
		repo:            repo,
		refreshInterval: refreshInterval,
	}
}

func (s *dashboardService) RefreshDashboard(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// // Skip if refreshed recently
	// if time.Since(s.lastRefresh) < s.refreshInterval {
	// 	return nil
	// }

	start := time.Now()
	err := s.repo.RefreshView(ctx)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to refresh dashboard view")
		return err
	}

	s.lastRefresh = time.Now()
	logger.Log.WithField("duration", time.Since(start)).Info("Dashboard view refreshed")
	return nil
}

func (s *dashboardService) SchedulePeriodicRefresh(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.RefreshDashboard(ctx); err != nil {
				logger.Log.WithError(err).Error("Dashboard refresh failed")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *dashboardService) GetDashboardData(
	ctx context.Context,
	filter dto.DashboardFilter,
) ([]dto.DashboardRecord, int64, error) {
	// Convert service filter to repository filter
	repoFilter := repository.DashboardFilter{
		PlatformUserIDs: filter.PlatformUserIDs,
		Statuses:        filter.Statuses,
		Platforms:       filter.Platforms,
		PlanTypes:       filter.PlanTypes,
		DateFrom:        filter.DateFrom,
		DateTo:          filter.DateTo,
		Page:            filter.Page,
		PageSize:        filter.PageSize,
	}

	// Optionally refresh before querying
	if err := s.RefreshDashboard(ctx); err != nil {
		logger.Log.WithError(err).Warn("Couldn't refresh dashboard before query")
	}

	records, total, err := s.repo.GetDashboardData(ctx, repoFilter)

	// Helper function to handle *time.Time to time.Time conversion
	derefTime := func(t *time.Time) time.Time {
		if t != nil {
			return *t
		}
		return time.Time{}
	}
	if err != nil {
		return nil, 0, err
	}

	// Convert repository models to service DTOs with all fields
	var result []dto.DashboardRecord
	for _, r := range records {
		result = append(result, dto.DashboardRecord{
			SubscriptionID: r.SubscriptionID,
			PlatformUserID: r.PlatformUserID,
			ProductID:      r.ProductID,
			BasePlanID:     r.BasePlanID,
			ActiveOfferID:  r.ActiveOfferID,
			Platform:       r.Platform,
			Status:         r.Status,
			PlanType:       r.PlanType,
			StartDate:      derefTime(r.StartDate),
			RenewalDate:    derefTime(r.RenewalDate),
			ExpirationDate: r.ExpirationDate,
			LatestOrderID:  r.LatestOrderID,
			PurchaseToken:  r.PurchaseToken,
			TotalAmount:    r.TotalAmount,
			Currency:       r.Currency,
			LastModified:   r.LastModified,
		})
	}

	return result, total, nil
}
