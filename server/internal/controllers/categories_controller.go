package controllers

import (
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type CategoriesController struct {
	service *services.CategoriesService
}

func NewCategoriesController(service *services.CategoriesService) *CategoriesController {
	return &CategoriesController{
		service: service,
	}
}

func (cc *CategoriesController) GetCategories(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	categories, err := cc.service.GetCategories(userID.(string))
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch categories", err)
		return
	}

	utils.OK(c, "Categories fetched successfully", categories)
}

func (cc *CategoriesController) UpdateCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	categoryID := c.Param("id")
	var req services.UpdateCategoryParams

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	err := cc.service.UpdateCategory(userID.(string), categoryID, req)
	if err != nil {
		utils.InternalServerError(c, "Failed to update category", err)
		return
	}

	utils.OK(c, "Category updated successfully", nil)
}

func (cc *CategoriesController) CreateCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	var req services.CreateCategoryParams
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	category, err := cc.service.CreateCategory(userID.(string), req)
	if err != nil {
		utils.InternalServerError(c, "Failed to create category", err)
		return
	}

	utils.Created(c, "Category created successfully", category)
}

func (cc *CategoriesController) DeleteCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	categoryID := c.Param("id")

	err := cc.service.DeleteCategory(userID.(string), categoryID)
	if err != nil {
		utils.InternalServerError(c, "Failed to delete category", err)
		return
	}

	utils.OK(c, "Category deleted successfully", nil)
}
