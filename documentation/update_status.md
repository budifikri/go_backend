# Update Plan: Migrate `status` -> `is_active` (boolean)

## Goal

### Master Data
For master data tables, replace `status` with `is_active` (boolean, default `true`) like `categories` and `units_of_measure`.

Target master entities:
- `supplier`, `user`, `warehouse`, `product`, `customer`, `company`

### Transaction Data
Transaction tables keep multi-status fields (no change), e.g. purchase orders, GRN, stock transfer, stock opname, sales, returns, exchanges, invoices, cash drawer, etc.

## Important Note (Master Data Non-Boolean States)
Some master entities currently carry multi-state `status` values (examples seen in UI/API):
- supplier: `active | inactive | blacklisted`
- warehouse: `active | inactive | maintenance`
- customer: `active | inactive | blocked`
- company: `active | inactive | suspended`
- product: `active | inactive | discontinued`

If master data is standardized to boolean `is_active`, these non-boolean states must be handled explicitly.

Options:
1) Pure boolean (simplest): master tables only have `is_active`; any legacy non-`active` state becomes `is_active=false`.
2) Boolean + flags (optional, later): keep `is_active` + dedicated flags to preserve special states.

This update plan is compatible with both; implement (1) first for consistency.

## Phased Implementation (Recommended, Low Risk)

### Phase 0: Inventory + Decisions
- List all usages of `status` across:
  - DB schema (tables/columns/defaults/indexes)
  - Go models/repos/services/handlers
  - Swagger schemas
  - Frontend UI (filters/toggles/badges)
  - TS `backend/test/*` scripts (expected fields)
- Decide per-entity mapping for master data:
  - `is_active` backfill rule (typically `status = 'active'`)
  - whether to add extra boolean flags (recommended) or map special states to `is_active=false`

### Phase 1: DB Migration (Non-breaking)

Master tables only:
- Add `is_active boolean NOT NULL DEFAULT true`.
- Backfill:
  - `is_active = (status = 'active')`
- Add indexes where filtering by active is common:
  - `(company_id, is_active)` for multi-tenant master tables
- Keep old `status` column for now.

Transaction tables: no schema changes.

### Phase 2: Go Backend Changes (Dual Read/Write)

Master endpoints only:
- Models:
  - Add `IsActive bool `json:"is_active"``
  - Keep `Status string` temporarily for backward compatibility.
- Repositories/services:
  - Read filters: support `is_active` (boolean) and ignore `status` for new code.
  - Write paths:
    - toggles should update `is_active`.
    - create defaults set `is_active=true`.
    - any remaining status transitions map into `is_active` (+ dedicated flags).
- Responses:
  - Return `is_active` (and dedicated flags if used).
  - Keep returning `status` temporarily if existing clients/tests still rely on it.

Transaction endpoints: keep existing multi-status logic.

### Phase 3: API Contract + Swagger
- Update request/response DTOs and swag annotations.
- Document transition:
  - for master data: `status` deprecated; prefer `is_active`.
- Regenerate swagger (`swag init`).

### Phase 4: Frontend Migration

Master screens:
- Replace status rendering/toggles/filters to use `is_active`.
- Update generated API types (`bun run api:gen`) after swagger changes.

Transaction screens: no change.

### Phase 5: TS Tests / Compatibility
- Update TS test scripts to accept `is_active`.
- If keeping `status` temporarily, tests can be migrated gradually.

### Phase 6: Breaking Cleanup (Optional, Later)

Master tables only:
- After all clients migrate:
  - Remove `status` from API.
  - Drop `status` column from DB.
- Add a version note / changelog entry.

## Per-Entity Checklist (Master Data)
- users: table + model + list/filter + create/update defaults + UI toggle
- customers: table + model + list/filter + UI toggle (+ optional `is_blocked`)
- suppliers: table + model + list/filter + UI toggle (+ optional `is_blacklisted`)
- warehouses: table + model + list/filter + UI toggle (+ optional `is_maintenance`)
- products: table + model + list/filter + UI toggle (+ optional `is_discontinued`)
- companies: table + model + list/filter + UI toggle (+ optional `is_suspended`)

## Proposed Mapping Rules (Draft)

Master data backfill (simple + consistent):
- `is_active = (status = 'active')`

Special states become dedicated flags:
- `blacklisted -> is_blacklisted=true`
- `maintenance -> is_maintenance=true`
- `blocked -> is_blocked=true`
- `suspended -> is_suspended=true`
- `discontinued -> is_discontinued=true`

Whether special states also imply `is_active=false` is a business rule decision.

## Rollback Strategy
- Since Phase 1 is additive, rollback is safe:
  - stop writing `is_active` and ignore it.
  - no data loss if `status/state` retained.
