package handler

import (
	"context"
	"errors"
	"fmt"

	"subsnotifpro-go/internal/pkg/apperrors"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/pkg/validation"
	"subsnotifpro-go/internal/playstore/api/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// EnhancedPlaystoreApiHandler manages Play Store client API endpoints with enhanced error handling,
// validation, logging, and resilience patterns.
type EnhancedPlaystoreApiHandler struct {
	service service.PlaystoreApiService // Use interface here
}

// NewEnhancedPlaystoreApiHandler creates a new instance of the enhanced handler.
func NewEnhancedPlaystoreApiHandler(svc service.PlaystoreApiService) *EnhancedPlaystoreApiHandler {
	return &EnhancedPlaystoreApiHandler{service: svc}
}

// -------------------------
// 🚀 GetUserSubscriptionPurchase - Enhanced Version
// -------------------------

func (h *EnhancedPlaystoreApiHandler) GetUserSubscriptionPurchase(c *gin.Context) {
	ctx := c.Request.Context()

	// Use direct query parameters like the original handler
	purchaseToken := c.Query("purchase_token")
	packageName := c.Query("package_name")

	// Enhanced validation with proper error handling
	requiredFields := map[string]string{
		"purchase_token": purchaseToken,
		"package_name":   packageName,
	}
	if err := validation.ValidateRequiredFields(requiredFields); err != nil {
		apperrors.HandleAppError(c, err)
		return
	}

	// Validate package name format
	if err := validation.ValidatePackageNameManual(packageName); err != nil {
		apperrors.HandleAppError(c, err)
		return
	}

	// Enhanced logging with context
	contextLogger := logger.FromContext(ctx)
	contextLogger.WithFields(logrus.Fields{
		"package_name":   packageName,
		"purchase_token": logger.MaskSensitiveData(purchaseToken, 10), // Mask sensitive data
		"handler":        "GetUserSubscriptionPurchase",
	}).Info("Processing subscription purchase request")

	// Fetch subscription details with enhanced error handling
	subscriptionPurchase, err := h.service.GetUserSubscriptionPurchase(ctx, purchaseToken, packageName)
	if err != nil {
		// Enhanced error logging with correlation ID
		logger.LogError(ctx, err, "Failed to fetch subscription details", logrus.Fields{
			"package_name":   packageName,
			"purchase_token": logger.MaskSensitiveData(purchaseToken, 10),
		})

		// Enhanced error handling with proper status codes
		if errors.Is(err, context.DeadlineExceeded) {
			apperrors.HandleAppError(c, apperrors.TimeoutError("Request timed out", err))
			return
		}
		if errors.Is(err, context.Canceled) {
			apperrors.HandleAppError(c, apperrors.ValidationError("Request canceled by client", err.Error()))
			return
		}

		// Check if it's an external API error
		apperrors.HandleAppError(c, apperrors.ExternalAPIError("Failed to fetch subscription details", err))
		return
	}

	if subscriptionPurchase == nil {
		contextLogger.WithField("purchase_token", logger.MaskSensitiveData(purchaseToken, 10)).Warn("No subscription details found")
		apperrors.HandleAppError(c, apperrors.NotFoundError("No subscription details found for this purchase token", nil))
		return
	}

	// Success logging
	contextLogger.WithFields(logrus.Fields{
		"package_name":       packageName,
		"subscription_found": true,
	}).Info("Subscription details fetched successfully")

	// Use standardized success response
	c.JSON(200, apperrors.StandardSuccessResponse(subscriptionPurchase, "Subscription details fetched successfully"))
}

// -------------------------
// 🚀 ListSubscriptionProducts - Enhanced Version
// -------------------------
func (h *EnhancedPlaystoreApiHandler) ListSubscriptionProducts(c *gin.Context) {
	ctx := c.Request.Context()

	// Use direct query parameters like the original handler
	packageName := c.Query("package_name")

	// Enhanced validation
	requiredFields := map[string]string{
		"package_name": packageName,
	}
	if err := validation.ValidateRequiredFields(requiredFields); err != nil {
		apperrors.HandleAppError(c, err)
		return
	}

	if err := validation.ValidatePackageNameManual(packageName); err != nil {
		apperrors.HandleAppError(c, err)
		return
	}

	// Enhanced logging with business context
	ctx = logger.WithPackageName(ctx, packageName)
	contextLogger := logger.FromContext(ctx)
	contextLogger.WithFields(logrus.Fields{
		"package_name": packageName,
		"handler":      "ListSubscriptionProducts",
	}).Info("Processing subscription products list request")

	// Fetch products with enhanced error handling
	products, err := h.service.ListSubscriptionProducts(ctx, packageName)
	if err != nil {
		logger.LogError(ctx, err, "Failed to fetch subscription products", logrus.Fields{
			"package_name": packageName,
		})

		if errors.Is(err, context.DeadlineExceeded) {
			apperrors.HandleAppError(c, apperrors.TimeoutError("Request timed out", err))
			return
		}
		if errors.Is(err, context.Canceled) {
			apperrors.HandleAppError(c, apperrors.ValidationError("Request canceled by client", err.Error()))
			return
		}

		apperrors.HandleAppError(c, apperrors.ExternalAPIError("Failed to fetch subscription products", err))
		return
	}

	// Success logging with metrics
	contextLogger.WithFields(logrus.Fields{
		"package_name":   packageName,
		"products_found": products != nil,
	}).Info("Subscription products fetched successfully")

	c.JSON(200, apperrors.StandardSuccessResponse(products, "Subscription products fetched successfully"))
}

// -------------------------
// 🚀 GetSubscriptionProduct - Enhanced Version
// -------------------------
func (h *EnhancedPlaystoreApiHandler) GetSubscriptionProduct(c *gin.Context) {
	ctx := c.Request.Context()

	// Enhanced validation for query parameters
	var params validation.PlaystoreQueryParams

	// Check if middleware has already validated parameters
	if validatedParams := validation.GetValidatedParams(c); validatedParams != nil {
		if p, ok := validatedParams.(*validation.PlaystoreQueryParams); ok {
			params = *p
		} else {
			// Type assertion failed - log and fallback
			logger.LogError(ctx, nil, "Unexpected validated params type", logrus.Fields{
				"expected": "*validation.PlaystoreQueryParams",
				"actual":   fmt.Sprintf("%T", validatedParams),
			})
			apperrors.HandleAppError(c, apperrors.InternalError("Parameter validation type mismatch", nil))
			return
		}
	} else {
		// No pre-validated params - perform manual validation
		if err := c.ShouldBindQuery(&params); err != nil {
			logger.LogError(ctx, err, "Failed to bind query parameters", logrus.Fields{
				"handler": "GetSubscriptionProduct",
			})
			apperrors.HandleAppError(c, apperrors.ValidationError("Invalid query parameters", err.Error()))
			return
		}

		// Sanitize and validate the parameters
		validation.SanitizePlaystoreParams(&params)

		// Validate required fields
		requiredFields := map[string]string{
			"package_name": params.PackageName,
			"product_id":   params.ProductID,
		}
		if err := validation.ValidateRequiredFields(requiredFields); err != nil {
			apperrors.HandleAppError(c, err)
			return
		}

		// Additional validation for package name format
		if err := validation.ValidatePackageNameManual(params.PackageName); err != nil {
			apperrors.HandleAppError(c, err)
			return
		}
	}

	// Enhanced logging with business context
	ctx = logger.WithPackageName(ctx, params.PackageName)
	ctx = logger.WithProductID(ctx, params.ProductID)
	contextLogger := logger.FromContext(ctx)
	contextLogger.WithFields(logrus.Fields{
		"package_name": params.PackageName,
		"product_id":   params.ProductID,
		"handler":      "GetSubscriptionProduct",
	}).Info("Processing subscription product request")

	// Fetch product with enhanced error handling
	product, err := h.service.GetSubscriptionProductDetails(ctx, params.ProductID, params.PackageName)
	if err != nil {
		logger.LogError(ctx, err, "Failed to fetch subscription product", logrus.Fields{
			"package_name": params.PackageName,
			"product_id":   params.ProductID,
		})

		if errors.Is(err, context.DeadlineExceeded) {
			apperrors.HandleAppError(c, apperrors.TimeoutError("Request timed out", err))
			return
		}
		if errors.Is(err, context.Canceled) {
			apperrors.HandleAppError(c, apperrors.ValidationError("Request canceled by client", err.Error()))
			return
		}

		apperrors.HandleAppError(c, apperrors.ExternalAPIError("Failed to fetch subscription product", err))
		return
	}

	if product == nil {
		contextLogger.WithFields(logrus.Fields{
			"package_name": params.PackageName,
			"product_id":   params.ProductID,
		}).Warn("Subscription product not found")
		apperrors.HandleAppError(c, apperrors.NotFoundError("Subscription product not found", nil))
		return
	}

	// Success logging
	contextLogger.WithFields(logrus.Fields{
		"package_name":  params.PackageName,
		"product_id":    params.ProductID,
		"product_found": true,
	}).Info("Subscription product fetched successfully")

	c.JSON(200, apperrors.StandardSuccessResponse(product, "Subscription product fetched successfully"))
}

// Health check endpoint with enhanced monitoring
func (h *EnhancedPlaystoreApiHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	contextLogger := logger.FromContext(ctx)

	// Perform service health check
	if h.service == nil {
		contextLogger.Error("Service dependency is nil")
		apperrors.HandleAppError(c, apperrors.InternalError("Service unavailable", nil))
		return
	}

	contextLogger.Info("PlayStore API handler health check passed")
	c.JSON(200, apperrors.StandardSuccessResponse(map[string]string{
		"status":  "healthy",
		"service": "playstore-api",
	}, "Service is healthy"))
}
