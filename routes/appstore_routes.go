package routes

import "github.com/gin-gonic/gin"

func registerAppStoreRoutes(router *gin.Engine, deps *RouteDependencies) {
	group := router.Group("/api/app-store")

	// 🟢 RTDN Webhooks and DLQ
	group.POST("/webhooks", func(c *gin.Context) {
		deps.AppStoreWebhookHandler.NotificationHandler(c)
	})

	// 🔵 App Store Settings Routes
	settingsGroup := group.Group("/settings")
	{
		settingsGroup.GET("", deps.AppStoreSettingsHandler.GetSettings) // Requires bundle_id query param
		settingsGroup.GET("/all", deps.AppStoreSettingsHandler.GetAllSettings)
		settingsGroup.POST("", deps.AppStoreSettingsHandler.UpdateSettings)
		settingsGroup.PUT("", deps.AppStoreSettingsHandler.UpdateSettings)
		settingsGroup.DELETE("", deps.AppStoreSettingsHandler.DeleteSettings) // Requires bundle_id query param
		settingsGroup.GET("/jwt/generate", deps.AppStoreSettingsHandler.GenerateJWT)
		settingsGroup.POST("/send-test-notification", deps.AppStoreSettingsHandler.SendTestNotification)

	}

}
