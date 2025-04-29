// internal/api/dto/dashboard.go
package dto

import (
	"time"
)

type GetDashboardRequest struct {
	PlatformUserIDs []string  `form:"platform_user_ids"` // Comma-separated in query
	Statuses        []string  `form:"statuses"`          // Comma-separated
	Platforms       []string  `form:"platforms"`         // Comma-separated
	PlanTypes       []string  `form:"plan_types"`        // Comma-separated
	DateFrom        time.Time `form:"date_from" time_format:"2006-01-02"`
	DateTo          time.Time `form:"date_to" time_format:"2006-01-02"`
	Page            int       `form:"page" default:"1"`
	PageSize        int       `form:"page_size" default:"20"`
}

type DashboardRecord struct {
	SubscriptionID string     `json:"subscription_id"`
	PlatformUserID string     `json:"platform_user_id"`
	ProductID      string     `json:"product_id"`
	BasePlanID     *string    `json:"base_plan_id,omitempty"`
	ActiveOfferID  *string    `json:"active_offer_id,omitempty"`
	Platform       string     `json:"platform"`
	Status         string     `json:"status"`
	PlanType       string     `json:"plan_type"`
	StartDate      time.Time  `json:"start_date"`
	RenewalDate    time.Time  `json:"renewal_date"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
	LatestOrderID  string     `json:"latest_order_id"`
	PurchaseToken  string     `json:"purchase_token"`
	TotalAmount    *float64   `json:"total_amount"`
	Currency       *string    `json:"currency"`
	LastModified   time.Time  `json:"last_modified"`
}

type GetDashboardResponse struct {
	Data       []DashboardRecord `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
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

func NewDashboardFilterFromRequest(req GetDashboardRequest) DashboardFilter {
	return DashboardFilter(req)
}
