package handler

import (
	"errors"
	"net/http"
	"strconv"

	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/subscription/service"
	unifiedSubscriptionService "subsnotifpro-go/internal/subscription/service"
	"subsnotifpro-go/internal/utils"

	"github.com/gin-gonic/gin"
)

type UnifiedSubscriptionsHandler struct {
	service unifiedSubscriptionService.UnifiedSubscriptionService
}

func NewHandler(service unifiedSubscriptionService.UnifiedSubscriptionService) *UnifiedSubscriptionsHandler {
	return &UnifiedSubscriptionsHandler{
		service: service,
	}
}

// GetUserSubscriptions godoc
// @Summary Get all subscriptions for a user
// @Description Retrieves all subscriptions associated with a user ID
// @Tags subscriptions
// @Accept  json
// @Produce  json
// @Param user_id path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} response.PaginatedResponse{data=[]models.UnifiedSubscription}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/users/{user_id}/subscriptions [get]
func (h *UnifiedSubscriptionsHandler) GetUserSubscriptions(c *gin.Context) {
	// Get user_id from query parameters
	platformUserId := c.Query("platform_user_id") // This is the critical change
	if platformUserId == "" {
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "platform_user_id is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	subscriptions, total, err := h.service.GetSubscriptionsByUserID(c.Request.Context(), platformUserId, page, pageSize)
	if err != nil {
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteGinPaginatedResponse(c, http.StatusOK, subscriptions, total, page, pageSize)
}

// GetSubscriptionEvents godoc
// @Summary Get events for a subscription
// @Description Retrieves all events associated with a subscription ID from the appropriate platform
// @Tags subscriptions
// @Accept  json
// @Produce  json
// @Param subscription_id path string true "Subscription ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} response.PaginatedResponse{data=[]models.SubscriptionEvent}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/subscriptions/{subscription_id}/events [get]
func (h *UnifiedSubscriptionsHandler) GetSubscriptionEvents(c *gin.Context) {
	// Start request logging
	logger.Log.WithFields(map[string]interface{}{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
	}).Info("Handling GetSubscriptionEvents request")

	subscriptionID := c.Param("subscription_id")
	if subscriptionID == "" {
		logger.Log.Warn("Missing subscription_id parameter")
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "subscription_id is required")
		return
	}

	// Parse pagination parameters with validation
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		logger.Log.WithField("page", c.Query("page")).Warn("Invalid page parameter")
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "page must be a positive integer")
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		logger.Log.WithField("page_size", c.Query("page_size")).Warn("Invalid page_size parameter")
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "page_size must be between 1 and 100")
		return
	}

	// Get events from service
	events, total, err := h.service.GetSubscriptionEvents(c.Request.Context(), subscriptionID, page, pageSize)
	if err != nil {
		logger.Log.WithError(err).WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
		}).Error("Failed to get subscription events")

		if errors.Is(err, service.ErrSubscriptionNotFound) {
			utils.WriteGinErrorResponse(c, http.StatusNotFound, "subscription not found")
		} else if errors.Is(err, service.ErrUnsupportedPlatform) {
			utils.WriteGinErrorResponse(c, http.StatusBadRequest, "unsupported platform")
		} else {
			utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "failed to get subscription events")
		}
		return
	}

	logger.Log.WithFields(map[string]interface{}{
		"subscription_id": subscriptionID,
		"event_count":     len(events),
		"total_events":    total,
	}).Debug("Successfully retrieved subscription events")

	utils.WriteGinPaginatedResponse(c, http.StatusOK, events, int64(total), page, pageSize)
}
