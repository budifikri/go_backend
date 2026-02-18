package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/pos-retail/go_backend/internal/config"
	"github.com/pos-retail/go_backend/internal/database"
	"github.com/pos-retail/go_backend/internal/handlers"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate models
	if err := database.AutoMigrate(
		&models.User{},
		&models.UserSession{},
		&models.EmailVerification{},
		&models.PasswordReset{},
		&models.Company{},
		&models.Unit{},
		&models.Category{},
		&models.Warehouse{},
		&models.Product{},
		&models.PriceTier{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	jwtUtil := utils.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.ExpiresIn)

	// Initialize services
	authService := services.NewAuthService(db, jwtUtil)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtUtil)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Public routes
	api := app.Group("/api")
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/register", authHandler.Register)
	auth.Post("/logout", authHandler.Logout)

	// Protected routes
	protected := api.Group("", authMiddleware.Handler())
	protected.Get("/users", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true, "message": "Users list"})
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
