package routes

import (
	"time"

	"github.com/gin-gonic/gin"

	"subsnotifpro-go/internal/middleware"
	"subsnotifpro-go/routes"
)

// EnhancedPlaystoreRoutesWithCache demonstrates how to integrate caching middleware
// with existing PlayStore routes for optimal performance
func RegisterEnhancedPlaystoreRoutesWithCache(router *gin.Engine, deps *routes.RouteDependencies) {
	// Create cache middleware instances with different TTLs
	shortCache := middleware.NewCacheMiddleware(2 * time.Minute)   // For dynamic data
	mediumCache := middleware.NewCacheMiddleware(10 * time.Minute) // For semi-static data
	longCache := middleware.NewCacheMiddleware(1 * time.Hour)      // For static data

	group := router.Group("/api/google-play")

	// 🟢 RTDN Webhooks and DLQ (No caching for webhooks and write operations)
	group.POST("/rtdn/webhooks", func(c *gin.Context) {
		deps.PlaystoreRTDNHandler.WebhookHandler(c)
	})

	// Cache DLQ size with short TTL (data changes frequently)
	group.GET("/rtdn/dlq/size",
		shortCache.CacheResponseWithTTL(30*time.Second), // Very short cache for queue size
		deps.PlaystoreRTDNHandler.GetDLQSize)

	group.POST("/rtdn/dlq/retry",
		shortCache.CacheInvalidate("/api/google-play/rtdn/dlq/*"), // Invalidate DLQ cache after retry
		deps.PlaystoreRTDNHandler.RetryDLQHandler)

	// 🟢 Playstore Settings (Short cache, frequently updated)
	group.POST("/save-settings",
		shortCache.CacheInvalidate("/api/google-play/get-settings*"), // Invalidate settings cache
		deps.PlaystoreSettingsHandler.SaveSettingsHandler)

	group.GET("/get-settings",
		shortCache.CacheResponse(), // Cache for 2 minutes
		deps.PlaystoreSettingsHandler.GetSettingsHandler)

	group.DELETE("/delete-settings",
		shortCache.CacheInvalidate("/api/google-play/get-settings*"), // Invalidate settings cache
		deps.PlaystoreSettingsHandler.DeleteSettingsHandler)

	// 🟢 Playstore Subscription API (Medium cache, moderately static)
	group.GET("/fetch-user-subscription-purchase",
		mediumCache.CacheResponseWithTTL(5*time.Minute), // User-specific, medium cache
		deps.PlaystoreApiHandler.GetUserSubscriptionPurchase)

	group.GET("/fetch-list-subscription-products",
		longCache.CacheResponse(), // Product list changes rarely, long cache
		deps.PlaystoreApiHandler.ListSubscriptionProducts)

	group.GET("/fetch-subscription-product-details",
		longCache.CacheResponse(), // Product details change rarely
		deps.PlaystoreApiHandler.GetSubscriptionProductDetails)

	group.GET("/fetch-subscription-offers",
		mediumCache.CacheResponse(), // Offers change more frequently than products
		deps.PlaystoreApiHandler.GetSubscriptionOffers)

	// 🟢 Playstore Subscription Catalog Sync (Invalidate caches after sync)
	group.POST("/subscription-catalog/sync",
		// Invalidate all catalog-related caches after sync
		longCache.CacheInvalidate("/api/google-play/get-subscription-product*"),
		longCache.CacheInvalidate("/api/google-play/list-*"),
		longCache.CacheInvalidate("/api/google-play/fetch-*"),
		deps.PlaystoreSubscriptionCatalogHandler.SyncSubscriptionCatalogHandler)

	// 🟢 Subscription Product Management (Long cache for static data)
	group.GET("/get-subscription-product-details",
		longCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.GetSubscriptionProductDetailsHandler)

	group.GET("/check-subscription-product-exists",
		longCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.CheckSubscriptionProductExistsHandler)

	group.GET("/list-all-subscription-products",
		longCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.ListAllSubscriptionProductsHandler)

	// 🟢 Base Plan Management (Long cache for static configuration)
	group.GET("/get-subscription-baseplan-details",
		longCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.GetBasePlanDetailsHandler)

	group.GET("/check-subscription-baseplan-active",
		longCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.CheckBasePlanActiveHandler)

	group.GET("/check-subscription-baseplan-availability-in-region",
		longCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.CheckBasePlanAvailabilityInRegionHandler)

	group.GET("/list-subscription-baseplan-names",
		longCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.ListBasePlanNamesHandler)

	// 🟢 Pricing Information (Medium cache, prices can change but not frequently)
	group.GET("/get-regional-baseplan-price",
		mediumCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.GetRegionalBasePlanPriceHandler)

	group.GET("/get-other-regions-baseplan-price",
		mediumCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.GetOtherRegionsBasePlanPriceHandler)

	// 🟢 Subscription Offers (Medium cache, offers change more frequently)
	group.GET("/get-subscription-offer-details",
		mediumCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.GetSubscriptionOfferDetailsHandler)

	group.GET("/list-subscription-offer-names",
		mediumCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.ListOfferNamesForBasePlanHandler)

	group.GET("/check-subscription-offer-active",
		mediumCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.CheckSubscriptionOfferActiveHandler)

	// 🟢 Offer Phases (Medium cache, can change during promotions)
	group.GET("/get-subscription-offer-phases",
		mediumCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.GetOfferPhasesHandler)

	group.GET("/check-offer-phase-exists",
		mediumCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.CheckSubscriptionOfferPhaseExistsHandler)

	group.GET("/get-current-offer-phase",
		mediumCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.GetCurrentOfferPhaseHandler)

	group.GET("/get-regional-offer-phase-price",
		mediumCache.CacheResponse(),
		deps.PlaystoreSubscriptionCatalogHandler.GetRegionalOfferPhasePriceHandler)

	// 🟢 Cache Management Endpoints
	group.GET("/cache/stats",
		shortCache.CacheStats()) // Real-time cache statistics

	// Manual cache invalidation endpoint (for admin use)
	group.POST("/cache/invalidate", func(c *gin.Context) {
		pattern := c.Query("pattern")
		if pattern == "" {
			pattern = "/api/google-play/*" // Default to invalidate all Google Play caches
		}

		// Invalidate across all cache instances
		shortCache.InvalidatePattern(pattern)
		mediumCache.InvalidatePattern(pattern)
		longCache.InvalidatePattern(pattern)

		c.JSON(200, gin.H{
			"message": "Cache invalidated",
			"pattern": pattern,
		})
	})
}

// CacheConfiguration represents the caching strategy for different endpoint types
type CacheConfiguration struct {
	EndpointType string        `json:"endpoint_type"`
	TTL          time.Duration `json:"ttl"`
	Strategy     string        `json:"strategy"`
	Description  string        `json:"description"`
}

// GetCacheStrategy returns the recommended caching strategy for the PlayStore module
func GetPlayStoreCacheStrategy() []CacheConfiguration {
	return []CacheConfiguration{
		{
			EndpointType: "Webhooks",
			TTL:          0,
			Strategy:     "No Cache",
			Description:  "Real-time webhook processing, no caching required",
		},
		{
			EndpointType: "DLQ Status",
			TTL:          30 * time.Second,
			Strategy:     "Very Short Cache",
			Description:  "Queue sizes change rapidly, minimal caching",
		},
		{
			EndpointType: "Settings",
			TTL:          2 * time.Minute,
			Strategy:     "Short Cache",
			Description:  "User settings can change frequently",
		},
		{
			EndpointType: "User Subscriptions",
			TTL:          5 * time.Minute,
			Strategy:     "Medium Cache",
			Description:  "User-specific data, moderate caching",
		},
		{
			EndpointType: "Product Catalog",
			TTL:          1 * time.Hour,
			Strategy:     "Long Cache",
			Description:  "Product information changes rarely",
		},
		{
			EndpointType: "Pricing",
			TTL:          10 * time.Minute,
			Strategy:     "Medium Cache",
			Description:  "Prices can change but not frequently",
		},
		{
			EndpointType: "Offers & Promotions",
			TTL:          10 * time.Minute,
			Strategy:     "Medium Cache",
			Description:  "Promotional offers change more frequently than products",
		},
	}
}
