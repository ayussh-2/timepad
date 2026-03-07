package services

import (
	"errors"
	"time"

	"github.com/ayussh-2/timepad/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type SettingsService struct {
	db *gorm.DB
}

func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{
		db: db,
	}
}

// FullSettings is the combined response for GET /settings — it includes the
// UserSetting row plus the user's IANA timezone from the users table.
type FullSettings struct {
	UserID            uuid.UUID      `json:"user_id"`
	ExcludedApps      pq.StringArray `json:"excluded_apps"`
	ExcludedUrls      pq.StringArray `json:"excluded_urls"`
	IdleThreshold     int            `json:"idle_threshold"`
	TrackingEnabled   bool           `json:"tracking_enabled"`
	DataRetentionDays int            `json:"data_retention_days"`
	UpdatedAt         time.Time      `json:"updated_at"`
	Timezone          string         `json:"timezone"`
}

type UpdateSettingsParams struct {
	ExcludedApps      *[]string `json:"excluded_apps"`
	ExcludedUrls      *[]string `json:"excluded_urls"`
	IdleThreshold     *int      `json:"idle_threshold"`
	TrackingEnabled   *bool     `json:"tracking_enabled"`
	DataRetentionDays *int      `json:"data_retention_days"`
	Timezone          *string   `json:"timezone"`
}

func (s *SettingsService) GetSettings(userID string) (*FullSettings, error) {
	parsedID, parseErr := uuid.Parse(userID)
	if parseErr != nil {
		return nil, errors.New("invalid user ID")
	}

	// Load the user's IANA timezone.
	var user models.User
	tz := "UTC"
	if err := s.db.Select("timezone").Where("id = ?", userID).First(&user).Error; err == nil && user.Timezone != "" {
		tz = user.Timezone
	}

	var settings models.UserSetting
	err := s.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No row yet — return defaults without persisting.
			return &FullSettings{
				UserID:            parsedID,
				ExcludedApps:      pq.StringArray{},
				ExcludedUrls:      pq.StringArray{},
				IdleThreshold:     300,
				TrackingEnabled:   true,
				DataRetentionDays: 365,
				Timezone:          tz,
			}, nil
		}
		return nil, errors.New("failed to fetch settings")
	}

	return &FullSettings{
		UserID:            settings.UserID,
		ExcludedApps:      settings.ExcludedApps,
		ExcludedUrls:      settings.ExcludedUrls,
		IdleThreshold:     settings.IdleThreshold,
		TrackingEnabled:   settings.TrackingEnabled,
		DataRetentionDays: settings.DataRetentionDays,
		UpdatedAt:         settings.UpdatedAt,
		Timezone:          tz,
	}, nil
}

func (s *SettingsService) UpdateSettings(userID string, params UpdateSettingsParams) error {
	// Validate and apply timezone to the users table first (independent of UserSetting).
	if params.Timezone != nil {
		if _, err := time.LoadLocation(*params.Timezone); err != nil {
			return errors.New("invalid timezone: must be a valid IANA timezone name (e.g. Asia/Kolkata)")
		}
		if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("timezone", *params.Timezone).Error; err != nil {
			return errors.New("failed to update timezone")
		}
	}

	var settings models.UserSetting
	err := s.db.Where("user_id = ?", userID).First(&settings).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Settings don't exist yet – create them with defaults then apply params.
			parsedID, parseErr := uuid.Parse(userID)
			if parseErr != nil {
				return errors.New("invalid user ID")
			}
			settings = models.UserSetting{
				UserID:            parsedID,
				IdleThreshold:     300,
				TrackingEnabled:   true,
				DataRetentionDays: 365,
			}
			if params.ExcludedApps != nil {
				settings.ExcludedApps = *params.ExcludedApps
			}
			if params.ExcludedUrls != nil {
				settings.ExcludedUrls = *params.ExcludedUrls
			}
			if params.IdleThreshold != nil {
				settings.IdleThreshold = *params.IdleThreshold
			}
			if params.TrackingEnabled != nil {
				settings.TrackingEnabled = *params.TrackingEnabled
			}
			if params.DataRetentionDays != nil {
				settings.DataRetentionDays = *params.DataRetentionDays
			}
			if createErr := s.db.Create(&settings).Error; createErr != nil {
				return errors.New("failed to create settings")
			}
			return nil
		}
		return errors.New("failed to fetch prior settings")
	}

	updates := map[string]interface{}{}

	if params.ExcludedApps != nil {
		updates["excluded_apps"] = pq.StringArray(*params.ExcludedApps)
	}
	if params.ExcludedUrls != nil {
		updates["excluded_urls"] = pq.StringArray(*params.ExcludedUrls)
	}
	if params.IdleThreshold != nil {
		updates["idle_threshold"] = *params.IdleThreshold
	}
	if params.TrackingEnabled != nil {
		updates["tracking_enabled"] = *params.TrackingEnabled
	}
	if params.DataRetentionDays != nil {
		updates["data_retention_days"] = *params.DataRetentionDays
	}

	if len(updates) > 0 {
		if err := s.db.Model(&settings).Updates(updates).Error; err != nil {
			return errors.New("failed to update settings")
		}
	}

	return nil
}
