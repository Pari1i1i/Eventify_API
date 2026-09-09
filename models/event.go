package models

import (
	"time"

	"gorm.io/gorm"
)

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusCompleted EventStatus = "completed"
	EventStatusCancelled EventStatus = "cancelled"
)

type Event struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	CreatedBy       *uint64        `gorm:"type:bigint unsigned;index:events_created_by_foreign" json:"created_by"`
	Name            string         `gorm:"type:varchar(200);not null" json:"name"`
	Slug            string         `gorm:"type:varchar(200);not null;uniqueIndex:events_slug_unique" json:"slug"`
	Description     string         `gorm:"type:longtext;not null" json:"description"`
	TermsConditions *string        `gorm:"type:text" json:"terms_conditions"`
	Location        string         `gorm:"type:varchar(255);not null" json:"location"`
	StartAt         time.Time      `gorm:"type:datetime;not null;index:events_status_start_at_index,priority:2" json:"start_at"`
	EndAt           time.Time      `gorm:"type:datetime;not null" json:"end_at"`
	BannerPath      *string        `gorm:"type:varchar(255)" json:"banner_path"`
	Status          EventStatus    `gorm:"type:enum('draft','published','completed','cancelled');not null;default:'draft';index:events_status_start_at_index,priority:1" json:"status"`
	CreatedAt       time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"type:datetime;index" json:"deleted_at,omitempty"`

	Creator     *User        `gorm:"foreignKey:CreatedBy;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"creator,omitempty"`
	TicketTiers []TicketTier `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"ticket_tiers,omitempty"`
}

func (Event) TableName() string {
	return "events"
}
