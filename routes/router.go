package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(deps *RouteDependencies) *gin.Engine {
	router := gin.Default()

	// Global Middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(cors.Default())

	// Health check endpoints
	if deps.HealthChecker != nil {
		router.GET("/health", gin.WrapH(deps.HealthChecker.HTTPHealthHandler()))
		router.GET("/health/ready", gin.WrapH(deps.HealthChecker.ReadinessHandler()))
		router.GET("/health/live", gin.WrapH(deps.HealthChecker.LivenessHandler()))
	}

	// Legacy health check for backwards compatibility
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	// Register all grouped routes

	registerGooglePlayRoutes(router, deps)
	registerAppStoreRoutes(router, deps)
	registerAuthRoutes(router, deps)
	registerDashboardRoutes(router, deps)
	registerUnifiedSubscriptionRoutes(router, deps)

	return router
}
