package routes

import (
	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterEventsRoutes(rg *gin.RouterGroup, eventsController *controllers.EventsController) {
	eventsGroup := rg.Group("/events")
	{
		eventsGroup.POST("", eventsController.IngestEvents)
		eventsGroup.GET("", eventsController.GetEvents)
		eventsGroup.PATCH("/:id", eventsController.EditEvent)
		eventsGroup.DELETE("/:id", eventsController.DeleteEvent)
	}

	rg.GET("/timeline", eventsController.GetTimeline)
}
