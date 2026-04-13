# Rencana Perbaikan: Cash-In Penjualan Tunai ke Cash Drawer

## Masalah
Cash-in dari pembayaran penjualan Tunai tidak masuk ke tabel:
- `cash_drawer_transactions` - tidak ada record transaksi
- `cash_drawers` - saldo tidak ter-update

## Penyebab
Di file `internal/services/sales_service.go` (lines 320-333), saat membuat transaksi penjualan, pembayaran di-proses hanya dengan membuat record `SalePayment`:

```go
for _, p := range input.Payments {
    sp := models.SalePayment{...}
    tx.Create(&sp)
}
```

**Tidak ada integrasi/hook ke cash drawer** ketika payment method adalah "CASH".

---

## Solusi

### Opsi yang Dipilih: Modifikasi sales_service.go

#### Lokasi File
`D:\Project\pos_retail\go_backend\internal\services\sales_service.go`

#### Perubahan yang Diperlukan

**1. Tambahkan import cash drawer service (jika belum ada)**

**2. Modifikasi fungsi CreateSale**

Tambahkan logic setelah pembuatan SalePayment untuk setiap pembayaran CASH:

```go
// Untuk setiap pembayaran CASH, catat ke cash drawer
if p.Method == "CASH" {
    // 1. Dapatkan atau buat cash drawer aktif
    drawer, err := cashDrawerSvc.GetCurrentDrawer(input.CompanyID)
    if err != nil {
        // Jika tidak ada, buat drawer baru
        openInput := types.OpenDrawerInput{
            OpeningBalance: p.Amount,
            WarehouseID:   input.WarehouseID,
        }
        drawer, err = cashDrawerSvc.OpenDrawer(input.CompanyID, userID, openInput)
    }
    
    // 2. Catat sebagai cash-in
    cashInInput := types.CashInOutInput{
        Amount: p.Amount,
        Reason: "Penjualan #" + createdSale.SaleNumber,
    }
    cashDrawerSvc.AddCashIn(drawer.ID, cashInInput, input.CompanyID, createdSale.CashierID)
    
    // 3. Update Sale dengan CashDrawerID
    createdSale.CashDrawerID = &drawer.ID
}
```

---

## Struktur CashDrawerTransaction

**Untuk record di `cash_drawer_transactions`:**

| Field | Value |
|-------|-------|
| Type | "SALE_IN" atau "CASH_IN" |
| Amount | Nominal pembayaran Tunai |
| Reason | "Penjualan #SALE_NUMBER" |
| SaleID | ID transaksi pembelian |
| PaymentMethod | "CASH" |
| BalanceAfter | Otomatis terhitung |
| CreatedBy | ID kasir |

**Untuk update `cash_drawers`:**
- Kolom `expected_balance` +=nominal cash payment

---

## File yang Perlu Diedit

1. `D:\Project\pos_retail\go_backend\internal\services\sales_service.go`
   - Modifikasi fungsi `CreateSale()` (sekitar line 320-333)

---

## Testing

1. **Manual Test:**
   - Buat penjualan dengan pembayaran Tunai
   - Cek `cash_drawer_transactions` - harus ada record baru
   - Cek `cash_drawers.expected_balance` - harus ter-update

2. **Edge Cases:**
   - Tidak ada cash drawer aktif → buat baru otomatis
   - Cash drawer sudah CLOSED → buat error atau buat baru
   - Multi-payment (CASH + Debit) → hanya CASH yang di-cash-in

---

## Estimasi Waktu
   
- Eksplorasi kode: 1 jam
- Implementasi: 2-3 jam
- Testing: 1 jam
- **Total: 4-5 jam**

---

*Catatan: Rencana ini dibuat berdasarkan eksplorasi kode go_backend. Detail implementasi mungkin perlu disesuaikan sesuai kode yang sebenarnya.*