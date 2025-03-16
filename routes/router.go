package routes

import (
	playstoreClientHandler "subsnotifpro-go/internal/google_playstore/client/handler"

	playstoreRTDNHandler "subsnotifpro-go/internal/google_playstore/rtdn/handler"
	playstoreSettingsHandler "subsnotifpro-go/internal/google_playstore/settings/handler"

	playstoreSubscriptionCatalogHandler "subsnotifpro-go/internal/google_playstore/subscription_catalog/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

// SetupRouter initializes and returns a Gin router with grouped routes
func SetupRouter(ch *amqp.Channel, psRtdnHandler *playstoreRTDNHandler.RTDNHandler, psSettingsHandler *playstoreSettingsHandler.PlaystoreSettingsHandler, psClientHandler *playstoreClientHandler.PlaystoreClientHandler, psSubscriptionCatalogHandler *playstoreSubscriptionCatalogHandler.SubscriptionCatalogHandler) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(gin.Recovery()) // Auto-handle panics
	router.Use(gin.Logger())   // Log requests
	router.Use(cors.Default()) // Enable CORS for frontend access

	// Health check route
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "message": "SubsNotifPro backend is running!"})
	})

	// Create Google Play API group
	googlePlayGroup := router.Group("/api/google-play")
	registerGooglePlayRoutes(ch, googlePlayGroup, psRtdnHandler, psSettingsHandler, psClientHandler, psSubscriptionCatalogHandler)

	return router
}

// registerGooglePlayRoutes registers all Google Play-related routes
func registerGooglePlayRoutes(ch *amqp.Channel, r *gin.RouterGroup, psRtdnHandler *playstoreRTDNHandler.RTDNHandler, psSettingsHandler *playstoreSettingsHandler.PlaystoreSettingsHandler, psClientHandler *playstoreClientHandler.PlaystoreClientHandler, psSubscriptionCatalogHandler *playstoreSubscriptionCatalogHandler.SubscriptionCatalogHandler) {
	// Register RTDN routes
	playstoreRTDNHandler.RegisterRTDNRoutes(r, ch, psRtdnHandler)

	// Register Playstore Settings routes
	playstoreSettingsHandler.RegisterPlaystoreSettingsRoutes(r, psSettingsHandler)

	// Register Client Subscription routes
	playstoreClientHandler.RegisterClientRoutes(r, psClientHandler)

	// Register Subscription Sync routes
	playstoreSubscriptionCatalogHandler.RegisterPlaystoreSubscriptionCatalogRoutes(r, psSubscriptionCatalogHandler)
}
