Logging CRUD Plan

1. Tujuan
- Menyimpan log operasi CRUD untuk semua table endpoint backend.
- Menyimpan file log per bulan dengan format folder `yyyy_mm`.
- Menyediakan API untuk membaca log per table dengan `limit` dan `offset`.
- Menyediakan API untuk menyimpan ringkasan log ke `summary.txt`.

2. Struktur Folder
- Root log: `logs/`
- Folder bulanan: `logs/yyyy_mm/` (contoh: `logs/2026_03/`)
- File per table: `logs/yyyy_mm/<table>.log` (contoh: `users.log`)
- File error: `logs/yyyy_mm/error.log`
- File ringkasan: `logs/yyyy_mm/summary.txt`

3. Auto Create Folder
- Folder `logs/` dan `logs/yyyy_mm/` dibuat otomatis saat aplikasi start.
- Tidak perlu membuat folder manual.

4. Sumber Logging
- Logging dilakukan di Service layer, bukan di middleware.
- Workflow update/delete:
  - Ambil old data dari database.
  - Jalankan update/delete.
  - Ambil new data (khusus update) dari database.
  - Kirim old/new ke logger.
- Helper wrapper tersedia di `internal/logger/audit.go`:
  - `AuditCreate(...)`
  - `AuditUpdate(...)`
  - `AuditDelete(...)`
- Jika proses gagal, error dicatat ke `error.log`.

5. Format Log
- Format plain text per baris.
- CREATE:
  - `[2026-03-04 10:30:00] [INFO] [CREATE] [units] user_id=abc company_id=xyz new={"name":"PCS","code":"PCS","is_active":true}`
- UPDATE:
  - `[2026-03-04 10:30:05] [INFO] [UPDATE] [units] user_id=abc company_id=xyz record_id=123 old={"name":"PCS","code":"PCS","is_active":true} new={"name":"PCS","code":"PCS","is_active":false}`
- DELETE:
  - `[2026-03-04 10:30:10] [INFO] [DELETE] [units] user_id=abc company_id=xyz record_id=123 old={"name":"PCS","code":"PCS","is_active":true}`
- ERROR:
  - `[2026-03-04 10:30:12] [ERROR] [UPDATE] [units] user_id=abc company_id=xyz record_id=123 error=...`

6. Summary
- `summary.txt` berisi total CREATE, UPDATE, DELETE, ERROR per table dan TOTAL keseluruhan.
- Ringkasan disimpan saat shutdown server dan bisa dipanggil manual via API.

7. API Logs
- `GET /api/logs/summary`
  - Menampilkan counter summary yang sedang berjalan di memory.
- `POST /api/logs/save`
  - Menyimpan ringkasan ke `logs/yyyy_mm/summary.txt`.
- `GET /api/logs/files`
  - Menampilkan daftar folder bulan dan file log yang tersedia.
- `GET /api/logs/:tahun_bulan/:table?limit=50&offset=0`
  - Membaca log CRUD per table berdasarkan bulan.
  - Format `:tahun_bulan` wajib `yyyy_mm`.
- `GET /api/logs/:tahun_bulan/error?limit=50&offset=0`
  - Membaca `error.log` berdasarkan bulan.

8. Catatan
- Semua endpoint logs berada di route protected dan role `admin` atau `manager`.
- Logging bisa dimatikan dengan env `LOG_ENABLE_CRUD=false`.
- Lokasi log bisa diubah dengan env `LOG_DIR`.
