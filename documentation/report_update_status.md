# Report: Update Master `status` -> `is_active`

Last updated: 2026-02-20

## Status

- [x] 1. Inventory current `status` usage
- [x] 2. Confirm mapping rules
- [x] 3. DB migration (master tables only)
- [x] 4. Update Go models
- [x] 5. Update repositories
- [x] 6. Update services
- [x] 7. Update handlers + request DTOs
- [x] 8. Update Swagger + regenerate docs
- [x] 9. Verify

## Notes

- Master data scope: users, customers, suppliers, warehouses, products, companies
- Transaction data unchanged: purchases/GRN/sales/returns/exchanges/stock transfer/opname/invoices/cash drawer

Mapping rule applied:
- master `is_active = (LOWER(status) = 'active')`
- special legacy states (blocked/blacklisted/maintenance/suspended/discontinued) are treated as `is_active=false`

## Change Log

- Added `is_active` to master models and hid legacy `status` from JSON
- Updated list filters to support `is_active` (kept `status=active|inactive` as backward-compat mapping for supplier/customer/product list endpoints)
- Updated create/update/delete flows to write `is_active` (and keep legacy `status` in sync as `active|inactive`)
- Updated auth login/register response to return `is_active` instead of `status`
- Added best-effort DB backfill on AutoMigrate for master tables
- Regenerated Swagger docs (`go_backend/docs/*`)
- Verified build: `go test ./...`
