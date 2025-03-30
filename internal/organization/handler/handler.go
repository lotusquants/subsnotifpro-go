package handler

import (
	"net/http"
	"subsnotifpro-go/internal/organization/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
}

func NewHandler(s service.Service) *Handler {
	return &Handler{s}
}

func (h *Handler) GetMyOrganization(c *gin.Context) {
	ownerID := c.GetString("user_id") // from JWT middleware

	org, err := h.service.GetOrganizationByOwner(c.Request.Context(), ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	c.JSON(http.StatusOK, org)
}
