package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"eventifyApi/config"
	"eventifyApi/dto"
	"eventifyApi/models"
	"eventifyApi/repositories"
	"eventifyApi/utils"

	"github.com/google/uuid"
)

type EventService interface {
	CreateEvent(userID uint64, req dto.CreateEventRequest) (*dto.EventDetailResponse, error)
	GetEventByID(id uint64) (*dto.EventDetailResponse, error)
	GetPublishedEventByID(id uint64) (*dto.EventDetailResponse, error)
	GetPublishedEventBySlug(slug string) (*dto.EventDetailResponse, error)
	GetOrganizerEventDetail(userID uint64, userRole uint8, id uint64) (*dto.EventDetailResponse, error)
	GetAllEvents(filter dto.EventFilterQuery, onlyPublished bool) (*utils.PaginatedData, error)
	GetMyEvents(userID uint64, page, limit int) (*utils.PaginatedData, error)
	UpdateEvent(userID uint64, userRole uint8, eventID uint64, req dto.UpdateEventRequest) (*dto.EventDetailResponse, error)
	DeleteEvent(userID uint64, userRole uint8, eventID uint64) error
	UpdateBanner(userID uint64, userRole uint8, eventID uint64, bannerPath string) error
	AdminUpdateEventStatus(eventID uint64, status string) (*dto.EventDetailResponse, error)

	// Ticket Tier
	CreateTicketTier(userID uint64, userRole uint8, req dto.CreateTicketTierRequest) (*dto.TicketTierResponse, error)
	UpdateTicketTier(userID uint64, userRole uint8, tierID uint64, req dto.UpdateTicketTierRequest) (*dto.TicketTierResponse, error)
	DeleteTicketTier(userID uint64, userRole uint8, tierID uint64) error
}

type eventService struct {
	eventRepo repositories.EventRepository
	cfg       *config.Config
}

func NewEventService(eventRepo repositories.EventRepository, cfg *config.Config) EventService {
	return &eventService{eventRepo: eventRepo, cfg: cfg}
}

func slugify(text string) string {
	reg, _ := regexp.Compile("[^a-z0-9]+")
	slug := strings.ToLower(strings.TrimSpace(text))
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func (s *eventService) CreateEvent(userID uint64, req dto.CreateEventRequest) (*dto.EventDetailResponse, error) {
	if req.EndAt.Before(req.StartAt) {
		return nil, errors.New("end_at must be after start_at")
	}

	baseSlug := slugify(req.Name)
	if baseSlug == "" {
		baseSlug = "event"
	}
	slug := fmt.Sprintf("%s-%s", baseSlug, uuid.New().String()[:8])

	status := models.EventStatusDraft
	if req.Status != nil {
		status = models.EventStatus(*req.Status)
	}

	category := models.EventCategoryUmum
	if req.Category != "" {
		category = models.ParseEventCategory(req.Category)
	}

	event := &models.Event{
		CreatedBy:       &userID,
		Name:            req.Name,
		Category:        category,
		Slug:            slug,
		Description:     req.Description,
		TermsConditions: req.TermsConditions,
		Location:        req.Location,
		StartAt:         req.StartAt,
		EndAt:           req.EndAt,
		Status:          status,
	}

	if err := s.eventRepo.Create(event); err != nil {
		return nil, err
	}

	for _, t := range req.TicketTiers {
		isActive := true
		if t.IsActive != nil {
			isActive = *t.IsActive
		}
		tier := &models.TicketTier{
			EventID:     event.ID,
			Name:        t.Name,
			Price:       t.Price,
			Quota:       t.Quota,
			Description: t.Description,
			IsActive:    isActive,
			StartSaleAt: t.StartSaleAt,
			EndSaleAt:   t.EndSaleAt,
		}
		_ = s.eventRepo.CreateTier(tier)
	}

	return s.GetEventByID(event.ID)
}

func (s *eventService) GetEventByID(id uint64) (*dto.EventDetailResponse, error) {
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.mapEventToDetailResponse(event)
}

func (s *eventService) GetPublishedEventByID(id uint64) (*dto.EventDetailResponse, error) {
	event, err := s.eventRepo.FindPublishedByID(id)
	if err != nil {
		return nil, errors.New("event not found or not published")
	}
	return s.mapEventToDetailResponse(event)
}

func (s *eventService) GetPublishedEventBySlug(slug string) (*dto.EventDetailResponse, error) {
	event, err := s.eventRepo.FindPublishedBySlug(slug)
	if err != nil {
		return nil, errors.New("event not found or not published")
	}
	return s.mapEventToDetailResponse(event)
}

func (s *eventService) GetOrganizerEventDetail(userID uint64, userRole uint8, id uint64) (*dto.EventDetailResponse, error) {
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if userRole != 1 && (event.CreatedBy == nil || *event.CreatedBy != userID) {
		return nil, errors.New("forbidden: you do not own this event")
	}

	return s.mapEventToDetailResponse(event)
}

func (s *eventService) GetAllEvents(filter dto.EventFilterQuery, onlyPublished bool) (*utils.PaginatedData, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	events, total, err := s.eventRepo.FindAll(filter, onlyPublished)
	if err != nil {
		return nil, err
	}

	var list []dto.EventDetailResponse
	for _, e := range events {
		resp, _ := s.mapEventToDetailResponse(&e)
		if resp != nil {
			list = append(list, *resp)
		}
	}

	totalPages := int((total + int64(filter.Limit) - 1) / int64(filter.Limit))

	return &utils.PaginatedData{
		Items:      list,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *eventService) GetMyEvents(userID uint64, page, limit int) (*utils.PaginatedData, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	events, total, err := s.eventRepo.FindByCreatorID(userID, page, limit)
	if err != nil {
		return nil, err
	}

	var list []dto.EventDetailResponse
	for _, e := range events {
		resp, _ := s.mapEventToDetailResponse(&e)
		if resp != nil {
			list = append(list, *resp)
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &utils.PaginatedData{
		Items:      list,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *eventService) UpdateEvent(userID uint64, userRole uint8, eventID uint64, req dto.UpdateEventRequest) (*dto.EventDetailResponse, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, err
	}

	// role 1 is admin, others can only update their own events
	if userRole != 1 && (event.CreatedBy == nil || *event.CreatedBy != userID) {
		return nil, errors.New("forbidden: you do not own this event")
	}

	if req.EndAt.Before(req.StartAt) {
		return nil, errors.New("end_at must be after start_at")
	}

	event.Name = req.Name
	if req.Category != "" {
		event.Category = models.ParseEventCategory(req.Category)
	}
	event.Description = req.Description
	event.TermsConditions = req.TermsConditions
	event.Location = req.Location
	event.StartAt = req.StartAt
	event.EndAt = req.EndAt
	event.Status = models.EventStatus(req.Status)

	if err := s.eventRepo.Update(event); err != nil {
		return nil, err
	}

	return s.GetEventByID(eventID)
}

func (s *eventService) DeleteEvent(userID uint64, userRole uint8, eventID uint64) error {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return err
	}

	if userRole != 1 && (event.CreatedBy == nil || *event.CreatedBy != userID) {
		return errors.New("forbidden: you do not own this event")
	}

	return s.eventRepo.Delete(eventID)
}

func (s *eventService) UpdateBanner(userID uint64, userRole uint8, eventID uint64, bannerPath string) error {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return err
	}

	if userRole != 1 && (event.CreatedBy == nil || *event.CreatedBy != userID) {
		return errors.New("forbidden: you do not own this event")
	}

	event.BannerPath = &bannerPath
	return s.eventRepo.Update(event)
}

func (s *eventService) AdminUpdateEventStatus(eventID uint64, status string) (*dto.EventDetailResponse, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, err
	}

	validStatuses := map[string]bool{
		string(models.EventStatusDraft):     true,
		string(models.EventStatusPublished): true,
		string(models.EventStatusCompleted): true,
		string(models.EventStatusCancelled): true,
	}

	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid event status: %s", status)
	}

	event.Status = models.EventStatus(status)
	if err := s.eventRepo.Update(event); err != nil {
		return nil, err
	}

	return s.GetEventByID(eventID)
}

func (s *eventService) CreateTicketTier(userID uint64, userRole uint8, req dto.CreateTicketTierRequest) (*dto.TicketTierResponse, error) {
	event, err := s.eventRepo.FindByID(req.EventID)
	if err != nil {
		return nil, err
	}

	if userRole != 1 && (event.CreatedBy == nil || *event.CreatedBy != userID) {
		return nil, errors.New("forbidden: you do not own this event")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	tier := &models.TicketTier{
		EventID:     req.EventID,
		Name:        req.Name,
		Price:       req.Price,
		Quota:       req.Quota,
		Description: req.Description,
		IsActive:    isActive,
		StartSaleAt: req.StartSaleAt,
		EndSaleAt:   req.EndSaleAt,
	}

	if err := s.eventRepo.CreateTier(tier); err != nil {
		return nil, err
	}

	return &dto.TicketTierResponse{
		ID:             tier.ID,
		EventID:        tier.EventID,
		Name:           tier.Name,
		Price:          tier.Price,
		Quota:          tier.Quota,
		RemainingQuota: tier.Quota,
		Description:    tier.Description,
		IsActive:       tier.IsActive,
		StartSaleAt:    tier.StartSaleAt,
		EndSaleAt:      tier.EndSaleAt,
		IsAvailable:    tier.IsActive,
	}, nil
}

func (s *eventService) UpdateTicketTier(userID uint64, userRole uint8, tierID uint64, req dto.UpdateTicketTierRequest) (*dto.TicketTierResponse, error) {
	tier, err := s.eventRepo.FindTierByID(tierID)
	if err != nil {
		return nil, err
	}

	if userRole != 1 && (tier.Event == nil || tier.Event.CreatedBy == nil || *tier.Event.CreatedBy != userID) {
		return nil, errors.New("forbidden: you do not own this event")
	}

	tier.Name = req.Name
	tier.Price = req.Price
	tier.Quota = req.Quota
	tier.Description = req.Description
	if req.IsActive != nil {
		tier.IsActive = *req.IsActive
	}
	tier.StartSaleAt = req.StartSaleAt
	tier.EndSaleAt = req.EndSaleAt

	if err := s.eventRepo.UpdateTier(tier); err != nil {
		return nil, err
	}

	sold, _ := s.eventRepo.GetSoldTierCount(tier.ID)
	remaining := uint32(0)
	if tier.Quota > uint32(sold) {
		remaining = tier.Quota - uint32(sold)
	}

	return &dto.TicketTierResponse{
		ID:             tier.ID,
		EventID:        tier.EventID,
		Name:           tier.Name,
		Price:          tier.Price,
		Quota:          tier.Quota,
		RemainingQuota: remaining,
		Description:    tier.Description,
		IsActive:       tier.IsActive,
		StartSaleAt:    tier.StartSaleAt,
		EndSaleAt:      tier.EndSaleAt,
		IsAvailable:    tier.IsActive && remaining > 0,
	}, nil
}

func (s *eventService) DeleteTicketTier(userID uint64, userRole uint8, tierID uint64) error {
	tier, err := s.eventRepo.FindTierByID(tierID)
	if err != nil {
		return err
	}

	if userRole != 1 && (tier.Event == nil || tier.Event.CreatedBy == nil || *tier.Event.CreatedBy != userID) {
		return errors.New("forbidden: you do not own this event")
	}

	sold, err := s.eventRepo.GetSoldTierCount(tierID)
	if err != nil {
		return err
	}
	if sold > 0 {
		return errors.New("cannot delete ticket tier because tickets have already been ordered for this tier")
	}

	return s.eventRepo.DeleteTier(tierID)
}

func (s *eventService) mapEventToDetailResponse(event *models.Event) (*dto.EventDetailResponse, error) {
	var creatorName string
	if event.Creator != nil {
		creatorName = event.Creator.Name
	}

	var bannerURL *string
	if event.BannerPath != nil && *event.BannerPath != "" {
		url := fmt.Sprintf("%s/%s", s.cfg.AppURL, strings.TrimPrefix(*event.BannerPath, "./"))
		bannerURL = &url
	}

	var tierResponses []dto.TicketTierResponse
	now := time.Now()

	for _, t := range event.TicketTiers {
		sold, _ := s.eventRepo.GetSoldTierCount(t.ID)
		remaining := uint32(0)
		if t.Quota > uint32(sold) {
			remaining = t.Quota - uint32(sold)
		}

		isAvailable := t.IsActive && remaining > 0
		if t.StartSaleAt != nil && now.Before(*t.StartSaleAt) {
			isAvailable = false
		}
		if t.EndSaleAt != nil && now.After(*t.EndSaleAt) {
			isAvailable = false
		}

		tierResponses = append(tierResponses, dto.TicketTierResponse{
			ID:             t.ID,
			EventID:        t.EventID,
			Name:           t.Name,
			Price:          t.Price,
			Quota:          t.Quota,
			RemainingQuota: remaining,
			Description:    t.Description,
			IsActive:       t.IsActive,
			StartSaleAt:    t.StartSaleAt,
			EndSaleAt:      t.EndSaleAt,
			IsAvailable:    isAvailable,
		})
	}

	category := string(event.Category)
	if category == "" {
		category = string(models.EventCategoryUmum)
	}

	return &dto.EventDetailResponse{
		ID:              event.ID,
		CreatedBy:       event.CreatedBy,
		CreatorName:     creatorName,
		Name:            event.Name,
		Category:        category,
		Slug:            event.Slug,
		Description:     event.Description,
		TermsConditions: event.TermsConditions,
		Location:        event.Location,
		StartAt:         event.StartAt,
		EndAt:           event.EndAt,
		BannerPath:      event.BannerPath,
		BannerURL:       bannerURL,
		Status:          string(event.Status),
		CreatedAt:       event.CreatedAt,
		UpdatedAt:       event.UpdatedAt,
		TicketTiers:     tierResponses,
	}, nil
}
