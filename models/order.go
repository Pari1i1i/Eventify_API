package models

import "time"

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusFree      PaymentStatus = "free"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
	PaymentStatusExpired   PaymentStatus = "expired"
)

type Order struct {
	ID            uint64        `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	UserID        uint64        `gorm:"type:bigint unsigned;not null;index:orders_user_id_foreign" json:"user_id"`
	EventID       uint64        `gorm:"type:bigint unsigned;not null;index:orders_event_id_foreign" json:"event_id"`
	OrderCode     string        `gorm:"type:varchar(50);not null;uniqueIndex:orders_order_code_unique" json:"order_code"`
	TotalTickets  uint16        `gorm:"type:smallint unsigned;not null;default:1" json:"total_tickets"`
	TotalAmount   float64       `gorm:"type:decimal(12,2);not null;default:0.00" json:"total_amount"`
	PaymentStatus PaymentStatus `gorm:"type:enum('pending','paid','free','failed','cancelled','expired');not null;default:'pending';index:orders_payment_status_index" json:"payment_status"`
	ExpiresAt     *time.Time    `gorm:"type:datetime" json:"expires_at"`
	CreatedAt     time.Time     `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"type:timestamp;default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`

	User           *User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`
	Event          *Event          `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"event,omitempty"`
	OrderItems     []OrderItem     `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"order_items,omitempty"`
	PaymentDetails *PaymentDetails `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"payment_details,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}
