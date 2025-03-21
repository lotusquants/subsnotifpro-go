package handler

import (
	"log"
	"net/http"

	"subsnotifpro-go/internal/auth/dto" // NEW: DTOs for request structs
	"subsnotifpro-go/internal/auth/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.IAdminService
}

func NewAuthHandler(s service.IAdminService) *AuthHandler {
	return &AuthHandler{service: s}
}

func RegisterAuthRoutes(rg *gin.RouterGroup, handler *AuthHandler) {
	rg.POST("/register", handler.Register)
	rg.POST("/login", handler.Login)
}

func RegisterAdminRoutes(rg *gin.RouterGroup, handler *AuthHandler) {
	rg.GET("/dashboard", handler.Dashboard)
	rg.GET("/me", handler.Me)
}

func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user_id":   c.GetString("user_id"),
		"role":      c.GetString("role"),
		"tenant_id": c.GetString("tenant_id"),
		"success":   true,
	})
}

func (h *AuthHandler) Dashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Access granted to admin dashboard!",
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin, err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.TenantID)
	if err != nil {
		log.Printf("register error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"admin":   admin,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("login error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
	})
}
