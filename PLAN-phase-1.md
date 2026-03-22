# Phase 1 — Foundation (Config + Targets) ✅

> Back to [overview](PLAN.md)

## Checklist

1. ✅ **Tenant config system** — `internal/config/` package, JSON loading, validation, tests
2. ✅ **Config API endpoint** — `GET /api/v1/config`
3. ✅ **Target domain + store** — `Target` entity, PostgreSQL repo with JSONB fields, migration 006
4. ✅ **Target API** — CRUD handlers, RBAC (rep sees own, manager sees team)
5. ✅ **Target import endpoint** — bulk upsert for admin/scripts
6. ✅ **Frontend: Target list + detail** — config-driven list with type filter, dynamic columns, detail page with resolved option labels
7. ✅ **Remove dead code** — dropped Lead, lead_events, events, metrics, dashboard code + frontend pages
8. ✅ **Seed script** — 12 doctor targets, 8 pharmacy targets, territory assignments, updated calendar events

---

## 1. Tenant Configuration System ✅

### 1.1 Tenant Config File ✅

`config/tenant.json` loaded at startup via `--tenant-config` flag.

Defines: tenant info, target types (doctor/pharmacy) with dynamic fields, activity types with field schemas, statuses + transitions, option lists (specialties, classifications, products), business rules (frequency, max activities/day, durations).

### 1.2 Go Implementation ✅

**Package:** `internal/config/`

- `tenant.go` — Structs: `TenantConfig`, `AccountTypeConfig`, `ActivityTypeConfig`, `FieldConfig`, `OptionDef`, `StatusDef`, `RulesConfig`
- `loader.go` — `Load(path string) (*TenantConfig, error)` — reads JSON, validates internal consistency
- `validator.go` — `ValidateActivity()`, `ValidateStatus()`, `ValidateStatusTransition()`, `ValidateDuration()`
- Full test coverage for loading and validation

Config is injected into services/handlers via constructor (DI). Read once at startup.

### 1.3 Config API Endpoint ✅

```
GET /api/v1/config
```

Implemented in `internal/api/config_handler.go`. Returns tenant config for frontend to render dynamic forms, dropdowns, required field indicators, and locale labels.

---

## 2. Target Domain ✅

### 2.1 Domain Model ✅

Replaces the old `Customer` entity. Gets `target_type` (doctor/pharmacy from config) and JSONB dynamic fields.

```go
// internal/domain/target.go
type Target struct {
    ID         uuid.UUID
    TargetType string            // "doctor", "pharmacy" — key from config
    Name       string
    Fields     map[string]any    // dynamic fields (specialitate, potential, tip, etc.)
    AssigneeID uuid.UUID         // rep who owns this target (territory)
    TeamID     *uuid.UUID
    ImportedAt *time.Time
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

### 2.2 Database Migration 006 ✅

Drops the old `customers` table; creates `targets` table with:
- Core columns: id, target_type, name, assignee_id, team_id, imported_at, timestamps
- `fields JSONB` for tenant-specific dynamic fields
- Indexes on type, assignee, team, GIN on fields
- RLS policy: reps see own targets, managers/admins see all

### 2.3 Full Stack ✅

- **Store:** `TargetRepository` interface + PostgreSQL implementation with RBAC-scoped queries (`TargetScope`)
- **Service:** `TargetService` — CRUD with RBAC enforcement and tenant config validation
- **Handler:** `target_handler.go` — List, Get, Create, Update
- **RBAC:** `CanViewTarget`, `CanUpdateTarget`, `ScopeTargetQuery` in enforcer
- **Frontend service:** `web/src/services/targets.ts` — `useTargets`, `useTarget`, `useCreateTarget`, `useUpdateTarget`
- **Frontend types:** `web/src/types/target.ts`

### 2.4 API Routes ✅

```
GET    /api/v1/targets                  # list (filtered by type, assignee)
GET    /api/v1/targets/{id}             # detail
POST   /api/v1/targets                  # create
PUT    /api/v1/targets/{id}             # update editable fields
```

---

## 3. Import, Frontend, Cleanup & Seed ✅

### 3.1 Target Import Endpoint ✅

Implemented:

- `POST /api/v1/targets/import` — accepts JSON array, upserts by `(target_type, external_id)`
- Admin-only, triggered by script/job
- `imported_at` timestamp set automatically on upsert
- Migration 007 adds `external_id` column with partial unique index
- Response includes `created`/`updated` counts and the imported targets
- Full test coverage at service and handler layers

### 3.2 Frontend: Target List + Detail ✅

| Route                          | Component           | Description                            |
| ------------------------------ | ------------------- | -------------------------------------- |
| `/targets`                     | `TargetsPage`       | Filterable list of all targets         |
| `/targets?type=doctor`         | (same, filtered)    | Doctor list                            |
| `/targets?type=pharmacy`       | (same, filtered)    | Pharmacy list                          |
| `/targets/:id`                 | `TargetDetailPage`  | Target detail + associated activities  |

**Implemented:**
- `web/src/types/config.ts` — TypeScript types mirroring Go `TenantConfig`
- `web/src/services/config.ts` — `useConfig()` hook (`GET /api/v1/config`, infinite staleTime)
- `web/src/routes/targets/index.tsx` — Target list with:
  - Config-driven type filter dropdown (doctor/pharmacy from tenant config)
  - Dynamic columns that change based on selected type filter
  - Config-resolved option labels (e.g. key "cardiology" → label "Cardiology")
  - Location display from JSONB fields (city, county)
  - Pagination, stats cards, empty/loading/error states
- `web/src/routes/targets/$targetId.tsx` — Target detail with:
  - Config-driven field rendering for all dynamic fields per target type
  - Resolved option labels from config options map
  - Location composite (address + city + county), type badge
  - Activities section placeholder (wired up in Phase 2)
  - Back-to-list navigation
- Routes registered in `App.tsx`, "Targets" added to `Sidebar.tsx`
- Full test coverage: `index.test.tsx` (14 tests), `$targetId.test.tsx` (8 tests)

### 3.3 Remove Dead Code ✅

Removed all Lead-domain scaffolding that was replaced by the Target domain:

**Backend removed:**
- `internal/domain/lead.go`, `lead_status.go`
- `internal/service/lead_service.go`, `dashboard_service.go`
- `internal/api/lead_handler.go`, `dashboard_handler.go` (+ tests)
- `internal/store/lead_store.go`, `event_store.go` + postgres implementations
- `internal/events/` — entire package
- `internal/metrics/` — entire package (lead-specific)
- Lead RBAC methods (`CanViewLead`, `CanAssignLead`, `CanUpdateLead`, `CanDeleteLead`, `ScopeLeadQuery`, `LeadScope`)
- Lead routes, dashboard routes, metrics routes from `router.go`

**Backend additions during cleanup:**
- `internal/service/errors.go` — extracted `ErrForbidden`/`ErrInvalidInput` sentinel errors (were in deleted `lead_service.go`)
- Test helper files for `api` and `service` packages
- Migration 008: `DROP TABLE leads, lead_events`

**Frontend removed:**
- `web/src/routes/leads/`, `web/src/routes/my-leads/`
- `web/src/services/leads.ts`, `web/src/services/dashboard.ts`
- `web/src/types/lead.ts`, `web/src/types/dashboard.ts`
- `web/src/components/dashboard/UnassignedLeadCard.tsx`
- Sidebar entries for Leads/My Leads + "New Lead" button

**Frontend updated:**
- Dashboard simplified (team performance only, no lead stats)
- `App.tsx` route tree trimmed
- `Sidebar.tsx` cleaned to Dashboard/Targets/Calendar/Team

**Deferred to Phase 2** (CalendarEvent replaced by Activity domain):
- `internal/domain/calendar_event.go`, service, store, handler, routes
- `web/src/routes/calendar/`, `web/src/services/calendar.ts`, `web/src/types/calendar.ts`
- Migration to drop `calendar_events` table

### 3.4 Seed Script ✅

Updated `scripts/seed-data.sql` with pharma-specific target data:

- **12 doctor targets** — specialties from config (cardiology, internal medicine, family medicine, gastroenterology, neurology, pulmonology, geriatrics, pediatrics, emergency medicine), A/B/C classifications, Czech cities (Prague, Brno, Olomouc, Ostrava, Zlin)
- **8 pharmacy targets** — Dr.Max chain locations across both regions
- All targets have `external_id` (DOC-xxx, PHR-xxx) for import compatibility
- Territory assignments: Alice + Bob → North Region (Prague), Carol + Dan → South Region (Brno, Olomouc, Ostrava, Zlin)
- Calendar events updated to reference doctors and pharmacies
- Removed all legacy lead, lead_event, and lead enrichment inserts
- Users, teams, and team members unchanged
