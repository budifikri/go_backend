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
	"github.com/pos-retail/go_backend/internal/repository"
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
		&models.Inventory{},
		&models.StockMovement{},
		&models.StockTransfer{},
		&models.StockTransferItem{},
		&models.StockOpname{},
		&models.StockOpnameItem{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	jwtUtil := utils.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.ExpiresIn)

	// Initialize repositories
	authService := services.NewAuthService(db, jwtUtil)
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	unitRepo := repository.NewUnitRepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)

	// Initialize services
	productService := services.NewProductService(productRepo, categoryRepo, unitRepo)
	inventoryService := services.NewInventoryService(inventoryRepo, productRepo, warehouseRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	productHandler := handlers.NewProductHandler(productService)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)

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

	// Product routes
	products := protected.Group("/products")
	products.Get("/", productHandler.GetProducts)
	products.Get("/:id", productHandler.GetProduct)
	products.Post("/", productHandler.CreateProduct)
	products.Put("/:id", productHandler.UpdateProduct)
	products.Delete("/:id", productHandler.DeleteProduct)

	// Category routes
	protected.Get("/categories", productHandler.GetCategories)

	// Unit routes
	protected.Get("/units", productHandler.GetUnits)

	// Inventory routes
	inventory := protected.Group("/inventory")
	inventory.Get("/", inventoryHandler.GetInventory)
	inventory.Get("/stock-card", inventoryHandler.GetStockCard)
	inventory.Post("/adjust", inventoryHandler.AdjustInventory)

	// Stock transfer routes
	stockTransfers := protected.Group("/stock-transfers")
	stockTransfers.Post("/", inventoryHandler.CreateStockTransfer)
	stockTransfers.Put("/:id/receive", inventoryHandler.ReceiveStockTransfer)

	// Stock opname routes
	stockOpname := protected.Group("/stock-opname")
	stockOpname.Post("/", inventoryHandler.CreateStockOpname)
	stockOpname.Get("/", inventoryHandler.GetStockOpnames)
	stockOpname.Get("/:id", inventoryHandler.GetStockOpname)
	stockOpname.Put("/:id/status", inventoryHandler.UpdateStockOpnameStatus)

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
