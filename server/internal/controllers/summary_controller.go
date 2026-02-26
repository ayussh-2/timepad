package controllers

import (
	"time"

	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type SummaryController struct {
	service *services.SummaryService
}

func NewSummaryController(service *services.SummaryService) *SummaryController {
	return &SummaryController{
		service: service,
	}
}

func (sc *SummaryController) GetDailySummary(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	summary, err := sc.service.GetDailySummary(userID.(string), date)
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch daily summary", err)
		return
	}

	utils.OK(c, "Daily summary fetched successfully", summary)
}

func (sc *SummaryController) GetWeeklySummary(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	summary, err := sc.service.GetWeeklySummary(userID.(string), date)
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch weekly summary", err)
		return
	}

	utils.OK(c, "Weekly summary fetched successfully", summary)
}
