package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null" json:"-"`
	DisplayName  string
	Timezone     string `gorm:"default:'UTC';not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Devices  []Device    `json:"-"`
	Settings UserSetting `json:"-"`
}

type Device struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"not null"`
	User       User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Name       string    `gorm:"not null"`
	Platform   string    `gorm:"check:platform IN ('android', 'windows', 'browser');not null"`
	DeviceKey  string    `gorm:"unique;not null"`
	LastSeenAt *time.Time
	CreatedAt  time.Time
}

type Category struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       *uuid.UUID
	User         User   `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Name         string `gorm:"not null"`
	Color        string `gorm:"default:'#6B7280';not null"`
	Icon         string
	IsSystem     bool           `gorm:"default:false"`
	IsProductive *bool          `gorm:"default:null"`
	Rules        datatypes.JSON `gorm:"type:jsonb;default:'[]'"`
}

type ActivityEvent struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID `gorm:"not null;index:idx_events_user_start,priority:1;index:idx_events_app_name,priority:1"`
	User         User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	DeviceID     uuid.UUID `gorm:"not null;index:idx_events_device"`
	Device       Device    `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	AppName      string    `gorm:"not null;index:idx_events_app_name,priority:2"`
	WindowTitle  string
	Url          string
	CategoryID   *uuid.UUID     `gorm:"index:idx_events_category"`
	Category     Category       `json:"-"`
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
	User              *User          `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	ExcludedApps      pq.StringArray `gorm:"type:text[];default:'{}'"`
	ExcludedUrls      pq.StringArray `gorm:"type:text[];default:'{}'"`
	IdleThreshold     int            `gorm:"default:300"`
	TrackingEnabled   bool           `gorm:"default:true"`
	DataRetentionDays int            `gorm:"default:365"`
	UpdatedAt         time.Time
}

// BeforeCreate hooks auto-generate UUID primary keys when the ID is the zero value.
// This makes the models portable across PostgreSQL (gen_random_uuid()) and SQLite.

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

func (a *ActivityEvent) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
