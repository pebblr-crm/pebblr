# Phase 1 — Foundation (Config + Targets) ✅

> Back to [overview](PLAN.md)

## Checklist

1. ✅ **Tenant config system** — `internal/config/` package, JSON loading, validation, tests
2. ✅ **Config API endpoint** — `GET /api/v1/config`
3. ✅ **Target domain + store** — `Target` entity, PostgreSQL repo with JSONB fields, migration 006
4. ✅ **Target API** — CRUD handlers, RBAC (rep sees own, manager sees team)
5. ✅ **Target import endpoint** — bulk upsert for admin/scripts
6. ❌ **Frontend: Target list + detail** — reuse `DataTable`, status badge, pagination from leads pages
7. ❌ **Remove dead code** — drop Lead, CalendarEvent, lead_events code + frontend pages
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

### 3.2 Frontend: Target List + Detail ❌

| Route                          | Component        | Description                            |
| ------------------------------ | ---------------- | -------------------------------------- |
| `/targets`                     | `TargetList`     | Filterable list of all targets         |
| `/targets?type=doctor`         | (same, filtered) | Doctor list                            |
| `/targets?type=pharmacy`       | (same, filtered) | Pharmacy list                          |
| `/targets/:id`                 | `TargetDetail`   | Target detail + associated activities  |

**Reuse from existing frontend:**
- `DataTable.tsx` — generic TanStack Table wrapper (used by leads page, works as-is for targets)
- Pagination UI pattern from leads list
- Status badge styling approach
- Initials avatar color hashing
- Layout/Sidebar/TopBar unchanged

**New:**
- Dynamic field rendering from tenant config (`useConfig()` hook → render fields per target type)
- Type filter (doctor/pharmacy) dropdown
- Target detail page with associated activities (placeholder until Phase 2)

### 3.3 Remove Dead Code ❌

Remove the following unused code from the generic CRM scaffold:

**Backend (remove now — Lead domain is unused):**
- `internal/domain/lead.go`, `lead_status.go`
- `internal/service/lead_service.go`
- `internal/store/lead_store.go`, `postgres/lead_repository.go`
- `internal/api/lead_handler.go`
- `internal/events/` — entire package (lead-specific event types, recorder, querier)
- `internal/store/event_store.go`, `postgres/event_repository.go`
- `internal/rbac/` — remove `CanViewLead`, `CanAssignLead`, `CanUpdateLead`, `CanDeleteLead`, `ScopeLeadQuery`, `LeadScope`
- Lead routes from `router.go`
- `internal/metrics/` — if lead-specific, remove; if generic, keep

**Backend (remove in Phase 2 when Activity domain replaces it):**
- `internal/domain/calendar_event.go`
- `internal/service/calendar_event_service.go`
- `internal/store/calendar_event_store.go`, `postgres/calendar_event_repository.go`
- `internal/api/calendar_event_handler.go`
- Calendar event routes from `router.go`

**Database:**
- Migration to drop `leads`, `lead_events` tables (new migration 007, before activities migration)
- Migration to drop `calendar_events` table (with activities migration in Phase 2)

**Frontend (remove now):**
- `web/src/routes/leads/` — leads list + detail pages
- `web/src/routes/my-leads/` — my leads page
- `web/src/services/leads.ts`
- `web/src/types/lead.ts`
- `web/src/components/dashboard/UnassignedLeadCard.tsx` — lead-specific widget
- Sidebar navigation entries for leads/my-leads

**Frontend (remove in Phase 2):**
- `web/src/routes/calendar/` — calendar page (replaced by Planner)
- `web/src/services/calendar.ts`
- `web/src/types/calendar.ts`
- `web/src/components/calendar/` — CalendarGrid, EventCard (reuse patterns in Planner, delete originals)

### 3.4 Seed Script 🔧

`scripts/seed.sh` and `scripts/seed-data.sql` exist with sample users/teams. Needs:
- DrMax-specific doctor targets (sample doctors with specialties, classifications)
- DrMax-specific pharmacy targets (sample pharmacies with types)
- Territory assignments (doctors/pharmacies assigned to sample reps)
