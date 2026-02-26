package services

import (
	"errors"
	"strings"
	"time"

	"github.com/ayussh-2/timepad/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventsService struct {
	db *gorm.DB
}

func NewEventsService(db *gorm.DB) *EventsService {
	return &EventsService{
		db: db,
	}
}

type EventInput struct {
	AppName     string    `json:"app_name" binding:"required"`
	WindowTitle string    `json:"window_title"`
	URL         string    `json:"url"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
	IsIdle      bool      `json:"is_idle"`
}

type IngestEventsParams struct {
	UserID    string
	DeviceKey string
	Events    []EventInput
}

type UpdateEventParams struct {
	CategoryID *string `json:"category_id"`
	IsPrivate  *bool   `json:"is_private"`
}

func (s *EventsService) IngestEvents(params IngestEventsParams) (int, error) {
	var device models.Device
	if err := s.db.Where("device_key = ? AND user_id = ?", params.DeviceKey, params.UserID).First(&device).Error; err != nil {
		return 0, errors.New("unknown device")
	}

	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		return 0, errors.New("invalid user ID")
	}

	// Load user settings for excluded apps/urls filtering
	var settings models.UserSetting
	hasSettings := true
	if err := s.db.Where("user_id = ?", params.UserID).First(&settings).Error; err != nil {
		hasSettings = false
	}

	// Build lookup sets for O(1) filtering
	excludedApps := make(map[string]bool)
	excludedUrls := make(map[string]bool)
	if hasSettings {
		for _, app := range settings.ExcludedApps {
			excludedApps[strings.ToLower(app)] = true
		}
		for _, url := range settings.ExcludedUrls {
			excludedUrls[strings.ToLower(url)] = true
		}
	}

	events := make([]models.ActivityEvent, 0, len(params.Events))
	for _, e := range params.Events {
		duration := int(e.EndTime.Sub(e.StartTime).Seconds())
		if duration <= 0 {
			continue
		}

		// Filter excluded apps (case-insensitive)
		if excludedApps[strings.ToLower(e.AppName)] {
			continue
		}

		// Filter excluded URLs (case-insensitive hostname match)
		if e.URL != "" && excludedUrls[strings.ToLower(e.URL)] {
			continue
		}

		events = append(events, models.ActivityEvent{
			UserID:       userID,
			DeviceID:     device.ID,
			AppName:      e.AppName,
			WindowTitle:  e.WindowTitle,
			Url:          e.URL,
			StartTime:    e.StartTime,
			EndTime:      e.EndTime,
			DurationSecs: duration,
			IsIdle:       e.IsIdle,
		})
	}

	if len(events) == 0 {
		return 0, nil
	}

	if err := s.db.CreateInBatches(events, 100).Error; err != nil {
		return 0, errors.New("failed to save events")
	}

	// Update device LastSeenAt
	now := time.Now()
	s.db.Model(&device).Update("last_seen_at", now)

	return len(events), nil
}

func (s *EventsService) GetEvents(userID string, limit int, offset int) ([]models.ActivityEvent, error) {
	var events []models.ActivityEvent
	err := s.db.Where("user_id = ?", userID).
		Order("start_time desc").
		Limit(limit).
		Offset(offset).
		Find(&events).Error

	if err != nil {
		return nil, errors.New("failed to fetch events")
	}
	return events, nil
}

func (s *EventsService) GetTimeline(userID string, date string) ([]models.ActivityEvent, error) {
	var events []models.ActivityEvent

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	startOfDay := parsedDate.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	err = s.db.Where("user_id = ? AND start_time >= ? AND start_time < ?", userID, startOfDay, endOfDay).
		Preload("Category").
		Preload("Device").
		Order("start_time asc").
		Find(&events).Error

	if err != nil {
		return nil, errors.New("failed to fetch timeline")
	}
	return events, nil
}

func (s *EventsService) EditEvent(userID string, eventID string, params UpdateEventParams) error {
	var event models.ActivityEvent
	if err := s.db.Where("id = ? AND user_id = ?", eventID, userID).First(&event).Error; err != nil {
		return errors.New("event not found or unauthorized")
	}

	updates := map[string]interface{}{}
	if params.CategoryID != nil {
		if *params.CategoryID == "" {
			updates["category_id"] = nil
		} else {
			catID, err := uuid.Parse(*params.CategoryID)
			if err != nil {
				return errors.New("invalid category ID")
			}
			updates["category_id"] = catID
		}
	}
	if params.IsPrivate != nil {
		updates["is_private"] = *params.IsPrivate
	}

	if len(updates) > 0 {
		if err := s.db.Model(&event).Updates(updates).Error; err != nil {
			return errors.New("failed to update event")
		}
	}

	return nil
}

func (s *EventsService) DeleteEvent(userID string, eventID string) error {
	result := s.db.Where("id = ? AND user_id = ?", eventID, userID).Delete(&models.ActivityEvent{})
	if result.Error != nil {
		return errors.New("failed to delete event")
	}
	if result.RowsAffected == 0 {
		return errors.New("event not found or unauthorized")
	}
	return nil
}
