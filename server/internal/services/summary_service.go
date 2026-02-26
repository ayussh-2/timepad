package services

import (
	"errors"
	"time"

	"github.com/ayussh-2/timepad/internal/models"
	"gorm.io/gorm"
)

type SummaryService struct {
	db *gorm.DB
}

func NewSummaryService(db *gorm.DB) *SummaryService {
	return &SummaryService{
		db: db,
	}
}

type AppUsage struct {
	AppName   string           `json:"app_name"`
	Category  *models.Category `json:"category,omitempty"`
	TotalSecs int              `json:"total_secs"`
}

type DeviceUsage struct {
	DeviceName string `json:"device_name"`
	Platform   string `json:"platform"`
	TotalSecs  int    `json:"total_secs"`
}

type DailySummary struct {
	Date            string        `json:"date"`
	TotalActiveSecs int           `json:"total_active_secs"`
	TotalIdleSecs   int           `json:"total_idle_secs"`
	ProductiveSecs  int           `json:"productive_secs"`
	DistractionSecs int           `json:"distraction_secs"`
	TopApps         []AppUsage    `json:"top_apps"`
	PeakHour        int           `json:"peak_hour"`
	DeviceBreakdown []DeviceUsage `json:"device_breakdown"`
}

func (s *SummaryService) GetDailySummary(userID string, date string) (*DailySummary, error) {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	startOfDay := parsedDate.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var events []models.ActivityEvent
	err = s.db.Where("user_id = ? AND start_time >= ? AND start_time < ?", userID, startOfDay, endOfDay).
		Preload("Category").
		Preload("Device").
		Find(&events).Error
	if err != nil {
		return nil, errors.New("failed to fetch daily events")
	}

	summary := &DailySummary{
		Date:            date,
		TopApps:         []AppUsage{},
		DeviceBreakdown: []DeviceUsage{},
	}

	appUsageMap := make(map[string]int)
	appCategoryMap := make(map[string]*models.Category)
	deviceUsageMap := make(map[string]int)
	deviceDetailsMap := make(map[string]models.Device)

	hourUsageMap := make(map[int]int)

	for _, e := range events {
		if e.IsIdle {
			summary.TotalIdleSecs += e.DurationSecs
			continue
		}

		summary.TotalActiveSecs += e.DurationSecs

		if e.CategoryID != nil && e.Category.IsProductive != nil {
			if *e.Category.IsProductive {
				summary.ProductiveSecs += e.DurationSecs
			} else {
				summary.DistractionSecs += e.DurationSecs
			}
		}

		appUsageMap[e.AppName] += e.DurationSecs
		if e.CategoryID != nil {
			appCategoryMap[e.AppName] = &e.Category
		}

		deviceUsageMap[e.Device.ID.String()] += e.DurationSecs
		deviceDetailsMap[e.Device.ID.String()] = e.Device

		startHour := e.StartTime.Hour()
		hourUsageMap[startHour] += e.DurationSecs
	}

	for app, secs := range appUsageMap {
		summary.TopApps = append(summary.TopApps, AppUsage{
			AppName:   app,
			TotalSecs: secs,
			Category:  appCategoryMap[app],
		})
	}

	for devID, secs := range deviceUsageMap {
		dev := deviceDetailsMap[devID]
		summary.DeviceBreakdown = append(summary.DeviceBreakdown, DeviceUsage{
			DeviceName: dev.Name,
			Platform:   dev.Platform,
			TotalSecs:  secs,
		})
	}

	maxHourSecs := -1
	for hour, secs := range hourUsageMap {
		if secs > maxHourSecs {
			maxHourSecs = secs
			summary.PeakHour = hour
		}
	}

	return summary, nil
}

type WeeklySummary struct {
	StartDate       string         `json:"start_date"`
	EndDate         string         `json:"end_date"`
	TotalActiveSecs int            `json:"total_active_secs"`
	TotalIdleSecs   int            `json:"total_idle_secs"`
	ProductiveSecs  int            `json:"productive_secs"`
	DistractionSecs int            `json:"distraction_secs"`
	TopApps         []AppUsage     `json:"top_apps"`
	DailyBreakdown  []DailySummary `json:"daily_breakdown"`
}

func (s *SummaryService) GetWeeklySummary(userID string, date string) (*WeeklySummary, error) {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	// Calculate start (Monday) and end of the week
	offset := int(time.Monday - parsedDate.Weekday())
	if offset > 0 {
		offset = -6
	}

	startOfWeek := parsedDate.AddDate(0, 0, offset).Truncate(24 * time.Hour)
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	var events []models.ActivityEvent
	err = s.db.Where("user_id = ? AND start_time >= ? AND start_time < ?", userID, startOfWeek, endOfWeek).
		Preload("Category").
		Preload("Device").
		Find(&events).Error
	if err != nil {
		return nil, errors.New("failed to fetch weekly events")
	}

	summary := &WeeklySummary{
		StartDate:      startOfWeek.Format("2006-01-02"),
		EndDate:        endOfWeek.AddDate(0, 0, -1).Format("2006-01-02"),
		TopApps:        []AppUsage{},
		DailyBreakdown: make([]DailySummary, 7),
	}

	// Initialize DailyBreakdown array
	for i := 0; i < 7; i++ {
		summary.DailyBreakdown[i] = DailySummary{
			Date:            startOfWeek.AddDate(0, 0, i).Format("2006-01-02"),
			TopApps:         []AppUsage{},
			DeviceBreakdown: []DeviceUsage{},
		}
	}

	appUsageMap := make(map[string]int)
	appCategoryMap := make(map[string]*models.Category)

	// Map to hold daily usages [dayIndex] -> maps
	dailyAppUsageMaps := make([]map[string]int, 7)
	dailyDeviceUsageMaps := make([]map[string]int, 7)
	deviceDetailsMap := make(map[string]models.Device)
	dailyHourUsageMaps := make([]map[int]int, 7)

	for i := 0; i < 7; i++ {
		dailyAppUsageMaps[i] = make(map[string]int)
		dailyDeviceUsageMaps[i] = make(map[string]int)
		dailyHourUsageMaps[i] = make(map[int]int)
	}

	for _, e := range events {
		dayIndex := int(e.StartTime.Sub(startOfWeek).Hours() / 24)
		if dayIndex < 0 || dayIndex >= 7 {
			continue
		}

		if e.IsIdle {
			summary.TotalIdleSecs += e.DurationSecs
			summary.DailyBreakdown[dayIndex].TotalIdleSecs += e.DurationSecs
			continue
		}

		summary.TotalActiveSecs += e.DurationSecs
		summary.DailyBreakdown[dayIndex].TotalActiveSecs += e.DurationSecs

		if e.CategoryID != nil && e.Category.IsProductive != nil {
			if *e.Category.IsProductive {
				summary.ProductiveSecs += e.DurationSecs
				summary.DailyBreakdown[dayIndex].ProductiveSecs += e.DurationSecs
			} else {
				summary.DistractionSecs += e.DurationSecs
				summary.DailyBreakdown[dayIndex].DistractionSecs += e.DurationSecs
			}
		}

		// Weekly Aggregations
		appUsageMap[e.AppName] += e.DurationSecs
		if e.CategoryID != nil {
			appCategoryMap[e.AppName] = &e.Category
		}

		// Daily Aggregations
		dailyAppUsageMaps[dayIndex][e.AppName] += e.DurationSecs
		dailyDeviceUsageMaps[dayIndex][e.Device.ID.String()] += e.DurationSecs
		deviceDetailsMap[e.Device.ID.String()] = e.Device
		dailyHourUsageMaps[dayIndex][e.StartTime.Hour()] += e.DurationSecs
	}

	for app, secs := range appUsageMap {
		summary.TopApps = append(summary.TopApps, AppUsage{
			AppName:   app,
			TotalSecs: secs,
			Category:  appCategoryMap[app],
		})
	}

	for i := 0; i < 7; i++ {
		for app, secs := range dailyAppUsageMaps[i] {
			summary.DailyBreakdown[i].TopApps = append(summary.DailyBreakdown[i].TopApps, AppUsage{
				AppName:   app,
				TotalSecs: secs,
				Category:  appCategoryMap[app],
			})
		}

		for devID, secs := range dailyDeviceUsageMaps[i] {
			dev := deviceDetailsMap[devID]
			summary.DailyBreakdown[i].DeviceBreakdown = append(summary.DailyBreakdown[i].DeviceBreakdown, DeviceUsage{
				DeviceName: dev.Name,
				Platform:   dev.Platform,
				TotalSecs:  secs,
			})
		}

		maxHourSecs := -1
		for hour, secs := range dailyHourUsageMaps[i] {
			if secs > maxHourSecs {
				maxHourSecs = secs
				summary.DailyBreakdown[i].PeakHour = hour
			}
		}
	}

	return summary, nil
}
