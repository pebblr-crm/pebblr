# Clean Code Review: `internal/store/`, `internal/service/`, `internal/geo/`

**Reviewer persona:** The Clean Code Geek (SOLID, SRP, DRY, meaningful naming, small functions, clean abstractions)

**Date:** 2026-04-02

---

## Overall Assessment

The codebase is **remarkably well-structured** for an early-stage project. Interface segregation is clean, dependency injection is consistent, error wrapping is disciplined, and the repository/service separation is textbook. That said, several DRY violations and SRP stretches have crept in as the codebase grew. The findings below are ordered by severity.

---

## 1. DRY Violation: Duplicated `fieldActivityTypes()` (HIGH)

The exact same function -- "collect keys of field-category activity types from config" -- is copy-pasted in **three** places:

| Location | Function name |
|---|---|
| `internal/service/activity_service.go:729` | `(s *ActivityService) fieldActivityTypes()` |
| `internal/service/target_service.go:307` | `(s *TargetService) fieldActivityTypes()` |
| `internal/service/dashboard_service.go:236` | `fieldActivityTypeKeys(cfg)` (package-level) |

All three iterate `cfg.Activities.Types`, filter by `Category == "field"`, and collect keys.

**Recommendation:** Move to a single method on `*config.TenantConfig`:

```go
// In internal/config/tenant.go (or wherever TenantConfig lives):
func (c *TenantConfig) FieldActivityTypeKeys() []string {
    var keys []string
    for i := range c.Activities.Types {
        if c.Activities.Types[i].Category == "field" {
            keys = append(keys, c.Activities.Types[i].Key)
        }
    }
    return keys
}
```

Then all three call sites become `s.cfg.FieldActivityTypeKeys()`. Similarly, `blockingActivityTypes()` in `activity_service.go` could be `c.BlockingActivityTypeKeys()`.

---

## 2. DRY Violation: Duplicated `monthsInRange` / `frequencyMonths` (HIGH)

Two identical functions exist:

| Location | Name |
|---|---|
| `internal/service/dashboard_service.go:267` | `monthsInRange(from, to)` |
| `internal/service/target_service.go:295` | `frequencyMonths(from, to)` |

They are **character-for-character identical** in logic. One name is better than the other (`monthsInRange` is more descriptive).

**Recommendation:** Keep one, delete the other. Since both are package-private in the same package, a single `monthsInRange` suffices.

---

## 3. DRY Violation: Duplicated RBAC Scope Builder Logic (MEDIUM)

The `store/postgres/` package has **three** parallel mechanisms for building RBAC scope SQL:

1. **`activityQueryBuilder.applyScope()`** -- method on the builder struct (used in `activity_repository.go`)
2. **`activityScopeConditionsAliased()`** -- standalone function (used in `dashboard_repository.go`)
3. **`targetQueryBuilder.applyScope()`** -- method on the builder struct (used in `target_repository.go`)
4. **`targetScopeConditionsAliased()`** -- standalone function (used in `dashboard_repository.go`)
5. **`buildTargetScopeConditions()`** -- yet another standalone function with a result struct (used in `target_repository.go` for `VisitStatus`/`FrequencyStatus`)

That is **five** implementations of fundamentally two operations (scope activities, scope targets). The `activityQueryBuilder.applyScope` and `activityScopeConditionsAliased` do the same thing with different signatures. Same for the target variants.

**Recommendation:** Converge on **one** builder abstraction. A `scopeBuilder` struct with `addActivityScope()` and `addTargetScope()` methods would unify all five. The query builders already exist; extend them to be the single source of truth.

---

## 4. DRY Violation: Territory Boundary Scanning (MEDIUM)

`territory_repository.go` repeats the boundary scan-and-unmarshal pattern **four times** (Get, List loop, Create, Update). Each one does:

```go
var boundaryJSON []byte
// ... Scan(..., &boundaryJSON, ...)
if len(boundaryJSON) > 0 {
    t.Boundary = make(map[string]any)
    if err := json.Unmarshal(boundaryJSON, &t.Boundary); err != nil {
        return nil, fmt.Errorf("unmarshalling territory boundary: %w", err)
    }
}
```

**Recommendation:** Extract a `scanTerritory(row pgx.Row) (*domain.Territory, error)` function (like `scanTarget` and `scanActivity` already do for their entities). The territory repository is the only one that lacks this pattern.

---

## 5. SRP Concern: `ActivityService` is Overloaded (MEDIUM)

`ActivityService` at 857 lines handles:
- CRUD (Create, Get, List, Update, Delete)
- Submit workflow
- PartialUpdate (PATCH)
- PatchStatus (status transitions)
- CloneWeek (batch duplication with dedup)
- Recovery balance validation
- Business day calculation helpers
- Auto-completion of past non-field activities

That is at least **four** distinct responsibilities. The `CloneWeek` feature alone is ~80 lines of its own business logic with `validateCloneWeekInputs`, `buildExistingTargetIndex`, and `cloneActivities`.

**Recommendation:** Extract `CloneWeek` and its helpers into a dedicated `ActivityCloneService` or at minimum a separate file `activity_clone.go`. Similarly, the recovery-balance validation (`checkRecoveryBalance`, `isRecoveryActivity`, `hasUnclaimedWindow`, `isWindowClaimed`) could live in `recovery_validation.go` within the same package.

---

## 6. SRP Concern: `DashboardService` Mixes Aggregation with Business Day Math (LOW)

`dashboard_service.go` contains `nextBusinessDay()` and `addBusinessDays()` which are general-purpose date utilities. They are also called from `activity_service.go` (via `hasUnclaimedWindow`), meaning these are **shared utilities living in a service file**.

**Recommendation:** Move `nextBusinessDay` and `addBusinessDays` to a shared package-level file like `internal/service/dateutil.go` or even `internal/domain/dates.go`.

---

## 7. Naming: `containsString` Misplaced and Generic (LOW)

`containsString` in `collection_service.go:146` is called from both `collection_service.go` and `territory_service.go`. It is a general utility, not specific to collections.

**Recommendation:** Move to a `internal/service/helpers.go` or similar. With Go 1.21+, consider `slices.Contains` from the standard library, which would eliminate this function entirely.

---

## 8. Naming: Inconsistent Error Format Constants (LOW)

Error format constants use two different naming conventions:

| File | Name | Convention |
|---|---|---|
| `activity_service.go` | `errFmtGettingActivity` | camelCase with `Fmt` |
| `target_service.go` | `errGettingTarget` | camelCase, no `Fmt` |
| `target_repository.go` | `errFmtMarshalTargetFields` | camelCase with `Fmt` |

Pick one convention (`errFmt*` or `errMsg*` or just `err*`) and apply it consistently.

---

## 9. Naming: `dbPool` Interface Could Be More Descriptive (LOW)

`postgres.go` defines `dbPool` as the mockable subset of `*pgxpool.Pool`. The name `dbPool` does not convey that it is an interface. Go convention for interfaces is typically the method name + `-er` suffix, or a noun describing the capability. Consider `querier` or `dbQuerier`, which is also what the `pgx` ecosystem commonly uses.

---

## 10. Audit Repository: `WHERE 1=1` Pattern (LOW)

`audit_repository.go:60` and `territory_repository.go:42` use the `WHERE 1=1` trick for conditional clause building. The other repositories use the query builder pattern. The `WHERE 1=1` approach works but is inconsistent with the rest of the codebase.

**Recommendation:** Either adopt the query builder pattern everywhere (activity/target repos already do this) or at least be consistent. The query builder is cleaner and avoids the `_ = argIdx` dead-assignment that appears in `territory_repository.go:56` and `audit_repository.go:106`.

---

## 11. `store_impl.go`: Repository Allocation on Every Call (INFO)

Each accessor method (`Users()`, `Teams()`, etc.) allocates a new repository struct on every call:

```go
func (db *DB) Users() store.UserRepository {
    return &userRepository{pool: db.pool}
}
```

This is fine for correctness (the structs are tiny, stateless), but if these methods are called in hot paths (e.g., per-request), you could cache the repository instances on `DB` construction. This is a micro-optimization and not urgent -- just noting it for when performance tuning becomes relevant.

---

## 12. `GoogleGeocoder`: API Key in Memory (INFO)

`geo/google.go` stores the API key in a plain `string` field. The project convention is file-mounted secrets. If `NewGoogleGeocoder` receives the key as a string read from a file at startup, this is fine -- just confirm the caller reads from a file mount, not from `os.Getenv`.

---

## 13. Test Helpers: Well Done (POSITIVE)

- `test_helpers_test.go` in both `postgres/` and `service/` centralizes fixtures, mock pools, and user factories
- `anyArgs(n)` helper simplifies pgxmock expectations
- Stub repos in service tests are lean and focused
- `t.Parallel()` used consistently

This is clean test infrastructure. No notes.

---

## 14. Interface Segregation: Well Done (POSITIVE)

- `store.Store` aggregates per-entity repository interfaces
- Each repository interface is in its own file with its own filter/page types
- Services depend on specific repository interfaces, not on `store.Store`
- `geo.Geocoder` is a minimal single-method interface

This is excellent ISP adherence.

---

## Summary of Actionable Items

| # | Severity | Finding | Effort |
|---|---|---|---|
| 1 | HIGH | `fieldActivityTypes()` duplicated 3x | Small |
| 2 | HIGH | `monthsInRange`/`frequencyMonths` duplicated | Trivial |
| 3 | MEDIUM | Five RBAC scope builder variants | Medium |
| 4 | MEDIUM | Territory boundary scan repeated 4x | Small |
| 5 | MEDIUM | `ActivityService` SRP -- 857 lines, 4+ responsibilities | Medium |
| 6 | LOW | Business day helpers in wrong file | Trivial |
| 7 | LOW | `containsString` misplaced, use `slices.Contains` | Trivial |
| 8 | LOW | Inconsistent error format constant naming | Trivial |
| 9 | LOW | `dbPool` interface naming | Trivial |
| 10 | LOW | `WHERE 1=1` inconsistent with query builder pattern | Small |
| 11 | INFO | Repository allocation per call | N/A |
| 12 | INFO | Geocoder API key sourcing | N/A |
