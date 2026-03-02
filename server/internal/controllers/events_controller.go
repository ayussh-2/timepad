package controllers

import (
	"strconv"

	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type EventsController struct {
	service *services.EventsService
}

func NewEventsController(service *services.EventsService) *EventsController {
	return &EventsController{
		service: service,
	}
}

type IngestPayload struct {
	DeviceKey string                `json:"device_key" binding:"required"`
	Events    []services.EventInput `json:"events" binding:"required,min=1"`
}

func (ec *EventsController) IngestEvents(c *gin.Context) {
	var req IngestPayload

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	params := services.IngestEventsParams{
		UserID:    userID.(string),
		DeviceKey: req.DeviceKey,
		Events:    req.Events,
	}

	result, err := ec.service.IngestEvents(params)
	if err != nil {
		utils.HandleError(c, "Failed to ingest events", err)
		return
	}

	if result.Queued {
		utils.Accepted(c, "Events queued for processing", gin.H{"queued": result.Count})
	} else {
		utils.Created(c, "Events ingested successfully", gin.H{"inserted": result.Count})
	}
}

func (ec *EventsController) GetEvents(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	events, err := ec.service.GetEvents(userID.(string), limit, offset)
	if err != nil {
		utils.HandleError(c, "Failed to fetch events", err)
		return
	}

	utils.OK(c, "Events fetched successfully", events)
}

func (ec *EventsController) GetTimeline(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	date := c.Query("date")
	if date == "" {
		utils.BadRequest(c, "Date parameter is required (YYYY-MM-DD)", "")
		return
	}

	cursor := c.Query("cursor")
	appName := c.Query("app_name")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 {
		if l > 500 {
			l = 500
		}
		limit = l
	}

	page, err := ec.service.GetTimeline(userID.(string), date, cursor, appName, limit)
	if err != nil {
		utils.HandleError(c, "Failed to fetch timeline", err)
		return
	}

	utils.OK(c, "Timeline fetched successfully", page)
}

func (ec *EventsController) EditEvent(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	eventID := c.Param("id")
	var req services.UpdateEventParams

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	err := ec.service.EditEvent(userID.(string), eventID, req)
	if err != nil {
		utils.HandleError(c, "Failed to update event", err)
		return
	}

	utils.OK(c, "Event updated successfully", nil)
}

func (ec *EventsController) DeleteEvent(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	eventID := c.Param("id")

	err := ec.service.DeleteEvent(userID.(string), eventID)
	if err != nil {
		utils.HandleError(c, "Failed to delete event", err)
		return
	}

	utils.OK(c, "Event deleted successfully", nil)
}

type ClassifyAppPayload struct {
	AppName      string `json:"app_name" binding:"required"`
	IsProductive *bool  `json:"is_productive"` // null = clear / neutral
}

func (ec *EventsController) ClassifyApp(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	var req ClassifyAppPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	cat, err := ec.service.ClassifyAppProductivity(userID.(string), req.AppName, req.IsProductive)
	if err != nil {
		utils.HandleError(c, "Failed to classify app", err)
		return
	}

	utils.OK(c, "App classified", gin.H{"category": cat})
}

type BulkCategorizePayload struct {
	AppName    string  `json:"app_name" binding:"required"`
	CategoryID *string `json:"category_id"` // null to clear
}

func (ec *EventsController) BulkCategorizeEvents(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	var req BulkCategorizePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	catID := ""
	if req.CategoryID != nil {
		catID = *req.CategoryID
	}

	count, err := ec.service.BulkCategorizeApp(userID.(string), req.AppName, catID)
	if err != nil {
		utils.HandleError(c, "Failed to categorize app events", err)
		return
	}

	utils.OK(c, "App events categorized", gin.H{"updated": count})
}
