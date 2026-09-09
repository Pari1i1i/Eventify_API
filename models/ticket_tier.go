package models

import "time"

type TicketTier struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	EventID     uint64     `gorm:"type:bigint unsigned;not null;index:ticket_tiers_event_id_foreign" json:"event_id"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Price       float64    `gorm:"type:decimal(12,2);not null;default:0.00" json:"price"`
	Quota       uint32     `gorm:"type:int unsigned;not null;default:0" json:"quota"`
	Description *string    `gorm:"type:varchar(255)" json:"description"`
	IsActive    bool       `gorm:"type:tinyint(1);not null;default:1;index:ticket_tiers_active_index" json:"is_active"`
	StartSaleAt *time.Time `gorm:"type:datetime" json:"start_sale_at"`
	EndSaleAt   *time.Time `gorm:"type:datetime" json:"end_sale_at"`
	CreatedAt   time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`

	Event *Event `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"event,omitempty"`
}

func (TicketTier) TableName() string {
	return "ticket_tiers"
}
