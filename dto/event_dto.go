package dto

import "time"

type CreateEventRequest struct {
	Name            string                  `json:"name" binding:"required,max=200" example:"Tech Conference 2026"`
	Category        string                  `json:"category" binding:"omitempty" example:"Teknologi"`
	Description     string                  `json:"description" binding:"required" example:"Biggest tech summit in Indonesia"`
	TermsConditions *string                 `json:"terms_conditions" example:"Tickets are non-refundable."`
	Location        string                  `json:"location" binding:"required,max=255" example:"Jakarta Convention Center"`
	StartAt         time.Time               `json:"start_at" binding:"required" example:"2026-10-15T09:00:00Z"`
	EndAt           time.Time               `json:"end_at" binding:"required" example:"2026-10-15T18:00:00Z"`
	Status          *string                 `json:"status" binding:"omitempty,oneof=draft published completed cancelled" example:"draft"`
	TicketTiers     []CreateTicketTierChild `json:"ticket_tiers" binding:"omitempty,dive"`
}

type UpdateEventRequest struct {
	Name            string    `json:"name" binding:"required,max=200" example:"Tech Conference 2026 Updated"`
	Category        string    `json:"category" binding:"omitempty" example:"Teknologi"`
	Description     string    `json:"description" binding:"required" example:"Updated description"`
	TermsConditions *string   `json:"terms_conditions" example:"Updated terms"`
	Location        string    `json:"location" binding:"required,max=255" example:"Bali Nusa Dua"`
	StartAt         time.Time `json:"start_at" binding:"required" example:"2026-11-01T09:00:00Z"`
	EndAt           time.Time `json:"end_at" binding:"required" example:"2026-11-01T17:00:00Z"`
	Status          string    `json:"status" binding:"required,oneof=draft published completed cancelled" example:"published"`
}

type CreateTicketTierChild struct {
	Name        string     `json:"name" binding:"required,max=100" example:"Early Bird"`
	Price       float64    `json:"price" binding:"min=0" example:"150000"`
	Quota       uint32     `json:"quota" binding:"min=1" example:"100"`
	Description *string    `json:"description" example:"Limited early bird tickets"`
	IsActive    *bool      `json:"is_active" example:"true"`
	StartSaleAt *time.Time `json:"start_sale_at"`
	EndSaleAt   *time.Time `json:"end_sale_at"`
}

type CreateTicketTierRequest struct {
	EventID     uint64     `json:"event_id" binding:"required" example:"1"`
	Name        string     `json:"name" binding:"required,max=100" example:"VIP"`
	Price       float64    `json:"price" binding:"min=0" example:"350000"`
	Quota       uint32     `json:"quota" binding:"min=1" example:"50"`
	Description *string    `json:"description" example:"Free snacks and front row seat"`
	IsActive    *bool      `json:"is_active" example:"true"`
	StartSaleAt *time.Time `json:"start_sale_at"`
	EndSaleAt   *time.Time `json:"end_sale_at"`
}

type UpdateTicketTierRequest struct {
	Name        string     `json:"name" binding:"required,max=100" example:"VIP Pass"`
	Price       float64    `json:"price" binding:"min=0" example:"400000"`
	Quota       uint32     `json:"quota" binding:"min=1" example:"60"`
	Description *string    `json:"description" example:"Front row VIP seats"`
	IsActive    *bool      `json:"is_active" example:"true"`
	StartSaleAt *time.Time `json:"start_sale_at"`
	EndSaleAt   *time.Time `json:"end_sale_at"`
}

type UpdateEventStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=draft published completed cancelled" example:"published"`
}

type EventFilterQuery struct {
	Search    string `form:"search"`
	Category  string `form:"category"`
	Status    string `form:"status"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=10"`
}

type EventDetailResponse struct {
	ID              uint64               `json:"id"`
	CreatedBy       *uint64              `json:"created_by"`
	CreatorName     string               `json:"creator_name,omitempty"`
	Name            string               `json:"name"`
	Category        string               `json:"category"`
	Slug            string               `json:"slug"`
	Description     string               `json:"description"`
	TermsConditions *string              `json:"terms_conditions"`
	Location        string               `json:"location"`
	StartAt         time.Time            `json:"start_at"`
	EndAt           time.Time            `json:"end_at"`
	BannerPath      *string              `json:"banner_path"`
	BannerURL       *string              `json:"banner_url"`
	Status          string               `json:"status"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	TicketTiers     []TicketTierResponse `json:"ticket_tiers"`
}

type TicketTierResponse struct {
	ID             uint64     `json:"id"`
	EventID        uint64     `json:"event_id"`
	Name           string     `json:"name"`
	Price          float64    `json:"price"`
	Quota          uint32     `json:"quota"`
	RemainingQuota uint32     `json:"remaining_quota"`
	Description    *string    `json:"description"`
	IsActive       bool       `json:"is_active"`
	StartSaleAt    *time.Time `json:"start_sale_at"`
	EndSaleAt      *time.Time `json:"end_sale_at"`
	IsAvailable    bool       `json:"is_available"`
}
