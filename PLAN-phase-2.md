# Phase 2 — Activities (Core Workflow) ❌

> Back to [overview](PLAN.md)

## Checklist

8. ❌ **Activity domain + store** — `Activity` entity, PostgreSQL repo, migration 007
9. ❌ **Activity API** — CRUD + status transitions + submit, all validated against config
10. ❌ **Audit log** — migration 008, generic audit recording on activity changes
11. ❌ **Business rules enforcement** — max activities/day, blocked days (vacation/holiday), status transitions
12. ❌ **Frontend: Activity form** — dynamic form from config, per-type field rendering
13. ❌ **Frontend: Planner** — weekly/monthly calendar view with activities
14. ❌ **Frontend: Activity detail** — view + report/submit flow

---

## 1. Activity Domain

### 1.1 Domain Model

```go
// internal/domain/activity.go
type Activity struct {
    ID             uuid.UUID
    ActivityType   string            // "visit", "administrative", etc. — key from config
    Status         string            // "planificat", "realizat", "anulat" — key from config
    DueDate        time.Time         // the scheduled date
    Duration       string            // "full_day", "half_day" — key from config
    Routing        *string           // optional routing week
    Fields         map[string]any    // dynamic fields (detalii, feedback, produse_promovate, etc.)
    AccountID      *uuid.UUID        // linked account (required for visits, null for time-off)
    CreatorID      uuid.UUID         // the rep who created it
    JointVisitUID  *uuid.UUID        // optional co-visitor
    TeamID         *uuid.UUID
    SubmittedAt    *time.Time        // when the report was submitted (locks editing)
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      *time.Time        // soft delete
}
```

### 1.2 Database Migration 007

```sql
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_type TEXT NOT NULL,
    status TEXT NOT NULL,
    due_date DATE NOT NULL,
    duration TEXT NOT NULL,
    routing TEXT,
    fields JSONB NOT NULL DEFAULT '{}',
    account_id UUID REFERENCES accounts(id),
    creator_id UUID NOT NULL REFERENCES users(id),
    joint_visit_user_id UUID REFERENCES users(id),
    team_id UUID REFERENCES teams(id),
    submitted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_activities_type ON activities(activity_type);
CREATE INDEX idx_activities_status ON activities(status);
CREATE INDEX idx_activities_due_date ON activities(due_date);
CREATE INDEX idx_activities_creator ON activities(creator_id);
CREATE INDEX idx_activities_account ON activities(account_id);
CREATE INDEX idx_activities_team ON activities(team_id);
CREATE INDEX idx_activities_fields ON activities USING GIN(fields);

ALTER TABLE activities ENABLE ROW LEVEL SECURITY;
CREATE POLICY activities_rep ON activities FOR ALL
    USING (current_setting('app.user_role') IN ('manager','admin')
           OR creator_id = current_setting('app.user_id')::uuid
           OR joint_visit_user_id = current_setting('app.user_id')::uuid);
```

---

## 2. Activity API

### 2.1 Routes

```
GET    /api/v1/activities                # list (filtered by type, status, date range, creator)
GET    /api/v1/activities/{id}           # detail
POST   /api/v1/activities                # create (validated against config)
PUT    /api/v1/activities/{id}           # update (validated, blocked if submitted)
DELETE /api/v1/activities/{id}           # soft delete (blocked if submitted)
POST   /api/v1/activities/{id}/submit    # submit report (locks activity, validates submit_required fields)
PATCH  /api/v1/activities/{id}/status    # status transition (validated against config transitions)
```

### 2.2 Validation Flow

```
Client POST/PUT /activities
  → Handler: parse JSON body
  → Handler: call config.ValidateActivity(cfg, activityType, fields, "save")
    → Check activity_type exists in config
    → Check status is valid
    → Check all required fields for this type are present
    → Check select/multi_select values are within allowed options
    → Check account_id is present if activity type requires it
    → Return []FieldError (field key + error message)
  → If errors: return 422 with field-level errors
  → Service: RBAC check (creator owns account, or is admin/manager)
  → Service: business rules (max activities/day, blocked days, status transitions)
  → Store: save to DB
```

Submit flow adds `submit_required` field validation and sets `submitted_at`, making the activity read-only.

> **Status:** Validation functions (`ValidateActivity`, `ValidateStatus`, `ValidateStatusTransition`, `ValidateDuration`) are implemented in `internal/config/validator.go` with full test coverage. Integration into handlers is pending.

### 2.3 Backend Package Structure

```
internal/
├── domain/activity.go               # Activity entity (NEW)
├── domain/audit.go                  # AuditEntry entity (NEW)
├── service/activity_service.go      # Activity CRUD + validation + submit (NEW)
├── store/activity.go                # ActivityRepository interface (NEW)
├── store/audit.go                   # AuditRepository interface (NEW)
├── store/postgres/activity.go       # PostgreSQL implementation (NEW)
├── store/postgres/audit.go          # PostgreSQL implementation (NEW)
├── api/activity_handler.go          # HTTP handlers (NEW)
└── api/router.go                    # Add new routes (MODIFY)
```

---

## 3. Audit Log

### 3.1 Migration 008

```sql
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type TEXT NOT NULL,          -- "activity", "account"
    entity_id UUID NOT NULL,
    event_type TEXT NOT NULL,           -- "created", "status_changed", "submitted", "field_updated"
    actor_id UUID NOT NULL REFERENCES users(id),
    old_value JSONB,
    new_value JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_actor ON audit_log(actor_id);
CREATE INDEX idx_audit_created ON audit_log(created_at);
```

> Note: `lead_events` table is kept as-is. New audit entries go to `audit_log`. If leads are fully deprecated later, we can drop `lead_events`.

---

## 4. Business Rules

- **Max activities per day:** `rules.max_activities_per_day` (default 10)
- **Blocked days:** Activity types with `blocks_field_activities: true` (vacation, public_holiday) prevent scheduling field activities on that day
- **Status transitions:** Only allowed transitions from `activities.status_transitions` config
- **Submit lock:** Once `submitted_at` is set, activity cannot be edited or deleted
- **Account required:** Visit-type activities require `account_id`

---

## 5. Frontend: Activities ❌

### 5.1 New Routes

| Route                    | Component        | Description                                        |
| ------------------------ | ---------------- | -------------------------------------------------- |
| `/planner`               | `Planner`        | Weekly/monthly calendar view of activities          |
| `/planner/daily`         | `PlannerDaily`   | Daily agenda view with time slots                   |
| `/activities/new`        | `ActivityForm`   | Create activity (dynamic form from config)          |
| `/activities/:id`        | `ActivityDetail` | Activity detail + report/submit                     |
| `/activities/:id/edit`   | `ActivityForm`   | Edit activity (blocked if submitted)                |

### 5.2 Dynamic Form Rendering

The `ActivityForm` component does not hardcode fields. It:

1. Fetches tenant config via `GET /api/v1/config` (cached with TanStack Query, staleTime: Infinity)
2. On activity type selection, looks up the type's `fields` array from config
3. Renders each field based on its `type`:
   - `text` → `<input type="text">`
   - `select` → `<select>` with options from config (resolved via `options_ref`)
   - `multi_select` → multi-select component with search
   - `relation` → lookup/search component (accounts or users)
   - `date` → date picker
4. Marks required fields with visual indicator
5. On save: sends `{ activity_type, status, due_date, duration, fields: { ... } }` — backend validates

### 5.3 TypeScript Types

```typescript
// types/activity.ts
interface Activity {
  id: string
  activity_type: string
  status: string
  due_date: string
  duration: string
  routing?: string
  fields: Record<string, unknown>
  account_id?: string
  account?: Account  // populated on detail
  creator_id: string
  joint_visit_user_id?: string
  team_id?: string
  submitted_at?: string
  created_at: string
  updated_at: string
}
```

### 5.4 TanStack Query Hooks

```typescript
// services/activities.ts
useActivities(filters)               // GET /activities
useActivity(id)                      // GET /activities/:id
useCreateActivity()                  // POST /activities
useUpdateActivity()                  // PUT /activities/:id
useDeleteActivity()                  // DELETE /activities/:id
useSubmitActivity()                  // POST /activities/:id/submit
usePatchActivityStatus()             // PATCH /activities/:id/status
```
