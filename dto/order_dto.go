package dto

import "time"

type CreateOrderRequest struct {
	EventID       uint64           `json:"event_id" binding:"required" example:"1"`
	PaymentMethod *string          `json:"payment_method" binding:"omitempty" example:"qris"`
	Items         []OrderItemInput `json:"items" binding:"required,min=1,dive"`
}

type OrderItemInput struct {
	TicketTierID uint64 `json:"ticket_tier_id" binding:"required" example:"1"`
	Quantity     uint16 `json:"quantity" binding:"required,min=1,max=10" example:"2"`
}

type OrderResponse struct {
	ID             uint64                 `json:"id"`
	OrderCode      string                 `json:"order_code"`
	UserID         uint64                 `json:"user_id"`
	EventID        uint64                 `json:"event_id"`
	EventName      string                 `json:"event_name,omitempty"`
	TotalTickets   uint16                 `json:"total_tickets"`
	TotalAmount    float64                `json:"total_amount"`
	PaymentStatus  string                 `json:"payment_status"`
	ExpiresAt      *time.Time             `json:"expires_at"`
	CreatedAt      time.Time              `json:"created_at"`
	Items          []OrderItemResponse    `json:"items,omitempty"`
	PaymentDetails *PaymentDetailResponse `json:"payment_details,omitempty"`
}

type OrderItemResponse struct {
	ID             uint64           `json:"id"`
	TicketTierID   uint64           `json:"ticket_tier_id"`
	TicketTierName string           `json:"ticket_tier_name"`
	Quantity       uint16           `json:"quantity"`
	Price          float64          `json:"price"`
	Subtotal       float64          `json:"subtotal"`
	Tickets        []TicketResponse `json:"tickets,omitempty"`
}

type PaymentDetailResponse struct {
	PaymentMethod    *string    `json:"payment_method,omitempty"`
	GatewayReference *string    `json:"gateway_reference,omitempty"`
	VaNumber         *string    `json:"va_number,omitempty"`
	QrURL            *string    `json:"qr_code_url,omitempty"`
	SimulationKey    *string    `json:"simulation_key,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
}

type PaymentWebhookRequest struct {
	OrderCode         string  `json:"order_code"`
	OrderID           *string `json:"order_id"`
	Status            *string `json:"status"`
	TransactionStatus *string `json:"transaction_status"`
	TransactionID     *string `json:"transaction_id"`
	PaymentType       *string `json:"payment_type"`
	PaymentMethod     *string `json:"payment_method"`
	ReferenceID       *string `json:"reference_id"`
	// Midtrans notification fields
	StatusCode   *string `json:"status_code"`
	GrossAmount  *string `json:"gross_amount"`
	SignatureKey *string `json:"signature_key"`
	// Xendit compatibility fields
	CallbackToken *string `json:"callback_token"`
}

type TicketResponse struct {
	ID             uint64     `json:"id"`
	EventID        uint64     `json:"event_id,omitempty"`
	Code           string     `json:"code"`
	Status         string     `json:"status"`
	CheckedInAt    *time.Time `json:"checked_in_at"`
	EventName      string     `json:"event_name,omitempty"`
	TicketTierName string     `json:"ticket_tier_name,omitempty"`
	CustomerName   string     `json:"customer_name,omitempty"`
	StartAt        *time.Time `json:"start_at,omitempty"`
	EndAt          *time.Time `json:"end_at,omitempty"`
	EventDate      *time.Time `json:"event_date,omitempty"`
	BannerURL      *string    `json:"banner_url,omitempty"`
	VenueName      string     `json:"venue_name,omitempty"`
	IsExpired      bool       `json:"is_expired"`
}

type CheckInRequest struct {
	Code string `json:"code" binding:"required" example:"EVT-TK-1A2B3C4D5E6F"`
}

type CheckInResponse struct {
	TicketID     uint64    `json:"ticket_id"`
	TicketCode   string    `json:"ticket_code"`
	EventName    string    `json:"event_name"`
	TierName     string    `json:"tier_name"`
	CustomerName string    `json:"customer_name"`
	CheckedInAt  time.Time `json:"checked_in_at"`
	CheckedBy    string    `json:"checked_by"`
}

type DashboardStatsResponse struct {
	TotalEvents      int64   `json:"total_events"`
	PublishedEvents  int64   `json:"published_events"`
	TotalOrders      int64   `json:"total_orders"`
	PaidOrders       int64   `json:"paid_orders"`
	TotalRevenue     float64 `json:"total_revenue"`
	TotalTicketsSold int64   `json:"total_tickets_sold"`
	TotalCheckedIn   int64   `json:"total_checked_in"`
}
