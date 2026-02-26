package routes

import (
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterCategoriesRoutes(rg *gin.RouterGroup, categoriesController *controllers.CategoriesController) {
	rg.GET("/categories", categoriesController.GetCategories)
	rg.POST("/categories", categoriesController.CreateCategory)
	rg.PATCH("/categories/:id", categoriesController.UpdateCategory)
	rg.DELETE("/categories/:id", categoriesController.DeleteCategory)
}
