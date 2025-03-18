package handler

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	"subsnotifpro-go/internal/google_playstore/client/service"
	"subsnotifpro-go/internal/google_playstore/client/utils"
	"subsnotifpro-go/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SubscriptionClientHandler manages subscription purchase API endpoints.
type PlaystoreClientHandler struct {
	service service.PlaystoreClientService // Use interface here
}

// NewSubscriptionClientHandler creates a new instance of the handler.
func NewPlaystoreClientHandler(svc service.PlaystoreClientService) *PlaystoreClientHandler {
	return &PlaystoreClientHandler{service: svc}
}

// RegisterRoutes registers all subscription client-related routes.
func RegisterClientRoutes(router *gin.RouterGroup, handler *PlaystoreClientHandler) {
	router.GET("/fetch-user-subscription-purchase", handler.GetUserSubscriptionPurchase)
	router.GET("/fetch-list-subscription-products", handler.ListSubscriptionProducts)
	router.GET("/fetch-subscription-product-details", handler.GetSubscriptionProductDetails)
	router.GET("/fetch-subscription-offers", handler.GetSubscriptionOffers)
}

// Regex for validating purchase tokens (alphanumeric, dashes, underscores)
var validTokenRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]{5,}$`)

// GetSubscriptionDetails fetches subscription purchase details for a given token.
func (h *PlaystoreClientHandler) GetUserSubscriptionPurchase(c *gin.Context) {
	purchaseToken := c.Query("purchase_token")
	packageName := c.Query("package_name")

	if purchaseToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "purchase_token is required"})
		return
	}

	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "package_name is required"})
		return
	}

	if !validTokenRegex.MatchString(purchaseToken) {
		logger.Log.WithField("token", purchaseToken).Debug("🔍 Invalid purchase token format")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid purchase_token format"})
		return
	}

	subscriptionPurchase, err := h.service.GetUserSubscriptionPurchase(purchaseToken, packageName)

	if errors.Is(err, context.DeadlineExceeded) {
		logger.Log.WithFields(logrus.Fields{
			"token": purchaseToken,
			"url":   c.Request.URL.Path,
		}).Warn("⚠️ Subscription request timed out")
		c.JSON(http.StatusGatewayTimeout, gin.H{"success": false, "error": "Request timed out"})
		return
	}

	if errors.Is(err, context.Canceled) {
		logger.Log.WithField("token", purchaseToken).Info("🚨 Client canceled request")
		c.JSON(499, gin.H{"success": false, "error": "Request canceled by client"})
		return
	}

	if err != nil {
		status := http.StatusInternalServerError
		errMsg := "Failed to fetch subscription details"

		if utils.IsInvalidTokenError(err) {
			status = http.StatusBadRequest
			errMsg = "Invalid purchase token"
			logger.Log.WithField("token", purchaseToken).Warn("⚠️ Invalid purchase token")
		} else {
			logger.Log.WithFields(logrus.Fields{
				"token": purchaseToken,
				"error": err.Error(),
			}).Error("❌ Internal error fetching subscription details")

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "An internal server error occurred. Please try again later.",
			})
			return
		}

		c.JSON(status, gin.H{"success": false, "error": errMsg})
		return
	}

	logger.Log.WithField("token", purchaseToken).Info("✅ Subscription details fetched successfully")

	// **✅ FIX: Ensure subscriptionDetails is properly checked**
	if subscriptionPurchase == nil {
		logger.Log.WithField("token", purchaseToken).Warn("⚠️ No subscription details found for token")
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "No subscription details found for this purchase token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription details fetched successfully",
		"data":    subscriptionPurchase,
	})
}

// ListSubscriptionProducts fetches all available subscription products (returns raw response).
func (h *PlaystoreClientHandler) ListSubscriptionProducts(c *gin.Context) {
	packageName := c.Query("package_name")

	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "package_name is required"})
		return
	}

	// Fetch raw subscription data
	subscriptions, err := h.service.ListSubscriptionProducts(packageName)
	if err != nil {
		logger.Log.Errorf("❌ Failed to fetch subscription products: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list subscription products",
		})
		return
	}

	if subscriptions == nil || len(subscriptions.Subscriptions) == 0 {
		logger.Log.Warnf("⚠️ No subscriptions found for package: %s", packageName)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "No subscription products found",
		})
		return
	}

	// ✅ Return the **raw API response** from Google Play without modification
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription products listed successfully",
		"data":    subscriptions, // 🔥 Raw response from Google Play API
	})
}

// GetSubscriptionProductDetails fetches details of a specific subscription product.
func (h *PlaystoreClientHandler) GetSubscriptionProductDetails(c *gin.Context) {
	productID := c.Query("product_id")
	packageName := c.Query("package_name")

	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "product_id is required"})
		return
	}

	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "package_name is required"})
		return
	}

	subscriptionDetails, err := h.service.GetSubscriptionProductDetails(productID, packageName)
	if err != nil {
		logger.Log.Errorf("❌ Failed to fetch subscription product details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch subscription product details",
		})
		return
	}

	if subscriptionDetails == nil {
		logger.Log.Warnf("⚠️ No details found for subscription product ID: %s", productID)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "No details found for the given subscription product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription product details fetched successfully",
		"data":    subscriptionDetails,
	})
}

// GetSubscriptionOffers fetches offers for a specific base plan in a subscription product.
func (h *PlaystoreClientHandler) GetSubscriptionOffers(c *gin.Context) {
	productID := c.Query("product_id")
	packageName := c.Query("package_name")
	basePlanID := c.Query("base_plan_id")

	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "product_id is required"})
		return
	}

	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "package_name is required"})
		return
	}

	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "base_plan_id is required"})
		return
	}

	offers, err := h.service.GetSubscriptionOffers(packageName, productID, basePlanID)
	if err != nil {
		logger.Log.Errorf("❌ Failed to fetch subscription offers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch subscription offers",
		})
		return
	}

	if len(offers) == 0 {
		logger.Log.Warnf("⚠️ No offers found for subscription product ID: %s and base plan ID: %s", productID, basePlanID)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "No offers found for the given subscription product and base plan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription offers fetched successfully",
		"data":    offers,
	})
}
