package repositories

import (
	"eventifyApi/models"

	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrderWithTx(order *models.Order, items []models.OrderItem, payment *models.PaymentDetails) error
	FindByID(id uint64) (*models.Order, error)
	FindByOrderCode(orderCode string) (*models.Order, error)
	FindByUserID(userID uint64, page, limit int) ([]models.Order, int64, error)
	FindAll(page, limit int) ([]models.Order, int64, error)
	UpdatePaymentStatus(orderID uint64, status models.PaymentStatus) error
	GetDB() *gorm.DB

	// Tickets
	CreateTickets(tickets []models.Ticket) error
	FindTicketByCode(code string) (*models.Ticket, error)
	FindTicketsByUserID(userID uint64, page, limit int) ([]models.Ticket, int64, error)
	UpdateTicket(ticket *models.Ticket) error

	// Dashboard metrics
	GetDashboardStats() (totalEvents, publishedEvents, totalOrders, paidOrders, totalTicketsSold, totalCheckedIn int64, totalRevenue float64, err error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *orderRepository) CreateOrderWithTx(order *models.Order, items []models.OrderItem, payment *models.PaymentDetails) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		for i := range items {
			items[i].OrderID = order.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}

		if payment != nil {
			payment.OrderID = order.ID
			if err := tx.Create(payment).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *orderRepository) FindByID(id uint64) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("User").
		Preload("Event").
		Preload("PaymentDetails").
		Preload("OrderItems.TicketTier").
		Preload("OrderItems.Tickets").
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByOrderCode(orderCode string) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("User").
		Preload("Event").
		Preload("PaymentDetails").
		Preload("OrderItems.TicketTier").
		Preload("OrderItems.Tickets").
		Where("order_code = ?", orderCode).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByUserID(userID uint64, page, limit int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("Event").
		Preload("PaymentDetails").
		Preload("OrderItems.TicketTier").
		Limit(limit).Offset(offset).
		Order("id DESC").
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepository) FindAll(page, limit int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	if err := r.db.Model(&models.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := r.db.Preload("User").
		Preload("Event").
		Preload("PaymentDetails").
		Preload("OrderItems.TicketTier").
		Limit(limit).Offset(offset).
		Order("id DESC").
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepository) UpdatePaymentStatus(orderID uint64, status models.PaymentStatus) error {
	return r.db.Model(&models.Order{}).Where("id = ?", orderID).Update("payment_status", status).Error
}

func (r *orderRepository) CreateTickets(tickets []models.Ticket) error {
	return r.db.Create(&tickets).Error
}

func (r *orderRepository) FindTicketByCode(code string) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.Preload("OrderItem.TicketTier.Event").
		Preload("OrderItem.Order.User").
		Preload("Checker").
		Where("code = ?", code).
		First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *orderRepository) FindTicketsByUserID(userID uint64, page, limit int) ([]models.Ticket, int64, error) {
	var tickets []models.Ticket
	var total int64

	query := r.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.user_id = ? AND orders.payment_status IN ('paid', 'free')", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("OrderItem.TicketTier.Event").
		Preload("OrderItem.Order.User").
		Limit(limit).Offset(offset).
		Order("tickets.id DESC").
		Find(&tickets).Error

	return tickets, total, err
}

func (r *orderRepository) UpdateTicket(ticket *models.Ticket) error {
	return r.db.Save(ticket).Error
}

func (r *orderRepository) GetDashboardStats() (totalEvents, publishedEvents, totalOrders, paidOrders, totalTicketsSold, totalCheckedIn int64, totalRevenue float64, err error) {
	err = r.db.Model(&models.Event{}).Count(&totalEvents).Error
	if err != nil {
		return
	}

	err = r.db.Model(&models.Event{}).Where("status = ?", models.EventStatusPublished).Count(&publishedEvents).Error
	if err != nil {
		return
	}

	err = r.db.Model(&models.Order{}).Count(&totalOrders).Error
	if err != nil {
		return
	}

	err = r.db.Model(&models.Order{}).Where("payment_status IN ('paid', 'free')").Count(&paidOrders).Error
	if err != nil {
		return
	}

	err = r.db.Model(&models.Order{}).Where("payment_status = 'paid'").Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue).Error
	if err != nil {
		return
	}

	err = r.db.Model(&models.Ticket{}).Count(&totalTicketsSold).Error
	if err != nil {
		return
	}

	err = r.db.Model(&models.Ticket{}).Where("status = ?", models.TicketStatusCheckedIn).Count(&totalCheckedIn).Error
	return
}
