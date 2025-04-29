// internal/api/handler/dashboard_handler.go
package handler

import (
	"net/http"
	"strings"

	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/subscription/dto"
	"subsnotifpro-go/internal/subscription/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

// GetDashboard godoc
// @Summary Get subscription dashboard data
// @Description Returns paginated subscription dashboard data
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param platform_user_ids query string false "Comma-separated platform user IDs"
// @Param statuses query string false "Comma-separated statuses (ACTIVE,EXPIRED,etc)"
// @Param platforms query string false "Comma-separated platforms (PLAY_STORE,APP_STORE)"
// @Param plan_types query string false "Comma-separated plan types"
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} dto.GetDashboardResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/dashboard [get]
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	var req dto.GetDashboardRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request parameters"})
		return
	}

	// Parse comma-separated lists
	if ids := c.Query("platform_user_ids"); ids != "" {
		req.PlatformUserIDs = strings.Split(ids, ",")
	}
	if statuses := c.Query("statuses"); statuses != "" {
		req.Statuses = strings.Split(statuses, ",")
	}
	if platforms := c.Query("platforms"); platforms != "" {
		req.Platforms = strings.Split(platforms, ",")
	}
	if planTypes := c.Query("plan_types"); planTypes != "" {
		req.PlanTypes = strings.Split(planTypes, ",")
	}

	// Convert to service filter
	filter := dto.NewDashboardFilterFromRequest(req)

	// Get data from service
	data, total, err := h.service.GetDashboardData(c.Request.Context(), filter)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to get dashboard data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch dashboard data"})
		return
	}

	// Convert to API response
	response := dto.GetDashboardResponse{
		Data:       make([]dto.DashboardRecord, 0, len(data)),
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: calculateTotalPages(total, req.PageSize),
	}

	for _, record := range data {
		response.Data = append(response.Data, dto.DashboardRecord{
			SubscriptionID: record.SubscriptionID,
			PlatformUserID: record.PlatformUserID,
			ProductID:      record.ProductID,
			BasePlanID:     record.BasePlanID,
			ActiveOfferID:  record.ActiveOfferID,
			Platform:       record.Platform,
			Status:         record.Status,
			PlanType:       record.PlanType,
			StartDate:      record.StartDate,
			RenewalDate:    record.RenewalDate,
			ExpirationDate: record.ExpirationDate,
			LatestOrderID:  record.LatestOrderID,
			PurchaseToken:  record.PurchaseToken,
			TotalAmount:    record.TotalAmount,
			Currency:       record.Currency,
			LastModified:   record.LastModified,
		})
	}

	c.JSON(http.StatusOK, response)
}

func calculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	return totalPages
}
