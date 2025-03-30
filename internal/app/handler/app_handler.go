package handler

import (
	"net/http"
	"subsnotifpro-go/internal/app/models"
	"subsnotifpro-go/internal/app/service"

	"github.com/gin-gonic/gin"
)

type AppHandler interface {
	CreateApp(c *gin.Context)
	GetAppWithSetup(c *gin.Context)
	UpdateSetupStatus(c *gin.Context)
}

type appHandler struct {
	appService service.AppServiceInterface
}

func NewAppHandler(appService service.AppServiceInterface) AppHandler {
	return &appHandler{appService: appService}
}
func (h *appHandler) CreateApp(c *gin.Context) {
	var app models.App
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.appService.CreateAppWithSetup(c.Request.Context(), &app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create app and setup"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "App created successfully", "app_id": app.ID})
}

func (h *appHandler) GetAppWithSetup(c *gin.Context) {
	appID := c.Query("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id is required"})
		return
	}

	app, setup, err := h.appService.GetAppWithSetup(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get app or setup"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"app":   app,
		"setup": setup,
	})
}

func (h *appHandler) UpdateSetupStatus(c *gin.Context) {
	var req struct {
		AppID       string             `json:"app_id" binding:"required"`
		SetupStatus models.SetupStatus `json:"setup_status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id and setup_status are required"})
		return
	}

	if err := h.appService.UpdateSetupStatus(c.Request.Context(), req.AppID, req.SetupStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setup status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setup status updated successfully"})
}
