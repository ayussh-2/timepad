package routes

import (
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterReportsRoutes(rg *gin.RouterGroup, reportsController *controllers.ReportsController) {
	rg.GET("/reports", reportsController.GetReports)
}
