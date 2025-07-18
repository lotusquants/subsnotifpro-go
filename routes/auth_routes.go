package routes

import (
	"subsnotifpro-go/internal/auth"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(r *gin.Engine, deps *RouteDependencies) {
	// Check if auth dependencies are available
	if deps.AuthHandler == nil || deps.AuthMiddleware == nil {
		return // Skip auth routes if dependencies are not available
	}

	// Public routes (no authentication required)
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", deps.AuthHandler.Register)
		authGroup.POST("/login", deps.AuthHandler.Login)
		authGroup.POST("/refresh", deps.AuthHandler.RefreshToken)
	}

	// Protected routes (authentication required)
	protectedGroup := r.Group("/api/auth")
	protectedGroup.Use(deps.AuthMiddleware.RequireAuth())
	{
		protectedGroup.GET("/profile", deps.AuthHandler.GetProfile)
		protectedGroup.PUT("/profile", deps.AuthHandler.UpdateProfile)
	}

	// Admin routes (admin role required)
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(deps.AuthMiddleware.RequireAuth())
	adminGroup.Use(deps.AuthMiddleware.RequireRole(auth.RoleAdmin))
	{
		adminGroup.GET("/users", deps.AuthHandler.ListUsers)
		adminGroup.PUT("/users/:id/roles", deps.AuthHandler.UpdateUserRoles)
		adminGroup.DELETE("/users/:id", deps.AuthHandler.DeleteUser)
	}
}
