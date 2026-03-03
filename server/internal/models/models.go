package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Email        string    `json:"email" gorm:"unique;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	DisplayName  string    `json:"display_name"`
	Timezone     string    `json:"timezone" gorm:"default:'UTC';not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Devices  []Device    `json:"-"`
	Settings UserSetting `json:"-"`
}

type Device struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID  `json:"user_id" gorm:"not null"`
	User       User       `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	Name       string     `json:"name" gorm:"not null"`
	Platform   string     `json:"platform" gorm:"check:platform IN ('android', 'windows', 'browser');not null"`
	DeviceKey  string     `json:"-" gorm:"unique;not null"`
	LastSeenAt *time.Time `json:"last_seen_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Category struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID       *uuid.UUID     `json:"user_id"`
	User         User           `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	Name         string         `json:"name" gorm:"not null"`
	Color        string         `json:"color" gorm:"default:'#6B7280';not null"`
	Icon         string         `json:"icon"`
	IsSystem     bool           `json:"is_system" gorm:"default:false"`
	IsProductive *bool          `json:"is_productive" gorm:"default:null"`
	Rules        datatypes.JSON `json:"rules" gorm:"type:jsonb;default:'[]'"`
}

// App represents a unique application per user, across all devices and platforms.
// Category assignment lives here — not on individual events.
type App struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID      `json:"user_id" gorm:"not null;uniqueIndex:idx_apps_user_name"`
	User        User           `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	Name        string         `json:"name" gorm:"not null;uniqueIndex:idx_apps_user_name"`
	Icon        string         `json:"icon"`
	Platforms   pq.StringArray `json:"platforms" gorm:"type:text[];default:'{}'"`
	IsSystem    bool           `json:"is_system" gorm:"default:false"`
	CategoryID  *uuid.UUID     `json:"category_id" gorm:"index"`
	Category    *Category      `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	FirstSeenAt time.Time      `json:"first_seen_at"`
	LastSeenAt  time.Time      `json:"last_seen_at"`
}

type ActivityEvent struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID      `json:"user_id" gorm:"not null;index:idx_events_user_start,priority:1"`
	User         User           `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	DeviceID     uuid.UUID      `json:"device_id" gorm:"not null;index:idx_events_device"`
	Device       Device         `json:"device,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	AppID        *uuid.UUID     `json:"app_id" gorm:"index:idx_events_app"`
	App          App            `json:"app,omitempty" gorm:"foreignKey:AppID"`
	AppName      string         `json:"app_name" gorm:"not null;index:idx_events_app_name"`
	WindowTitle  string         `json:"window_title"`
	Url          string         `json:"url"`
	StartTime    time.Time      `json:"start_time" gorm:"not null;index:idx_events_user_start,priority:2,sort:desc"`
	EndTime      time.Time      `json:"end_time" gorm:"not null"`
	DurationSecs int            `json:"duration_secs" gorm:"not null"`
	IsIdle       bool           `json:"is_idle" gorm:"default:false"`
	IsPrivate    bool           `json:"is_private" gorm:"default:false"`
	RawMeta      datatypes.JSON `json:"raw_meta,omitempty" gorm:"type:jsonb"`
	CreatedAt    time.Time      `json:"created_at"`
}

type UserSetting struct {
	UserID            uuid.UUID      `json:"user_id" gorm:"type:uuid;primaryKey"`
	User              *User          `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	ExcludedApps      pq.StringArray `json:"excluded_apps" gorm:"type:text[];default:'{}'"`
	ExcludedUrls      pq.StringArray `json:"excluded_urls" gorm:"type:text[];default:'{}'"`
	IdleThreshold     int            `json:"idle_threshold" gorm:"default:300"`
	TrackingEnabled   bool           `json:"tracking_enabled" gorm:"default:true"`
	DataRetentionDays int            `json:"data_retention_days" gorm:"default:365"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// BeforeCreate hooks auto-generate UUID primary keys when the ID is the zero value.
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (d *Device) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (c *Category) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (a *App) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (e *ActivityEvent) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
