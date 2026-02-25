package controllers

import (
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type HealthController struct {
	service *services.HealthService
}

func NewHealthController(service *services.HealthService) *HealthController {
	return &HealthController{service: service}
}

func (h *HealthController) GetHealth(c *gin.Context) {
	status := h.service.GetHealth()
	utils.OK(c, "Server is healthy", status)
}

func (h *HealthController) Ping(c *gin.Context) {
	utils.OK(c, h.service.Ping(), nil)
}
