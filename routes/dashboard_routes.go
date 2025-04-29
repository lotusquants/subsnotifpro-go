package routes

import (
	"github.com/gin-gonic/gin"
)

func registerDashboardRoutes(router *gin.Engine, deps *RouteDependencies) {
	group := router.Group("/api/dashboard")

	group.GET("/get", deps.DashboardHandler.GetDashboard)

}
