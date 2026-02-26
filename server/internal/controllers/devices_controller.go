package controllers

import (
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type DevicesController struct {
	service *services.DevicesService
}

func NewDevicesController(service *services.DevicesService) *DevicesController {
	return &DevicesController{
		service: service,
	}
}

func (dc *DevicesController) GetDevices(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	devices, err := dc.service.GetDevices(userID.(string))
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch devices", err)
		return
	}

	utils.OK(c, "Devices fetched successfully", devices)
}

func (dc *DevicesController) RegisterDevice(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	var req services.RegisterDeviceParams
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	device, err := dc.service.RegisterDevice(userID.(string), req)
	if err != nil {
		utils.InternalServerError(c, "Failed to register device", err)
		return
	}

	utils.Created(c, "Device registered successfully", device)
}

func (dc *DevicesController) DeleteDevice(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	deviceID := c.Param("id")

	err := dc.service.DeleteDevice(userID.(string), deviceID)
	if err != nil {
		utils.InternalServerError(c, "Failed to delete device", err)
		return
	}

	utils.OK(c, "Device deleted successfully", nil)
}
