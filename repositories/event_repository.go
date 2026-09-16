package repositories

import (
	"strings"
	"time"

	"eventifyApi/dto"
	"eventifyApi/models"

	"gorm.io/gorm"
)

type EventRepository interface {
	Create(event *models.Event) error
	FindByID(id uint64) (*models.Event, error)
	FindPublishedByID(id uint64) (*models.Event, error)
	FindBySlug(slug string) (*models.Event, error)
	FindPublishedBySlug(slug string) (*models.Event, error)
	FindAll(filter dto.EventFilterQuery, onlyPublished bool) ([]models.Event, int64, error)
	FindByCreatorID(creatorID uint64, page, limit int) ([]models.Event, int64, error)
	Update(event *models.Event) error
	Delete(id uint64) error

	// Ticket Tier
	CreateTier(tier *models.TicketTier) error
	FindTierByID(id uint64) (*models.TicketTier, error)
	UpdateTier(tier *models.TicketTier) error
	DeleteTier(id uint64) error
	GetSoldTierCount(tierID uint64) (int64, error)
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) FindByID(id uint64) (*models.Event, error) {
	var event models.Event
	err := r.db.Preload("Creator").Preload("TicketTiers").First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) FindPublishedByID(id uint64) (*models.Event, error) {
	var event models.Event
	err := r.db.Preload("Creator").Preload("TicketTiers").
		Where("id = ? AND status = ?", id, models.EventStatusPublished).
		First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) FindBySlug(slug string) (*models.Event, error) {
	var event models.Event
	err := r.db.Preload("Creator").Preload("TicketTiers").Where("slug = ?", slug).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) FindPublishedBySlug(slug string) (*models.Event, error) {
	var event models.Event
	err := r.db.Preload("Creator").Preload("TicketTiers").
		Where("slug = ? AND status = ?", slug, models.EventStatusPublished).
		First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) FindAll(filter dto.EventFilterQuery, onlyPublished bool) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	query := r.db.Model(&models.Event{})

	if onlyPublished {
		query = query.Where("status = ?", models.EventStatusPublished)
	} else if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Category != "" && strings.ToLower(strings.TrimSpace(filter.Category)) != "semua" {
		query = query.Where("category = ?", models.ParseEventCategory(filter.Category))
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("name LIKE ? OR location LIKE ? OR description LIKE ? OR category LIKE ?", searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if filter.StartDate != "" {
		if t, err := time.Parse("2006-01-02", filter.StartDate); err == nil {
			query = query.Where("start_at >= ?", t)
		}
	}

	if filter.EndDate != "" {
		if t, err := time.Parse("2006-01-02", filter.EndDate); err == nil {
			query = query.Where("end_at <= ?", t.Add(24*time.Hour))
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.Preload("TicketTiers").Preload("Creator").
		Limit(filter.Limit).Offset(offset).
		Order("start_at ASC").
		Find(&events).Error

	return events, total, err
}

func (r *eventRepository) FindByCreatorID(creatorID uint64, page, limit int) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	query := r.db.Model(&models.Event{}).Where("created_by = ?", creatorID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("TicketTiers").Preload("Creator").
		Limit(limit).Offset(offset).
		Order("id DESC").
		Find(&events).Error

	return events, total, err
}

func (r *eventRepository) Update(event *models.Event) error {
	return r.db.Save(event).Error
}

func (r *eventRepository) Delete(id uint64) error {
	return r.db.Delete(&models.Event{}, id).Error
}

func (r *eventRepository) CreateTier(tier *models.TicketTier) error {
	return r.db.Create(tier).Error
}

func (r *eventRepository) FindTierByID(id uint64) (*models.TicketTier, error) {
	var tier models.TicketTier
	err := r.db.Preload("Event").First(&tier, id).Error
	if err != nil {
		return nil, err
	}
	return &tier, nil
}

func (r *eventRepository) UpdateTier(tier *models.TicketTier) error {
	return r.db.Save(tier).Error
}

func (r *eventRepository) DeleteTier(id uint64) error {
	return r.db.Delete(&models.TicketTier{}, id).Error
}

func (r *eventRepository) GetSoldTierCount(tierID uint64) (int64, error) {
	var soldCount int64
	// count orders where payment_status IN ('paid', 'free', 'pending')
	err := r.db.Table("order_items").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("order_items.ticket_tier_id = ? AND orders.payment_status IN ('paid', 'free', 'pending')", tierID).
		Select("COALESCE(SUM(order_items.quantity), 0)").
		Scan(&soldCount).Error
	return soldCount, err
}
