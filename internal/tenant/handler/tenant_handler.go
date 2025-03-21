package handler

import (
	"net/http"

	"subsnotifpro-go/internal/tenant/service"
	"subsnotifpro-go/internal/utils"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	service service.ITenantService
}

func NewTenantHandler(service service.ITenantService) *TenantHandler {
	return &TenantHandler{service: service}
}

func RegisterTenantRoutes(rg *gin.RouterGroup, handler *TenantHandler) {
	rg.POST("/create_tenant", handler.CreateTenantHandler)
	rg.GET("/tenant/:id", handler.GetTenantHandler)
}

// ✅ CreateTenantHandler - Handles POST /create_tenant
func (h *TenantHandler) CreateTenantHandler(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tenant, err := h.service.CreateTenant(c.Request.Context(), req.Name)
	if err != nil {
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{
		"success": true,
		"tenant":  tenant,
	})
}

// ✅ GetTenantHandler - Handles GET /tenant/:id
func (h *TenantHandler) GetTenantHandler(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "tenant ID is required")
		return
	}

	tenant, err := h.service.GetTenantByID(c.Request.Context(), tenantID)
	if err != nil {
		utils.WriteGinErrorResponse(c, http.StatusNotFound, "tenant not found")
		return
	}

	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{
		"success": true,
		"tenant":  tenant,
	})
}
