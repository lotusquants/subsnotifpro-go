package handler

import (
	"net/http"

	"subsnotifpro-go/internal/tenant/dto"
	"subsnotifpro-go/internal/tenant/models"
	"subsnotifpro-go/internal/tenant/service"
	"subsnotifpro-go/internal/utils"

	"github.com/gin-gonic/gin"
)

type AppHandler struct {
	service service.IAppService
}

func NewAppHandler(s service.IAppService) *AppHandler {
	return &AppHandler{service: s}
}

func RegisterAppRoutes(rg *gin.RouterGroup, handler *AppHandler) {
	rg.POST("/create_app", handler.CreateApp)
	rg.GET("/tenant/:tenant_id/apps", handler.ListAppsByTenant)
}

// ✅ Create a new app for a tenant
func (h *AppHandler) CreateApp(c *gin.Context) {
	var req dto.CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// ✅ Basic platform validation
	if req.Platform != "play_store" && req.Platform != "app_store" {
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "Invalid platform: must be 'play_store' or 'app_store'")
		return
	}

	app := &models.App{
		TenantID:      req.TenantID,
		Platform:      req.Platform,
		AppIdentifier: req.PackageName,
		Name:          req.Name,
	}

	newApp, err := h.service.CreateApp(c.Request.Context(), app)
	if err != nil {
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteGinJSONResponse(c, http.StatusCreated, gin.H{
		"success": true,
		"app":     newApp,
	})
}

// ✅ List apps for a specific tenant
func (h *AppHandler) ListAppsByTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "tenant_id is required")
		return
	}

	apps, err := h.service.ListApps(c.Request.Context(), tenantID)
	if err != nil {
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{
		"success": true,
		"apps":    apps,
	})
}
