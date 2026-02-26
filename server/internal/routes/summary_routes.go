package routes

import (
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterSummaryRoutes(rg *gin.RouterGroup, summaryController *controllers.SummaryController) {
	summaryGroup := rg.Group("/summary")
	{
		summaryGroup.GET("/daily", summaryController.GetDailySummary)
		summaryGroup.GET("/weekly", summaryController.GetWeeklySummary)
	}
}
