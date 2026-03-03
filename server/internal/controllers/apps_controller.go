package controllers

import (
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type AppsController struct {
	service *services.AppsService
}

func NewAppsController(service *services.AppsService) *AppsController {
	return &AppsController{service: service}
}

// ListApps returns all tracked apps for the authenticated user.
func (ac *AppsController) ListApps(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	apps, err := ac.service.ListApps(userID.(string))
	if err != nil {
		utils.HandleError(c, "Failed to fetch apps", err)
		return
	}

	utils.OK(c, "Apps fetched successfully", apps)
}

// SetAppCategory assigns (or clears) a category for an app by ID.
// Body: { "category_id": "uuid" | null }
type setAppCategoryPayload struct {
	CategoryID *string `json:"category_id"`
}

func (ac *AppsController) SetAppCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	appID := c.Param("id")
	var req setAppCategoryPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	app, err := ac.service.SetAppCategory(userID.(string), appID, req.CategoryID)
	if err != nil {
		utils.HandleError(c, "Failed to set app category", err)
		return
	}

	utils.OK(c, "App category updated", app)
}

// SetAppSystem marks or unmarks an app as a system app.
// Body: { "is_system": true | false }
type setAppSystemPayload struct {
	IsSystem bool `json:"is_system"`
}

func (ac *AppsController) SetAppSystem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	appID := c.Param("id")
	var req setAppSystemPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	app, err := ac.service.SetAppSystem(userID.(string), appID, req.IsSystem)
	if err != nil {
		utils.HandleError(c, "Failed to update app system flag", err)
		return
	}

	utils.OK(c, "App system flag updated", app)
}

// ClassifyApp finds-or-creates a Productive/Distraction category and assigns it.
// Body: { "is_productive": true | false | null }
type classifyAppPayload struct {
	IsProductive *bool `json:"is_productive"`
}

func (ac *AppsController) ClassifyApp(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	appID := c.Param("id")
	var req classifyAppPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	app, err := ac.service.ClassifyApp(userID.(string), appID, req.IsProductive)
	if err != nil {
		utils.HandleError(c, "Failed to classify app", err)
		return
	}

	utils.OK(c, "App classified", app)
}
