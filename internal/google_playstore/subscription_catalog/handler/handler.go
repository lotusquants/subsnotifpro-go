package handler

import (
	"log"
	"net/http"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/service"

	"github.com/gin-gonic/gin"
)

// SubscriptionCatalogHandler handles subscription sync-related endpoints.
type SubscriptionCatalogHandler struct {
	service service.SubscriptionCatalogService
}

// NewSubscriptionCatalogHandler initializes the handler with the service dependency.
func NewSubscriptionCatalogHandler(svc service.SubscriptionCatalogService) *SubscriptionCatalogHandler {
	return &SubscriptionCatalogHandler{service: svc}
}

// RegisterPlaystoreSubscriptionCatalogRoutes registers all routes for managing subscription products, base plans, and offers.
func RegisterPlaystoreSubscriptionCatalogRoutes(r *gin.RouterGroup, handler *SubscriptionCatalogHandler) {
	r.POST("/sync-subscription-catalog", handler.SyncSubscriptionCatalog)
}

// -------------------------
// ✅ HTTP Handler: Sync Subscription Catalog  (Runs in a Goroutine)
// -------------------------
func (h *SubscriptionCatalogHandler) SyncSubscriptionCatalog(c *gin.Context) {
	// ✅ Bind JSON Body
	type SyncRequest struct {
		PackageName string `json:"package_name" binding:"required"`
	}

	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package_name is required"})
		return
	}
	packageName := req.PackageName

	// ✅ Run in a background Goroutine
	go func() {
		log.Printf("🚀 Running subscription sync in background for package: %s", req.PackageName)
		if err := h.service.SyncSubscriptionCatalog(req.PackageName); err != nil {
			log.Printf("❌ Subscription sync failed for package: %s - %v", req.PackageName, err)
		} else {
			log.Printf("✅ Subscription sync completed for package: %s", req.PackageName)
		}
	}()

	// ✅ Immediately respond with "Accepted"
	c.JSON(http.StatusAccepted, gin.H{"message": "Subscription sync started", "packageName": packageName})
}
