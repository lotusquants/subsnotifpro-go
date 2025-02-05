package routes

import (
	"subsnotifpro-go/internal/google_play"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and returns a Gin router with grouped routes
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Health check route
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "message": "SubsNotifPro backend is running!"})
	})

	// Google Play Service Account Routes
	googlePlayGroup := router.Group("/api/google-play")
	{
		googlePlayGroup.POST("/upload-service-account", google_play.UploadServiceAccountHandler)
		googlePlayGroup.GET("/validate-service-account", google_play.ValidateServiceAccountHandler)
	}

	return router
}
