package routes

import (
	"time"

	"github.com/gin-gonic/gin"
)

func registerGooglePlayRoutes(router *gin.Engine, deps *RouteDependencies) {
	group := router.Group("/api/google-play")

	// 🟢 RTDN Webhooks and DLQ
	group.POST("/rtdn/webhooks", func(c *gin.Context) {
		deps.PlaystoreRTDNHandler.WebhookHandler(c)
	})
	group.GET("/rtdn/dlq/size", deps.PlaystoreRTDNHandler.GetDLQSize)
	group.POST("/rtdn/dlq/retry", deps.PlaystoreRTDNHandler.RetryDLQHandler)

	// 🟢 Playstore Settings
	group.POST("/save-settings", deps.PlaystoreSettingsHandler.SaveSettingsHandler)
	group.GET("/get-settings", deps.PlaystoreSettingsHandler.GetSettingsHandler)
	group.DELETE("/delete-settings", deps.PlaystoreSettingsHandler.DeleteSettingsHandler)

	// 🔥 Enhanced Playstore Subscription Api (with new middleware stack)
	enhancedGroup := router.Group("/api/google-play/enhanced")
	if deps.EnhancedMiddleware != nil {
		enhancedGroup.Use(deps.EnhancedMiddleware.RequestLogging())
		enhancedGroup.Use(deps.EnhancedMiddleware.CorrelationID())
		enhancedGroup.Use(deps.EnhancedMiddleware.RateLimit())
		enhancedGroup.Use(deps.EnhancedMiddleware.CircuitBreakerSimple())
		enhancedGroup.Use(deps.EnhancedMiddleware.Timeout(30 * time.Second))
		enhancedGroup.Use(deps.EnhancedMiddleware.ErrorHandler())

		enhancedGroup.GET("/fetch-user-subscription-purchase", deps.EnhancedPlaystoreApiHandler.GetUserSubscriptionPurchase)
		enhancedGroup.GET("/fetch-list-subscription-products", deps.EnhancedPlaystoreApiHandler.ListSubscriptionProducts)
		enhancedGroup.GET("/fetch-subscription-product-details", deps.EnhancedPlaystoreApiHandler.GetSubscriptionProduct)
		enhancedGroup.GET("/health", deps.EnhancedPlaystoreApiHandler.HealthCheck)
	}

	// � Legacy Playstore Subscription Api (keeping for backwards compatibility)
	group.GET("/fetch-user-subscription-purchase", deps.PlaystoreApiHandler.GetUserSubscriptionPurchase)
	group.GET("/fetch-list-subscription-products", deps.PlaystoreApiHandler.ListSubscriptionProducts)
	group.GET("/fetch-subscription-product-details", deps.PlaystoreApiHandler.GetSubscriptionProductDetails)
	group.GET("/fetch-subscription-offers", deps.PlaystoreApiHandler.GetSubscriptionOffers)

	// 🟢 Playstore Subscription Catalog Sync
	group.POST("/subscription-catalog/sync", deps.PlaystoreSubscriptionCatalogHandler.SyncSubscriptionCatalogHandler)

	group.GET("/get-subscription-product-details", deps.PlaystoreSubscriptionCatalogHandler.GetSubscriptionProductDetailsHandler)
	group.GET("/check-subscription-product-exists", deps.PlaystoreSubscriptionCatalogHandler.CheckSubscriptionProductExistsHandler)
	group.GET("/list-all-subscription-products", deps.PlaystoreSubscriptionCatalogHandler.ListAllSubscriptionProductsHandler)

	group.GET("/get-subscription-baseplan-details", deps.PlaystoreSubscriptionCatalogHandler.GetBasePlanDetailsHandler)
	group.GET("/check-subscription-baseplan-active", deps.PlaystoreSubscriptionCatalogHandler.CheckBasePlanActiveHandler)
	group.GET("/check-subscription-baseplan-availability-in-region", deps.PlaystoreSubscriptionCatalogHandler.CheckBasePlanAvailabilityInRegionHandler)
	group.GET("/list-subscription-baseplan-names", deps.PlaystoreSubscriptionCatalogHandler.ListBasePlanNamesHandler)
	group.GET("/get-regional-baseplan-price", deps.PlaystoreSubscriptionCatalogHandler.GetRegionalBasePlanPriceHandler)
	group.GET("/get-other-regions-baseplan-price", deps.PlaystoreSubscriptionCatalogHandler.GetOtherRegionsBasePlanPriceHandler)

	group.GET("/get-subscription-offer-details", deps.PlaystoreSubscriptionCatalogHandler.GetSubscriptionOfferDetailsHandler)
	group.GET("/list-subscription-offer-names", deps.PlaystoreSubscriptionCatalogHandler.ListOfferNamesForBasePlanHandler)
	group.GET("/check-subscription-offer-active", deps.PlaystoreSubscriptionCatalogHandler.CheckSubscriptionOfferActiveHandler)

	group.GET("/get-subscription-offer-phases", deps.PlaystoreSubscriptionCatalogHandler.GetOfferPhasesHandler)
	group.GET("/check-offer-phase-exists", deps.PlaystoreSubscriptionCatalogHandler.CheckSubscriptionOfferPhaseExistsHandler)
	group.GET("/get-current-offer-phase", deps.PlaystoreSubscriptionCatalogHandler.GetCurrentOfferPhaseHandler)
	group.GET("/get-regional-offer-phase-price", deps.PlaystoreSubscriptionCatalogHandler.GetRegionalOfferPhasePriceHandler)
}
