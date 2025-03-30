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

	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	// Register all grouped routes

	registerGooglePlayRoutes(router, deps)
	registerAuthRoutes(router, deps)

	return router
}
