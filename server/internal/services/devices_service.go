package services

import (
	"errors"

	"github.com/ayussh-2/timepad/internal/models"
	"github.com/ayussh-2/timepad/internal/utils"
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

type RegisterDeviceResponse struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	Name       string  `json:"name"`
	Platform   string  `json:"platform"`
	DeviceKey  string  `json:"device_key"`
	LastSeenAt *string `json:"last_seen_at"`
	CreatedAt  string  `json:"created_at"`
}

func (s *DevicesService) RegisterDevice(userID string, params RegisterDeviceParams) (*RegisterDeviceResponse, error) {
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

	var lastSeen *string
	if device.LastSeenAt != nil {
		s := device.LastSeenAt.Format("2006-01-02T15:04:05Z07:00")
		lastSeen = &s
	}
	return &RegisterDeviceResponse{
		ID:         device.ID.String(),
		UserID:     device.UserID.String(),
		Name:       device.Name,
		Platform:   device.Platform,
		DeviceKey:  device.DeviceKey,
		LastSeenAt: lastSeen,
		CreatedAt:  device.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

type RenameDeviceParams struct {
	Name string `json:"name" binding:"required"`
}

func (s *DevicesService) RenameDevice(userID string, deviceID string, params RenameDeviceParams) (*models.Device, error) {
	var device models.Device
	result := s.db.Where("id = ? AND user_id = ?", deviceID, userID).First(&device)
	if result.Error != nil {
		return nil, utils.NewNotFoundError("device not found or unauthorized")
	}
	device.Name = params.Name
	if err := s.db.Save(&device).Error; err != nil {
		return nil, errors.New("failed to rename device")
	}
	return &device, nil
}

func (s *DevicesService) DeleteDevice(userID string, deviceID string) error {
	result := s.db.Where("id = ? AND user_id = ?", deviceID, userID).Delete(&models.Device{})
	if result.Error != nil {
		return errors.New("failed to delete device")
	}
	if result.RowsAffected == 0 {
		return utils.NewNotFoundError("device not found or unauthorized")
	}
	return nil
}
