# Laporan Perbaikan Foreign Key Constraint pada PurchaseOrder

## Tanggal Perbaikan
10 Maret 2026

## Masalah yang Diperbaiki
Foreign key constraint violation pada tabel `purchase_orders` yang menyebabkan error:
```
ERROR: insert or update on table "purchase_orders" violates foreign key constraint "fk_purchase_orders_company" (SQLSTATE 23503)
```

## Akar Masalah
1. Model `PurchaseOrder` tidak memiliki foreign key constraint yang benar
2. Service layer tidak memvalidasi keberadaan company ID sebelum membuat purchase order
3. Tidak ada error handling yang spesifik untuk foreign key violation

## Solusi yang Diterapkan

### 1. Perbaikan Model PurchaseOrder
- Menambahkan foreign key constraint pada field `CompanyID`:
  ```go
  CompanyID uuid.UUID `gorm:"column:company_id;type:uuid;notNull;index;references:companies(id)" json:"company_id"`
  ```
- Menambahkan relationship struct untuk navigasi antar tabel:
  ```go
  Company Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"-"`
  ```

### 2. Validasi di Service Layer
- Menambahkan validasi keberadaan company ID sebelum create purchase order
- Error message yang spesifik untuk company ID tidak ditemukan
- Validasi format UUID untuk semua ID yang diterima

### 3. Testing
- Membuat test case untuk validasi company ID
- Test scenario: create purchase order dengan company ID invalid vs valid
- Test transaction rollback pada foreign key violation

## File yang Diubah
1. `internal/models/purchase.go` - Menambahkan foreign key constraint
2. `internal/services/purchase_service.go` - Menambahkan validasi
3. `internal/integration/auth_integration_test.go` - Menambahkan test case

## Pedoman yang Dibuat
Membuat file aturan Cline `.clinerules/database-foreign-key.md` yang berisi:
- Best practices untuk foreign key constraint
- Format yang benar untuk tag GORM
- Strategi error handling
- Validasi di service layer
- Migration strategy
- Testing guidelines

## Dampak Perubahan
- Mencegah foreign key constraint violation
- Memberikan error message yang jelas dan spesifik
- Meningkatkan data integrity
- Memudahkan debugging dan maintenance

## Langkah Selanjutnya
1. Testing menyeluruh pada semua CRUD operations
2. Dokumentasi API untuk error handling
3. Monitoring production untuk error foreign key
4. Review model-model lain yang mungkin memerlukan foreign key constraint

## Kesimpulan
Perbaikan ini menyelesaikan masalah foreign key constraint violation dan meningkatkan robustness dari sistem dengan menambahkan validasi yang tepat dan error handling yang jelas.