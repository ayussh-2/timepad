package routes

import (
	"github.com/ayussh-2/timepad/config"
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/ayussh-2/timepad/internal/middleware"
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS())

	// Initialize services
	healthService := services.NewHealthService()

	// Initialize controllers
	healthController := controllers.NewHealthController(healthService)

	v1 := r.Group("/api/v1")

	RegisterHealthRoutes(r, v1, healthController)

	return r
}
