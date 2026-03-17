# Master Plan: API Remove Data

## 1. Overview

**Purpose:** API untuk menghapus data testing (master data dan transaksi) berdasarkan company user yang login.

**Use Case:**
- Development/Testing: Engineer perlu reset data untuk testing
- Demo: Reset data ke kondisi awal
- Data Migration: Bersihkan data lama

**Requirements:**
- Authentication wajib (Bearer JWT)
- Role: admin only
- Delete permanen (hard delete)
- Filter berdasarkan company dari JWT
- Partial delete (tabel spesifik) didukung
- Logging setiap operasi delete

---

## 2. API Specification

### 2.1 Endpoints

| No | Endpoint | Method | Description |
|----|----------|--------|-------------|
| 1 | `/api/remove-data/master` | DELETE | Hapus semua master data |
| 2 | `/api/remove-data/transactions` | DELETE | Hapus semua transaksi |
| 3 | `/api/remove-data/table` | DELETE | Hapus tabel spesifik (partial) |

### 2.2 Authentication

```
Authorization: Bearer <jwt_token>
```

**JWT Payload Required:**
- `userId`: ID user yang melakukan aksi
- `companyId`: Company ID utama
- `companyAccess`: Array company IDs yang bisa diakses (untuk multi-company)
- `role`: Harus "admin"

### 2.3 Request/Response Format

**DELETE `/api/remove-data/table`** (Request Body)
```json
{
  "tables": ["sales", "products", "customers"]
}
```

**Success Response**
```json
{
  "success": true,
  "message": "3 tables cleared",
  "data": {
    "sales": 150,
    "products": 45,
    "customers": 20
  }
}
```

**Error Response**
```json
{
  "success": false,
  "error": "Unauthorized: Admin role required"
}
```

---

## 3. Allowed Tables (Whitelist)

### 3.1 Master Data

| Table Key | Model | GORM Table | Company Filter |
|-----------|-------|------------|----------------|
| `units` | Unit | `units` | ✅ |
| `categories` | Category | `categories` | ✅ |
| `warehouses` | Warehouse | `warehouses` | ✅ |
| `products` | Product | `products` | ✅ |
| `price_tiers` | PriceTier | `price_tiers` | ✅ |
| `customers` | Customer | `customers` | ✅ |
| `suppliers` | Supplier | `suppliers` | ✅ |
| `promotions` | Promotion | `promotions` | ✅ |
| `promotion_products` | PromotionProduct | `promotion_products` | ✅ |
| `promotion_categories` | PromotionCategory | `promotion_categories` | ✅ |
| `promotion_customers` | PromotionCustomer | `promotion_customers` | ✅ |

### 3.2 Transaction Data

| Table Key | Model | GORM Table | Company Filter |
|-----------|-------|------------|----------------|
| `inventories` | Inventory | `inventories` | ✅ |
| `stock_movements` | StockMovement | `stock_movements` | ✅ |
| `stock_transfers` | StockTransfer | `stock_transfers` | ✅ |
| `stock_transfer_items` | StockTransferItem | `stock_transfer_items` | ✅ |
| `stock_opnames` | StockOpname | `stock_opnames` | ✅ |
| `stock_opname_items` | StockOpnameItem | `stock_opname_items` | ✅ |
| `sales` | Sale | `sales` | ✅ |
| `sale_items` | SaleItem | `sale_items` | ✅ |
| `sale_payments` | SalePayment | `sale_payments` | ✅ |
| `sales_returns` | SalesReturn | `sales_returns` | ✅ |
| `sales_return_items` | SalesReturnItem | `sales_return_items` | ✅ |
| `item_exchanges` | ItemExchange | `item_exchanges` | ✅ |
| `exchange_items` | ExchangeItem | `exchange_items` | ✅ |
| `purchase_orders` | PurchaseOrder | `purchase_orders` | ✅ |
| `purchase_order_items` | PurchaseOrderItem | `purchase_order_items` | ✅ |
| `purchase_returns` | PurchaseReturn | `purchase_returns` | ✅ |
| `purchase_return_items` | PurchaseReturnItem | `purchase_return_items` | ✅ |
| `invoices_incoming` | IncomingInvoice | `invoices_incoming` | ✅ |
| `invoices_outgoing` | OutgoingInvoice | `invoices_outgoing` | ✅ |
| `invoice_items` | InvoiceItem | `invoice_items` | ✅ |
| `invoice_payments` | InvoicePayment | `invoice_payments` | ✅ |
| `cash_drawers` | CashDrawer | `cash_drawers` | ✅ |
| `cash_drawer_transactions` | CashDrawerTransaction | `cash_drawer_transactions` | ✅ |

### 3.3 Protected Tables (Cannot Delete)

- `users`
- `user_sessions`
- `email_verifications`
- `password_resets`
- `companies`

---

## 4. Deletion Order

### 4.1 Master Data (11 tables)

```
DELETE ORDER:
1. promotion_customers
2. promotion_categories
3. promotion_products
4. promotions
5. price_tiers
6. products
7. customers
8. suppliers
9. warehouses
10. categories
11. units
```

### 4.2 Transaction Data (23 tables)

```
DELETE ORDER (child → parent):
1. sale_payments
2. sale_items
3. sales
4. sales_return_items
5. sales_returns
6. exchange_items
7. item_exchanges
8. invoice_payments
9. invoice_items
10. invoices_incoming
11. invoices_outgoing
12. cash_drawer_transactions
13. cash_drawers
14. purchase_order_items
15. purchase_orders
16. purchase_return_items
17. purchase_returns
18. stock_opname_items
19. stock_opnames
20. stock_transfer_items
21. stock_transfers
22. stock_movements
23. inventories
```

**Rationale:** Child tables dihapus dulu sebelum parent untuk menghindari foreign key constraint violation.

---

## 5. File Structure

### 5.1 New Files

```
internal/
├── services/
│   └── test_data_service.go      # Business logic
├── handlers/
│   └── test_data_handler.go      # HTTP handlers
└── types/
    └── request/
        └── delete_table_request.go  # Request DTO
```

### 5.2 Modified Files

```
cmd/server/main.go               # Register routes & DI
```

---

## 6. Implementation Details

### 6.1 Service Layer (`test_data_service.go`)

**Struct:**
```go
type TestDataService struct {
    db *gorm.DB
}
```

**Methods:**
| Method | Input | Output |
|--------|-------|--------|
| `DeleteMasterData(companyIDs []uuid.UUID, actorUserID, actorCompanyID string)` | company IDs, user info | `(map[string]int64, ApiResponse)` |
| `DeleteTransactionData(companyIDs []uuid.UUID, actorUserID, actorCompanyID string)` | company IDs, user info | `(map[string]int64, ApiResponse)` |
| `DeleteTableData(tables []string, companyIDs []uuid.UUID, actorUserID, actorCompanyID string)` | tables, company IDs, user info | `(map[string]int64, ApiResponse)` |
| `GetCompanyIDs(user *utils.JWTPayload)` | JWT payload | `[]uuid.UUID` |

### 6.2 Handler Layer (`test_data_handler.go`)

**Endpoints:**

| Handler | Route | Auth | Role |
|---------|-------|------|------|
| `DeleteMasterData` | `DELETE /master` | ✅ | admin |
| `DeleteTransactionData` | `DELETE /transactions` | ✅ | admin |
| `DeleteTableData` | `DELETE /table` | ✅ | admin |

### 6.3 Request DTO (`delete_table_request.go`)

```go
type DeleteTableRequest struct {
    Tables []string `json:"tables" validate:"required,min=1"`
}
```

---

## 7. Security

### 7.1 Authentication
- Bearer JWT token wajib di `Authorization` header
- Token divalidasi via `AuthMiddleware`

### 7.2 Authorization
- Role `admin` wajib
- `RoleMiddleware("admin")` diterapkan

### 7.3 Data Isolation
- Semua delete difilter berdasarkan `company_id`
- Company IDs diambil dari JWT (`CompanyID` dan `CompanyAccess`)
- User tidak bisa menghapus data company lain

### 7.4 Input Validation
- Table name divalidasi against whitelist
- Request body divalidasi menggunakan `ValidateBody` middleware

---

## 8. Logging

Menggunakan existing `applogger` yang sama dengan fitur lain.

```go
if l := applogger.Default(); l != nil {
    l.Log(applogger.ActionDelete, tableName, actorUserID, actorCompanyID, tableName, nil, count)
}
```

**Log Fields:**
| Field | Value |
|-------|-------|
| Action | `delete` |
| Table | Nama tabel yang dihapus |
| Actor User ID | User ID dari JWT |
| Actor Company ID | Company ID dari JWT |
| Record ID | Nama tabel |
| Old Value | `nil` |
| New Value | Jumlah record deleted |

---

## 9. Route Registration

```go
// main.go

// Initialize service
testDataService := services.NewTestDataService(db)

// Initialize handler
testDataHandler := handlers.NewTestDataHandler(testDataService)

// Register routes
removeData := protected.Group("/remove-data", middleware.RoleMiddleware("admin"))
removeData.Delete("/master", testDataHandler.DeleteMasterData)
removeData.Delete("/transactions", testDataHandler.DeleteTransactionData)
removeData.Delete("/table", testDataHandler.DeleteTableData)
```

---

## 10. Transaction Safety

Semua operasi delete dibungkus dalam GORM transaction:

```go
err := s.db.Transaction(func(tx *gorm.DB) error {
    for _, tableKey := range tables {
        count, err := s.deleteTableByKey(tx, tableKey, companyIDs, ...)
        if err != nil {
            return fmt.Errorf("failed to delete %s: %w", tableKey, err)
        }
        results[tableKey] = count
    }
    return nil
})
```

Jika salah satu tabel gagal dihapus, semua perubahan akan di-rollback.

---

## 11. Build & Run

```bash
# Build
go build -o bin/server ./cmd/server

# Run
go run ./cmd/server
```

---

## 12. Acceptance Criteria

- [x] API `/remove-data/master` menghapus 11 master tables
- [x] API `/remove-data/transactions` menghapus 23 transaction tables
- [x] API `/remove-data/table` menghapus tabel spesifik sesuai request
- [x] Hanya tabel dalam whitelist bisa dihapus
- [x] Data company & user tidak terhapus
- [x] Semua delete difilter berdasarkan company dari JWT
- [x] Semua operasi logged menggunakan applogger
- [x] Build success tanpa error

---

## 13. Implementation Status

| Component | Status |
|-----------|--------|
| `delete_table_request.go` | ✅ Done |
| `test_data_service.go` | ✅ Done |
| `test_data_handler.go` | ✅ Done |
| `main.go` routes | ✅ Done |
| Build verification | ✅ Success |

---

## 14. Swagger Documentation

Swagger annotations sudah ditambahkan di handler:

- `DeleteMasterData` - `/api/remove-data/master`
- `DeleteTransactionData` - `/api/remove-data/transactions`
- `DeleteTableData` - `/api/remove-data/table`

Jalankan `swag init` untuk menggenerate swagger.json terbaru.
