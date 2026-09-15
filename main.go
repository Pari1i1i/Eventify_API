package main

import (
	"log"
	"os"
	"time"

	"eventifyApi/config"
	"eventifyApi/cron"
	_ "eventifyApi/docs"
	"eventifyApi/handlers"
	"eventifyApi/middlewares"
	"eventifyApi/models"
	"eventifyApi/repositories"
	"eventifyApi/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Eventify API
// @version 1.0
// @description Production-Ready RESTful API for Eventify Platform (Web Admin, Panitia Organizer, Customer Portal, & Flutter Mobile QR Scanner).
// @termsOfService http://swagger.io/terms/

// @contact.name Eventify Support Team
// @contact.email support@eventify.local

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token. Example: "Bearer eyJhbGciOi..."
func main() {
	cfg := config.LoadConfig()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	_ = os.MkdirAll(cfg.UploadDir, os.ModePerm)

	db, err := config.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("Fatal Database Error: %v\nPlease make sure MySQL is running and the database '%s' exists.", err, cfg.DBName)
	}

	// Auto-migrate additional runtime tables (e.g. password reset tokens)
	_ = db.AutoMigrate(&models.PasswordResetToken{})

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	eventRepo := repositories.NewEventRepository(db)
	orderRepo := repositories.NewOrderRepository(db)

	// Services
	authService := services.NewAuthService(userRepo, cfg)
	eventService := services.NewEventService(eventRepo, cfg)
	orderService := services.NewOrderService(orderRepo, eventRepo, userRepo, cfg)

	// Start Background Cron Worker to sweep and expire unpaid stale orders every 1 minute
	cronWorker := cron.NewWorker(orderService)
	cronWorker.Start(1 * time.Minute)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	eventHandler := handlers.NewEventHandler(eventService, cfg)
	orderHandler := handlers.NewOrderHandler(orderService)

	router := gin.Default()
	router.Use(middlewares.CORSMiddleware())

	// Serve uploaded banner files
	router.Static("/uploads", cfg.UploadDir)

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": cfg.AppName,
			"env":     cfg.AppEnv,
		})
	})

	apiV1 := router.Group("/api/v1")
	{
		// ------------------ Authentication (Public) ------------------
		authRoutes := apiV1.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/forgot-password", authHandler.ForgotPassword)
			authRoutes.POST("/reset-password", authHandler.ResetPassword)
			authRoutes.GET("/me", middlewares.AuthMiddleware(cfg), authHandler.GetProfile)
			authRoutes.PUT("/me", middlewares.AuthMiddleware(cfg), authHandler.UpdateProfile)
			authRoutes.PUT("/change-password", middlewares.AuthMiddleware(cfg), authHandler.ChangePassword)
		}

		// ------------------ Events (Public - Strictly Published Events Only) ------------------
		publicEventRoutes := apiV1.Group("/events")
		{
			publicEventRoutes.GET("", eventHandler.GetAllEvents)
			publicEventRoutes.GET("/:slug", eventHandler.GetEventBySlug)
			publicEventRoutes.GET("/id/:id", eventHandler.GetEventByID)
		}

		// ------------------ Payment Webhook (Midtrans / Xendit Signature Verified) ------------------
		apiV1.POST("/payments/webhook", orderHandler.PaymentWebhook)
		apiV1.POST("/payments/notification", orderHandler.PaymentNotification)

		// ------------------ Authenticated User / Customer Routes ------------------
		authenticated := apiV1.Group("")
		authenticated.Use(middlewares.AuthMiddleware(cfg))
		{
			// Orders
			authenticated.POST("/orders", orderHandler.CreateOrder)
			authenticated.GET("/orders/my-orders", orderHandler.GetMyOrders)
			authenticated.GET("/orders/:code", orderHandler.GetOrderByCode)

			// Customer Tickets
			authenticated.GET("/tickets/my-tickets", orderHandler.GetMyTickets)
			authenticated.GET("/tickets/:code", orderHandler.GetTicketByCode)
		}

		// ------------------ Organizer Routes (Panitia & Admin: Scoped To Owner) ------------------
		organizer := apiV1.Group("/organizer")
		organizer.Use(middlewares.AuthMiddleware(cfg), middlewares.RequireRole(middlewares.RoleAdmin, middlewares.RolePanitia))
		{
			organizer.POST("/events", eventHandler.CreateEvent)
			organizer.GET("/my-events", eventHandler.GetMyEvents)
			organizer.GET("/events/:id", eventHandler.GetOrganizerEventDetail) // Draft preview
			organizer.PUT("/events/:id", eventHandler.UpdateEvent)
			organizer.DELETE("/events/:id", eventHandler.DeleteEvent)
			organizer.POST("/events/:id/banner", eventHandler.UploadBanner)

			// Ticket Tiers
			organizer.POST("/ticket-tiers", eventHandler.CreateTicketTier)
			organizer.PUT("/ticket-tiers/:id", eventHandler.UpdateTicketTier)
			organizer.DELETE("/ticket-tiers/:id", eventHandler.DeleteTicketTier)
		}

		// ------------------ Entrance Scanner Routes (Panitia & Admin: Event-Scoped) ------------------
		scanner := apiV1.Group("/scanner")
		scanner.Use(middlewares.AuthMiddleware(cfg), middlewares.RequireRole(middlewares.RoleAdmin, middlewares.RolePanitia))
		{
			scanner.POST("/check-in", orderHandler.CheckInTicket)
		}

		// ------------------ Admin Only Routes (Platform Super-Access & Moderation) ------------------
		admin := apiV1.Group("/admin")
		admin.Use(middlewares.AuthMiddleware(cfg), middlewares.RequireRole(middlewares.RoleAdmin))
		{
			admin.GET("/dashboard", orderHandler.GetDashboardStats)
			admin.GET("/users", authHandler.GetAllUsers)
			admin.PUT("/users/:id/role", authHandler.UpdateUserRole)
			admin.GET("/orders", orderHandler.GetAllOrders)
			admin.GET("/events", eventHandler.AdminGetAllEvents)
			admin.PUT("/events/:id/status", eventHandler.AdminUpdateEventStatus)
		}
	}

	log.Printf("%s running on http://localhost:%s", cfg.AppName, cfg.AppPort)
	log.Printf("Swagger documentation available at: http://localhost:%s/swagger/index.html", cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

