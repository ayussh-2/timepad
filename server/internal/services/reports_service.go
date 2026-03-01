package services

import (
	"errors"
	"time"

	"github.com/ayussh-2/timepad/internal/models"
	"gorm.io/gorm"
)

type ReportsService struct {
	db *gorm.DB
}

func NewReportsService(db *gorm.DB) *ReportsService {
	return &ReportsService{
		db: db,
	}
}

type ReportParams struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

type ReportData struct {
	TotalActiveSecs  int            `json:"total_active_secs"`
	TotalIdleSecs    int            `json:"total_idle_secs"`
	CategoryUsage    map[string]int `json:"category_usage"`
	AppUsage         map[string]int `json:"app_usage"`
	DeviceUsage      map[string]int `json:"device_usage"`
	DailyActiveTrend map[string]int `json:"daily_active_trend"`
}

func (s *ReportsService) GetReports(userID string, params ReportParams) (*ReportData, error) {
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	loc := s.userLocation(userID)
	query := s.db.Where("user_id = ? AND is_private = false", userID)

	if params.StartDate != "" {
		if startDate, err := time.ParseInLocation("2006-01-02", params.StartDate, loc); err == nil {
			query = query.Where("start_time >= ?", startDate)
		}
	}

	if params.EndDate != "" {
		if endDate, err := time.ParseInLocation("2006-01-02", params.EndDate, loc); err == nil {
			query = query.Where("start_time < ?", endDate.AddDate(0, 0, 1))
		}
	}

	var events []models.ActivityEvent
	err := query.Preload("Category").Preload("Device").Find(&events).Error
	if err != nil {
		return nil, errors.New("failed to fetch events for report")
	}

	report := &ReportData{
		CategoryUsage:    make(map[string]int),
		AppUsage:         make(map[string]int),
		DeviceUsage:      make(map[string]int),
		DailyActiveTrend: make(map[string]int),
	}

	for _, e := range events {
		dateKey := e.StartTime.Format("2006-01-02")

		if e.IsIdle {
			report.TotalIdleSecs += e.DurationSecs
			continue
		}

		report.TotalActiveSecs += e.DurationSecs
		report.DailyActiveTrend[dateKey] += e.DurationSecs
		report.AppUsage[e.AppName] += e.DurationSecs
		report.DeviceUsage[e.Device.Name] += e.DurationSecs

		categoryName := "Uncategorized"
		if e.CategoryID != nil && e.Category.Name != "" {
			categoryName = e.Category.Name
		}
		report.CategoryUsage[categoryName] += e.DurationSecs
	}

	return report, nil
}

// userLocation loads the user's IANA timezone from the DB.
// Falls back to UTC on any error.
func (s *ReportsService) userLocation(userID string) *time.Location {
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
