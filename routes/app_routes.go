package routes

import (
	"subsnotifpro-go/internal/app/handler"

	"github.com/gin-gonic/gin"
)

func RegisterAppRoutes(r *gin.RouterGroup, handler handler.AppHandler) {
	r.POST("/app/create", handler.CreateApp)
	r.GET("/app/with-setup", handler.GetAppWithSetup)
	r.POST("/app/update-setup-status", handler.UpdateSetupStatus)
}

func RegisterPlaystoreSetupRoutes(r *gin.RouterGroup, handler handler.PlaystoreSetupHandlerInterface) {
	r.POST("/setup/save", handler.SaveSetup)
	r.POST("/setup/validate-service-account", handler.ValidateServiceAccount)
	r.GET("/setup", handler.GetSetupByAppID)
	r.POST("/setup/update-bucket-id", handler.UpdateBucketID)
}
