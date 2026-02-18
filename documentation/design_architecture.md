# POS Retail Backend - Go Implementation

## Design Architecture

---

## 1. System Overview

### Purpose
Reimplement the existing POS Retail Backend from TypeScript (Elysia.js) to Go (Fiber) while maintaining full API parity.

### Key Requirements
- **API Parity**: All endpoints, payloads, and response formats must match the existing TypeScript backend
- **Database Parity**: Same PostgreSQL database, same schema, same tables
- **Authentication**: JWT-based auth with session management
- **Multi-company**: Support multi-tenant operations

### Technology Stack
| Component | Choice | Rationale |
|-----------|--------|-----------|
| Web Framework | Fiber | High performance, similar API to Express |
| ORM | GORM | Mature, auto-migrations, widely used |
| Database | PostgreSQL | Same as existing backend |
| Validation | go-playground/validator | Industry standard |
| JWT | golang-jwt/jwt | Standard JWT library |
| Password Hashing | golang.org/x/crypto/bcrypt | Secure, same as bcrypt |

---

## 2. Project Structure

```
go_backend/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── database/
│   │   └── database.go            # GORM database connection
│   ├── models/
│   │   ├── core.go                # User, Session models
│   │   ├── company.go             # Company model
│   │   ├── product.go             # Product, Category, Unit, Warehouse models
│   │   ├── inventory.go           # Inventory, StockMovement models
│   │   ├── sales.go               # Sale, SaleItem, SalePayment models
│   │   ├── purchase.go            # Purchase, GRN models
│   │   ├── promotion.go           # Promotion models
│   │   ├── customer.go            # Customer model
│   │   ├── supplier.go            # Supplier model
│   │   ├── finance.go             # Invoice models
│   │   └── cashdrawer.go          # CashDrawer models
│   ├── repository/
│   │   ├── auth_repository.go
│   │   ├── product_repository.go
│   │   ├── sales_repository.go
│   │   ├── inventory_repository.go
│   │   └── ...
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── product_service.go
│   │   ├── sales_service.go
│   │   └── ...
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── product_handler.go
│   │   ├── sales_handler.go
│   │   └── ...
│   ├── middleware/
│   │   ├── auth_middleware.go
│   │   ├── cors_middleware.go
│   │   └── logging_middleware.go
│   ├── types/
│   │   ├── request/
│   │   │   ├── auth_request.go
│   │   │   ├── product_request.go
│   │   │   └── ...
│   │   └── response/
│   │       ├── api_response.go
│   │       └── pagination.go
│   └── utils/
│       ├── jwt.go
│       ├── password.go
│       └── helpers.go
├── documentation/
│   ├── design_architecture.md
│   └── task.md
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

---

## 3. Architecture Pattern

### Layered Architecture

```
┌─────────────────────────────────────────┐
│           Router (Fiber)                │
│   /api/auth, /api/products, etc.       │
└────────────────┬──────────────────────┘
                 │
┌────────────────▼──────────────────────┐
│         Middleware                     │
│   Auth, CORS, Logging, Validation     │
└────────────────┬──────────────────────┘
                 │
┌────────────────▼──────────────────────┐
│          Handlers                      │
│   HTTP Request/Response handling      │
└────────────────┬──────────────────────┘
                 │
┌────────────────▼──────────────────────┐
│          Services                       │
│   Business Logic                       │
└────────────────┬──────────────────────┘
                 │
┌────────────────▼──────────────────────┐
│          Repository                     │
│   Database Operations (GORM)           │
└────────────────┬──────────────────────┘
                 │
┌────────────────▼──────────────────────┐
│          Database (PostgreSQL)         │
└────────────────────────────────────────┘
```

### Request Flow

1. Client sends HTTP request with/without JWT token
2. Fiber router matches route to handler
3. Auth middleware validates JWT and populates context
4. Request body is validated
5. Handler calls Service method
6. Service calls Repository for database operations
7. Response flows back through layers

---

## 4. Database Schema Mapping

### Core Tables (from `backend/src/db/schema/core.ts`)

```go
// models/core.go
type User struct {
    ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Username  string     `gorm:"type:varchar(50);uniqueIndex;not null"`
    Email     string     `gorm:"type:varchar(100);uniqueIndex;not null"`
    Password  string     `gorm:"type:text;not null"`
    FullName  string     `gorm:"type:varchar(100);not null"`
    Role      string     `gorm:"type:user_role;not null;default:'staff'"`
    Status    string     `gorm:"type:user_status;not null;default:'active'"`
    CompanyID uuid.UUID  `gorm:"type:uuid;not null"`
    CreatedAt time.Time  `gorm:"autoCreateTime"`
    UpdatedAt time.Time  `gorm:"autoUpdateTime"`
    LastLogin *time.Time
}
```

### Enums (PostgreSQL Types)

The following enums must be created in PostgreSQL before running the application:

```sql
-- User enums
CREATE TYPE user_role AS ENUM ('admin', 'manager', 'cashier', 'staff');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');

-- Product enums
CREATE TYPE product_status AS ENUM ('active', 'inactive', 'discontinued');
CREATE TYPE warehouse_type AS ENUM ('MAIN', 'BRANCH', 'STORAGE', 'OUTLET');
CREATE TYPE warehouse_status AS ENUM ('active', 'inactive');

-- Sale enums
CREATE TYPE sale_status AS ENUM ('PENDING', 'COMPLETED', 'CANCELLED', 'REFUNDED');
CREATE TYPE payment_method AS ENUM ('CASH', 'CREDIT_CARD', 'DEBIT_CARD', 'EWALLET', 'BANK_TRANSFER', 'CUSTOMER_CREDIT');

-- Movement enums
CREATE TYPE movement_type AS ENUM ('SALE', 'PURCHASE', 'TRANSFER_OUT', 'TRANSFER_IN', 'ADJUSTMENT_IN', 'ADJUSTMENT_OUT', 'RETURN', 'DAMAGE', 'OPNAME', 'EXCHANGE_IN', 'EXCHANGE_OUT');

-- And more...
```

**Note**: GORM will use `gorm:"type:enum"` for enum fields but requires the PostgreSQL types to exist.

---

## 5. API Response Contract

### Standard Response Format

```go
// types/response/api_response.go
type ApiResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Message string      `json:"message,omitempty"`
}

type PaginatedResponse struct {
    Success  bool        `json:"success"`
    Data     interface{} `json:"data"`
    Total    int64       `json:"total"`
    Limit    int         `json:"limit"`
    Offset   int         `json:"offset"`
    HasMore  bool        `json:"has_more"`
}
```

### Response Examples

**Success Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "username": "john_doe",
    "full_name": "John Doe",
    "role": "staff"
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "Invalid credentials"
}
```

**Paginated Response:**
```json
{
  "success": true,
  "data": [...],
  "pagination": {
    "total": 150,
    "limit": 50,
    "offset": 0,
    "has_more": true
  }
}
```

---

## 6. Authentication & Authorization

### JWT Token Structure

Matches existing backend (`backend/src/utils/auth.ts`):

```go
type JWTPayload struct {
    UserID       string   `json:"userId"`
    Username     string   `json:"username"`
    Email        string   `json:"email"`
    CompanyID    string   `json:"companyId"`
    Role         string   `json:"role"`
    CompanyAccess []string `json:"companyAccess"`
    Permissions  []string `json:"permissions"`
    Iat          int64   `json:"iat,omitempty"`
    Exp          int64   `json:"exp,omitempty"`
}
```

### Auth Middleware Flow

```
Request → Extract Bearer Token → Verify JWT → Check User Status → Add User to Context → Next Handler
```

### Role-Based Access

```go
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userRole := c.Locals("userRole").(string)
        for _, role := range allowedRoles {
            if userRole == role {
                return c.Next()
            }
        }
        return c.Status(403).JSON(ApiResponse{
            Success: false,
            Error: "Forbidden: Insufficient permissions",
        })
    }
}
```

---

## 7. Configuration

### Environment Variables (.env)

```env
# Server
APP_HOST=0.0.0.0
APP_PORT=3000
APP_ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=pos_retail
DB_USER=postgres
DB_PASSWORD=postgres
DB_MAX_CONNECTIONS=10

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRES_IN=8h

# Email (optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

### Config Structure

```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Email    EmailConfig
}

type ServerConfig struct {
    Host string
    Port int
    Env  string
}

type DatabaseConfig struct {
    Host           string
    Port           int
    Name           string
    User           string
    Password       string
    MaxConnections int
}
```

---

## 8. Validation

### Request Validation

Using `go-playground/validator`:

```go
type LoginRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Password string `json:"password" validate:"required,min=6,max=128"`
}

type ProductCreateRequest struct {
    SKU         string  `json:"sku" validate:"required,max=50"`
    Barcode     string  `json:"barcode" validate:"max=50"`
    Name        string  `json:"name" validate:"required,max=200"`
    CategoryID  string  `json:"category_id" validate:"required,uuid"`
    UnitID      string  `json:"unit_id" validate:"required,uuid"`
    CostPrice   float64 `json:"cost_price" validate:"required,gt=0"`
    RetailPrice float64 `json:"retail_price" validate:"required,gt=0"`
    TaxRate     float64 `json:"tax_rate" validate:"omitempty,min=0,max=100"`
    ReorderPoint int    `json:"reorder_point" validate:"omitempty,min=0"`
}
```

---

## 9. Error Handling

### Error Codes

```go
const (
    ErrCodeNotFound         = "NOT_FOUND"
    ErrCodeUnauthorized    = "UNAUTHORIZED"
    ErrCodeForbidden       = "FORBIDDEN"
    ErrCodeValidation      = "VALIDATION_ERROR"
    ErrCodeDuplicate      = "DUPLICATE_ENTRY"
    ErrCodeInternal       = "INTERNAL_ERROR"
)
```

### Error Response

```go
type ErrorResponse struct {
    Success bool   `json:"success"`
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

---

## 10. Logging

### Structured Logging

Using `gofiber/logger` middleware and optional `zerolog`/`zap`:

```go
// Request logging example
logger := fiber.New(fiber.Config{
    Logger: logger.Config{
        Format:     "${time} | ${status} | ${latency} | ${method} ${path} ${error}",
        TimeFormat: "2006-01-02 15:04:05",
    },
})
```

---

## 11. Swagger / API Documentation

### OpenAPI Generation with Scalar UI

We use **Scalar** instead of default Swagger UI for a modern, clean API documentation experience.

**Installation:**
```bash
go install github.com/scalar/scalar/cmd/scalar@latest
```

**Dependencies:**
```go
require (
    github.com/scalar/scalar-go v0.0.1
    github.com/swaggo/swag v1.16.2
)
```

**Setup:**

1. Add Swagger annotations to handlers
2. Run `swag init -g cmd/server/main.go --parseDependency --parseInternal`
3. Serve Scalar UI at `/docs` endpoint

**Custom Integration:**

```go
// internal/docs/docs.go
//go:generate swag init -g cmd/server/main.go --parseDependency --parseInternal --output ./docs

package docs

import "github.com/swaggo/swag"

var SwaggerInfo = &swag.Spec{
    Version:          "1.0.0",
    Title:            "POS Retail API",
    Description:     "Point of Sale Retail Backend System API",
    TermsOfService:   "http://swagger.io/terms/",
    Contact:          &swag.Contact{Name: "API Support", Email: "support@pos-retail.com"},
    License:          &swag.License{Name: "MIT", URL: "https://opensource.org/licenses/MIT"},
    Host:             "localhost:3000",
    BasePath:         "/",
    Schemes:          []string{"http", "https"},
}

// ServeDocs serves the API documentation
func ServeDocs(app *fiber.App) {
    app.Get("/docs/*", func(c *fiber.Ctx) error {
        return c.SendFile("./docs/index.html")
    })
    
    app.Get("/docs/openapi.json", func(c *fiber.Ctx) error {
        return c.JSON(SwaggerInfo)
    })
}
```

**Scalar UI Configuration:**

```go
// In main.go or router setup
import (
    "github.com/gofiber/fiber/v2"
    scalar "github.com/scalar/scalar-go"
)

func main() {
    app := fiber.New()
    
    // ... other setup ...
    
    // Scalar UI at /docs
    scalarHandler := scalar.New(
        scalar.WithSpec("./docs/openapi.json"),
        scalar.WithDarkMode(true),
    )
    
    app.Get("/docs", scalarHandler)
    app.Get("/docs/*", scalarHandler)
    
    app.Listen(":3000")
}
```

**Scalar UI Features:**
- Modern, clean design
- Dark/light mode support
- Interactive API testing
- OpenAPI 3.0 support
- Request/response preview
- Copy code snippets

Example annotation:

```go
// @Summary Login user
// @Description Authenticate user and create session
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body request.LoginRequest true "Login credentials"
// @Success 200 {object} response.LoginResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/auth/login [post]
func LoginHandler(c *fiber.Ctx) error { ... }
```

---

## 12. Testing Strategy

### Unit Tests
- Repository layer: Mock GORM for database operations
- Service layer: Mock repository for business logic
- Handler layer: Test HTTP responses

### Integration Tests
- Test against actual PostgreSQL database
- Use test database or transactions with rollback

### API Testing
- Port existing TypeScript tests to `newman` (Postman collection)
- Or create Go test suite using `testify` + `gock`

---

## 13. Parity Checklist

To ensure full API parity with TypeScript backend:

| Component | Checkpoint |
|-----------|------------|
| Endpoints | All routes in `backend/src/routes/*.ts` mapped |
| Request Body | All DTOs match TypeScript interfaces |
| Response | JSON structure matches exactly |
| Auth | JWT payload, token format identical |
| Database | Same queries, same schema |
| Error Messages | Error text matches TypeScript version |
| Pagination | Same limit, offset, total format |
| Enums | PostgreSQL enum values match |

---

## 14. Dependencies (go.mod)

```go
module github.com/your-org/go_backend

go 1.21

require (
    github.com/gofiber/fiber/v2 v2.52.0
    github.com/golang-jwt/jwt/v5 v5.2.0
    gorm.io/driver/postgres v1.5.4
    gorm.io/gorm v1.25.5
    github.com/go-playground/validator/v10 v10.16.0
    golang.org/x/crypto v0.18.0
    github.com/google/uuid v1.6.0
    github.com/joho/godotenv v1.5.1
    github.com/swaggo/fiber v2.20.0
    github.com/swaggo/swag v1.16.2
    github.com/adam-hanna/bcrypt v0.0.0-20190803104444-7c46b7f3d998
)
```

---

## 15. Next Steps

1. Create `go.mod` and initialize project
2. Set up configuration and database connection
3. Implement models matching existing schema
4. Build auth system (JWT, middleware)
5. Implement each domain in order:
   - Auth → Products → Inventory → Sales → Purchases → Finance
6. Add request validation
7. Generate Swagger docs
8. Run parity tests

---

*Document Version: 1.0*
*Created: 2026-02-17*
*Based on: backend/src (TypeScript/Elysia)*
