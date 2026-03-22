# Pebblr — DrMax MVP Implementation Plan

## Context

**Client:** DrMax Romania — pharmaceutical field sales CRM for Medical Division Team (18 reps, 3 managers).

**Current state:** DrMax runs on Twenty CRM with a fragile per-user object duplication hack (54 custom objects, 126 workflows, PowerShell webhook) to work around Twenty's lack of row-level security. Pebblr replaces this with a proper multi-tenant CRM with native RBAC.

**Current Pebblr codebase:** Has generic domain entities (Lead, Customer, User, Team, CalendarEvent) with RBAC, event audit trail, PostgreSQL RLS, React+TanStack frontend. Needs domain evolution to support pharmaceutical field sales workflows.

**Key design constraint:** Nothing client-specific is hardcoded. Enums (statuses, activity types, specialties, products, etc.) and field-level requirements are driven by a YAML tenant configuration file. Validation happens at the API layer against this config, not via DB constraints on enum values.

---

## 1. Tenant Configuration System

### 1.1 YAML Config File

Create `config/tenant.yaml` (loaded at server startup, path via CLI flag `--tenant-config`).

```yaml
# config/tenant.yaml — DrMax example
tenant:
  name: "DrMax Romania"
  locale: "ro"

accounts:
  types:
    - key: doctor
      label: "Doctor"
      fields:
        - { key: name, type: text, required: true, editable: false }
        - { key: specialitate, type: select, required: false, editable: false, options_ref: specialties }
        - { key: potential, type: select, required: false, editable: true, options_ref: classifications }
        - { key: oras, type: text, required: false, editable: false }
        - { key: judet, type: text, required: false, editable: false }
        - { key: adresa, type: text, required: false, editable: false }
    - key: pharmacy
      label: "Farmacie"
      fields:
        - { key: name, type: text, required: true, editable: false }
        - { key: tip, type: select, required: false, editable: false, options_ref: pharmacy_types }
        - { key: oras, type: text, required: false, editable: false }
        - { key: judet, type: text, required: false, editable: false }
        - { key: adresa, type: text, required: false, editable: false }

activities:
  statuses:
    - { key: planificat, label: "Planificat", initial: true }
    - { key: realizat, label: "Realizat" }
    - { key: anulat, label: "Anulat" }
  status_transitions:
    planificat: [realizat, anulat]
    realizat: []       # terminal
    anulat: []         # terminal
  durations:
    - { key: full_day, label: "Full Day" }
    - { key: half_day, label: "Half Day" }
  types:
    - key: visit
      label: "Vizită"
      category: field          # field vs non_field — drives reporting splits
      fields:
        - { key: account_id, type: relation, required: true }
        - { key: tip_vizita, type: select, required: true, options: [f2f, remote] }
        - { key: produse_promovate, type: multi_select, required: false, options_ref: products }
        - { key: feedback, type: text, required: false }
        - { key: detalii, type: text, required: false }
        - { key: duration, type: select, required: true, options_ref: durations }
        - { key: partener_vizita, type: text, required: false }
        - { key: joint_visit_user_id, type: relation, required: false }
      submit_required:         # fields required at submit (report) time, beyond save-time requirements
        - produse_promovate
        - feedback
    - key: administrative
      label: "Administrative"
      category: non_field
      fields:
        - { key: duration, type: select, required: true, options_ref: durations }
        - { key: detalii, type: text, required: false }
    - key: business_travel
      label: "Business Travel"
      category: non_field
      fields:
        - { key: duration, type: select, required: true, options_ref: durations }
        - { key: detalii, type: text, required: false }
    - key: company_event
      label: "Company Event"
      category: non_field
      fields:
        - { key: duration, type: select, required: true, options_ref: durations }
        - { key: detalii, type: text, required: false }
    - key: cycle_meeting
      label: "Cycle Meeting"
      category: non_field
      fields:
        - { key: duration, type: select, required: true, options_ref: durations }
        - { key: detalii, type: text, required: false }
    - key: team_meeting
      label: "Team Meeting"
      category: non_field
      fields:
        - { key: duration, type: select, required: true, options_ref: durations }
        - { key: detalii, type: text, required: false }
    - key: training
      label: "Training"
      category: non_field
      fields:
        - { key: duration, type: select, required: true, options_ref: durations }
        - { key: detalii, type: text, required: false }
    - key: public_holiday
      label: "Public Holiday"
      category: non_field
      blocks_field_activities: true
      fields:
        - { key: duration, type: select, required: true, options_ref: durations }
    - key: vacation
      label: "Vacation"
      category: non_field
      blocks_field_activities: true
      fields:
        - { key: duration, type: select, required: true, options_ref: durations }
    - key: pauza_de_masa
      label: "Pauză de masă"
      category: non_field
      fields:
        - { key: duration, type: select, required: true, options_ref: durations }

  routing_options:
    - { key: saptamana_1, label: "Săptămâna 1" }
    - { key: saptamana_2, label: "Săptămâna 2" }
    - { key: saptamana_3, label: "Săptămâna 3" }

options:
  specialties:
    - { key: cardiologie, label: "Cardiologie" }
    - { key: internal_medicine, label: "Medicină Internă" }
    - { key: family_and_general_medicine, label: "Medicină de Familie" }
    - { key: emergency_medicine, label: "Medicină de Urgență" }
    - { key: gastroenterologie, label: "Gastroenterologie" }
    - { key: geriatrics, label: "Geriatrie" }
    - { key: neurologie, label: "Neurologie" }
    - { key: pediatrie, label: "Pediatrie" }
    - { key: pulmonology, label: "Pneumologie" }
    - { key: other, label: "Altele" }
    # ... (full list imported from TGA — ~48 more)

  classifications:
    - { key: a, label: "A" }
    - { key: b, label: "B" }
    - { key: c, label: "C" }

  pharmacy_types:
    - { key: lant, label: "Lanț" }

  products:
    # Per-quarter product list — managed by marketing/admin
    - { key: product_1, label: "Product 1" }
    # ... populated per deployment

rules:
  frequency:
    # Minimum visits per month per classification
    a: 4
    b: 2
    c: 1
  max_activities_per_day: 10
  default_visit_duration_minutes:
    doctor: 30
    pharmacy: 15
  visit_duration_step_minutes: 15
  recovery:
    weekend_activity_flag: true
    recovery_window_days: 5
    recovery_type: full_day
```

### 1.2 Go Implementation

**New package:** `internal/config/`

- `tenant.go` — Structs mirroring the YAML schema: `TenantConfig`, `AccountTypeConfig`, `ActivityTypeConfig`, `FieldConfig`, `OptionDef`, `StatusDef`, `RulesConfig`
- `loader.go` — `Load(path string) (*TenantConfig, error)` — reads YAML, validates internal consistency (e.g., `options_ref` keys resolve, status transitions reference valid statuses)
- `validator.go` — `ValidateActivity(cfg *TenantConfig, activityType string, fields map[string]any, phase string) []FieldError` — validates field values and required-ness against config. `phase` is `"save"` or `"submit"` (submit enforces additional required fields)

**Config is injected** into services and handlers via constructor (DI). It is **not** stored in the DB — it's read once at startup. The API exposes it read-only so the frontend can render dynamic forms.

### 1.3 API Endpoint

```
GET /api/v1/config
```

Returns the tenant config (sans internal-only fields). Frontend uses this to:
- Render dropdown options dynamically
- Know which fields are required per activity type
- Show/hide fields based on activity type selection
- Display labels in the configured locale

---

## 2. Domain Model Evolution

### 2.1 What Changes

The current generic Lead/Customer/CalendarEvent model evolves to support the pharmaceutical field sales domain. The key insight: **Accounts** (doctors, pharmacies) are contacts to visit; **Activities** are the work units (visits, time-off, etc.).

| Current Entity     | Evolution                                  | Notes                                                    |
| ------------------ | ------------------------------------------ | -------------------------------------------------------- |
| `Customer`         | → **`Account`**                            | Renamed. Gets `account_type` (doctor/pharmacy from config). Custom fields stored as JSONB. |
| `Lead`             | → **Removed** (or kept for future use)     | DrMax doesn't have "leads" — they have accounts + activities. Keep the table/code but deprioritize. |
| `CalendarEvent`    | → **`Activity`**                           | Becomes the core entity. Gets `activity_type`, `status`, `duration`, `routing`, dynamic fields as JSONB. |
| `User`             | Unchanged                                  | Already has role, team, external ID.                     |
| `Team`             | Unchanged                                  | Already has manager.                                     |
| (new)              | **`Territory`**                            | Assignment of accounts to users. A user's territory = the set of accounts assigned to them. |
| (new)              | **`ActivityReport`**                       | The "call report" / "store visit report" — submitted after an activity is realized. Locks the activity. |
| `lead_events`      | → **`audit_log`** (rename)                 | Generalize to audit any entity, not just leads.          |

### 2.2 New Domain Types

```go
// internal/domain/account.go
type Account struct {
    ID          uuid.UUID
    AccountType string            // "doctor", "pharmacy" — key from config
    Name        string
    Fields      map[string]any    // dynamic fields from config (specialitate, potential, tip, etc.)
    AssigneeID  uuid.UUID         // rep who owns this account (territory)
    TeamID      *uuid.UUID
    ImportedAt  *time.Time        // when imported from external system
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

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

// internal/domain/territory.go
type Territory struct {
    UserID    uuid.UUID
    AccountID uuid.UUID
    AssignedAt time.Time
}
```

### 2.3 Dynamic Fields Strategy

**Problem:** Different tenants need different fields on accounts and activities. Hardcoding columns is not extensible.

**Solution:** Core columns (ID, type, status, dates, foreign keys) are real columns. Tenant-specific fields live in a `fields JSONB` column.

- The **config** defines which fields exist and their types
- The **API handler** validates field values against config before saving
- The **DB** stores them as JSONB — no schema migration needed when a tenant changes their config
- The **frontend** renders forms dynamically from the config

**Why not EAV?** JSONB is simpler, faster to query (GIN index), and sufficient for single-tenant deployments. If multi-tenant with shared DB becomes a requirement, we can revisit.

---

## 3. Database Migrations

### Migration 006: Accounts table

```sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_type TEXT NOT NULL,          -- "doctor", "pharmacy"
    name TEXT NOT NULL,
    fields JSONB NOT NULL DEFAULT '{}',  -- dynamic fields from config
    assignee_id UUID REFERENCES users(id),
    team_id UUID REFERENCES teams(id),
    imported_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_type ON accounts(account_type);
CREATE INDEX idx_accounts_assignee ON accounts(assignee_id);
CREATE INDEX idx_accounts_team ON accounts(team_id);
CREATE INDEX idx_accounts_fields ON accounts USING GIN(fields);

-- RLS
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;
-- rep: own accounts only
CREATE POLICY accounts_rep ON accounts FOR ALL
    USING (current_setting('app.user_role') IN ('manager','admin')
           OR assignee_id = current_setting('app.user_id')::uuid);
```

### Migration 007: Activities table

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

### Migration 008: Audit log (generalized from lead_events)

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

> Note: `lead_events` table is kept as-is for backward compatibility. New audit entries go to `audit_log`. If leads are fully deprecated later, we can drop `lead_events`.

---

## 4. Backend Implementation

### 4.1 Package Structure (new/modified)

```
internal/
├── config/
│   ├── tenant.go          # TenantConfig structs
│   ├── loader.go          # YAML loading
│   └── validator.go       # Field validation against config
├── domain/
│   ├── account.go         # Account entity (NEW)
│   ├── activity.go        # Activity entity (NEW)
│   ├── audit.go           # AuditEntry entity (NEW)
│   ├── user.go            # (unchanged)
│   ├── team.go            # (unchanged)
│   └── ...                # lead.go, customer.go kept but deprioritized
├── service/
│   ├── account_service.go  # Account CRUD + RBAC (NEW)
│   ├── activity_service.go # Activity CRUD + validation + submit (NEW)
│   └── ...
├── store/
│   ├── account.go          # AccountRepository interface (NEW)
│   ├── activity.go         # ActivityRepository interface (NEW)
│   ├── audit.go            # AuditRepository interface (NEW)
│   └── postgres/
│       ├── account.go      # (NEW)
│       ├── activity.go     # (NEW)
│       └── audit.go        # (NEW)
├── api/
│   ├── account_handler.go  # (NEW)
│   ├── activity_handler.go # (NEW)
│   ├── config_handler.go   # GET /config (NEW)
│   └── router.go           # add new routes
```

### 4.2 API Routes (new)

```
GET    /api/v1/config                    # tenant config for frontend
GET    /api/v1/accounts                  # list (filtered by type, assignee, territory)
GET    /api/v1/accounts/{id}             # detail
POST   /api/v1/accounts                  # create (admin/import)
PUT    /api/v1/accounts/{id}             # update editable fields
GET    /api/v1/activities                # list (filtered by type, status, date range, creator)
GET    /api/v1/activities/{id}           # detail
POST   /api/v1/activities                # create (validated against config)
PUT    /api/v1/activities/{id}           # update (validated, blocked if submitted)
DELETE /api/v1/activities/{id}           # soft delete (blocked if submitted)
POST   /api/v1/activities/{id}/submit    # submit report (locks activity, validates submit_required fields)
PATCH  /api/v1/activities/{id}/status    # status transition (validated against config transitions)
```

### 4.3 Validation Flow

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
  → Service: business rules (not exceeding max activities/day, no field activities on blocked days, etc.)
  → Store: save to DB
```

Submit flow adds `submit_required` field validation and sets `submitted_at`, making the activity read-only.

### 4.4 Account Import

Accounts (doctors, pharmacies) come from external data (TGA exports, Dr. Max internal DB). For MVP:

- `POST /api/v1/accounts/import` — accepts a JSON array of accounts, upserts by external ID
- Triggered by a script/job, not by reps
- Only admins can import
- `imported_at` timestamp tracks last import
- Imported fields marked `editable: false` in config cannot be changed by reps

---

## 5. Frontend Implementation

### 5.1 New Pages / Routes

| Route                          | Component              | Description                                                |
| ------------------------------ | ---------------------- | ---------------------------------------------------------- |
| `/accounts`                    | `AccountList`          | Filterable list of all accounts (doctors + pharmacies)     |
| `/accounts?type=doctor`        | (same, filtered)       | Doctor list                                                |
| `/accounts?type=pharmacy`      | (same, filtered)       | Pharmacy list                                              |
| `/accounts/:id`                | `AccountDetail`        | Account detail + associated activities                     |
| `/planner`                     | `Planner`              | Weekly/monthly calendar view of activities                 |
| `/planner/daily`               | `PlannerDaily`         | Daily agenda view with time slots                          |
| `/activities/new`              | `ActivityForm`         | Create activity (dynamic form from config)                 |
| `/activities/:id`              | `ActivityDetail`       | Activity detail + report/submit                            |
| `/activities/:id/edit`         | `ActivityForm`         | Edit activity (blocked if submitted)                       |
| `/dashboard`                   | `Dashboard`            | Updated: planned vs realized, coverage, field vs non-field |

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

### 5.3 New TypeScript Types

```typescript
// types/config.ts
interface TenantConfig {
  tenant: { name: string; locale: string }
  accounts: { types: AccountTypeConfig[] }
  activities: {
    statuses: StatusDef[]
    status_transitions: Record<string, string[]>
    durations: OptionDef[]
    types: ActivityTypeConfig[]
    routing_options: OptionDef[]
  }
  options: Record<string, OptionDef[]>
  rules: RulesConfig
}

interface ActivityTypeConfig {
  key: string
  label: string
  category: 'field' | 'non_field'
  fields: FieldConfig[]
  submit_required?: string[]
  blocks_field_activities?: boolean
}

interface FieldConfig {
  key: string
  type: 'text' | 'select' | 'multi_select' | 'relation' | 'date'
  required: boolean
  options?: string[]
  options_ref?: string
}

// types/account.ts
interface Account {
  id: string
  account_type: string
  name: string
  fields: Record<string, unknown>
  assignee_id?: string
  team_id?: string
  imported_at?: string
  created_at: string
  updated_at: string
}

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

### 5.4 New TanStack Query Hooks

```typescript
// services/config.ts
useConfig()                                    // GET /config, staleTime: Infinity

// services/accounts.ts
useAccounts(filters)                           // GET /accounts
useAccount(id)                                 // GET /accounts/:id
useCreateAccount()                             // POST /accounts
useUpdateAccount()                             // PUT /accounts/:id

// services/activities.ts
useActivities(filters)                         // GET /activities
useActivity(id)                                // GET /activities/:id
useCreateActivity()                            // POST /activities
useUpdateActivity()                            // PUT /activities/:id
useDeleteActivity()                            // DELETE /activities/:id
useSubmitActivity()                            // POST /activities/:id/submit
usePatchActivityStatus()                       // PATCH /activities/:id/status
```

---

## 6. Implementation Phases

### Phase 1 — Foundation (config + accounts)

1. **Tenant config system** — `internal/config/` package, YAML loading, validation, tests
2. **Config API endpoint** — `GET /api/v1/config`
3. **Account domain + store** — `Account` entity, PostgreSQL repo with JSONB fields, migration 006
4. **Account API** — CRUD handlers, RBAC (rep sees own, manager sees team)
5. **Account import endpoint** — bulk upsert for admin/scripts
6. **Frontend: Account list + detail** — dynamic field rendering from config
7. **Seed script** — import sample DrMax doctor/pharmacy data

### Phase 2 — Activities (core workflow)

8. **Activity domain + store** — `Activity` entity, PostgreSQL repo, migration 007
9. **Activity API** — CRUD + status transitions + submit, all validated against config
10. **Audit log** — migration 008, generic audit recording on activity changes
11. **Business rules enforcement** — max activities/day, blocked days (vacation/holiday), status transitions
12. **Frontend: Activity form** — dynamic form from config, per-type field rendering
13. **Frontend: Planner** — weekly/monthly calendar view with activities
14. **Frontend: Activity detail** — view + report/submit flow

### Phase 3 — Reporting & Dashboard

15. **Dashboard stats API** — planned vs realized, coverage, field vs non-field, per user/team/period
16. **Frontend: Dashboard** — KPI cards, charts, filters
17. **Joint visit** — co-visitor association, activity visible to both users
18. **Frequency tracking** — visits per account vs target from config rules

### Phase 4 — Phase 2 Optimizations (post go-live)

19. Weekend activity + recovery days
20. Drag & drop calendar
21. Copy-paste activities
22. Advanced filtering with saved filters
23. Target group management (quarterly)
24. Plan generation (rule-based monthly plan proposal)

---

## 7. Data Migration from Twenty CRM

For go-live, existing Twenty CRM data needs to be imported:

- **Doctors** — export from per-user `*Doctori` objects → `POST /accounts/import`
- **Pharmacies** — export from per-user `*Farmacii` objects → `POST /accounts/import`
- **Historical activities** — export from per-user `*Tasks` + master `tasks` → `POST /activities/import` (new endpoint, admin only)

Script in `scripts/import-twenty.sh` or a Go CLI tool under `cmd/import/`.

---

## 8. What Stays, What Goes

| Current Code                | Decision     | Rationale                                                    |
| --------------------------- | ------------ | ------------------------------------------------------------ |
| `internal/domain/lead.go`   | **Keep**     | May be useful later; not in the way                          |
| `internal/domain/customer.go`| **Keep**    | Account replaces it for DrMax, but no need to delete          |
| `internal/api/lead_handler` | **Keep**     | Still functional, just not used by DrMax frontend             |
| `internal/api/calendar_*`   | **Keep**     | Activity replaces it, but keep for backward compat            |
| `internal/events/`          | **Keep**     | Event types still useful; audit_log extends this concept      |
| `internal/rbac/`            | **Extend**   | Add `CanViewAccount`, `CanUpdateActivity`, `ScopeAccountQuery` |
| `migrations/001-005`        | **Keep**     | Don't touch existing schema; add new tables alongside         |
| Frontend routes             | **Extend**   | Add new routes; existing ones stay but won't be in DrMax nav  |

---

## 9. Open Questions

1. **Account territory assignment** — Is territory = "the accounts assigned to this user"? Or is there a geographic territory entity? → For MVP: territory = set of accounts with `assignee_id` = user. No separate territory table.

2. **Data import frequency** — One-time import or periodic sync? → Start with one-time + manual re-import. Automated sync is Phase 4+.

3. **CLM (Closed Loop Marketing)** — Material tracking during visits. → Deprioritize for MVP. Can be added as another dynamic field type later.

4. **IQVIA integration** — External market data for doctor potential scoring. → Out of scope for MVP. Potential (A/B/C) is manually set or imported.

5. **Plan generation algorithm** — Auto-propose monthly plan based on frequency rules + priority. → Phase 4. For MVP, reps create activities manually.

6. **Retrospective edit restrictions** — Can reps edit past activities? → Configurable in `rules` section of tenant YAML. Default: allow edits up to N days back.
