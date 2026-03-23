# Phase 2 — Activities (Core Workflow) 🔧

> Back to [overview](PLAN.md)

## Checklist

9. ✅ **Activity domain + store** — `Activity` + `AuditEntry` entities, PostgreSQL repos, migrations 009–011, RBAC enforcer
10. ✅ **Activity API** — CRUD + status transitions + submit, all validated against config
11. ✅ **Audit log** — migration 011, `AuditRepository` interface + PostgreSQL impl
12. ✅ **Business rules enforcement** — max activities/day, blocked days, target required, status transitions, submit lock
13. ✅ **Frontend: Activity form** — config-driven dynamic form, create/edit routes, target search, multi-select fields, 12 tests
14. ❌ **Frontend: Planner** — weekly/monthly calendar view (replaces old calendar page)
15. ✅ **Frontend: Activity detail** — view + report/submit flow, status transitions, edit link
16. ✅ **Remove calendar_event code** — backend fully removed; frontend calendar files remain until Planner lands

---

## 1. Activity Domain ✅

### 1.1 Domain Model ✅

Implemented in `internal/domain/activity.go` and `internal/domain/audit.go`.

Uses string IDs (matching Target pattern), JSONB dynamic fields, soft delete. `IsSubmitted()` helper method checks lock state.

### 1.2 Database Migrations ✅

- **Migration 009** (`009_drop_calendar_events`) — drops `calendar_events` table
- **Migration 010** (`010_activities`) — creates `activities` table with all indexes + RLS policy (reps see own + joint visits, managers/admins see all)
- **Migration 011** (`011_audit_log`) — creates `audit_log` table with entity/actor/time indexes

### 1.3 Repository Interfaces ✅

- `store/activity_store.go` — `ActivityRepository`: Get, List (scoped), Create, Update, SoftDelete, CountByDate
- `store/audit_store.go` — `AuditRepository`: Record, ListByEntity
- `store/postgres/activity_repository.go` — full PostgreSQL implementation with RBAC-scoped queries
- `store/postgres/audit_repository.go` — PostgreSQL implementation

### 1.4 RBAC ✅

Added to `rbac.Enforcer` interface and `policyEnforcer`:
- `CanViewActivity` — rep sees own + joint visits, manager sees team, admin sees all
- `CanUpdateActivity` — rep sees own only (not joint), manager sees team, admin sees all
- `CanDeleteActivity` — same as update
- `ScopeActivityQuery` — returns `ActivityScope` for list queries
- 7 new tests covering all roles and edge cases

---

## 2. Activity API ✅

### 2.1 Routes ✅

All 7 endpoints implemented in `internal/api/activity_handler.go` and wired in `internal/api/router.go`:

```
GET    /api/v1/activities                # list (filtered by type, status, date range, creator, target, team)
GET    /api/v1/activities/{id}           # detail
POST   /api/v1/activities                # create (validated against config)
PUT    /api/v1/activities/{id}           # update (validated, blocked if submitted)
DELETE /api/v1/activities/{id}           # soft delete (blocked if submitted) → 204
POST   /api/v1/activities/{id}/submit    # submit report (locks activity, validates submit_required fields)
PATCH  /api/v1/activities/{id}/status    # status transition (validated against config transitions)
```

### 2.2 Validation Flow ✅

```
Client POST/PUT /activities
  → Handler: parse JSON body, validate activityType + dueDate presence
  → Service: validateCore — checks type exists in config, status valid, duration valid
  → Service: config.ValidateActivity(cfg, activityType, fields, "save")
    → Check all required fields for this type are present
    → Check select/multi_select values are within allowed options
    → Return []FieldError (field key + error message)
  → If errors: return 422 with field-level errors in { error, fields } response
  → Service: RBAC check (creator owns activity, or is admin/manager)
  → Service: business rules (max activities/day, status transitions)
  → Store: save to DB
  → Audit: record mutation
```

Submit flow adds `submit_required` field validation and sets `submitted_at`, locking the activity.

### 2.3 Service Layer ✅

**`internal/service/activity_service.go`** — `ActivityService` with DI constructor:

- `Create` — sets creator/team from actor, validates core + config fields, checks max/day, creates, audits
- `Get` — fetches + RBAC CanViewActivity check
- `List` — RBAC-scoped via ScopeActivityQuery + filters
- `Update` — fetches existing, RBAC check, submit lock check, validates, preserves immutable fields, audits
- `Delete` — RBAC + submit lock check, soft deletes, audits
- `Submit` — submit-phase validation (stricter required fields), sets SubmittedAt, audits
- `PatchStatus` — validates status + transition against config, audits with old/new values

New sentinel errors: `ErrSubmitted`, `ErrMaxActivities`, `ValidationErrors` (wraps `[]config.FieldError`)

### 2.4 Handler Layer ✅

**`internal/api/activity_handler.go`** — `ActivityHandler` backed by `ActivityServicer` interface:

- Date filters parsed as `YYYY-MM-DD`
- Validation errors return 422 with `{ error, fields }` envelope
- Submitted conflicts return 409
- Delete returns 204 No Content
- All responses use consistent JSON envelopes (`{ activity }`, `{ items, total, page, limit }`)

### 2.5 Backend Package Structure ✅

```
internal/
├── domain/activity.go               # Activity entity ✅
├── domain/audit.go                  # AuditEntry entity ✅
├── service/activity_service.go      # Activity CRUD + validation + submit ✅
├── service/activity_service_test.go # 22 tests ✅
├── store/activity_store.go          # ActivityRepository interface ✅
├── store/audit_store.go             # AuditRepository interface ✅
├── store/postgres/activity_repository.go  # PostgreSQL implementation ✅
├── store/postgres/audit_repository.go     # PostgreSQL implementation ✅
├── api/activity_handler.go          # HTTP handlers ✅
├── api/activity_handler_test.go     # 16 tests ✅
└── api/router.go                    # Activity routes wired ✅
```

Wiring: `cmd/api/main.go` instantiates `ActivityService(db.Activities(), db.Audit(), enforcer, tenantCfg)` → `ActivityHandler` → `RouterConfig.ActivityHandler`.

RBAC: ✅ `CanViewActivity`, `CanUpdateActivity`, `CanDeleteActivity`, `ScopeActivityQuery` added to enforcer with tests.

---

## 3. Business Rules ✅

- ✅ **Max activities per day:** `rules.max_activities_per_day` — enforced in `ActivityService.Create` via `CountByDate`
- ✅ **Blocked days:** Activity types with `blocks_field_activities: true` (vacation, public_holiday) prevent scheduling field activities on that day — enforced in `ActivityService.Create` via `HasActivityWithTypes`
- ✅ **Status transitions:** Only allowed transitions from `activities.status_transitions` config — enforced in `PatchStatus`
- ✅ **Submit lock:** Once `submitted_at` is set, activity cannot be edited or deleted — enforced in Update, Delete, Submit, PatchStatus
- ✅ **Target required:** Field-category activities require `target_id` — enforced in `ActivityService.Create` based on activity type category

---

## 4. Frontend: Activities 🔧

### 4.1 Routes

| Route                    | Component           | Status | Description                                        |
| ------------------------ | ------------------- | ------ | -------------------------------------------------- |
| `/activities/new`        | `NewActivityPage`   | ✅     | Create activity (dynamic form from config)          |
| `/activities/:id`        | `ActivityDetailPage`| ✅     | Activity detail + report/submit + status transitions|
| `/activities/:id/edit`   | `EditActivityPage`  | ✅     | Edit activity (blocked if submitted)                |
| `/planner`               | `Planner`           | ❌     | Weekly/monthly calendar view of activities          |
| `/planner/daily`         | `PlannerDaily`      | ❌     | Daily agenda view with time slots                   |

### 4.2 Dynamic Form Rendering ✅

The `ActivityForm` component (`web/src/components/ActivityForm.tsx`) does not hardcode fields. It:

1. Fetches tenant config via `useConfig()` (cached with TanStack Query, staleTime: Infinity)
2. On activity type selection, looks up the type's `fields` array from config
3. Renders each field based on its `type`:
   - `text` → `<input type="text">`
   - `select` → `<select>` with options from config (resolved via `options_ref`)
   - `multi_select` → toggle buttons with options from config
   - `date` → `<input type="date">`
4. Marks required fields with visual `*` indicator
5. Shows target search with dropdown for field-category activities
6. On save: sends `{ activityType, status, dueDate, duration, fields: { ... } }` — backend validates
7. 12 component tests in `ActivityForm.test.tsx`

### 4.3 Activity Detail ✅

The `ActivityDetailPage` (`web/src/routes/activities/$activityId.tsx`) shows:
- Activity header with type, status badge, date, duration
- Submitted lock indicator
- Edit button + Submit Report button (hidden when submitted)
- Dynamic fields resolved via config options
- Status transition buttons (from config `status_transitions`)
- Link to target detail page

### 4.4 TypeScript Types ✅

```typescript
// web/src/types/activity.ts
interface Activity {
  id: string
  activityType: string
  status: string
  dueDate: string
  duration: string
  routing?: string
  fields: Record<string, unknown>
  targetId?: string
  creatorId: string
  jointVisitUserId?: string
  teamId?: string
  submittedAt?: string
  createdAt: string
  updatedAt: string
}

// web/src/types/config.ts — extended with:
interface ActivitiesConfig { statuses, status_transitions, durations, types, routing_options }
interface ActivityTypeConfig { key, label, category, fields, submit_required, blocks_field_activities }
interface RulesConfig { frequency, max_activities_per_day, ... }
```

### 4.5 TanStack Query Hooks ✅

All 7 hooks implemented in `web/src/services/activities.ts`:

```typescript
useActivities(filters)               // GET /activities
useActivity(id)                      // GET /activities/:id
useCreateActivity()                  // POST /activities
useUpdateActivity()                  // PUT /activities/:id
useDeleteActivity()                  // DELETE /activities/:id
useSubmitActivity()                  // POST /activities/:id/submit
usePatchActivityStatus()             // PATCH /activities/:id/status
```

### 4.6 Planner — Next ❌

The old `CalendarGrid.tsx` and `EventCard.tsx` provide useful patterns:
- Month grid layout logic
- Color-coded event cards by type
- Date navigation (prev/next month)

Reuse these patterns in the Planner component, then delete the originals. The Planner adds:
- Week view (in addition to month)
- Daily agenda view
- Activity type color coding from config
- Status indicators (planned/realized/cancelled)
- Drag-and-drop (Phase 4)

### 4.7 Frontend Cleanup (with Planner)

Remove when Planner lands:
- `web/src/routes/calendar/` → replaced by Planner
- `web/src/services/calendar.ts` → replaced by activity service
- `web/src/types/calendar.ts` → replaced by activity types
- `web/src/components/calendar/` → patterns reused in Planner, originals deleted
- Sidebar navigation: replace "Calendar" with "Planner"
