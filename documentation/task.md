# POS Retail Backend - Go Implementation

## Task Plan

---

## 1. Project Setup & Bootstrapping

### 1.1 Initialize Go Project
- [ ] Create `go.mod` with module name and Go version
- [ ] Install core dependencies (Fiber, GORM, pgx, JWT, validator, godotenv)
- [ ] Create `.env.example` file with all required variables
- [ ] Set up folder structure (cmd, internal, documentation)

### 1.2 Configuration
- [ ] Create `internal/config/config.go` with Config struct
- [ ] Implement environment variable loading
- [ ] Create database configuration loader

### 1.3 Database Connection
- [ ] Create `internal/database/database.go`
- [ ] Implement GORM PostgreSQL connection with connection pooling
- [ ] Add auto-migration support for all models
- [ ] Test database connection

**Subtasks:**
- Configure `DBMaxConnections` from env
- Handle connection errors gracefully
- Add connection health check endpoint

---

## 2. Models (Database Schema)

### 2.1 Core Models
- [ ] Create `internal/models/core.go`
  - [ ] User model (id, username, email, password_hash, full_name, role, status, company_id, created_at, updated_at, last_login)
  - [ ] UserSession model
  - [ ] EmailVerification model
  - [ ] PasswordReset model
- [ ] Add GORM tags matching PostgreSQL schema
- [ ] Add enum support for role and status

### 2.2 Company Model
- [ ] Create `internal/models/company.go`
- [ ] Company model (id, code, name, address, phone, email, tax_id, logo, status, created_at, updated_at)

### 2.3 Product Models
- [ ] Create `internal/models/product.go`
- [ ] Unit model (id, code, name, description, is_active)
- [ ] Category model (id, code, name, description, parent_id, company_id, is_active, created_at)
- [ ] Product model (id, sku, barcode, name, description, category_id, unit_id, cost_price, retail_price, status, tax_rate, is_trackable, reorder_point, company_id, created_at, updated_at)
- [ ] PriceTier model (id, product_id, tier_name, min_quantity, max_quantity, unit_price, is_active, created_at)

### 2.4 Warehouse Models
- [ ] Add warehouse model to product.go or create separate file
- [ ] Warehouse fields: id, code, name, type, status, address, city, phone, manager_id, company_id, created_at, updated_at

### 2.5 Inventory Models
- [ ] Create `internal/models/inventory.go`
- [ ] Inventory model (id, product_id, warehouse_id, quantity, reserved_quantity, available_quantity, min_stock_level, max_stock_level, last_stock_take)
- [ ] StockMovement model (id, product_id, warehouse_id, movement_type, quantity, reference_type, reference_id, notes, created_at)
- [ ] StockTransfer model (id, transfer_number, from_warehouse_id, to_warehouse_id, user_id, expected_arrival, actual_arrival, status, notes, created_at, updated_at)
- [ ] StockTransferItem model
- [ ] StockOpname model (id, opname_number, warehouse_id, user_id, opname_date, status, notes, created_at, updated_at)
- [ ] StockOpnameItem model

### 2.6 Sales Models
- [ ] Create `internal/models/sales.go`
- [ ] Customer model (id, code, name, email, phone, address, loyalty_points, tier, status, company_id, created_at, updated_at)
- [ ] Sale model (id, sale_number, warehouse_id, customer_id, cashier_id, sale_date, status, subtotal, discount_amount, tax_amount, total_amount, paid_amount, change_amount, loyalty_points_earned, loyalty_points_redeemed, notes, created_at, updated_at)
- [ ] SaleItem model (id, sale_id, product_id, quantity, unit_price, original_price, discount_amount, tax_rate, line_total, price_tier_id, promotion_id)
- [ ] SalePayment model (id, sale_id, method, amount, reference_number, card_last_4, status)
- [ ] SalesReturn model (id, return_number, sale_id, warehouse_id, customer_id, return_date, status, reason, total_amount, refund_method, processed_by, created_at, updated_at)
- [ ] SalesReturnItem model
- [ ] ItemExchange model (id, exchange_number, sale_id, warehouse_id, customer_id, exchange_date, status, reason, total_returned_value, total_received_value, price_difference, processed_by, created_at, updated_at)
- [ ] ExchangeItem model

### 2.7 Purchase Models
- [ ] Create `internal/models/purchase.go`
- [ ] Supplier model (id, code, name, email, phone, address, contact_person, notes, status, company_id, created_at, updated_at)
- [ ] PurchaseOrder model (id, po_number, supplier_id, warehouse_id, expected_delivery, payment_terms, status, total_amount, notes, created_by, created_at, updated_at)
- [ ] PurchaseOrderItem model
- [ ] GoodsReceivedNote model (id, grn_number, po_id, supplier_id, warehouse_id, received_date, status, notes, received_by, created_at, updated_at)
- [ ] GRNItem model

### 2.8 Promotion Models
- [ ] Create `internal/models/promotion.go`
- [ ] Promotion model (id, code, name, type, scope, value, min_purchase, max_discount, start_date, end_date, usage_limit, usage_count, is_active, created_at, updated_at)
- [ ] PromotionProduct model
- [ ] PromotionCategory model
- [ ] PromotionCustomer model

### 2.9 Finance Models
- [ ] Create `internal/models/finance.go`
- [ ] InvoiceIncoming model (id, invoice_number, supplier_id, grn_id, amount, tax_amount, total_amount, due_date, status, notes, created_at, updated_at)
- [ ] InvoiceOutgoing model (id, invoice_number, customer_id, sale_id, amount, tax_amount, total_amount, due_date, status, notes, created_at, updated_at)
- [ ] InvoiceItem model
- [ ] InvoicePayment model

### 2.10 Cash Drawer Models
- [ ] Create `internal/models/cashdrawer.go`
- [ ] CashDrawer model (id, warehouse_id, user_id, opening_balance, closing_balance, expected_balance, status, opened_at, closed_at, created_at, updated_at)
- [ ] CashDrawerTransaction model (id, cash_drawer_id, type, amount, notes, created_at)

---

## 3. Utilities & Types

### 3.1 Response Types
- [ ] Create `internal/types/response/api_response.go`
- [ ] ApiResponse struct
- [ ] PaginatedResponse struct
- [ ] ErrorResponse struct

### 3.2 Request Types
- [ ] Create `internal/types/request/auth_request.go`
- [ ] Create `internal/types/request/product_request.go`
- [ ] Create `internal/types/request/sales_request.go`
- [ ] Create `internal/types/request/inventory_request.go`
- [ ] Create `internal/types/request/purchase_request.go`
- [ ] Add validation tags using go-playground/validator

### 3.3 Utility Functions
- [ ] Create `internal/utils/jwt.go` (generate, verify, payload)
- [ ] Create `internal/utils/password.go` (hash, compare using bcrypt)
- [ ] Create `internal/utils/helpers.go` (UUID generation, time helpers)
- [ ] Create `internal/utils/validation.go` (custom validators)

---

## 4. Middleware

### 4.1 Authentication Middleware
- [ ] Create `internal/middleware/auth_middleware.go`
- [ ] JWT verification function
- [ ] Extract user info from token
- [ ] Add user to Locals context
- [ ] Handle expired/invalid tokens

### 4.2 Role-Based Access
- [ ] Create role checking middleware
- [ ] Support for admin, manager, cashier, staff roles

### 4.3 CORS Middleware
- [ ] Configure CORS for allowed origins
- [ ] Handle preflight requests

### 4.4 Logging Middleware
- [ ] Request/response logging
- [ ] Error logging
- [ ] Performance timing

### 4.5 Validation Middleware
- [ ] Request body validation
- [ ] Query parameter validation

---

## 5. Repository Layer

### 5.1 Auth Repository
- [ ] Create `internal/repository/auth_repository.go`
- [ ] FindUserByUsername
- [ ] FindUserByEmail
- [ ] FindUserByID
- [ ] CreateUser
- [ ] UpdateUser
- [ ] FindSessionByToken
- [ ] CreateSession
- [ ] DeleteSession

### 5.2 Product Repository
- [ ] Create `internal/repository/product_repository.go`
- [ ] FindProducts (with filters: status, category_id, search, company_id)
- [ ] FindProductByID
- [ ] CreateProduct
- [ ] UpdateProduct
- [ ] DeleteProduct (soft delete)
- [ ] FindCategories
- [ ] FindCategoryByID
- [ ] CreateCategory
- [ ] FindUnits
- [ ] FindPriceTiersByProductID
- [ ] CreatePriceTier

### 5.3 Warehouse Repository
- [ ] Create `internal/repository/warehouse_repository.go`
- [ ] FindWarehouses
- [ ] FindWarehouseByID
- [ ] CreateWarehouse
- [ ] UpdateWarehouse

### 5.4 Inventory Repository
- [ ] Create `internal/repository/inventory_repository.go`
- [ ] FindInventory (filters: warehouse_id, product_id, status)
- [ ] FindInventoryByProductWarehouse
- [ ] UpdateInventoryQuantity
- [ ] CreateStockMovement
- [ ] GetStockCard (product_id, warehouse_id, date range)
- [ ] FindStockTransfers
- [ ] CreateStockTransfer
- [ ] UpdateStockTransferStatus
- [ ] FindStockOpnames
- [ ] CreateStockOpname

### 5.5 Sales Repository
- [ ] Create `internal/repository/sales_repository.go`
- [ ] FindSales (filters: warehouse_id, customer_id, cashier_id, status, date range)
- [ ] FindSaleByID
- [ ] CreateSale (with items and payments)
- [ ] UpdateSaleStatus
- [ ] FindSaleItems
- [ ] FindSalePayments

### 5.6 Returns Repository
- [ ] Create `internal/repository/returns_repository.go`
- [ ] FindReturns
- [ ] FindReturnByID
- [ ] CreateReturn

### 5.7 Exchanges Repository
- [ ] Create `internal/repository/exchanges_repository.go`
- [ ] FindExchanges
- [ ] FindExchangeByID
- [ ] CreateExchange

### 5.8 Purchase Repository
- [ ] Create `internal/repository/purchase_repository.go`
- [ ] FindPurchaseOrders
- [ ] FindPOByID
- [ ] CreatePurchaseOrder
- [ ] UpdatePOStatus
- [ ] FindGRNs
- [ ] FindGRNByID
- [ ] CreateGRN

### 5.9 Customer Repository
- [ ] Create `internal/repository/customer_repository.go`
- [ ] FindCustomers
- [ ] FindCustomerByID
- [ ] CreateCustomer
- [ ] UpdateCustomer
- [ ] UpdateLoyaltyPoints

### 5.10 Supplier Repository
- [ ] Create `internal/repository/supplier_repository.go`
- [ ] FindSuppliers
- [ ] FindSupplierByID
- [ ] CreateSupplier
- [ ] UpdateSupplier

### 5.11 Promotion Repository
- [ ] Create `internal/repository/promotion_repository.go`
- [ ] FindActivePromotions
- [ ] FindPromotionByCode
- [ ] CreatePromotion

### 5.12 Finance Repository
- [ ] Create `internal/repository/finance_repository.go`
- [ ] FindInvoices
- [ ] CreateInvoice

### 5.13 Cash Drawer Repository
- [ ] Create `internal/repository/cashdrawer_repository.go`
- [ ] FindCashDrawers
- [ ] FindCashDrawerByID
- [ ] CreateCashDrawer
- [ ] OpenCashDrawer
- [ ] CloseCashDrawer
- [ ] AddTransaction

---

## 6. Service Layer

### 6.1 Auth Service
- [ ] Create `internal/services/auth_service.go`
- [ ] Login (validate credentials, generate JWT, create session)
- [ ] Register (validate, hash password, create user)
- [ ] Logout (invalidate session)
- [ ] VerifyEmail
- [ ] ForgotPassword
- [ ] ResetPassword
- [ ] ValidateToken

### 6.2 Product Service
- [ ] Create `internal/services/product_service.go`
- [ ] GetProducts (with pagination, filters)
- [ ] GetProductByID (with price tiers, promotions)
- [ ] CreateProduct (validate SKU uniqueness)
- [ ] UpdateProduct
- [ ] UpdateProductPrice
- [ ] UpdateProductStock
- [ ] UpdateProductStatus
- [ ] DeleteProduct (soft delete)
- [ ] GetCategories
- [ ] GetUnits

### 6.3 Warehouse Service
- [ ] Create `internal/services/warehouse_service.go`
- [ ] GetWarehouses
- [ ] GetWarehouseByID
- [ ] CreateWarehouse
- [ ] UpdateWarehouse

### 6.4 Inventory Service
- [ ] Create `internal/services/inventory_service.go`
- [ ] GetInventory (with filters)
- [ ] AdjustInventory
- [ ] GetStockCard
- [ ] CreateStockTransfer
- [ ] ReceiveStockTransfer
- [ ] CreateStockOpname
- [ ] GetStockOpnames
- [ ] UpdateStockOpnameStatus

### 6.5 Sales Service
- [ ] Create `internal/services/sales_service.go`
- [ ] CreateSale (validate stock, calculate prices, apply promotions, create records)
- [ ] GetSales
- [ ] GetSaleByID
- [ ] ValidateStockAvailability

### 6.6 Returns Service
- [ ] Create `internal/services/returns_service.go`
- [ ] CreateReturn (validate, restore inventory, process refund)
- [ ] GetReturns
- [ ] GetReturnByID

### 6.7 Exchanges Service
- [ ] Create `internal/services/exchanges_service.go`
- [ ] CreateExchange (calculate difference, handle inventory)
- [ ] GetExchanges
- [ ] GetExchangeByID

### 6.8 Purchase Service
- [ ] Create `internal/services/purchase_service.go`
- [ ] CreatePurchaseOrder
- [ ] GetPurchaseOrders
- [ ] GetPOByID
- [ ] CreateGRN
- [ ] GetGRNs
- [ ] GetGRNByID

### 6.9 Customer Service
- [ ] Create `internal/services/customer_service.go`
- [ ] GetCustomers
- [ ] GetCustomerByID
- [ ] CreateCustomer
- [ ] UpdateCustomer

### 6.10 Supplier Service
- [ ] Create `internal/services/supplier_service.go`
- [ ] GetSuppliers
- [ ] GetSupplierByID
- [ ] CreateSupplier
- [ ] UpdateSupplier

### 6.11 Promotion Service
- [ ] Create `internal/services/promotion_service.go`
- [ ] GetActivePromotions
- [ ] ApplyPromotion

### 6.12 Finance Service
- [ ] Create `internal/services/finance_service.go`
- [ ] GetInvoices
- [ ] CreateInvoice

### 6.13 Cash Drawer Service
- [ ] Create `internal/services/cashdrawer_service.go`
- [ ] GetCashDrawers
- [ ] OpenCashDrawer
- [ ] CloseCashDrawer
- [ ] AddTransaction

---

## 7. Handlers (HTTP Layer)

### 7.1 Auth Handler
- [ ] Create `internal/handlers/auth_handler.go`
- [ ] POST /api/auth/login
- [ ] POST /api/auth/register
- [ ] POST /api/auth/logout
- [ ] GET /api/auth/verify-email
- [ ] POST /api/auth/forgot-password
- [ ] GET /api/auth/reset-password
- [ ] POST /api/auth/reset-password

### 7.2 User Handler
- [ ] GET /api/users (list users)
- [ ] GET /api/users/:id
- [ ] PUT /api/users/:id

### 7.3 Product Handler
- [ ] Create `internal/handlers/product_handler.go`
- [ ] GET /api/products
- [ ] GET /api/products/:id
- [ ] POST /api/products
- [ ] PUT /api/products/:id
- [ ] PUT /api/products/:id/price
- [ ] PUT /api/products/:id/stock
- [ ] PUT /api/products/:id/status
- [ ] DELETE /api/products/:id

### 7.4 Category Handler
- [ ] GET /api/categories
- [ ] POST /api/categories
- [ ] PUT /api/categories/:id
- [ ] DELETE /api/categories/:id

### 7.5 Unit Handler
- [ ] GET /api/units
- [ ] POST /api/units
- [ ] PUT /api/units/:id
- [ ] DELETE /api/units/:id

### 7.6 Warehouse Handler
- [ ] Create `internal/handlers/warehouse_handler.go`
- [ ] GET /api/warehouses
- [ ] GET /api/warehouses/:id
- [ ] POST /api/warehouses
- [ ] PUT /api/warehouses/:id
- [ ] DELETE /api/warehouses/:id

### 7.7 Inventory Handler
- [ ] Create `internal/handlers/inventory_handler.go`
- [ ] GET /api/inventory
- [ ] GET /api/inventory/stock-card
- [ ] POST /api/inventory/adjust

### 7.8 Stock Transfer Handler
- [ ] POST /api/stock-transfers
- [ ] PUT /api/stock-transfers/:id/receive

### 7.9 Stock Opname Handler
- [ ] POST /api/stock-opname
- [ ] GET /api/stock-opname
- [ ] GET /api/stock-opname/:id
- [ ] PUT /api/stock-opname/:id/status

### 7.10 Sales Handler
- [ ] Create `internal/handlers/sales_handler.go`
- [ ] POST /api/sales
- [ ] GET /api/sales
- [ ] GET /api/sales/:id

### 7.11 Returns Handler
- [ ] Create `internal/handlers/returns_handler.go`
- [ ] POST /api/returns
- [ ] GET /api/returns
- [ ] GET /api/returns/:id

### 7.12 Exchanges Handler
- [ ] Create `internal/handlers/exchanges_handler.go`
- [ ] POST /api/exchanges
- [ ] GET /api/exchanges
- [ ] GET /api/exchanges/:id

### 7.13 Purchase Handler
- [ ] Create `internal/handlers/purchase_handler.go`
- [ ] POST /api/purchase-orders
- [ ] GET /api/purchase-orders
- [ ] GET /api/purchase-orders/:id
- [ ] POST /api/grn
- [ ] GET /api/grn
- [ ] GET /api/grn/:id

### 7.14 Customer Handler
- [ ] Create `internal/handlers/customer_handler.go`
- [ ] GET /api/customers
- [ ] GET /api/customers/:id
- [ ] POST /api/customers
- [ ] PUT /api/customers/:id
- [ ] DELETE /api/customers/:id

### 7.15 Supplier Handler
- [ ] Create `internal/handlers/supplier_handler.go`
- [ ] GET /api/suppliers
- [ ] GET /api/suppliers/:id
- [ ] POST /api/suppliers
- [ ] PUT /api/suppliers/:id
- [ ] DELETE /api/suppliers/:id

### 7.16 Promotion Handler
- [ ] Create `internal/handlers/promotion_handler.go`
- [ ] GET /api/promotions
- [ ] POST /api/promotions

### 7.17 Price Tier Handler
- [ ] Create `internal/handlers/pricetier_handler.go`
- [ ] GET /api/price-tiers
- [ ] POST /api/price-tiers

### 7.18 Company Handler
- [ ] Create `internal/handlers/company_handler.go`
- [ ] GET /api/companies
- [ ] GET /api/companies/:id

### 7.19 Finance Handler
- [ ] Create `internal/handlers/finance_handler.go`
- [ ] GET /api/invoices/incoming
- [ ] GET /api/invoices/outgoing

### 7.20 Cash Drawer Handler
- [ ] Create `internal/handlers/cashdrawer_handler.go`
- [ ] GET /api/cash-drawers
- [ ] POST /api/cash-drawers
- [ ] POST /api/cash-drawers/:id/open
- [ ] POST /api/cash-drawers/:id/close
- [ ] POST /api/cash-drawers/:id/transactions

---

## 8. Router Setup

### 8.1 Main Router
- [ ] Create `internal/router/router.go`
- [ ] Setup Fiber app
- [ ] Add CORS middleware
- [ ] Add logging middleware
- [ ] Setup API routes

### 8.2 Route Groups
- [ ] Group /api/auth
- [ ] Group /api/users (protected)
- [ ] Group /api/products (protected)
- [ ] Group /api/categories (protected)
- [ ] Group /api/units (protected)
- [ ] Group /api/warehouses (protected)
- [ ] Group /api/inventory (protected)
- [ ] Group /api/stock-transfers (protected)
- [ ] Group /api/stock-opname (protected)
- [ ] Group /api/sales (protected)
- [ ] Group /api/returns (protected)
- [ ] Group /api/exchanges (protected)
- [ ] Group /api/purchase-orders (protected)
- [ ] Group /api/grn (protected)
- [ ] Group /api/customers (protected)
- [ ] Group /api/suppliers (protected)
- [ ] Group /api/promotions (protected)
- [ ] Group /api/price-tiers (protected)
- [ ] Group /api/companies (protected)
- [ ] Group /api/invoices (protected)
- [ ] Group /api/cash-drawers (protected)

---

## 9. Swagger Documentation

### 9.1 Setup
- [ ] Install swaggo: `go install github.com/swaggo/swag/cmd/swag@latest`
- [ ] Install Scalar CLI: `go install github.com/scalar/scalar/cmd/scalar@latest`
- [ ] Create `docs/` folder for generated files
- [ ] Configure swag init in Makefile or go generate

### 9.2 Swag Annotations
- [ ] Add `@Summary`, `@Description`, `@Tags` to all handlers
- [ ] Add `@Param` for request parameters
- [ ] Add `@Success` and `@Failure` responses
- [ ] Add `@Router` for route definition
- [ ] Add `@Security` BearerAuth for protected endpoints

### 9.3 Generate Documentation
- [ ] Run `swag init -g cmd/server/main.go --parseDependency --parseInternal --output ./docs`
- [ ] Verify `docs/swagger.json` and `docs/docs.go` are generated

### 9.4 Scalar UI Integration
- [ ] Install Scalar Go package: `go get github.com/scalar/scalar-go`
- [ ] Create docs middleware in `internal/docs/docs.go`
- [ ] Configure Scalar with OpenAPI spec
- [ ] Enable dark mode support
- [ ] Serve Scalar at `/docs` endpoint
- [ ] Test interactive API documentation

### 9.5 Documentation Structure
```
docs/
├── docs.go           # Generated by swag
├── swagger.json      # OpenAPI 3.0 spec
├── swagger.yaml      # Alternative format
└── index.html        # Custom Scalar index (optional)
```

### 9.6 Annotation Examples

```go
// Auth Handler
// @Summary Login user
// @Description Authenticate user and create session
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body request.LoginRequest true "Login credentials"
// @Success 200 {object} response.LoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/auth/login [post]

// Product Handler
// @Summary List products
// @Description Get paginated list of products with optional filters
// @Tags Products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param status query string false "Filter by status"
// @Param category_id query string false "Filter by category"
// @Param search query string false "Search in name, SKU, barcode"
// @Success 200 {object} response.PaginatedResponse
// @Router /api/products [get]
```

---

## 10. Testing & Validation

### 10.1 Unit Tests
- [ ] Write unit tests for utility functions (JWT, password)
- [ ] Write unit tests for validation logic
- [ ] Mock repositories for service tests

### 10.2 Integration Tests
- [ ] Test database operations
- [ ] Test API endpoints with test database

### 10.3 Parity Testing
- [ ] Compare response format with TypeScript backend
- [ ] Test error messages match
- [ ] Test pagination format
- [ ] Test authentication flow

---

## 11. Production Readiness

### 11.1 Security
- [ ] Input sanitization
- [ ] SQL injection prevention (GORM handles this)
- [ ] Rate limiting (optional)
- [ ] Secure headers

### 11.2 Performance
- [ ] Database connection pooling
- [ ] Query optimization
- [ ] Proper indexing

### 11.3 Deployment
- [ ] Docker support (Dockerfile)
- [ ] Docker Compose for local dev
- [ ] Environment configuration
- [ ] Graceful shutdown

---

## Implementation Order

```
Phase 1: Foundation
1. Project Setup
2. Configuration & Database
3. Models
4. Middleware (Auth)

Phase 2: Core Features
5. Auth Handler & Service
6. Product Handler & Service
7. Warehouse Handler & Service

Phase 3: Business Logic
8. Inventory Handler & Service
9. Sales Handler & Service
10. Returns & Exchanges

Phase 4: Operations
11. Purchase Handler & Service
12. Customer Handler & Service
13. Supplier Handler & Service

Phase 5: Advanced
14. Promotions & Price Tiers
15. Finance & Cash Drawer

Phase 6: Polish
16. Swagger Documentation
17. Testing & Parity Check
18. Production Ready
```

---

## Notes

- All endpoints must match TypeScript backend exactly
- Response format must be identical
- Error messages should be consistent
- Use same JWT payload structure
- Follow Go best practices (Clean Code, proper error handling)
- Add proper logging throughout
- Update task_report.md after completing each task item

---

## Task Report

See [task_report.md](./task_report.md) for progress tracking.

---

*Document Version: 1.0*
*Created: 2026-02-17*
*Based on: backend/src (TypeScript/Elysia)*
