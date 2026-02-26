package services

import (
	"fmt"
	"log"
	"time"

	"github.com/ayussh-2/timepad/internal/models"
	"gorm.io/gorm"
)

type PurgeService struct {
	db *gorm.DB
}

func NewPurgeService(db *gorm.DB) *PurgeService {
	return &PurgeService{db: db}
}

func (s *PurgeService) PurgeExpiredEvents() (int64, error) {
	var settings []models.UserSetting
	if err := s.db.Where("data_retention_days > 0").Find(&settings).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch user settings: %w", err)
	}

	var totalPurged int64

	for _, setting := range settings {
		cutoff := time.Now().AddDate(0, 0, -setting.DataRetentionDays)

		result := s.db.Where("user_id = ? AND created_at < ?", setting.UserID, cutoff).
			Delete(&models.ActivityEvent{})

		if result.Error != nil {
			log.Printf("Failed to purge events for user %s: %v", setting.UserID, result.Error)
			continue
		}

		if result.RowsAffected > 0 {
			log.Printf("Purged %d events for user %s (retention: %d days)",
				result.RowsAffected, setting.UserID, setting.DataRetentionDays)
		}

		totalPurged += result.RowsAffected
	}

	return totalPurged, nil
}
