package routes

import (
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterHealthRoutes(r *gin.Engine, v1 *gin.RouterGroup, healthController *controllers.HealthController) {
	// Root health check endpoint
	r.GET("/health", healthController.GetHealth)

	// API v1 ping endpoint
	v1.GET("/ping", healthController.Ping)
}
