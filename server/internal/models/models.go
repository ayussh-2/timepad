package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null"`
	DisplayName  string
	Timezone     string `gorm:"default:'UTC';not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Devices  []Device
	Settings UserSetting
}

type Device struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"not null"`
	User       User      `gorm:"constraint:OnDelete:CASCADE;"`
	Name       string    `gorm:"not null"`
	Platform   string    `gorm:"check:platform IN ('android', 'windows', 'browser');not null"`
	DeviceKey  string    `gorm:"unique;not null"`
	LastSeenAt *time.Time
	CreatedAt  time.Time
}

type Category struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID   *uuid.UUID
	User     User   `gorm:"constraint:OnDelete:CASCADE;"`
	Name     string `gorm:"not null"`
	Color    string `gorm:"default:'#6B7280';not null"`
	Icon     string
	IsSystem bool           `gorm:"default:false"`
	Rules    datatypes.JSON `gorm:"type:jsonb;default:'[]'"`
}

type ActivityEvent struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID `gorm:"not null;index:idx_events_user_start,priority:1;index:idx_events_app_name,priority:1"`
	User         User      `gorm:"constraint:OnDelete:CASCADE;"`
	DeviceID     uuid.UUID `gorm:"not null;index:idx_events_device"`
	Device       Device    `gorm:"constraint:OnDelete:CASCADE;"`
	AppName      string    `gorm:"not null;index:idx_events_app_name,priority:2"`
	WindowTitle  string
	Url          string
	CategoryID   *uuid.UUID `gorm:"index:idx_events_category"`
	Category     Category
	StartTime    time.Time      `gorm:"not null;index:idx_events_user_start,priority:2,sort:desc"`
	EndTime      time.Time      `gorm:"not null"`
	DurationSecs int            `gorm:"not null"`
	IsIdle       bool           `gorm:"default:false"`
	IsPrivate    bool           `gorm:"default:false"`
	RawMeta      datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt    time.Time
}

type UserSetting struct {
	UserID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	User              *User          `gorm:"constraint:OnDelete:CASCADE;"`
	ExcludedApps      pq.StringArray `gorm:"type:text[];default:'{}'"`
	ExcludedUrls      pq.StringArray `gorm:"type:text[];default:'{}'"`
	IdleThreshold     int            `gorm:"default:300"`
	TrackingEnabled   bool           `gorm:"default:true"`
	DataRetentionDays int            `gorm:"default:365"`
	UpdatedAt         time.Time
}
