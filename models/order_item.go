package models

import "time"

type OrderItem struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	OrderID      uint64    `gorm:"type:bigint unsigned;not null;index:order_items_order_id_foreign" json:"order_id"`
	TicketTierID uint64    `gorm:"type:bigint unsigned;not null;index:order_items_ticket_tier_id_foreign" json:"ticket_tier_id"`
	Quantity     uint16    `gorm:"type:smallint unsigned;not null;default:1" json:"quantity"`
	Price        float64   `gorm:"type:decimal(12,2);not null;default:0.00" json:"price"`
	Subtotal     float64   `gorm:"type:decimal(12,2);not null;default:0.00" json:"subtotal"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`

	Order      *Order      `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"order,omitempty"`
	TicketTier *TicketTier `gorm:"foreignKey:TicketTierID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE" json:"ticket_tier,omitempty"`
	Tickets    []Ticket    `gorm:"foreignKey:OrderItemID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"tickets,omitempty"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
