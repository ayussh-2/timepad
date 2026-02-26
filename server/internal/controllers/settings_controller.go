package controllers

import (
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type SettingsController struct {
	service *services.SettingsService
}

func NewSettingsController(service *services.SettingsService) *SettingsController {
	return &SettingsController{
		service: service,
	}
}

func (sc *SettingsController) GetSettings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	settings, err := sc.service.GetSettings(userID.(string))
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch settings", err)
		return
	}

	utils.OK(c, "Settings fetched successfully", settings)
}

func (sc *SettingsController) UpdateSettings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	var req services.UpdateSettingsParams
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	err := sc.service.UpdateSettings(userID.(string), req)
	if err != nil {
		utils.InternalServerError(c, "Failed to update settings", err)
		return
	}

	utils.OK(c, "Settings updated successfully", nil)
}
