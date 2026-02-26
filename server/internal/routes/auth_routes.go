package routes

import (
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/ayussh-2/timepad/internal/middleware"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine, v1 *gin.RouterGroup, authController *controllers.AuthController, jwtUtil *utils.JWTUtil) {

	v1.POST("/register", authController.Register)
	v1.POST("/login", authController.Login)
	v1.GET("/refresh", middleware.Auth(jwtUtil), authController.Refresh)
}
