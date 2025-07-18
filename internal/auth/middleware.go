// Package auth provides authentication middleware for HTTP handlers
package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ContextKey represents the key used to store user claims in context
type ContextKey string

const (
	UserClaimsKey ContextKey = "user_claims"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	authService *AuthService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth middleware validates JWT tokens and adds user claims to context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromHeader(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "missing or invalid authorization header",
			})
			c.Abort()
			return
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid token",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Add claims to context
		c.Set(string(UserClaimsKey), claims)
		c.Next()
	}
}

// RequireRole middleware checks if user has a specific role
func (m *AuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(string(UserClaimsKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "user claims not found in context",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*UserClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid user claims",
			})
			c.Abort()
			return
		}

		if !m.authService.HasRole(userClaims, role) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "forbidden",
				"message":       "insufficient permissions",
				"required_role": role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware checks if user has any of the specified roles
func (m *AuthMiddleware) RequireAnyRole(roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(string(UserClaimsKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "user claims not found in context",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*UserClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid user claims",
			})
			c.Abort()
			return
		}

		if !m.authService.HasAnyRole(userClaims, roles) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "forbidden",
				"message":        "insufficient permissions",
				"required_roles": roles,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware validates JWT tokens if present but doesn't require them
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromHeader(c.GetHeader("Authorization"))
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			// Token is invalid, but we don't abort since it's optional
			c.Next()
			return
		}

		// Add claims to context
		c.Set(string(UserClaimsKey), claims)
		c.Next()
	}
}

// extractTokenFromHeader extracts the JWT token from the Authorization header
func (m *AuthMiddleware) extractTokenFromHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	// Expected format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// GetUserClaims retrieves user claims from the Gin context
func GetUserClaims(c *gin.Context) (*UserClaims, bool) {
	claims, exists := c.Get(string(UserClaimsKey))
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*UserClaims)
	if !ok {
		return nil, false
	}

	return userClaims, true
}

// GetUserClaimsFromContext retrieves user claims from a standard context
func GetUserClaimsFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*UserClaims)
	if !ok {
		return nil, false
	}
	return claims, true
}

// WithUserClaims adds user claims to a standard context
func WithUserClaims(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, UserClaimsKey, claims)
}

// IsAuthenticated checks if a user is authenticated in the current context
func IsAuthenticated(c *gin.Context) bool {
	_, exists := GetUserClaims(c)
	return exists
}

// HasRole checks if the current user has a specific role
func HasRole(c *gin.Context, role string) bool {
	claims, exists := GetUserClaims(c)
	if !exists {
		return false
	}

	for _, userRole := range claims.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the current user has any of the specified roles
func HasAnyRole(c *gin.Context, roles []string) bool {
	claims, exists := GetUserClaims(c)
	if !exists {
		return false
	}

	for _, role := range roles {
		for _, userRole := range claims.Roles {
			if userRole == role {
				return true
			}
		}
	}
	return false
}

// IsOwnerOrAdmin checks if the current user owns the resource or is an admin
func IsOwnerOrAdmin(c *gin.Context, resourceUserID string) bool {
	claims, exists := GetUserClaims(c)
	if !exists {
		return false
	}

	return claims.UserID == resourceUserID || HasRole(c, RoleAdmin)
}
