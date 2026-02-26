package controllers

import (
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type ReportsController struct {
	service *services.ReportsService
}

func NewReportsController(service *services.ReportsService) *ReportsController {
	return &ReportsController{
		service: service,
	}
}

func (rc *ReportsController) GetReports(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	var params services.ReportParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequest(c, "Invalid parameters", err.Error())
		return
	}

	reports, err := rc.service.GetReports(userID.(string), params)
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch reports", err)
		return
	}

	utils.OK(c, "Reports fetched successfully", reports)
}
