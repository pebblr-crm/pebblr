# Devil's Advocate Review: `internal/store/`, `internal/service/`, `internal/geo/`

Reviewed 2026-04-02. Every finding below is a genuine question about whether the
current design is earning its keep. Some are nitpicks; others are structural
concerns that will compound as the codebase grows.

---

## 1. The God-Store Aggregate

**File:** `internal/store/store.go`

```go
type Store interface {
    Users() UserRepository
    Teams() TeamRepository
    Targets() TargetRepository
    Activities() ActivityRepository
    Audit() AuditRepository
    Dashboard() DashboardRepository
    Collections() CollectionRepository
    Territories() TerritoryRepository
}
```

**Challenge:** Does this interface add value or does it just shuffle the problem?

Services already declare exactly which repositories they need in their
constructors (e.g. `NewActivityService(activities, targets, users, audit, ...)`).
The `Store` aggregate is never *consumed* by a service -- it is only used in
`store_impl.go` to wire things up, and every new repository means touching this
interface, the implementation, *and* the constructor.

This is a "header file" that provides no abstraction -- it merely groups things
that happen to live in the same database. If you add a second storage backend
(Redis cache, event bus), this interface actively misleads because not every
"store" method would live in Postgres.

**Recommendation:** Consider dropping `Store` entirely. Let `cmd/api/main.go`
(or a wire function) call `postgres.New(pool)` and pass individual repositories
to service constructors directly. You already do the granular injection in
tests -- production should match.

---

## 2. RBAC Scoping Is Split Between Two Worlds

**Where RBAC lives today:**

| Layer | What it does |
|-------|-------------|
| `rbac.Enforcer` (service layer) | Item-level checks (`CanViewTarget`, `CanUpdateActivity`) |
| `rbac.ActivityScope` / `rbac.TargetScope` (store layer) | List-query scoping pushed into SQL WHERE clauses |
| `store/postgres/*.go` | Translates scope structs into SQL conditions |

**Challenge:** The store layer has intimate knowledge of RBAC policy.

`activityScopeConditionsAliased` and `targetScopeConditionsAliased` in
`dashboard_repository.go` know the *meaning* of `CreatorIDs`, `TeamIDs`, and
`joint_visit_user_id`. That is policy, not plumbing. If policy changes (e.g.
"managers can see cross-team joint visits"), you have to update both the
`policyEnforcer` *and* every SQL scope builder.

Worse, there are **three separate implementations** of scope-to-SQL:
1. `activityQueryBuilder.applyScope` in `activity_repository.go`
2. `activityScopeConditionsAliased` in `dashboard_repository.go`
3. `buildTargetScopeConditions` in `target_repository.go`

Plus the query builder structs (`activityQueryBuilder`, `targetQueryBuilder`)
that do essentially the same thing with slightly different field names.

**Recommendation:**
- Extract a single `scopeToSQL(scope, alias, args, argIdx)` helper per entity
  type. There should be exactly one function that translates `ActivityScope` to
  SQL and exactly one for `TargetScope`.
- Move these helpers into `store/postgres/scope.go` so the policy-to-SQL
  translation is consolidated.

---

## 3. Services That Are Just Pass-Throughs

**`UserService`** (`user_service.go`):

```go
func (s *UserService) List(ctx context.Context) ([]*domain.User, error) {
    users, err := s.users.List(ctx)
    if err != nil {
        return nil, fmt.Errorf("listing users: %w", err)
    }
    return users, nil
}
```

This service wraps two repository methods with nothing but error re-wrapping.
No RBAC. No validation. No business logic. It exists solely to satisfy a
layering convention.

**`TeamService`** (`team_service.go`): Same story. Three methods, all
pass-through. No RBAC checks despite the fact that `UserRepository.List` is
documented as "Intended for admin/manager views" -- but the service does not
enforce that.

**Challenge:** If a service adds no behavior, it is ceremony, not
architecture. Worse, it creates a *false sense of security* -- a reader
assumes RBAC is enforced at the service layer because every other service does
it, but `UserService` and `TeamService` silently skip it.

**Recommendation:** Either:
- Add RBAC checks (reps should NOT see all users -- this is a data leak), or
- Delete these services and let handlers call the repository directly, with an
  explicit comment that these are admin-only endpoints guarded by middleware.

---

## 4. TerritoryService and CollectionService Roll Their Own RBAC

**Files:** `territory_service.go`, `collection_service.go`

Both services implement `canView`, `canModify`, and `scopeFilter` with inline
`switch actor.Role` logic -- duplicating the pattern that `rbac.Enforcer`
exists to centralize.

```go
func (s *TerritoryService) canView(actor *domain.User, t *domain.Territory) bool {
    switch actor.Role {
    case domain.RoleAdmin:
        return true
    default:
        return containsString(actor.TeamIDs, t.TeamID)
    }
}
```

This is a parallel RBAC system. If the role model changes (new role, new
visibility rule), you now have to update `rbac/policy.go` *and* two service
files.

**`containsString`** is also duplicated in `collection_service.go` AND
`rbac/policy.go`. Three copies of a helper that should exist once.

**Recommendation:**
- Add `CanViewTerritory`, `CanModifyTerritory`, `CanViewCollection`,
  `CanModifyCollection` to `rbac.Enforcer`.
- Add `ScopeTerritoryQuery` and `ScopeCollectionQuery` if list scoping is
  needed.
- Delete the private RBAC methods from the services.
- Move `containsString` to a shared `internal/sliceutil` package or use
  `slices.Contains` from the stdlib (Go 1.21+).

---

## 5. Dashboard Repository Queries Are Doing Too Much

**File:** `store/postgres/dashboard_repository.go` -- 434 lines.

`ActivityStats` fires **two separate queries** (by status, by type) that scan
the same rows. The database is doing the same WHERE + filter twice.

`CoverageStats` builds two entirely independent query pipelines (one for
target count, one for visited count) with their own scope builders, arg
arrays, and arg indices. That is ~80 lines of careful positional-arg
bookkeeping that is a bug magnet.

`FrequencyStats` joins targets and activities with scope conditions on both
aliases, managing a single shared `args` slice and `argIdx` counter across
both. One off-by-one in `argIdx` and you get a silent wrong-result bug.

**Challenge:** The positional-arg pattern (`$1`, `$2`, ...) with manual
`argIdx` tracking across multiple scope builders is the most fragile code in
the repository. Every new filter or scope condition risks an off-by-one.

**Recommendation:**
- `ActivityStats`: Combine into a single query with conditional aggregation:
  `SELECT status, activity_type, COUNT(*)... GROUP BY GROUPING SETS ((status), (activity_type))`.
  One query, one scan, half the code.
- Extract a `queryBuilder` type that encapsulates `args []any` and
  `argIdx int` behind `Add(val) string` (returns `$N`). This eliminates the
  manual `argIdx` tracking that is duplicated ~15 times across dashboard and
  activity repositories.

---

## 6. Audit Recording Silently Swallows Errors

**Files:** `activity_service.go`, `target_service.go`

```go
_ = s.audit.Record(ctx, &domain.AuditEntry{...})
```

Every audit call discards the error. Audit is described as a core invariant
("lead_events table captures all lead lifecycle events for audit and
telemetry"), yet failures are invisible.

**Challenge:** If the audit table is full, the connection drops, or the schema
drifts, you will lose audit records with zero signal. For a CRM where audit
trails may have compliance implications (pharmaceutical field sales for
DrMax), this is risky.

**Recommendation:** At minimum, log the error with `slog.Error`. Better: make
audit recording transactional with the primary write (wrap both in a single
`tx.Commit`) so either both succeed or both fail.

---

## 7. `autoCompleteNonFieldActivities` Has Side Effects in a Read Path

**File:** `activity_service.go`, line 204

```go
func (s *ActivityService) List(...) (*store.ActivityPage, error) {
    ...
    s.autoCompleteNonFieldActivities(ctx, result.Activities)
    return result, nil
}
```

A **list** operation triggers **writes** (status updates) as a side effect.
This means:

- Read-only consumers (dashboards, exports) can trigger mutations.
- Pagination is unreliable -- activities can change status between page
  fetches.
- Two concurrent List calls can race on the same activity.
- The method uses `time.Now()` internally, making it non-deterministic and
  untestable without time injection.

**Recommendation:** Extract auto-completion into a separate background job or
explicit `Reconcile()` method called at defined points (e.g., login, daily
cron), not hidden inside a read path.

---

## 8. The `geo` Package Is Fine, But the API Key Is Suspect

**File:** `internal/geo/google.go`

```go
func NewGoogleGeocoder(apiKey string) *GoogleGeocoder {
    return &GoogleGeocoder{
        apiKey:     apiKey,
        ...
    }
}
```

Where does `apiKey` come from? The constructor takes a plain string. The
project convention is "secrets from mounted files, never env vars." If the
caller reads this from an env var, the geo package has no way to enforce the
convention.

**Challenge:** The geo package is clean, but it is an invitation for the
caller to do `NewGoogleGeocoder(os.Getenv("GOOGLE_API_KEY"))`.

**Recommendation:** Either:
- Accept a file path and read the secret internally (matching `postgres.Connect`
  which takes `dsnFile`), or
- Document the expectation clearly and add a `geo.NewGoogleGeocoderFromFile(path)`
  constructor.

---

## 9. `TargetService.Import` Geocodes Synchronously in a Write Path

**File:** `target_service.go`, line 180

```go
if s.geocoder != nil {
    s.geocodeTargets(ctx, targets)
}
result, err := s.targets.Upsert(ctx, targets)
```

Geocoding makes external HTTP calls to Google Maps for *every un-geocoded
target* in a single request handler. A 500-target import could make 500
serial HTTP requests before the upsert even starts.

**Challenge:** This blocks the HTTP request for the duration of all geocoding
calls. Google Geocoding API has rate limits (50 QPS default). A large import
will either timeout or get rate-limited.

**Recommendation:** Geocode asynchronously:
1. Upsert targets immediately (without coordinates).
2. Queue un-geocoded targets for background processing.
3. Update coordinates as geocoding completes.

This also avoids the silent swallowing of geocoding errors -- the current code
logs and skips, which means the caller has no way to know which targets failed
geocoding.

---

## 10. Duplicate `monthsInRange` / `frequencyMonths`

**Files:** `dashboard_service.go` line 267, `target_service.go` line 295

Two identical functions:

```go
func monthsInRange(from, to time.Time) int { ... }
func frequencyMonths(from, to time.Time) int { ... }
```

Same logic, different names. One is used by `DashboardService`, the other by
`TargetService`.

**Recommendation:** Keep one, delete the other. Put it in a shared
`internal/timeutil` package or keep it unexported in `service/` but use it
from both callers.

---

## 11. Duplicate `fieldActivityTypes` / `fieldActivityTypeKeys`

**Files:** `activity_service.go` line 729, `target_service.go` line 307,
`dashboard_service.go` line 236

Three methods that extract field-category activity type keys from config:

- `ActivityService.fieldActivityTypes()`
- `TargetService.fieldActivityTypes()`
- `fieldActivityTypeKeys(cfg)` (package-level in dashboard_service.go)

All do the same loop over `cfg.Activities.Types` filtering by
`Category == "field"`.

**Recommendation:** Add a `FieldActivityTypeKeys()` method to
`config.TenantConfig` and call it from all three places.

---

## 12. `TeamRepository` Is a Stub That Lies

**File:** `store/postgres/team_repository.go`

```go
func (r *teamRepository) Get(_ context.Context, _ string) (*domain.Team, error) {
    return nil, store.ErrNotFound
}

func (r *teamRepository) List(_ context.Context) ([]*domain.Team, error) {
    return []*domain.Team{}, nil
}
```

`Get` returns `ErrNotFound` for *every* request. `List` returns an empty
slice. `Create`, `Update`, `Delete` return `errNotImplemented`. But
`errNotImplemented` is unexported and never checked by callers, so they will
get an opaque error.

**Challenge:** This is a stub that pretends to work for reads (returning empty
results instead of errors) while failing for writes. Any code that calls
`Teams().Get(id)` will silently get "not found" even if the team exists in
the database.

**Recommendation:** Either implement it or make it panic/fail explicitly with
a clear message. The current behavior will cause subtle bugs when team-related
features are built on top.

---

## 13. `AuditFilter` Embeds Pagination, Other Filters Don't

**File:** `store/audit_store.go`

```go
type AuditFilter struct {
    EntityType *string
    ActorID    *string
    Status     *string
    Page       int
    Limit      int
}
```

`ActivityFilter` and `TargetFilter` take `page, limit` as separate function
parameters. `AuditFilter` embeds them in the struct. `TerritoryFilter` and
`CollectionFilter` have no pagination at all.

This inconsistency means every handler has to know which pattern each filter
uses. It is a papercut that will confuse every new contributor.

**Recommendation:** Pick one pattern and apply it everywhere. The
`(filter, page, limit int)` signature is cleaner because it separates
"what to find" from "how many to return."

---

## 14. Missing RBAC on `ActivityRepository.Get`

**File:** `store/postgres/activity_repository.go`, line 74

```go
func (r *activityRepository) Get(ctx context.Context, id string) (*domain.Activity, error) {
    row := r.pool.QueryRow(ctx,
        `SELECT `+activityColumns+activityFrom+` WHERE a.id = $1::UUID AND a.deleted_at IS NULL`,
        id,
    )
    return scanActivity(row)
}
```

This returns *any* activity by ID with no scope restriction. The service
layer calls `enforcer.CanViewActivity` after the fetch -- but this means the
database serves the row first, and the app decides afterward.

**Challenge:** This is *by design* (fetch-then-check), and the CLAUDE.md says
"RBAC enforcement in the data layer + PostgreSQL RLS as defense-in-depth." But
there is no RLS policy visible in the codebase. The "defense-in-depth" layer
does not exist yet.

**Recommendation:** If RLS is planned, track it as a TODO. If it is not
planned, document why fetch-then-check is acceptable for single-record access
(it is, for correctness -- the service layer is the authority). But do NOT
claim defense-in-depth if the second layer does not exist.

---

## 15. `WHERE 1=1` Anti-Pattern

**Files:** `audit_repository.go` line 60, `territory_repository.go` line 42

```go
query := `SELECT ... FROM audit_log WHERE 1=1`
```

This is the "I don't want to track whether I need AND or WHERE" shortcut. It
works, but it is a code smell that signals the query builder is underpowered.

The `activityQueryBuilder` and `targetQueryBuilder` solve this properly with
`whereClause()` that handles the WHERE/AND logic. The audit and territory
repos did not get the same treatment.

**Recommendation:** Apply the query builder pattern consistently, or at least
use the same `whereClause()` helper everywhere.

---

## Summary of Priorities

| Priority | Finding | Effort |
|----------|---------|--------|
| HIGH | #3 UserService/TeamService missing RBAC (data leak) | Low |
| HIGH | #6 Audit errors silently swallowed | Low |
| HIGH | #12 TeamRepository stub returns wrong results | Medium |
| MEDIUM | #2 Triplicated scope-to-SQL logic | Medium |
| MEDIUM | #4 Parallel RBAC in Territory/Collection services | Medium |
| MEDIUM | #7 Write side effects in List read path | Medium |
| MEDIUM | #14 No RLS despite claiming defense-in-depth | Medium |
| LOW | #1 Store aggregate interface adds no value | Low |
| LOW | #5 Dashboard queries doing double work | Medium |
| LOW | #8 Geo API key source ambiguity | Low |
| LOW | #9 Synchronous geocoding in import | Medium |
| LOW | #10-11 Duplicated helpers | Low |
| LOW | #13 Inconsistent pagination patterns | Low |
| LOW | #15 WHERE 1=1 inconsistency | Low |
