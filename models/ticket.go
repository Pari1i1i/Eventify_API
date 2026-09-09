package models

import "time"

type TicketStatus string

const (
	TicketStatusPending   TicketStatus = "pending"
	TicketStatusCheckedIn TicketStatus = "checked_in"
)

type Ticket struct {
	ID          uint64       `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	OrderItemID uint64       `gorm:"type:bigint unsigned;not null;index:tickets_order_item_id_foreign" json:"order_item_id"`
	Code        string       `gorm:"type:varchar(64);not null;uniqueIndex:tickets_code_unique" json:"code"`
	Status      TicketStatus `gorm:"type:enum('pending','checked_in');not null;default:'pending';index:tickets_status_index" json:"status"`
	CheckedInAt *time.Time   `gorm:"type:datetime" json:"checked_in_at"`
	CheckedInBy *uint64      `gorm:"type:bigint unsigned;index:tickets_checked_in_by_foreign" json:"checked_in_by"`
	CreatedAt   time.Time    `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"type:timestamp;default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`

	OrderItem *OrderItem `gorm:"foreignKey:OrderItemID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"order_item,omitempty"`
	Checker   *User      `gorm:"foreignKey:CheckedInBy;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"checker,omitempty"`
}

func (Ticket) TableName() string {
	return "tickets"
}
