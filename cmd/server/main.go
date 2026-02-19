// @title POS Retail Backend API
// @version 1.0
// @description Go (Fiber) implementation of POS Retail Backend.
// @BasePath /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/pos-retail/go_backend/internal/config"
	"github.com/pos-retail/go_backend/internal/database"
	apidocs "github.com/pos-retail/go_backend/internal/docs"
	"github.com/pos-retail/go_backend/internal/handlers"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
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
		&models.IncomingInvoice{},
		&models.OutgoingInvoice{},
		&models.InvoiceItem{},
		&models.InvoicePayment{},
		&models.CashDrawer{},
		&models.CashDrawerTransaction{},
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
	financeRepo := repository.NewFinanceRepository(db)
	cashDrawerRepo := repository.NewCashDrawerRepository(db)

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
	financeService := services.NewFinanceService(db, financeRepo)
	cashDrawerService := services.NewCashDrawerService(db, cashDrawerRepo, financeService)
	companyService := services.NewCompanyService(db)

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
	financeHandler := handlers.NewFinanceHandler(financeService)
	cashDrawerHandler := handlers.NewCashDrawerHandler(cashDrawerService)
	companyHandler := handlers.NewCompanyHandler(companyService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtUtil)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		BodyLimit:    cfg.Server.BodyLimit,
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
	if cfg.Security.SecureHeadersEnabled {
		app.Use(helmet.New())
	}
	if cfg.Perf.CompressionEnabled {
		app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	}
	if cfg.Security.RateLimitEnabled {
		app.Use(limiter.New(limiter.Config{Max: cfg.Security.RateLimitMax, Expiration: cfg.Security.RateLimitWindow}))
	}

	// API docs (public)
	apidocs.Register(app)

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
	categories := protected.Group("/categories")
	categories.Get("/", productHandler.GetCategories)
	categories.Get("/:id", productHandler.GetCategory)
	categoriesAdmin := categories.Group("", middleware.RoleMiddleware("admin", "manager"))
	categoriesAdmin.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateCategoryRequest{} }), productHandler.CreateCategory)
	categoriesAdmin.Put("/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateCategoryRequest{} }), productHandler.UpdateCategory)
	categoriesAdmin.Delete("/:id", productHandler.DeleteCategory)

	// Unit routes
	units := protected.Group("/units")
	units.Get("/", productHandler.GetUnits)
	units.Get("/:id", productHandler.GetUnit)
	units.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateUnitRequest{} }), productHandler.CreateUnit)
	units.Put("/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateUnitRequest{} }), productHandler.UpdateUnit)
	units.Delete("/:id", productHandler.DeleteUnit)

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

	// Company routes
	companies := protected.Group("/companies")
	companies.Get("/current", companyHandler.GetCurrentCompany)
	companies.Get("/:id", companyHandler.GetCompany)
	companiesAdmin := companies.Group("", middleware.RoleMiddleware("admin"))
	companiesAdmin.Get("/", companyHandler.GetCompanies)
	companiesAdmin.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateCompanyRequest{} }), companyHandler.CreateCompany)
	companiesAdmin.Put("/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateCompanyRequest{} }), companyHandler.UpdateCompany)
	companiesAdmin.Delete("/:id", companyHandler.DeleteCompany)
	companiesAdmin.Post("/:id/logo", middleware.ValidateBody(func() interface{} { return &request.UploadCompanyLogoRequest{} }), companyHandler.UploadCompanyLogo)

	// Finance routes
	invoices := protected.Group("/invoices")
	invoices.Get("/incoming", financeHandler.GetIncomingInvoices)
	invoices.Get("/incoming/:id", financeHandler.GetIncomingInvoice)
	invoices.Post("/incoming", middleware.ValidateBody(func() interface{} { return &request.CreateIncomingInvoiceRequest{} }), financeHandler.CreateIncomingInvoice)
	invoices.Put("/incoming/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateIncomingInvoiceRequest{} }), financeHandler.UpdateIncomingInvoice)
	invoices.Post("/incoming/:id/send", financeHandler.SendIncomingInvoice)
	invoices.Post("/incoming/:id/payments", middleware.ValidateBody(func() interface{} { return &request.CreateInvoicePaymentRequest{} }), financeHandler.AddIncomingInvoicePayment)
	invoices.Post("/incoming/:id/cancel", financeHandler.CancelIncomingInvoice)

	invoices.Get("/outgoing", financeHandler.GetOutgoingInvoices)
	invoices.Get("/outgoing/:id", financeHandler.GetOutgoingInvoice)
	invoices.Post("/outgoing", middleware.ValidateBody(func() interface{} { return &request.CreateOutgoingInvoiceRequest{} }), financeHandler.CreateOutgoingInvoice)
	invoices.Put("/outgoing/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateOutgoingInvoiceRequest{} }), financeHandler.UpdateOutgoingInvoice)
	invoices.Post("/outgoing/:id/send", financeHandler.SendOutgoingInvoice)
	invoices.Post("/outgoing/:id/payments", middleware.ValidateBody(func() interface{} { return &request.CreateInvoicePaymentRequest{} }), financeHandler.AddOutgoingInvoicePayment)
	invoices.Post("/outgoing/:id/cancel", financeHandler.CancelOutgoingInvoice)

	invoices.Get("/summary", financeHandler.GetInvoiceSummary)

	// Cash drawer routes
	drawers := protected.Group("/cash-drawers")
	drawers.Post("/open", middleware.ValidateBody(func() interface{} { return &request.OpenCashDrawerRequest{} }), cashDrawerHandler.OpenCashDrawer)
	drawers.Get("/current", cashDrawerHandler.GetCurrentDrawer)
	drawers.Post("/:id/cash-in", middleware.ValidateBody(func() interface{} { return &request.CashInOutRequest{} }), cashDrawerHandler.CashIn)
	drawers.Post("/:id/cash-out", middleware.ValidateBody(func() interface{} { return &request.CashInOutRequest{} }), cashDrawerHandler.CashOut)
	drawers.Get("/:id/transactions", cashDrawerHandler.GetTransactions)
	drawers.Post("/:id/close", middleware.ValidateBody(func() interface{} { return &request.CloseCashDrawerRequest{} }), cashDrawerHandler.CloseCashDrawer)
	drawers.Get("/:id/summary", cashDrawerHandler.GetSummary)
	drawers.Get("/", cashDrawerHandler.ListCashDrawers)
	drawers.Get("/:id", cashDrawerHandler.GetCashDrawer)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listenErr := make(chan error, 1)
	go func() {
		err := app.Listen(addr)
		listenErr <- err
	}()

	select {
	case <-ctx.Done():
		log.Printf("Shutdown signal received")
	case err := <-listenErr:
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
		return
	}

	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	if err := database.Close(); err != nil {
		log.Printf("Database close error: %v", err)
	}
}
