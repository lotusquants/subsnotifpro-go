package routes

import (
	"github.com/gin-gonic/gin"
)

func registerDashboardRoutes(router *gin.Engine, deps *RouteDependencies) {
	group := router.Group("/api/dashboard")

	group.GET("/get", deps.DashboardHandler.GetDashboard)

}

func registerUnifiedSubscriptionRoutes(router *gin.Engine, deps *RouteDependencies) {
	group := router.Group("/api")

	group.GET("/user/unified_subscriptions", deps.UnifiedSubscriptionsHandler.GetUserSubscriptions)
	group.GET("/subscriptions/:subscription_id/events", deps.UnifiedSubscriptionsHandler.GetSubscriptionEvents)

}
