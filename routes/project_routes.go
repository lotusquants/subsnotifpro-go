package routes

import (
	"subsnotifpro-go/internal/project/handler"

	"github.com/gin-gonic/gin"
)

func RegisterProjectRoutes(rg *gin.RouterGroup, h *handler.Handler) {
	project := rg.Group("/projects")

	project.POST("", h.CreateProject)
	project.GET("", h.ListProjects)
}
