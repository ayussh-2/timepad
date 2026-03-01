package routes

import (
	"context"

	"github.com/ayussh-2/timepad/config"
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/ayussh-2/timepad/internal/middleware"
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB, jwtUtil *utils.JWTUtil, rdb *redis.Client) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(cfg.RateLimitRPM))

	// Initialize services
	healthService := services.NewHealthServiceWithRedis(rdb)
	authService := services.NewAuthService(db, jwtUtil)
	eventsService := services.NewEventsServiceWithQueue(db, rdb)
	go eventsService.StartIngestWorker(context.Background())
	summaryService := services.NewSummaryService(db)
	reportsService := services.NewReportsService(db)
	categoriesService := services.NewCategoriesService(db)
	devicesService := services.NewDevicesService(db)
	settingsService := services.NewSettingsService(db)

	// Initialize controllers
	healthController := controllers.NewHealthController(healthService)
	authController := controllers.NewAuthController(authService)
	eventsController := controllers.NewEventsController(eventsService)
	summaryController := controllers.NewSummaryController(summaryService)
	reportsController := controllers.NewReportsController(reportsService)
	categoriesController := controllers.NewCategoriesController(categoriesService)
	devicesController := controllers.NewDevicesController(devicesService)
	settingsController := controllers.NewSettingsController(settingsService)

	v1 := r.Group("/api/v1")

	// Public routes
	auth := v1.Group("/auth")
	RegisterAuthRoutes(r, auth, authController, jwtUtil)
	RegisterHealthRoutes(r, v1, healthController)

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.Auth(jwtUtil))
	{
		RegisterEventsRoutes(protected, eventsController)
		RegisterSummaryRoutes(protected, summaryController)
		RegisterReportsRoutes(protected, reportsController)
		RegisterCategoriesRoutes(protected, categoriesController)
		RegisterDevicesRoutes(protected, devicesController)
		RegisterSettingsRoutes(protected, settingsController)
	}

	return r
}
