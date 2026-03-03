package routes

import (
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterAppsRoutes(rg *gin.RouterGroup, appsController *controllers.AppsController) {
	appsGroup := rg.Group("/apps")
	{
		appsGroup.GET("", appsController.ListApps)
		appsGroup.PATCH("/:id/category", appsController.SetAppCategory)
		appsGroup.PATCH("/:id/classify", appsController.ClassifyApp)
		appsGroup.PATCH("/:id/system", appsController.SetAppSystem)
	}
}
