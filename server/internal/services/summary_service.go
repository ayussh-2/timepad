package services

import (
	"errors"
	"sort"
	"time"

	"github.com/ayussh-2/timepad/internal/models"
	"github.com/google/uuid"
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
	AppID     string           `json:"app_id"`
	AppName   string           `json:"app_name"`
	Category  *models.Category `json:"category,omitempty"`
	TotalSecs int              `json:"total_secs"`
	Platforms []string         `json:"platforms,omitempty"`
	Icon      string           `json:"icon,omitempty"`
	IsSystem  bool             `json:"is_system"`
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
	loc := s.userLocation(userID)
	parsedDate, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var events []models.ActivityEvent
	err = s.db.Where("user_id = ? AND start_time >= ? AND start_time < ? AND is_private = false", userID, startOfDay, endOfDay).
		Preload("App.Category").
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
	appIDLookup := make(map[string]*uuid.UUID)
	appCategoryMap := make(map[string]*models.Category)
	appIconMap := make(map[string]string)
	appSystemMap := make(map[string]bool)
	appPlatformMap := make(map[string]map[string]bool)
	deviceUsageMap := make(map[string]int)
	deviceDetailsMap := make(map[string]models.Device)

	hourUsageMap := make(map[int]int)

	for _, e := range events {
		if e.IsIdle {
			summary.TotalIdleSecs += e.DurationSecs
			continue
		}

		summary.TotalActiveSecs += e.DurationSecs

		if e.AppID != nil && e.App.CategoryID != nil && e.App.Category != nil && e.App.Category.IsProductive != nil {
			if *e.App.Category.IsProductive {
				summary.ProductiveSecs += e.DurationSecs
			} else {
				summary.DistractionSecs += e.DurationSecs
			}
		}

		appUsageMap[e.AppName] += e.DurationSecs
		if e.AppID != nil {
			appIDLookup[e.AppName] = e.AppID
			if e.App.CategoryID != nil {
				appCategoryMap[e.AppName] = e.App.Category
				appIconMap[e.AppName] = e.App.Icon
			}
			if e.App.IsSystem {
				appSystemMap[e.AppName] = true
			}
		}
		if appPlatformMap[e.AppName] == nil {
			appPlatformMap[e.AppName] = make(map[string]bool)
		}
		appPlatformMap[e.AppName][e.Device.Platform] = true

		deviceUsageMap[e.Device.ID.String()] += e.DurationSecs
		deviceDetailsMap[e.Device.ID.String()] = e.Device

		startHour := e.StartTime.In(loc).Hour()
		hourUsageMap[startHour] += e.DurationSecs
	}

	for app, secs := range appUsageMap {
		var platforms []string
		for p := range appPlatformMap[app] {
			platforms = append(platforms, p)
		}
		appID := ""
		if id := appIDLookup[app]; id != nil {
			appID = id.String()
		}
		summary.TopApps = append(summary.TopApps, AppUsage{
			AppID:     appID,
			AppName:   app,
			TotalSecs: secs,
			Category:  appCategoryMap[app],
			Platforms: platforms,
			Icon:      appIconMap[app],
			IsSystem:  appSystemMap[app],
		})
	}

	sort.Slice(summary.TopApps, func(i, j int) bool {
		return summary.TopApps[i].TotalSecs > summary.TopApps[j].TotalSecs
	})

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
	loc := s.userLocation(userID)
	parsedDate, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	// Calculate start (Monday) and end of the week
	offset := int(time.Monday - parsedDate.Weekday())
	if offset > 0 {
		offset = -6
	}

	monday := parsedDate.AddDate(0, 0, offset)
	startOfWeek := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, loc)
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	var events []models.ActivityEvent
	err = s.db.Where("user_id = ? AND start_time >= ? AND start_time < ? AND is_private = false", userID, startOfWeek, endOfWeek).
		Preload("App.Category").
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
	weeklyAppIDLookup := make(map[string]*uuid.UUID)
	appCategoryMap := make(map[string]*models.Category)
	weeklyAppIconMap := make(map[string]string)
	weeklyAppSystemMap := make(map[string]bool)
	weeklyAppPlatformMap := make(map[string]map[string]bool)

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

		if e.AppID != nil && e.App.CategoryID != nil && e.App.Category != nil && e.App.Category.IsProductive != nil {
			if *e.App.Category.IsProductive {
				summary.ProductiveSecs += e.DurationSecs
				summary.DailyBreakdown[dayIndex].ProductiveSecs += e.DurationSecs
			} else {
				summary.DistractionSecs += e.DurationSecs
				summary.DailyBreakdown[dayIndex].DistractionSecs += e.DurationSecs
			}
		}

		// Weekly Aggregations
		appUsageMap[e.AppName] += e.DurationSecs
		if e.AppID != nil {
			weeklyAppIDLookup[e.AppName] = e.AppID
			if e.App.CategoryID != nil {
				appCategoryMap[e.AppName] = e.App.Category
				weeklyAppIconMap[e.AppName] = e.App.Icon
			}
			if e.App.IsSystem {
				weeklyAppSystemMap[e.AppName] = true
			}
		}
		if weeklyAppPlatformMap[e.AppName] == nil {
			weeklyAppPlatformMap[e.AppName] = make(map[string]bool)
		}
		weeklyAppPlatformMap[e.AppName][e.Device.Platform] = true

		// Daily Aggregations
		dailyAppUsageMaps[dayIndex][e.AppName] += e.DurationSecs
		dailyDeviceUsageMaps[dayIndex][e.Device.ID.String()] += e.DurationSecs
		deviceDetailsMap[e.Device.ID.String()] = e.Device
		dailyHourUsageMaps[dayIndex][e.StartTime.In(loc).Hour()] += e.DurationSecs
	}

	for app, secs := range appUsageMap {
		var wplatforms []string
		for p := range weeklyAppPlatformMap[app] {
			wplatforms = append(wplatforms, p)
		}
		weeklyAppID := ""
		if id := weeklyAppIDLookup[app]; id != nil {
			weeklyAppID = id.String()
		}
		summary.TopApps = append(summary.TopApps, AppUsage{
			AppID:     weeklyAppID,
			AppName:   app,
			TotalSecs: secs,
			Category:  appCategoryMap[app],
			Platforms: wplatforms,
			Icon:      weeklyAppIconMap[app],
			IsSystem:  weeklyAppSystemMap[app],
		})
	}

	sort.Slice(summary.TopApps, func(i, j int) bool {
		return summary.TopApps[i].TotalSecs > summary.TopApps[j].TotalSecs
	})

	for i := 0; i < 7; i++ {
		for app, secs := range dailyAppUsageMaps[i] {
			dayAppID := ""
			if id := weeklyAppIDLookup[app]; id != nil {
				dayAppID = id.String()
			}
			summary.DailyBreakdown[i].TopApps = append(summary.DailyBreakdown[i].TopApps, AppUsage{
				AppID:     dayAppID,
				AppName:   app,
				TotalSecs: secs,
				Category:  appCategoryMap[app],
				Icon:      weeklyAppIconMap[app],
				IsSystem:  weeklyAppSystemMap[app],
			})
		}

		sort.Slice(summary.DailyBreakdown[i].TopApps, func(a, b int) bool {
			return summary.DailyBreakdown[i].TopApps[a].TotalSecs > summary.DailyBreakdown[i].TopApps[b].TotalSecs
		})

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

// userLocation loads the user's IANA timezone from the DB.
// Falls back to UTC on any error.
func (s *SummaryService) userLocation(userID string) *time.Location {
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
