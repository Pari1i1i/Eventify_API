package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"eventifyApi/config"
	"eventifyApi/dto"
	"eventifyApi/services"
	"eventifyApi/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventHandler struct {
	eventService services.EventService
	cfg          *config.Config
}

func NewEventHandler(eventService services.EventService, cfg *config.Config) *EventHandler {
	return &EventHandler{eventService: eventService, cfg: cfg}
}

// GetAllEvents godoc
// @Summary List published events (Public)
// @Description Browse all published events with filtering and search
// @Tags Public - Events
// @Produce json
// @Param search query string false "Search name, location, or description"
// @Param start_date query string false "Filter start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter end date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.APIResponse{data=utils.PaginatedData{items=[]dto.EventDetailResponse}} "Events retrieved"
// @Router /api/v1/events [get]
func (h *EventHandler) GetAllEvents(c *gin.Context) {
	var filter dto.EventFilterQuery
	_ = c.ShouldBindQuery(&filter)

	data, err := h.eventService.GetAllEvents(filter, true)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Events retrieved successfully", data)
}

// GetEventBySlug godoc
// @Summary Get event detail by slug (Public)
// @Description Retrieve single event detail and active ticket tiers by slug (Only published events)
// @Tags Public - Events
// @Produce json
// @Param slug path string true "Event Slug"
// @Success 200 {object} utils.APIResponse{data=dto.EventDetailResponse} "Event detail retrieved"
// @Failure 404 {object} utils.APIResponse "Event not found or not published"
// @Router /api/v1/events/{slug} [get]
func (h *EventHandler) GetEventBySlug(c *gin.Context) {
	slug := c.Param("slug")

	event, err := h.eventService.GetPublishedEventBySlug(slug)
	if err != nil {
		utils.JSONError(c, http.StatusNotFound, "Event not found or not published", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Event detail retrieved successfully", event)
}

// GetEventByID godoc
// @Summary Get published event detail by ID (Public)
// @Description Retrieve published event detail and tiers by numerical ID
// @Tags Public - Events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} utils.APIResponse{data=dto.EventDetailResponse} "Event detail retrieved"
// @Failure 404 {object} utils.APIResponse "Event not found or not published"
// @Router /api/v1/events/id/{id} [get]
func (h *EventHandler) GetEventByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid event ID", nil)
		return
	}

	event, err := h.eventService.GetPublishedEventByID(id)
	if err != nil {
		utils.JSONError(c, http.StatusNotFound, "Event not found or not published", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Event detail retrieved successfully", event)
}

// GetOrganizerEventDetail godoc
// @Summary Preview organizer event detail by ID (Panitia & Admin)
// @Description View event detail including draft/cancelled status for owner/admin preview
// @Tags Events Management
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} utils.APIResponse{data=dto.EventDetailResponse} "Event detail retrieved"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Failure 404 {object} utils.APIResponse "Event not found"
// @Router /api/v1/organizer/events/{id} [get]
func (h *EventHandler) GetOrganizerEventDetail(c *gin.Context) {
	userID := c.GetUint64("user_id")
	userRole := c.GetUint8("role_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid event ID", nil)
		return
	}

	event, err := h.eventService.GetOrganizerEventDetail(userID, userRole, id)
	if err != nil {
		utils.JSONError(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Event detail retrieved successfully", event)
}

// AdminGetAllEvents godoc
// @Summary List all events with all statuses (Admin only)
// @Description View and filter all events from all organizers
// @Tags Admin - Events
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search keyword"
// @Param status query string false "Filter status (draft, published, completed, cancelled)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.APIResponse{data=utils.PaginatedData{items=[]dto.EventDetailResponse}} "Events list"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/admin/events [get]
func (h *EventHandler) AdminGetAllEvents(c *gin.Context) {
	var filter dto.EventFilterQuery
	_ = c.ShouldBindQuery(&filter)

	data, err := h.eventService.GetAllEvents(filter, false)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "All events retrieved successfully", data)
}

// AdminUpdateEventStatus godoc
// @Summary Approve or change event status (Admin only)
// @Description Moderation tool to publish, complete, or cancel an event
// @Tags Admin - Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Param request body dto.UpdateEventStatusRequest true "New status payload"
// @Success 200 {object} utils.APIResponse{data=dto.EventDetailResponse} "Status updated"
// @Failure 400 {object} utils.APIResponse "Invalid request"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/admin/events/{id}/status [put]
func (h *EventHandler) AdminUpdateEventStatus(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid event ID", nil)
		return
	}

	var req dto.UpdateEventStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	resp, err := h.eventService.AdminUpdateEventStatus(eventID, req.Status)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Event status updated successfully", resp)
}

// CreateEvent godoc
// @Summary Create a new event (Panitia & Admin)
// @Description Create an event with draft or published status and optional initial ticket tiers
// @Tags Events Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateEventRequest true "Event Creation Details"
// @Success 201 {object} utils.APIResponse{data=dto.EventDetailResponse} "Event created successfully"
// @Failure 400 {object} utils.APIResponse "Validation error"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/organizer/events [post]
func (h *EventHandler) CreateEvent(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req dto.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	event, err := h.eventService.CreateEvent(userID, req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Event created successfully", event)
}

// GetMyEvents godoc
// @Summary List organizer's own events (Panitia & Admin)
// @Description Retrieve events created by logged-in organizer
// @Tags Events Management
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.APIResponse{data=utils.PaginatedData{items=[]dto.EventDetailResponse}} "Organizer events retrieved"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Router /api/v1/organizer/my-events [get]
func (h *EventHandler) GetMyEvents(c *gin.Context) {
	userID := c.GetUint64("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, err := h.eventService.GetMyEvents(userID, page, limit)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Organizer events retrieved successfully", data)
}

// UpdateEvent godoc
// @Summary Update event (Panitia & Admin)
// @Description Update existing event details
// @Tags Events Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Param request body dto.UpdateEventRequest true "Update Event Body"
// @Success 200 {object} utils.APIResponse{data=dto.EventDetailResponse} "Event updated successfully"
// @Failure 400 {object} utils.APIResponse "Invalid request"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/organizer/events/{id} [put]
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	userID := c.GetUint64("user_id")
	userRole := c.GetUint8("role_id")
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid event ID", nil)
		return
	}

	var req dto.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	updated, err := h.eventService.UpdateEvent(userID, userRole, eventID, req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Event updated successfully", updated)
}

// DeleteEvent godoc
// @Summary Soft delete event (Panitia & Admin)
// @Description Delete an event
// @Tags Events Management
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} utils.APIResponse "Event deleted successfully"
// @Failure 400 {object} utils.APIResponse "Invalid request"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/organizer/events/{id} [delete]
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	userID := c.GetUint64("user_id")
	userRole := c.GetUint8("role_id")
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid event ID", nil)
		return
	}

	if err := h.eventService.DeleteEvent(userID, userRole, eventID); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Event deleted successfully", nil)
}

// UploadBanner godoc
// @Summary Upload event banner image (Panitia & Admin)
// @Description Upload image file (jpg, jpeg, png, webp) for event banner
// @Tags Events Management
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Param banner formData file true "Banner Image File"
// @Success 200 {object} utils.APIResponse "Banner uploaded successfully"
// @Failure 400 {object} utils.APIResponse "Invalid file upload"
// @Router /api/v1/organizer/events/{id}/banner [post]
func (h *EventHandler) UploadBanner(c *gin.Context) {
	userID := c.GetUint64("user_id")
	userRole := c.GetUint8("role_id")
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid event ID", nil)
		return
	}

	file, err := c.FormFile("banner")
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Banner image file is required", err.Error())
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		utils.JSONError(c, http.StatusBadRequest, "Only .jpg, .jpeg, .png, and .webp files are allowed", nil)
		return
	}

	uploadDir := h.cfg.UploadDir
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to create upload directory", err.Error())
		return
	}

	filename := fmt.Sprintf("event-%d-%d-%s%s", eventID, time.Now().Unix(), uuid.New().String()[:6], ext)
	savePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to save file", err.Error())
		return
	}

	// Normalize relative path with forward slashes for cross-platform DB compatibility
	bannerRelativePath := filepath.ToSlash(savePath)
	if err := h.eventService.UpdateBanner(userID, userRole, eventID, bannerRelativePath); err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Banner uploaded successfully", gin.H{
		"banner_path": bannerRelativePath,
		"banner_url":  fmt.Sprintf("%s/%s", h.cfg.AppURL, strings.TrimPrefix(bannerRelativePath, "./")),
	})
}

// CreateTicketTier godoc
// @Summary Add ticket tier to event (Panitia & Admin)
// @Description Create tier like VIP, Regular, Early Bird
// @Tags Ticket Tiers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateTicketTierRequest true "Ticket Tier Details"
// @Success 201 {object} utils.APIResponse{data=dto.TicketTierResponse} "Ticket tier created"
// @Failure 400 {object} utils.APIResponse "Bad request"
// @Router /api/v1/organizer/ticket-tiers [post]
func (h *EventHandler) CreateTicketTier(c *gin.Context) {
	userID := c.GetUint64("user_id")
	userRole := c.GetUint8("role_id")

	var req dto.CreateTicketTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tier, err := h.eventService.CreateTicketTier(userID, userRole, req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Ticket tier created successfully", tier)
}

// UpdateTicketTier godoc
// @Summary Update ticket tier (Panitia & Admin)
// @Description Update tier pricing, quota, date sale window, or status
// @Tags Ticket Tiers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket Tier ID"
// @Param request body dto.UpdateTicketTierRequest true "Update Ticket Tier Body"
// @Success 200 {object} utils.APIResponse{data=dto.TicketTierResponse} "Ticket tier updated"
// @Failure 400 {object} utils.APIResponse "Bad request"
// @Router /api/v1/organizer/ticket-tiers/{id} [put]
func (h *EventHandler) UpdateTicketTier(c *gin.Context) {
	userID := c.GetUint64("user_id")
	userRole := c.GetUint8("role_id")
	tierID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid tier ID", nil)
		return
	}

	var req dto.UpdateTicketTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tier, err := h.eventService.UpdateTicketTier(userID, userRole, tierID, req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Ticket tier updated successfully", tier)
}

// DeleteTicketTier godoc
// @Summary Delete ticket tier (Panitia & Admin)
// @Description Delete ticket tier (only allowed if no tickets sold yet)
// @Tags Ticket Tiers
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket Tier ID"
// @Success 200 {object} utils.APIResponse "Ticket tier deleted"
// @Failure 400 {object} utils.APIResponse "Cannot delete tier with sold tickets"
// @Router /api/v1/organizer/ticket-tiers/{id} [delete]
func (h *EventHandler) DeleteTicketTier(c *gin.Context) {
	userID := c.GetUint64("user_id")
	userRole := c.GetUint8("role_id")
	tierID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid tier ID", nil)
		return
	}

	if err := h.eventService.DeleteTicketTier(userID, userRole, tierID); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Ticket tier deleted successfully", nil)
}
