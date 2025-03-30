package handler

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/playstore/api/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// PlaystoreApiHandler manages Play Store client API endpoints.
type PlaystoreApiHandler struct {
	service service.PlaystoreApiService // Use interface here
}

// NewPlaystoreClientHandler creates a new instance of the handler.
func NewPlaystoreClientHandler(svc service.PlaystoreApiService) *PlaystoreApiHandler {
	return &PlaystoreApiHandler{service: svc}
}

// Regex for validating purchase tokens (alphanumeric, dashes, underscores)
var validTokenRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]{5,}$`)

// -------------------------
// 🚀 GetUserSubscriptionPurchase
// -------------------------

func (h *PlaystoreApiHandler) GetUserSubscriptionPurchase(c *gin.Context) {
	ctx := c.Request.Context()
	purchaseToken := c.Query("purchase_token")
	packageName := c.Query("package_name")

	// Validate inputs
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

	// Fetch subscription details
	subscriptionPurchase, err := h.service.GetUserSubscriptionPurchase(ctx, purchaseToken, packageName)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"token": purchaseToken,
			"error": err.Error(),
		}).Error("❌ Failed to fetch subscription details")

		if errors.Is(err, context.DeadlineExceeded) {
			c.JSON(http.StatusGatewayTimeout, gin.H{"success": false, "error": "Request timed out"})
			return
		}
		if errors.Is(err, context.Canceled) {
			c.JSON(499, gin.H{"success": false, "error": "Request canceled by client"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch subscription details"})
		return
	}

	if subscriptionPurchase == nil {
		logger.Log.WithField("token", purchaseToken).Warn("⚠️ No subscription details found for token")
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "No subscription details found for this purchase token"})
		return
	}

	logger.Log.WithField("token", purchaseToken).Info("✅ Subscription details fetched successfully")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription details fetched successfully",
		"data":    subscriptionPurchase,
	})
}

// -------------------------
// 🚀 ListSubscriptionProducts
// -------------------------
func (h *PlaystoreApiHandler) ListSubscriptionProducts(c *gin.Context) {
	ctx := c.Request.Context()
	packageName := c.Query("package_name")

	// Validate inputs
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "package_name is required"})
		return
	}

	// Fetch subscription products
	subscriptions, err := h.service.ListSubscriptionProducts(ctx, packageName)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"package": packageName,
			"error":   err.Error(),
		}).Error("❌ Failed to fetch subscription products")

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to list subscription products"})
		return
	}

	if subscriptions == nil || len(subscriptions.Subscriptions) == 0 {
		logger.Log.WithField("package", packageName).Warn("⚠️ No subscriptions found for package")
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "No subscription products found"})
		return
	}

	logger.Log.WithField("package", packageName).Info("✅ Subscription products listed successfully")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription products listed successfully",
		"data":    subscriptions,
	})
}

// -------------------------
// 🚀 GetSubscriptionProductDetails
// -------------------------
func (h *PlaystoreApiHandler) GetSubscriptionProductDetails(c *gin.Context) {
	ctx := c.Request.Context()
	productID := c.Query("product_id")
	packageName := c.Query("package_name")

	// Validate inputs
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "product_id is required"})
		return
	}
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "package_name is required"})
		return
	}

	// Fetch subscription product details
	subscriptionDetails, err := h.service.GetSubscriptionProductDetails(ctx, productID, packageName)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"product": productID,
			"error":   err.Error(),
		}).Error("❌ Failed to fetch subscription product details")

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch subscription product details"})
		return
	}

	if subscriptionDetails == nil {
		logger.Log.WithField("product", productID).Warn("⚠️ No details found for subscription product")
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "No details found for the given subscription product"})
		return
	}

	logger.Log.WithField("product", productID).Info("✅ Subscription product details fetched successfully")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription product details fetched successfully",
		"data":    subscriptionDetails,
	})
}

// -------------------------
// 🚀 GetSubscriptionOffers
// -------------------------
func (h *PlaystoreApiHandler) GetSubscriptionOffers(c *gin.Context) {
	ctx := c.Request.Context()
	productID := c.Query("product_id")
	packageName := c.Query("package_name")
	basePlanID := c.Query("base_plan_id")

	// Validate inputs
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

	// Fetch subscription offers
	offers, err := h.service.GetSubscriptionOffers(ctx, packageName, productID, basePlanID)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"product":  productID,
			"basePlan": basePlanID,
			"error":    err.Error(),
		}).Error("❌ Failed to fetch subscription offers")

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch subscription offers"})
		return
	}

	if len(offers) == 0 {
		logger.Log.WithFields(logrus.Fields{
			"product":  productID,
			"basePlan": basePlanID,
		}).Warn("⚠️ No offers found for subscription product and base plan")

		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "No offers found for the given subscription product and base plan"})
		return
	}

	logger.Log.WithFields(logrus.Fields{
		"product":  productID,
		"basePlan": basePlanID,
	}).Info("✅ Subscription offers fetched successfully")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription offers fetched successfully",
		"data":    offers,
	})
}
