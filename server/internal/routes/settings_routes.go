package routes

import (
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterSettingsRoutes(rg *gin.RouterGroup, settingsController *controllers.SettingsController) {
	rg.GET("/settings", settingsController.GetSettings)
	rg.PUT("/settings", settingsController.UpdateSettings)
}
