package routes

import (
	"github.com/ayussh-2/timepad/config"
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/ayussh-2/timepad/internal/middleware"
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB, jwtUtil *utils.JWTUtil) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS())

	// Initialize services
	healthService := services.NewHealthService()
	authService := services.NewAuthService(db, jwtUtil)

	// Initialize controllers
	healthController := controllers.NewHealthController(healthService)
	authController := controllers.NewAuthController(authService)

	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")

	RegisterHealthRoutes(r, v1, healthController)
	RegisterAuthRoutes(r, auth, authController, jwtUtil)

	return r
}
