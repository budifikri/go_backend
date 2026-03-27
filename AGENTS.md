# AGENTS.md - Development Guide for POS Retail Backend

## Project Overview

Go 1.23 Fiber REST API with GORM/PostgreSQL. Layered architecture: handlers → services → repositories.

## Build, Lint, and Test Commands

### Build
```bash
go build -o bin/server ./cmd/server
```

### Run Development Server
```bash
go run ./cmd/server
```

### Run Tests
```bash
# All tests
go test ./...

# Single test
go test -run TestJWTGenerateAndVerify ./internal/utils/

# With verbose output
go test -v ./...

# With coverage
go test -cover ./...
```

### Linting
```bash
# Format code
go fmt ./...

# Vet checks
go vet ./...

# Run all (fmt + vet + test)
go vet ./... && go test ./...
```

### Database (Docker Compose)
```bash
docker-compose up -d db
```

### API Documentation (Swagger)
```bash
# IMPORTANT: Always output to docs/ folder, NOT internal/docs/
# The app reads swagger.json from docs/ folder (see docs/register_docs_stub.go)
# DO NOT use -o internal/docs or internal/swagger
swag init -g cmd/server/main.go -o docs
```

## Project Structure

```
cmd/server/main.go       # Application entry point
docs/                   # Swagger documentation (served at /docs)
internal/
├── config/              # Configuration loading (.env)
├── database/           # GORM database connection
├── handlers/           # HTTP handlers (Fiber)
├── middleware/         # Auth, validation, CORS
├── models/              # GORM database models
├── parity/              # TS parity tests
├── repository/          # Data access layer
├── services/            # Business logic
├── types/
│   ├── request/        # Request DTOs
│   └── response/       # Response DTOs
└── utils/               # JWT, password utilities
```

## Code Style Guidelines

### Naming Conventions

- **Files**: `snake_case.go` (e.g., `product_handler.go`)
- **Types**: `PascalCase` (e.g., `ProductHandler`, `ApiResponse`)
- **Variables/Functions**: `camelCase` (e.g., `getProducts`, `isActive`)
- **Constants**: `PascalCase` or `SCREAMING_SNAKE_CASE` for enum values
- **Interfaces**: `PascalCase` with `er` suffix where appropriate (e.g., `Reader`, `Writer`)

### Package Structure

- Handlers receive services via constructor dependency injection
- Services receive repositories via constructor
- Request DTOs in `internal/types/request/`
- Response DTOs in `internal/types/response/`

### Error Handling

```go
// Use response helpers
return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))

// Service errors return ApiResponse
return response.NewErrorResponse("Product not found")
return response.NewSuccessResponse(product, "Product created")
```

### Database Models

```go
type Product struct {
    ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Name      string     `gorm:"type:varchar(200);notNull" json:"name"`
    IsActive  bool       `gorm:"default:true;notNull" json:"is_active"`
    CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// Always define TableName for custom table names
func (Product) TableName() string {
    return "products"
}
```

### API Responses

Use the standard response types from `internal/types/response/`:

```go
// Success response
response.NewSuccessResponse(data, "message")

// Error response
response.NewErrorResponse("error message")

// Paginated response
response.NewPaginatedResponse(data, total, limit, offset)
```

### HTTP Handlers

```go
func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
    // Parse query params
    limit := c.QueryInt("limit", 50)
    offset := c.QueryInt("offset", 0)
    
    // Get user from context
    user := middleware.GetUserFromContext(c)
    
    // Call service
    result := h.productService.GetProducts(filters, limit, offset)
    return c.JSON(result)
}
```

### JWT Authentication

- Token passed via `Authorization: Bearer <token>` header
- Middleware validates token and stores payload in context
- Access via `middleware.GetUserFromContext(c)`

### Validation

Use the validation middleware for request body validation:

```go
products.Post("/", middleware.ValidateBody(func() interface{} { return &request.CreateProductRequest{} }), productHandler.CreateProduct)
```

### Testing

```go
func TestFunctionName(t *testing.T) {
    // Arrange
    util := NewJWTUtil("test-secret", 10*time.Minute)
    
    // Act
    result, err := util.VerifyToken(token)
    
    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
}
```

### Imports

Standard order (go fmt will handle this):

1. Standard library
2. Third-party packages
3. Local packages

### Configuration

- All config via environment variables
- Load with `config.Load()` in main.go
- Use `.env` file for local development (see `.env.example`)
- Required env vars: `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `JWT_SECRET`

### API Documentation

Add Swagger annotations to handler functions:

```go
// GetProducts godoc
// @Summary List products
// @Tags Products
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.PaginatedResponse
// @Security BearerAuth
// @Router /api/products [get]
func (h *ProductHandler) GetProducts(c *fiber.Ctx) error { ... }
```

## Common Tasks

### Adding a New Entity

1. Create model in `internal/models/`
2. Create repository in `internal/repository/`
3. Create service in `internal/services/`
4. Create request DTOs in `internal/types/request/`
5. Create handler in `internal/handlers/`
6. Register routes in `cmd/server/main.go`
7. Add auto-migration in main.go

### Adding a New Field

1. Update model struct
2. Update request DTOs
3. Update service to handle the field
4. Update repository queries if needed

## Database

- PostgreSQL 16 via Docker Compose
- GORM auto-migration on startup
- Use transactions for multi-table operations

## Dependencies

Key packages:
- `github.com/gofiber/fiber/v2` - HTTP framework
- `gorm.io/gorm` - ORM
- `gorm.io/driver/postgres` - PostgreSQL driver
- `github.com/google/uuid` - UUID generation
- `github.com/golang-jwt/jwt/v5` - JWT auth
- `github.com/go-playground/validator/v10` - Validation
- `github.com/swaggo/swag` - Swagger docs
