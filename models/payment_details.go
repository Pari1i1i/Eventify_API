package models

import "time"

type PaymentDetails struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	OrderID          uint64     `gorm:"type:bigint unsigned;not null;uniqueIndex:payment_details_order_id_unique" json:"order_id"`
	PaymentMethod    *string    `gorm:"type:varchar(50)" json:"payment_method"`
	GatewayReference *string    `gorm:"type:varchar(100)" json:"gateway_reference"`
	VaNumber         *string    `gorm:"type:varchar(50)" json:"va_number"`
	QrURL            *string    `gorm:"type:text" json:"qr_url"`
	PaidAt           *time.Time `gorm:"type:datetime" json:"paid_at"`
	CreatedAt        time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`

	Order *Order `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"order,omitempty"`
}

func (PaymentDetails) TableName() string {
	return "payment_details"
}
