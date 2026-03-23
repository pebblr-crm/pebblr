# Phase 1 — Foundation (Config + Targets) ✅

> Back to [overview](PLAN.md)

## Checklist

1. ✅ **Tenant config system** — `internal/config/` package, JSON loading, validation, tests
2. ✅ **Config API endpoint** — `GET /api/v1/config`
3. ✅ **Target domain + store** — `Target` entity, PostgreSQL repo with JSONB fields, migration 006
4. ✅ **Target API** — CRUD handlers, RBAC (rep sees own, manager sees team)
5. ✅ **Target import endpoint** — bulk upsert for admin/scripts
6. ✅ **Frontend: Target list + detail** — config-driven list with type filter, dynamic columns, detail page with resolved option labels
7. ✅ **Remove dead code** — Lead removed (PR #51, migration 008); CalendarEvent removed with Activity domain (migration 009)
8. 🔧 **Seed script** — exists with sample users/teams; needs DrMax-specific doctor/pharmacy target data

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

## 3. Remaining Work

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

All backend dead code has been removed:

**Backend (removed in PR #51):**
- Lead domain, service, store, handler, events, RBAC methods, routes

**Backend (removed with Activity domain landing):**
- CalendarEvent domain, service, store, handler, tests, routes
- Migration 008 drops `leads` + `lead_events`; migration 009 drops `calendar_events`

**Frontend (removed in PR #51):**
- Leads pages, my-leads page, leads service/types, sidebar entries

**Frontend (remaining — remove when Planner lands in Phase 2, item 14):**
- `web/src/routes/calendar/`, `web/src/services/calendar.ts`, `web/src/types/calendar.ts`, `web/src/components/calendar/`

### 3.4 Seed Script 🔧

`scripts/seed.sh` and `scripts/seed-data.sql` exist with sample users/teams. Needs:
- DrMax-specific doctor targets (sample doctors with specialties, classifications)
- DrMax-specific pharmacy targets (sample pharmacies with types)
- Territory assignments (doctors/pharmacies assigned to sample reps)
