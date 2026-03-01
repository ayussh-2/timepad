package services

import (
	"errors"

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

type UpdateSettingsParams struct {
	ExcludedApps      *[]string `json:"excluded_apps"`
	ExcludedUrls      *[]string `json:"excluded_urls"`
	IdleThreshold     *int      `json:"idle_threshold"`
	TrackingEnabled   *bool     `json:"tracking_enabled"`
	DataRetentionDays *int      `json:"data_retention_days"`
}

func (s *SettingsService) GetSettings(userID string) (*models.UserSetting, error) {
	var settings models.UserSetting

	err := s.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No row yet — return defaults without persisting.
			parsedID, parseErr := uuid.Parse(userID)
			if parseErr != nil {
				return nil, errors.New("invalid user ID")
			}
			return &models.UserSetting{
				UserID:            parsedID,
				ExcludedApps:      pq.StringArray{},
				ExcludedUrls:      pq.StringArray{},
				IdleThreshold:     300,
				TrackingEnabled:   true,
				DataRetentionDays: 365,
			}, nil
		}
		return nil, errors.New("failed to fetch settings")
	}

	return &settings, nil
}

func (s *SettingsService) UpdateSettings(userID string, params UpdateSettingsParams) error {
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
