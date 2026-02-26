package services

import (
	"errors"

	"github.com/ayussh-2/timepad/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DevicesService struct {
	db *gorm.DB
}

func NewDevicesService(db *gorm.DB) *DevicesService {
	return &DevicesService{
		db: db,
	}
}

func (s *DevicesService) GetDevices(userID string) ([]models.Device, error) {
	var devices []models.Device

	err := s.db.Where("user_id = ?", userID).Find(&devices).Error
	if err != nil {
		return nil, errors.New("failed to fetch devices")
	}

	return devices, nil
}

type RegisterDeviceParams struct {
	Name     string `json:"name" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=android windows browser"`
}

func (s *DevicesService) RegisterDevice(userID string, params RegisterDeviceParams) (*models.Device, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	deviceKey := params.Platform + "-" + uuid.New().String()

	device := models.Device{
		UserID:    parsedUserID,
		Name:      params.Name,
		Platform:  params.Platform,
		DeviceKey: deviceKey,
	}

	if err := s.db.Create(&device).Error; err != nil {
		return nil, errors.New("failed to register device")
	}

	return &device, nil
}

func (s *DevicesService) DeleteDevice(userID string, deviceID string) error {
	result := s.db.Where("id = ? AND user_id = ?", deviceID, userID).Delete(&models.Device{})
	if result.Error != nil {
		return errors.New("failed to delete device")
	}
	if result.RowsAffected == 0 {
		return errors.New("device not found or unauthorized")
	}
	return nil
}
