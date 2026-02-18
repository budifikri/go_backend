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
		&models.Customer{},
		&models.Sale{},
		&models.SaleItem{},
		&models.SalePayment{},
		&models.SalesReturn{},
		&models.SalesReturnItem{},
		&models.ItemExchange{},
		&models.ExchangeItem{},
		&models.Supplier{},
		&models.PurchaseOrder{},
		&models.PurchaseOrderItem{},
		&models.GoodsReceivedNote{},
		&models.GrnItem{},
		&models.Promotion{},
		&models.PromotionProduct{},
		&models.PromotionCategory{},
		&models.PromotionCustomer{},
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
	salesRepo := repository.NewSalesRepository(db)
	returnsRepo := repository.NewReturnsRepository(db)
	exchangesRepo := repository.NewExchangesRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	purchaseRepo := repository.NewPurchaseRepository(db)
	promotionRepo := repository.NewPromotionRepository(db)

	// Initialize services
	productService := services.NewProductService(productRepo, categoryRepo, unitRepo)
	inventoryService := services.NewInventoryService(inventoryRepo, productRepo, warehouseRepo)
	salesService := services.NewSalesService(db, salesRepo)
	returnsService := services.NewReturnsService(db, returnsRepo)
	exchangesService := services.NewExchangesService(db, exchangesRepo)
	customerService := services.NewCustomerService(customerRepo)
	supplierService := services.NewSupplierService(supplierRepo)
	purchaseService := services.NewPurchaseService(db, purchaseRepo)
	grnService := services.NewGrnService(db)
	promotionService := services.NewPromotionService(db, promotionRepo)
	priceTierService := services.NewPriceTierService(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	productHandler := handlers.NewProductHandler(productService)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)
	salesHandler := handlers.NewSalesHandler(salesService)
	returnsHandler := handlers.NewReturnsHandler(returnsService)
	exchangesHandler := handlers.NewExchangesHandler(exchangesService)
	customerHandler := handlers.NewCustomerHandler(customerService)
	supplierHandler := handlers.NewSupplierHandler(supplierService)
	purchaseHandler := handlers.NewPurchaseHandler(purchaseService)
	grnHandler := handlers.NewGrnHandler(grnService)
	promotionHandler := handlers.NewPromotionHandler(promotionService)
	priceTierHandler := handlers.NewPriceTierHandler(priceTierService)

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

	// Sales routes
	sales := protected.Group("/sales")
	sales.Post("/", salesHandler.CreateSale)
	sales.Get("/", salesHandler.GetSales)
	sales.Get("/:id", salesHandler.GetSale)

	// Returns routes
	returns := protected.Group("/returns")
	returns.Post("/", returnsHandler.CreateReturn)
	returns.Get("/", returnsHandler.GetReturns)
	returns.Get("/:id", returnsHandler.GetReturn)

	// Exchanges routes
	exchanges := protected.Group("/exchanges")
	exchanges.Post("/", exchangesHandler.CreateExchange)
	exchanges.Get("/", exchangesHandler.GetExchanges)
	exchanges.Get("/:id", exchangesHandler.GetExchange)

	// Customer routes (static paths first to avoid :id collisions)
	customers := protected.Group("/customers")
	customers.Get("/search/:term", customerHandler.SearchCustomers)
	customers.Get("/tier/:tier", customerHandler.GetCustomersByTier)
	customers.Get("/status/:status", customerHandler.GetCustomersByStatus)
	customers.Post("/", customerHandler.CreateCustomer)
	customers.Get("/", customerHandler.GetCustomers)
	customers.Get("/:id", customerHandler.GetCustomer)
	customers.Put("/:id", customerHandler.UpdateCustomer)
	customers.Delete("/:id", customerHandler.DeleteCustomer)

	// Supplier routes (static paths first)
	suppliers := protected.Group("/suppliers")
	suppliers.Get("/search/:term", supplierHandler.SearchSuppliers)
	suppliers.Get("/status/:status", supplierHandler.GetSuppliersByStatus)
	suppliers.Get("/payment-terms/:terms", supplierHandler.GetSuppliersByPaymentTerms)
	suppliers.Post("/", supplierHandler.CreateSupplier)
	suppliers.Get("/", supplierHandler.GetSuppliers)
	suppliers.Get("/:id", supplierHandler.GetSupplier)
	suppliers.Put("/:id", supplierHandler.UpdateSupplier)
	suppliers.Delete("/:id", supplierHandler.DeleteSupplier)

	// Purchase order routes (TS parity: /api/purchases)
	purchases := protected.Group("/purchases")
	purchases.Get("/", purchaseHandler.GetPurchaseOrders)
	purchases.Get("/:id", purchaseHandler.GetPurchaseOrder)
	purchases.Post("/", purchaseHandler.CreatePurchaseOrder)
	purchases.Put("/:id/status", purchaseHandler.UpdatePurchaseOrderStatus)
	purchases.Delete("/:id", purchaseHandler.CancelPurchaseOrder)

	// GRN routes
	grn := protected.Group("/grn")
	grn.Post("/", grnHandler.CreateGrn)
	grn.Get("/", grnHandler.ListGrns)
	grn.Get("/:id", grnHandler.GetGrn)
	grn.Put("/:id", grnHandler.UpdateGrn)
	grn.Delete("/:id", grnHandler.CancelGrn)
	grn.Put("/:id/verify", grnHandler.VerifyGrn)

	// Promotions routes
	promotions := protected.Group("/promotions")
	promotions.Get("/", promotionHandler.GetPromotions)
	promotions.Get("/:id", promotionHandler.GetPromotion)
	promotions.Post("/", promotionHandler.CreatePromotion)
	promotions.Put("/:id", promotionHandler.UpdatePromotion)
	promotions.Delete("/:id", promotionHandler.DeletePromotion)

	// Price tier routes
	priceTiers := protected.Group("/price-tiers")
	priceTiers.Get("/", priceTierHandler.GetPriceTiers)
	priceTiers.Get("/:id", priceTierHandler.GetPriceTier)
	priceTiers.Post("/", priceTierHandler.CreatePriceTier)
	priceTiers.Put("/:id", priceTierHandler.UpdatePriceTier)
	priceTiers.Delete("/:id", priceTierHandler.DeletePriceTier)

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
