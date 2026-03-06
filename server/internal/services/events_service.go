package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/ayussh-2/timepad/internal/models"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const ingestQueueKey = "timepad:ingest_queue"

type EventsService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewEventsService(db *gorm.DB) *EventsService {
	return &EventsService{db: db}
}

// NewEventsServiceWithQueue creates an EventsService backed by a Redis queue
// for asynchronous event ingestion. Pass a nil rdb to fall back to sync mode.
func NewEventsServiceWithQueue(db *gorm.DB, rdb *redis.Client) *EventsService {
	return &EventsService{db: db, rdb: rdb}
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
	IsPrivate *bool `json:"is_private"`
}

// IngestResult carries the outcome of an ingest call.
// Queued=true means events were placed on the async Redis queue rather than written to DB directly.
type IngestResult struct {
	Count  int
	Queued bool
}

// ingestJobPayload is the message serialized onto the Redis queue.
type ingestJobPayload struct {
	DeviceID string                 `json:"device_id"`
	UserID   string                 `json:"user_id"`
	Platform string                 `json:"platform"`
	Events   []models.ActivityEvent `json:"events"`
}

func (s *EventsService) IngestEvents(params IngestEventsParams) (IngestResult, error) {
	var device models.Device
	if err := s.db.Where("device_key = ? AND user_id = ?", params.DeviceKey, params.UserID).First(&device).Error; err != nil {
		return IngestResult{}, utils.NewNotFoundError("unknown device")
	}

	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		return IngestResult{}, errors.New("invalid user ID")
	}

	// TODO: load user settings for excluded apps/urls filtering when settings UI is ready
	excludedApps := make(map[string]bool)
	excludedUrls := make(map[string]bool)

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
		return IngestResult{Count: 0}, nil
	}

	// Async path: enqueue to Redis if available.
	if s.rdb != nil {
		payload := ingestJobPayload{
			DeviceID: device.ID.String(),
			UserID:   params.UserID,
			Platform: device.Platform,
			Events:   events,
		}
		if data, merr := json.Marshal(payload); merr == nil {
			if perr := s.rdb.LPush(context.Background(), ingestQueueKey, data).Err(); perr == nil {
				return IngestResult{Count: len(events), Queued: true}, nil
			} else {
				log.Printf("Redis enqueue failed, falling back to sync insert: %v", perr)
			}
		}
	}

	// Sync fallback path (no Redis or Redis unavailable).
	if err := s.processEvents(device.ID, params.UserID, device.Platform, events); err != nil {
		return IngestResult{}, err
	}
	return IngestResult{Count: len(events), Queued: false}, nil
}

// processEvents upserts apps for each unique app_name, resolves AppIDs onto events,
// batch-inserts events, and updates the device LastSeenAt.
func (s *EventsService) processEvents(deviceID uuid.UUID, userID, platform string, events []models.ActivityEvent) error {
	// Collect unique non-idle app names.
	appNameSet := make(map[string]struct{})
	for _, e := range events {
		if !e.IsIdle {
			appNameSet[e.AppName] = struct{}{}
		}
	}

	now := time.Now()
	appIDMap := make(map[string]uuid.UUID)

	for appName := range appNameSet {
		// Upsert: create app if new, otherwise extend platforms + update last_seen_at.
		err := s.db.Exec(`
			INSERT INTO apps (id, user_id, name, platforms, icon, first_seen_at, last_seen_at)
			VALUES (gen_random_uuid(), ?, ?, ?, '', ?, ?)
			ON CONFLICT (user_id, name) DO UPDATE SET
				platforms    = ARRAY(SELECT DISTINCT unnest(apps.platforms || EXCLUDED.platforms)),
				last_seen_at = GREATEST(apps.last_seen_at, EXCLUDED.last_seen_at)
		`, userID, appName, pq.StringArray{platform}, now, now).Error
		if err != nil {
			log.Printf("app upsert failed for %q: %v", appName, err)
			continue
		}

		var app models.App
		if err := s.db.Select("id").Where("user_id = ? AND name = ?", userID, appName).First(&app).Error; err != nil {
			continue
		}
		appIDMap[appName] = app.ID
	}

	// Assign AppID on every event.
	for i := range events {
		if id, ok := appIDMap[events[i].AppName]; ok {
			events[i].AppID = &id
		}
	}

	if err := s.db.CreateInBatches(events, 100).Error; err != nil {
		return errors.New("failed to save events")
	}
	s.db.Model(&models.Device{}).Where("id = ?", deviceID).Update("last_seen_at", now)
	return nil
}

// StartIngestWorker starts a blocking loop that pops ingest jobs from the Redis queue
// and processes them in the background. Must be called in a goroutine.
func (s *EventsService) StartIngestWorker(ctx context.Context) {
	if s.rdb == nil {
		return
	}
	log.Println("Ingest worker started, listening on", ingestQueueKey)
	for {
		select {
		case <-ctx.Done():
			log.Println("Ingest worker stopped")
			return
		default:
		}

		results, err := s.rdb.BRPop(ctx, 5*time.Second, ingestQueueKey).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			log.Printf("Ingest worker queue error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if len(results) < 2 {
			continue
		}

		var payload ingestJobPayload
		if err := json.Unmarshal([]byte(results[1]), &payload); err != nil {
			log.Printf("Ingest worker: failed to parse job: %v", err)
			continue
		}

		devID, err := uuid.Parse(payload.DeviceID)
		if err != nil {
			log.Printf("Ingest worker: invalid device ID %q: %v", payload.DeviceID, err)
			continue
		}

		if err := s.processEvents(devID, payload.UserID, payload.Platform, payload.Events); err != nil {
			log.Printf("Ingest worker: failed to process %d events for user %s: %v",
				len(payload.Events), payload.UserID, err)
		} else {
			log.Printf("Ingest worker: processed %d events for user %s", len(payload.Events), payload.UserID)
		}
	}
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

// TimelinePage is the paginated response for GetTimeline.
type TimelinePage struct {
	Events     []models.ActivityEvent `json:"events"`
	NextCursor string                 `json:"next_cursor,omitempty"`
}

// GetTimeline returns up to limit events for the given date, starting after cursor.
// Pass an empty cursor to retrieve the first page. Pass a non-empty appName to filter by app.
func (s *EventsService) GetTimeline(userID, date, cursor, appName string, limit int) (*TimelinePage, error) {
	loc := s.userLocation(userID)
	parsedDate, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := s.db.Where(
		"user_id = ? AND start_time >= ? AND start_time < ? AND is_private = false",
		userID, startOfDay, endOfDay,
	)

	// Apply cursor: only return events strictly after the cursor timestamp.
	if cursor != "" {
		if cursorTime, cerr := decodeCursor(cursor); cerr == nil {
			query = query.Where("start_time > ?", cursorTime)
		}
	}

	// Optional app name filter.
	if appName != "" {
		query = query.Where("app_name = ?", appName)
	}

	// Fetch one extra record to detect whether a next page exists.
	var events []models.ActivityEvent
	err = query.
		Preload("App.Category").
		Preload("Device").
		Order("start_time asc").
		Limit(limit + 1).
		Find(&events).Error
	if err != nil {
		return nil, errors.New("failed to fetch timeline")
	}

	page := &TimelinePage{}
	if len(events) > limit {
		page.NextCursor = encodeCursor(events[limit-1].StartTime)
		page.Events = events[:limit]
	} else {
		page.Events = events
	}
	if page.Events == nil {
		page.Events = []models.ActivityEvent{}
	}
	return page, nil
}

// encodeCursor encodes a time.Time as an opaque base64 cursor string.
func encodeCursor(t time.Time) string {
	return base64.StdEncoding.EncodeToString([]byte(t.UTC().Format(time.RFC3339Nano)))
}

// decodeCursor decodes a cursor string back into a time.Time.
func decodeCursor(cursor string) (time.Time, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, string(data))
}

func (s *EventsService) EditEvent(userID string, eventID string, params UpdateEventParams) error {
	var event models.ActivityEvent
	if err := s.db.Where("id = ? AND user_id = ?", eventID, userID).First(&event).Error; err != nil {
		return utils.NewNotFoundError("event not found or unauthorized")
	}

	if params.IsPrivate != nil {
		if err := s.db.Model(&event).Update("is_private", *params.IsPrivate).Error; err != nil {
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
		return utils.NewNotFoundError("event not found or unauthorized")
	}
	return nil
}

// userLocation loads the user's IANA timezone from the DB, falls back to UTC.
func (s *EventsService) userLocation(userID string) *time.Location {
	var user models.User
	if err := s.db.Select("timezone").Where("id = ?", userID).First(&user).Error; err != nil {
		return time.UTC
	}
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}
