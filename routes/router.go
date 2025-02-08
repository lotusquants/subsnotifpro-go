package routes

import (
	playstoresettings "subsnotifpro-go/internal/google_playstore/google_playstore_settings"
	"subsnotifpro-go/internal/google_playstore/rtdn"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

// SetupRouter initializes and returns a Gin router with grouped routes
func SetupRouter(ch *amqp.Channel) *gin.Engine {
	router := gin.Default()

	// Health check route
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "message": "SubsNotifPro backend is running!"})
	})

	// Google Play Service Account Routes
	googlePlayGroup := router.Group("/api/google-play")
	{
		// Service Account Management
		googlePlayGroup.POST("/upload-service-account", playstoresettings.UploadServiceAccountHandler)
		googlePlayGroup.GET("/validate-service-account", playstoresettings.ValidateServiceAccountHandler)
		googlePlayGroup.GET("/service-account-status", playstoresettings.GetServiceAccountStatusHandler)
		googlePlayGroup.DELETE("/delete-service-account", playstoresettings.DeleteServiceAccountHandler)

		// Package Name Management
		googlePlayGroup.POST("/set-package-name", playstoresettings.SetPackageNameHandler)
		googlePlayGroup.GET("/get-package-name", playstoresettings.GetPackageNameHandler)

		googlePlayGroup.POST("/recieve-rtdn", func(c *gin.Context) {
			rtdn.WebhookHandler(c, ch) // Pass the channel here
		})
		googlePlayGroup.GET("/rtdn/dlq/size", rtdn.GetDLQSize) // API to check RTDN Dead Letter Queue size
		googlePlayGroup.GET("/rtdn/dlq/retry", rtdn.RetryDLQHandler)
	}

	return router
}
