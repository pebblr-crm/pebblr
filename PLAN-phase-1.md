# Phase 1 — Foundation (Config + Accounts) 🔧

> Back to [overview](PLAN.md)

## Checklist

1. ✅ **Tenant config system** — `internal/config/` package, JSON loading, validation, tests
2. ❌ **Config API endpoint** — `GET /api/v1/config`
3. ❌ **Account domain + store** — `Account` entity, PostgreSQL repo with JSONB fields, migration 006
4. ❌ **Account API** — CRUD handlers, RBAC (rep sees own, manager sees team)
5. ❌ **Account import endpoint** — bulk upsert for admin/scripts
6. ❌ **Frontend: Account list + detail** — dynamic field rendering from config
7. 🔧 **Seed script** — exists with sample users/teams/customers/leads; needs DrMax-specific doctor/pharmacy account data

---

## 1. Tenant Configuration System ✅

### 1.1 Tenant Config File ✅

Created `config/tenant.yaml` (loaded at server startup, path via CLI flag `--tenant-config`).

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
    realizat: []
    anulat: []
  durations:
    - { key: full_day, label: "Full Day" }
    - { key: half_day, label: "Half Day" }
  types:
    - key: visit
      label: "Vizită"
      category: field
      fields:
        - { key: account_id, type: relation, required: true }
        - { key: tip_vizita, type: select, required: true, options: [f2f, remote] }
        - { key: produse_promovate, type: multi_select, required: false, options_ref: products }
        - { key: feedback, type: text, required: false }
        - { key: detalii, type: text, required: false }
        - { key: duration, type: select, required: true, options_ref: durations }
        - { key: partener_vizita, type: text, required: false }
        - { key: joint_visit_user_id, type: relation, required: false }
      submit_required:
        - produse_promovate
        - feedback
    # ... (administrative, business_travel, company_event, cycle_meeting,
    #      team_meeting, training, public_holiday, vacation, pauza_de_masa
    #      — all non_field, with duration + detalii fields)
  routing_options:
    - { key: saptamana_1, label: "Săptămâna 1" }
    - { key: saptamana_2, label: "Săptămâna 2" }
    - { key: saptamana_3, label: "Săptămâna 3" }

options:
  specialties: [cardiologie, internal_medicine, family_and_general_medicine, ...]
  classifications: [a, b, c]
  pharmacy_types: [lant]
  products: [product_1, ...]

rules:
  frequency: { a: 4, b: 2, c: 1 }
  max_activities_per_day: 10
  default_visit_duration_minutes: { doctor: 30, pharmacy: 15 }
  visit_duration_step_minutes: 15
  recovery:
    weekend_activity_flag: true
    recovery_window_days: 5
    recovery_type: full_day
```

### 1.2 Go Implementation ✅

**Package:** `internal/config/`

- ✅ `tenant.go` — Structs: `TenantConfig`, `AccountTypeConfig`, `ActivityTypeConfig`, `FieldConfig`, `OptionDef`, `StatusDef`, `RulesConfig`
- ✅ `loader.go` — `Load(path string) (*TenantConfig, error)` — reads JSON, validates internal consistency
- ✅ `loader_test.go` — Tests for loading and validation
- ✅ `validator.go` — `ValidateActivity()`, `ValidateStatus()`, `ValidateStatusTransition()`, `ValidateDuration()`
- ✅ `validator_test.go` — Tests for field-level validation

Config is injected into services/handlers via constructor (DI). Read once at startup, exposed read-only via API.

> **Note:** Loader currently reads JSON format rather than YAML.

### 1.3 Config API Endpoint ❌

```
GET /api/v1/config
```

Returns the tenant config (sans internal-only fields). Frontend uses this to render dynamic forms, dropdown options, required field indicators, and locale labels.

> **Status:** Handler (`config_handler.go`) and route registration not yet implemented.

---

## 2. Account Domain ❌

### 2.1 Domain Model

The current `Customer` entity evolves into `Account` — gets `account_type` (doctor/pharmacy from config) and JSONB dynamic fields.

```go
// internal/domain/account.go
type Account struct {
    ID          uuid.UUID
    AccountType string            // "doctor", "pharmacy" — key from config
    Name        string
    Fields      map[string]any    // dynamic fields (specialitate, potential, tip, etc.)
    AssigneeID  uuid.UUID         // rep who owns this account (territory)
    TeamID      *uuid.UUID
    ImportedAt  *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 2.2 Dynamic Fields Strategy

Core columns (ID, type, name, dates, FKs) are real columns. Tenant-specific fields live in `fields JSONB`.

- Config defines which fields exist and their types
- API handler validates field values against config before saving
- DB stores them as JSONB — no migration needed when config changes
- Frontend renders forms dynamically from config

> **Status:** JSONB pattern is proven on `leads` table (migration 003). Config-driven validation is implemented. Needs to be applied to Account table.

### 2.3 Database Migration 006

```sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_type TEXT NOT NULL,
    name TEXT NOT NULL,
    fields JSONB NOT NULL DEFAULT '{}',
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
CREATE POLICY accounts_rep ON accounts FOR ALL
    USING (current_setting('app.user_role') IN ('manager','admin')
           OR assignee_id = current_setting('app.user_id')::uuid);
```

### 2.4 Backend Package Structure

```
internal/
├── domain/account.go             # Account entity (NEW)
├── service/account_service.go    # Account CRUD + RBAC (NEW)
├── store/account.go              # AccountRepository interface (NEW)
├── store/postgres/account.go     # PostgreSQL implementation (NEW)
├── api/account_handler.go        # HTTP handlers (NEW)
├── api/config_handler.go         # GET /config (NEW)
└── api/router.go                 # Add new routes (MODIFY)
```

### 2.5 API Routes

```
GET    /api/v1/config                    # tenant config for frontend
GET    /api/v1/accounts                  # list (filtered by type, assignee, territory)
GET    /api/v1/accounts/{id}             # detail
POST   /api/v1/accounts                  # create (admin/import)
PUT    /api/v1/accounts/{id}             # update editable fields
```

### 2.6 Account Import

For MVP, accounts come from external data (TGA exports, DrMax internal DB):

- `POST /api/v1/accounts/import` — accepts JSON array, upserts by external ID
- Admin-only, triggered by script/job
- `imported_at` timestamp tracks last import
- Fields marked `editable: false` in config cannot be changed by reps

---

## 3. Frontend: Accounts ❌

### 3.1 New Routes

| Route                          | Component        | Description                            |
| ------------------------------ | ---------------- | -------------------------------------- |
| `/accounts`                    | `AccountList`    | Filterable list of all accounts        |
| `/accounts?type=doctor`        | (same, filtered) | Doctor list                            |
| `/accounts?type=pharmacy`      | (same, filtered) | Pharmacy list                          |
| `/accounts/:id`                | `AccountDetail`  | Account detail + associated activities |

### 3.2 TypeScript Types

```typescript
// types/config.ts
interface TenantConfig {
  tenant: { name: string; locale: string }
  accounts: { types: AccountTypeConfig[] }
  activities: { statuses: StatusDef[]; status_transitions: Record<string, string[]>; durations: OptionDef[]; types: ActivityTypeConfig[]; routing_options: OptionDef[] }
  options: Record<string, OptionDef[]>
  rules: RulesConfig
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
```

### 3.3 TanStack Query Hooks

```typescript
// services/config.ts
useConfig()                  // GET /config, staleTime: Infinity

// services/accounts.ts
useAccounts(filters)         // GET /accounts
useAccount(id)               // GET /accounts/:id
useCreateAccount()           // POST /accounts
useUpdateAccount()           // PUT /accounts/:id
```
