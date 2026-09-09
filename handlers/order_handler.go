package handlers

import (
	"net/http"
	"strconv"

	"eventifyApi/dto"
	"eventifyApi/services"
	"eventifyApi/utils"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService services.OrderService
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrder godoc
// @Summary Order event tickets
// @Description Book tickets for an event across one or more tiers (with atomic quota check)
// @Tags Orders & Bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateOrderRequest true "Order Details"
// @Success 201 {object} utils.APIResponse{data=dto.OrderResponse} "Order created successfully"
// @Failure 400 {object} utils.APIResponse "Quota exceeded or invalid input"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Router /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	order, err := h.orderService.CreateOrder(userID, req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Order created successfully", order)
}

// GetMyOrders godoc
// @Summary List customer's orders
// @Description Retrieve list of orders made by logged-in customer
// @Tags Orders & Bookings
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.APIResponse{data=utils.PaginatedData{items=[]dto.OrderResponse}} "Orders retrieved"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Router /api/v1/orders/my-orders [get]
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID := c.GetUint64("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, err := h.orderService.GetMyOrders(userID, page, limit)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Orders retrieved successfully", data)
}

// GetOrderByCode godoc
// @Summary Get order detail by code
// @Description View full order details and payment status by order code
// @Tags Orders & Bookings
// @Produce json
// @Security BearerAuth
// @Param code path string true "Order Code (e.g. ORD-20260909-XXXX)"
// @Success 200 {object} utils.APIResponse{data=dto.OrderResponse} "Order detail retrieved"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Failure 404 {object} utils.APIResponse "Order not found"
// @Router /api/v1/orders/{code} [get]
func (h *OrderHandler) GetOrderByCode(c *gin.Context) {
	userID := c.GetUint64("user_id")
	userRole := c.GetUint8("role_id")
	code := c.Param("code")

	order, err := h.orderService.GetOrderByCode(userID, userRole, code)
	if err != nil {
		utils.JSONError(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Order retrieved successfully", order)
}

// GetAllOrders godoc
// @Summary List all orders (Admin only)
// @Description Retrieve all orders across the entire platform
// @Tags Admin - Orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.APIResponse{data=utils.PaginatedData{items=[]dto.OrderResponse}} "All orders"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/admin/orders [get]
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, err := h.orderService.GetAllOrders(page, limit)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Orders retrieved successfully", data)
}

// PaymentWebhook godoc
// @Summary Payment gateway callback webhook
// @Description Handles notifications from payment gateways (Midtrans SHA512 signature or Xendit callback token) and generates tickets upon paid status
// @Tags Payments & Gateway
// @Accept json
// @Produce json
// @Param x-callback-token header string false "Xendit Callback Token or Dev Secret"
// @Param request body dto.PaymentWebhookRequest true "Payment Webhook Payload"
// @Success 200 {object} utils.APIResponse "Payment status updated"
// @Failure 400 {object} utils.APIResponse "Invalid request"
// @Failure 401 {object} utils.APIResponse "Unauthorized / Invalid signature"
// @Router /api/v1/payments/webhook [post]
func (h *OrderHandler) PaymentWebhook(c *gin.Context) {
	var req dto.PaymentWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid webhook body", err.Error())
		return
	}

	headerToken := c.GetHeader("x-callback-token")
	if headerToken == "" {
		headerToken = c.GetHeader("X-Callback-Token")
	}

	if err := h.orderService.HandlePaymentWebhook(headerToken, req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Payment status processed successfully", nil)
}

// GetMyTickets godoc
// @Summary List customer tickets (QR Codes)
// @Description Retrieve all issued tickets belonging to the logged-in customer for mobile app display
// @Tags Customer - Tickets
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.APIResponse{data=utils.PaginatedData{items=[]dto.TicketResponse}} "Tickets list"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Router /api/v1/tickets/my-tickets [get]
func (h *OrderHandler) GetMyTickets(c *gin.Context) {
	userID := c.GetUint64("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, err := h.orderService.GetMyTickets(userID, page, limit)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Tickets retrieved successfully", data)
}

// GetTicketByCode godoc
// @Summary Get ticket detail by QR Code string
// @Description View ticket details by ticket code string
// @Tags Tickets
// @Produce json
// @Security BearerAuth
// @Param code path string true "Ticket Code (e.g. EVT-TK-1A2B3C4D5E6F)"
// @Success 200 {object} utils.APIResponse{data=dto.TicketResponse} "Ticket details"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Failure 404 {object} utils.APIResponse "Ticket not found"
// @Router /api/v1/tickets/{code} [get]
func (h *OrderHandler) GetTicketByCode(c *gin.Context) {
	userID := c.GetUint64("user_id")
	userRole := c.GetUint8("role_id")
	code := c.Param("code")

	ticket, err := h.orderService.GetTicketByCode(userID, userRole, code)
	if err != nil {
		utils.JSONError(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Ticket retrieved successfully", ticket)
}

// CheckInTicket godoc
// @Summary Scan / Check-in ticket at entrance (Panitia & Admin)
// @Description Mobile QR scanner endpoint to check-in attendee and invalidate reuse
// @Tags Gate Check-In
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CheckInRequest true "Ticket QR String"
// @Success 200 {object} utils.APIResponse{data=dto.CheckInResponse} "Check-in successful"
// @Failure 400 {object} utils.APIResponse "Ticket invalid or already checked in"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/scanner/check-in [post]
func (h *OrderHandler) CheckInTicket(c *gin.Context) {
	checkerID := c.GetUint64("user_id")

	var req dto.CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	result, err := h.orderService.CheckInTicket(checkerID, req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Attendee successfully checked in", result)
}

// GetDashboardStats godoc
// @Summary Get admin analytics & overview metrics (Admin only)
// @Description Summary of events, orders, tickets sold, attendance, and revenue
// @Tags Admin - Analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.DashboardStatsResponse} "Dashboard metrics"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/admin/dashboard [get]
func (h *OrderHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.orderService.GetDashboardStats()
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Dashboard statistics retrieved successfully", stats)
}
