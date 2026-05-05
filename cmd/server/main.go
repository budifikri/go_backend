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

	apidocs "github.com/pos-retail/go_backend/docs"
	"github.com/pos-retail/go_backend/internal/config"
	"github.com/pos-retail/go_backend/internal/database"
	"github.com/pos-retail/go_backend/internal/handlers"
	applogger "github.com/pos-retail/go_backend/internal/logger"
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

	crudLogger := applogger.NewLogger(cfg.Log.LogDir, cfg.Log.EnableCRUD)
	applogger.SetDefault(crudLogger)

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
		&models.BusinessType{},
		&models.ModulePackage{},
		&models.CompanyModule{},
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
		&models.PurchaseReturn{},
		&models.PurchaseReturnItem{},
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
		&models.BackupLog{},
		&models.BackupSchedule{},
		&models.TelegramConfig{},
		&models.Dokter{},
		&models.JadwalDokter{},
		&models.Paket{},
		&models.DetailPaket{},
		&models.Appointment{},
		&models.Treatment{},
		&models.TreatmentTag{},
		&models.TreatmentTagRelation{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	if err := database.SeedModuleDefaults(db); err != nil {
		log.Fatalf("Failed to seed module defaults: %v", err)
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
	backupRepo := repository.NewBackupRepository(db)
	telegramRepo := repository.NewTelegramRepository(db)
	dokterRepo := repository.NewDokterRepository(db)
	jadwalDokterRepo := repository.NewJadwalDokterRepository(db)
	paketRepo := repository.NewPaketRepository(db)
	appointmentRepo := repository.NewAppointmentRepository(db)
	treatmentRepo := repository.NewTreatmentRepository(db)

	// Initialize services
	productService := services.NewProductService(productRepo, categoryRepo, unitRepo)
	warehouseService := services.NewWarehouseService(warehouseRepo)
	inventoryService := services.NewInventoryServiceWithTelegram(db, inventoryRepo, productRepo, warehouseRepo, purchaseRepo, telegramRepo)
	salesService := services.NewSalesServiceWithTelegram(db, salesRepo, cashDrawerRepo, telegramRepo)
	returnsService := services.NewReturnsService(db, returnsRepo)
	exchangesService := services.NewExchangesService(db, exchangesRepo)
	customerService := services.NewCustomerService(customerRepo)
	supplierService := services.NewSupplierService(supplierRepo)
	purchaseService := services.NewPurchaseServiceWithTelegram(db, purchaseRepo, inventoryRepo, telegramRepo)
	promotionService := services.NewPromotionService(db, promotionRepo)
	priceTierService := services.NewPriceTierService(db)
	financeService := services.NewFinanceService(db, financeRepo)
	cashDrawerService := services.NewCashDrawerService(db, cashDrawerRepo, financeService, telegramRepo)
	companyService := services.NewCompanyService(db)
	moduleService := services.NewModuleService(db)
	userService := services.NewUserService(db)
	testDataService := services.NewTestDataService(db)
	backupService := services.NewBackupService(db, backupRepo, cfg)
	telegramService := services.NewTelegramService(db, telegramRepo)
	dokterService := services.NewDokterService(dokterRepo)
	jadwalDokterService := services.NewJadwalDokterService(jadwalDokterRepo)
	paketService := services.NewPaketService(paketRepo)
	appointmentService := services.NewAppointmentService(appointmentRepo, customerRepo, dokterRepo, treatmentRepo)
	treatmentService := services.NewTreatmentService(treatmentRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	productHandler := handlers.NewProductHandler(productService)
	warehouseHandler := handlers.NewWarehouseHandler(warehouseService)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)
	salesHandler := handlers.NewSalesHandler(salesService)
	returnsHandler := handlers.NewReturnsHandler(returnsService)
	exchangesHandler := handlers.NewExchangesHandler(exchangesService)
	customerHandler := handlers.NewCustomerHandler(customerService)
	supplierHandler := handlers.NewSupplierHandler(supplierService)
	purchaseHandler := handlers.NewPurchaseHandler(purchaseService)
	promotionHandler := handlers.NewPromotionHandler(promotionService)
	priceTierHandler := handlers.NewPriceTierHandler(priceTierService)
	financeHandler := handlers.NewFinanceHandler(financeService)
	cashDrawerHandler := handlers.NewCashDrawerHandler(cashDrawerService)
	companyHandler := handlers.NewCompanyHandler(companyService)
	moduleHandler := handlers.NewModuleHandler(moduleService)
	userHandler := handlers.NewUserHandler(userService)
	healthHandler := handlers.NewHealthHandler(db, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	logHandler := handlers.NewLogHandler(crudLogger)
	testDataHandler := handlers.NewTestDataHandler(testDataService)
	backupHandler := handlers.NewBackupHandler(backupService)
	telegramHandler := handlers.NewTelegramHandler(telegramService)
	dokterHandler := handlers.NewDokterHandler(dokterService)
	jadwalDokterHandler := handlers.NewJadwalDokterHandler(jadwalDokterService)
	paketHandler := handlers.NewPaketHandler(paketService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)
	treatmentHandler := handlers.NewTreatmentHandler(treatmentService)

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
	api.Get("/health", healthHandler.GetHealth)

	// Protected routes
	protected := api.Group("", authMiddleware.Handler())
	// Log routes
	logs := protected.Group("/logs", middleware.RoleMiddleware("admin", "manager"))
	logs.Get("/summary", logHandler.GetSummary)
	logs.Post("/save", logHandler.SaveSummary)
	logs.Get("/files", logHandler.ListFiles)
	logs.Get("/:tahun_bulan/error", logHandler.GetErrorLogs)
	logs.Get("/:tahun_bulan/:table", logHandler.GetTableLogs)

	// User routes
	users := protected.Group("/users", middleware.RoleMiddleware("admin", "manager"))
	users.Get("/", userHandler.GetUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateUserRequest{} }), userHandler.CreateUser)
	users.Put("/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateUserRequest{} }), userHandler.UpdateUser)
	users.Patch("/:id/password", middleware.ValidateBody(func() interface{} { return &request.UpdateUserPasswordRequest{} }), userHandler.UpdateUserPassword)
	users.Delete("/:id", userHandler.DeleteUser)

	// Product routes
	products := protected.Group("/products")
	products.Get("/", productHandler.GetProducts)
	products.Get("/:id/hpp-trace", productHandler.GetProductHppTrace)
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

	// Warehouse routes
	warehouses := protected.Group("/warehouses")
	warehouses.Get("/", warehouseHandler.GetWarehouses)
	warehouses.Get("/:id", warehouseHandler.GetWarehouse)
	warehouses.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateWarehouseRequest{} }), warehouseHandler.CreateWarehouse)
	warehouses.Put("/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateWarehouseRequest{} }), warehouseHandler.UpdateWarehouse)
	warehouses.Delete("/:id", warehouseHandler.DeleteWarehouse)

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
	stockOpname.Put("/:id", inventoryHandler.UpdateStockOpname)
	stockOpname.Put("/:id/status", inventoryHandler.UpdateStockOpnameStatus)
	stockOpname.Delete("/:id", inventoryHandler.DeleteStockOpname)

	// Sales routes
	sales := protected.Group("/sales")
	sales.Post("/", salesHandler.CreateSale)
	sales.Get("/", salesHandler.GetSales)
	sales.Get("/summary", salesHandler.GetSalesSummary)
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

	// Dokter routes
	dokters := protected.Group("/dokters")
	dokters.Get("/", dokterHandler.GetDokters)
	dokters.Get("/:id", dokterHandler.GetDokter)
	dokters.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateDokterRequest{} }), dokterHandler.CreateDokter)
	dokters.Put("/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateDokterRequest{} }), dokterHandler.UpdateDokter)
	dokters.Delete("/:id", dokterHandler.DeleteDokter)

	// Jadwal Dokter routes
	jadwalDokter := protected.Group("/jadwal-dokter")
	jadwalDokter.Get("/", jadwalDokterHandler.GetJadwals)
	jadwalDokter.Get("/:id", jadwalDokterHandler.GetJadwalDokter)
	jadwalDokter.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateJadwalDokterRequest{} }), jadwalDokterHandler.CreateJadwalDokter)
	jadwalDokter.Put("/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateJadwalDokterRequest{} }), jadwalDokterHandler.UpdateJadwalDokter)
	jadwalDokter.Delete("/:id", jadwalDokterHandler.DeleteJadwalDokter)

	// Paket routes
	pakets := protected.Group("/paket")
	pakets.Get("/", paketHandler.GetPakets)
	pakets.Get("/:id", paketHandler.GetPaket)
	pakets.Post("/", paketHandler.CreatePaket)
	pakets.Put("/:id", paketHandler.UpdatePaket)
	pakets.Delete("/:id", paketHandler.DeletePaket)

	// Appointment routes
	appointments := protected.Group("/appointments")
	appointments.Get("/", appointmentHandler.GetAppointments)
	appointments.Get("/:id", appointmentHandler.GetAppointment)
	appointments.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateAppointmentRequest{} }), appointmentHandler.CreateAppointment)
	appointments.Put("/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateAppointmentRequest{} }), appointmentHandler.UpdateAppointment)
	appointments.Delete("/:id", appointmentHandler.DeleteAppointment)

	// Treatment routes
	treatments := protected.Group("/treatments")
	treatments.Get("/", treatmentHandler.GetTreatments)
	treatments.Get("/:id", treatmentHandler.GetTreatment)
	treatments.Post("/", treatmentHandler.CreateTreatment)
	treatments.Put("/:id", treatmentHandler.UpdateTreatment)
	treatments.Delete("/:id", treatmentHandler.DeleteTreatment)

	// Treatment Tags routes
	treatmentTags := protected.Group("/treatment-tags")
	treatmentTags.Get("/", treatmentHandler.GetTags)
	treatmentTags.Post("/", treatmentHandler.CreateTag)
	treatmentTags.Put("/:id", treatmentHandler.UpdateTag)
	treatmentTags.Delete("/:id", treatmentHandler.DeleteTag)

	// Purchase order routes (TS parity: /api/purchases)
	purchases := protected.Group("/purchases")
	purchases.Get("/", purchaseHandler.GetPurchaseOrders)
	purchases.Get("/:id", purchaseHandler.GetPurchaseOrder)
	purchases.Post("/", purchaseHandler.CreatePurchaseOrder)
	purchases.Put("/:id", purchaseHandler.UpdatePurchaseOrder)
	purchases.Put("/:id/status", purchaseHandler.UpdatePurchaseOrderStatus)
	purchases.Put("/:id/approve", purchaseHandler.ApprovePurchaseOrder)
	purchases.Put("/:id/pending", purchaseHandler.SetPendingPurchaseOrder)
	purchases.Put("/:id/receive", purchaseHandler.ReceivePurchaseOrder)
	purchases.Post("/:id/void", purchaseHandler.VoidPurchaseOrder)
	purchases.Post("/:id/cancel", purchaseHandler.CancelPurchaseOrder)
	purchases.Delete("/:id", purchaseHandler.DeletePurchaseOrder)

	// Purchase returns routes
	purchaseReturns := protected.Group("/purchase-returns")
	purchaseReturns.Get("/", purchaseHandler.GetPurchaseReturns)
	purchaseReturns.Get("/:id", purchaseHandler.GetPurchaseReturn)
	purchaseReturns.Post("/", purchaseHandler.CreatePurchaseReturn)
	purchaseReturns.Put("/:id", purchaseHandler.UpdatePurchaseReturn)
	purchaseReturns.Put("/:id/status", purchaseHandler.UpdatePurchaseReturnStatus)
	purchaseReturns.Delete("/:id", purchaseHandler.DeletePurchaseReturn)

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
	priceTiers.Get("/report/by-product", priceTierHandler.GetPriceTierReportByProduct)
	priceTiers.Get("/:id", priceTierHandler.GetPriceTier)
	priceTiers.Post("/", priceTierHandler.CreatePriceTier)
	priceTiers.Put("/:id", priceTierHandler.UpdatePriceTier)
	priceTiers.Delete("/:id", priceTierHandler.DeletePriceTier)
	priceTiers.Get("/product/:product_id", priceTierHandler.GetPriceTiersByProduct)
	priceTiers.Post("/product/:product_id", priceTierHandler.SaveProductPriceTiers)

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

	// Business type routes
	businessTypes := protected.Group("/business-types", middleware.RoleMiddleware("admin"))
	businessTypes.Get("/", moduleHandler.GetBusinessTypes)
	businessTypes.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateBusinessTypeRequest{} }), moduleHandler.CreateBusinessType)
	businessTypes.Put("/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateBusinessTypeRequest{} }), moduleHandler.UpdateBusinessType)

	// Module package routes
	modulePackages := protected.Group("/module-packages", middleware.RoleMiddleware("admin"))
	modulePackages.Get("/", moduleHandler.GetModulePackages)
	modulePackages.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateModulePackageRequest{} }), moduleHandler.CreateModulePackage)
	modulePackages.Put("/:id", middleware.ValidateBody(func() interface{} { return &request.UpdateModulePackageRequest{} }), moduleHandler.UpdateModulePackage)

	// Company module routes
	protected.Get("/me/modules", moduleHandler.GetMyModules)
	companiesAdmin.Get("/:id/modules", moduleHandler.GetCompanyModules)
	companiesAdmin.Patch("/:id/modules/:code/toggle", middleware.ValidateBody(func() interface{} { return &request.ToggleCompanyModuleRequest{} }), moduleHandler.ToggleCompanyModule)

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

	// Remove data routes
	removeData := protected.Group("/remove-data", middleware.RoleMiddleware("admin"))
	removeData.Delete("/master", testDataHandler.DeleteMasterData)
	removeData.Delete("/transactions", testDataHandler.DeleteTransactionData)
	removeData.Delete("/table", testDataHandler.DeleteTableData)

	// Backup routes
	backup := protected.Group("/backup", middleware.RoleMiddleware("admin", "manager"))
	backup.Post("/", backupHandler.CreateBackup)
	backup.Get("/list", backupHandler.ListBackups)
	backup.Get("/download/:filename", backupHandler.DownloadBackup)
	backup.Delete("/:filename", backupHandler.DeleteBackup)
	backup.Get("/schedule", backupHandler.GetSchedule)
	backup.Post("/schedule", backupHandler.UpdateSchedule)
	backup.Post("/delete", backupHandler.DeleteData)
	backup.Get("/count/:scope", backupHandler.GetTableCounts)

	// Telegram routes
	telegram := protected.Group("/telegram")
	telegram.Get("/", telegramHandler.GetConfig)
	telegram.Post("/", telegramHandler.SaveConfig)
	telegram.Post("/test", telegramHandler.TestConnection)

	// Restore routes
	restore := protected.Group("/restore", middleware.RoleMiddleware("admin"))
	restore.Post("/validate", backupHandler.ValidateRestore)
	restore.Post("/", backupHandler.RestoreBackup)
	restore.Get("/progress", backupHandler.RestoreProgress)

	// Health check
	app.Get("/health", healthHandler.GetHealth)

	// Start backup scheduler
	go backupService.StartAllSchedules()
	defer backupService.Stop()

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
	if err := crudLogger.SaveSummary(); err != nil {
		log.Printf("Failed to save log summary: %v", err)
	}
	if err := database.Close(); err != nil {
		log.Printf("Database close error: %v", err)
	}
}
