# TODO: Cash-In Penjualan Tunai ke Cash Drawer

## Progress
- [x] Status: IN PROGRESS
- [x] Assigned to: -
- [x] Due date: -

## Masalah
Cash-in dari pembayaran penjualan Tunai tidak masuk ke tabel:
- `cash_drawer_transactions`
- `cash_drawers` (saldo tidak ter-update)

## Estimasi Waktu
4-5 jam

## Tahapan Pengerjaan

### Tahap 1: Eksplorasi Kode
- [x] Review `sales_service.go` - fungsi CreateSale
- [x] Review `cashdrawer_service.go` - fungsi AddCashIn
- [x] Pahami alur data dan dependencies
- Status: DONE

### Tahap 2: Implementasi
- [x] Tambahkan cashDrawerRepo ke SalesService struct
- [x] Modifikasi fungsi NewSalesService untuk accept cashDrawerRepo
- [x] Modifikasi fungsi CreateSale untuk cash payment
- [x] Tambahkan logic GetCurrentDrawer / OpenDrawer jika belum ada
- [x] Tambahkan logic AddCashIn untuk setiap CASH payment
- [ ] Update main.go untuk pass cashDrawerRepo
- Status: IN PROGRESS

### Tahap 3: Testing
- [ ] Manual test: Buat penjualan Tunai
- [ ] Verifikasi: `cash_drawer_transactions` ada record baru
- [ ] Verifikasi: `cash_drawers.expected_balance` ter-update
- [ ] Edge case: Tidak ada drawer aktif
- [ ] Edge case: Multiple payment (CASH + Debit)
- Status: TODO

### Tahap 4: Finalisasi
- [ ] Update frontend laporan cash drawer
- [ ] Review dan cleanup kode
- [ ] Dokumentasi perubahan
- Status: TODO

---

## Catatan

Implementasi telah ditambahkan ke sales_service.go:

1. Di dalam loop payments, setelah buat SalePayment:
2. Cek jika Method == "CASH":
   - Cari cash drawer aktif (status OPEN)
   - Jika tidak ada, buat baru
   - Update expected_balance
   - Buat CashDrawerTransaction dengan Type="SALE_IN"

File yang diedit:
- `internal/services/sales_service.go`
- `cmd/server/main.go`

## Riwayat Perubahan
| Date | Description | Status |
|------|---------------|--------|
| 2026-04-13 | Rencana dibuat | DONE |
| 2026-04-13 | Implementasi sales_service.go | DONE |
| 2026-04-13 | Update main.go | DONE |
| 2026-04-13 | Build berhasil | DONE |