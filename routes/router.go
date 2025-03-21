package routes

import (
	playstoreClientHandler "subsnotifpro-go/internal/google_playstore/client/handler"

	playstoreRTDNHandler "subsnotifpro-go/internal/google_playstore/rtdn/handler"
	playstoreSettingsHandler "subsnotifpro-go/internal/google_playstore/settings/handler"

	authHandlerPkg "subsnotifpro-go/internal/auth/handler"
	authMiddleware "subsnotifpro-go/internal/auth/middleware"

	tenantHandlerPkg "subsnotifpro-go/internal/tenant/handler" // Tenant & App handlers

	playstoreSubscriptionCatalogHandler "subsnotifpro-go/internal/google_playstore/subscription_catalog/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // 👈 Required for Swagger UI
)

// SetupRouter initializes and returns a Gin router with grouped routes
func SetupRouter(ch *amqp.Channel,
	psRtdnHandler *playstoreRTDNHandler.RTDNHandler,
	psSettingsHandler *playstoreSettingsHandler.PlaystoreSettingsHandler,
	psClientHandler *playstoreClientHandler.PlaystoreClientHandler,
	psSubscriptionCatalogHandler *playstoreSubscriptionCatalogHandler.SubscriptionCatalogHandler,
	authHandler *authHandlerPkg.AuthHandler,
	tenantHandler *tenantHandlerPkg.TenantHandler,
	appHandler *tenantHandlerPkg.AppHandler) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(gin.Recovery()) // Auto-handle panics
	router.Use(gin.Logger())   // Log requests
	router.Use(cors.Default()) // Enable CORS for frontend access

	// Health check route
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "message": "SubsNotifPro backend is running!"})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authGroup := router.Group("/api/auth")

	authHandlerPkg.RegisterAuthRoutes(authGroup, authHandler)
	adminGroup := router.Group("/api/admin")
	adminGroup.Use(authMiddleware.JWTAuthMiddleware())
	adminGroup.Use(authMiddleware.RequireRole("admin", "owner"))
	authHandlerPkg.RegisterAdminRoutes(adminGroup, authHandler)

	// 🏢 Tenant routes (public)
	// ------------------------------
	tenantGroup := router.Group("/api/tenant")
	tenantHandlerPkg.RegisterTenantRoutes(tenantGroup, tenantHandler)

	// ------------------------------
	// 📱 App routes (protected)
	// ------------------------------
	appGroup := router.Group("/api/app")
	appGroup.Use(authMiddleware.JWTAuthMiddleware())
	appGroup.Use(authMiddleware.RequireRole("admin", "owner"))
	tenantHandlerPkg.RegisterAppRoutes(appGroup, appHandler)

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
