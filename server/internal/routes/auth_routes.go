package routes

import (
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine, v1 *gin.RouterGroup, authContoller *controllers.AuthController) {

	v1.POST("/register", authContoller.Register)

}
