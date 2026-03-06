package routes

import (
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterDevicesRoutes(rg *gin.RouterGroup, devicesController *controllers.DevicesController) {
	rg.GET("/devices", devicesController.GetDevices)
	rg.POST("/devices", devicesController.RegisterDevice)
	rg.PATCH("/devices/:id", devicesController.RenameDevice)
	rg.DELETE("/devices/:id", devicesController.DeleteDevice)
}
