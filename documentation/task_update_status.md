# Task: Update Master `status` -> `is_active`

## Scope

Master data only (convert to boolean `is_active` default `true`):
- users, customers, suppliers, warehouses, products, companies

Transaction data: keep existing multi-status fields (no change).

## Non-Goals
- Do not modify TS `backend/` or `frontend/` code.
- Do not change transaction status enums/logic.

## Risks / Decisions Needed

Some master tables currently use multi-state `status` in Go/DB/UI (e.g. `blocked`, `maintenance`, `blacklisted`, `discontinued`, `suspended`).

Decision options:
1) Pure boolean (apply now): map `status='active'` -> `is_active=true`, else `is_active=false`.
2) Boolean + flags (optional later): preserve special states with dedicated boolean flags.

This task applies option (1).

## Deliverables

- DB migration script(s) for adding `is_active` (+ dedicated flags) to master tables and backfilling from `status`.
- Go models updated to include `is_active` (and flags if applicable).
- Repositories/services/handlers updated to:
  - read/write `is_active`
  - accept filters by `is_active`
  - keep `status` temporarily for backward compatibility (optional but safest)
- Swagger updated and regenerated (`swag init`).
- Smoke tests: build + basic curl checks.

## Task List (Formatted)

- [ ] 1. Inventory current `status` usage
      - Grep master handlers/services/repositories for JSON field `status`
      - List endpoints that return/filter/mutate status (toggle/deactivate/delete)
      - Output: short table of endpoints + current behavior

- [ ] 2. Confirm mapping rules
      - Backfill rule for `is_active` from legacy `status`
      - Confirm whether any legacy non-`active` states should map to `is_active=false` (default: yes)

- [ ] 3. DB migration (master tables only)
      - Add `is_active BOOLEAN NOT NULL DEFAULT true`
      - Backfill from existing `status` (`is_active = (status = 'active')`)
      - Add indexes like `(company_id, is_active)` where applicable

- [ ] 4. Update Go models
      - Add `is_active` JSON field to each master model
      - Keep legacy `status` field DB-only (`json:"-"`) during transition

- [ ] 5. Update repositories
      - List queries support filtering by `is_active`
      - Deactivate operations set `is_active=false` (stop writing `status='inactive'`)
      - Ensure required hard delete endpoints remain hard delete

- [ ] 6. Update services
      - Create defaults: `is_active=true`
      - Update/toggle paths write `is_active`
      - Optional migration: accept `status=active|inactive` and translate to `is_active`

- [ ] 7. Update handlers + request DTOs
      - Accept `is_active` in update payloads
      - Keep routes stable if needed (API parity); only change the stored field

- [ ] 8. Update Swagger + regenerate docs
      - Schemas show `is_active` for master entities
      - Deprecate `status` if still returned
      - Run `swag init -g cmd/server/main.go -o docs`

- [ ] 9. Verify
      - Run `go test ./...`
      - Curl smoke checks: list returns `is_active`, toggle flips it, filter works

## Files to Touch (Expected)

- `go_backend/internal/models/*.go`
- `go_backend/internal/repository/*_repository.go`
- `go_backend/internal/services/*_service.go`
- `go_backend/internal/handlers/*_handler.go`
- `go_backend/internal/types/request/*.go`
- `go_backend/docs/*` (generated)

## Completion Criteria

- Master endpoints consistently use `is_active` for lifecycle.
- Transaction status logic unchanged.
- Swagger reflects `is_active`.
- Build passes (`go test ./...`).
