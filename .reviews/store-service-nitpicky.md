# Store & Service Layer Review -- The Nitpicky Ninny

Reviewed: `internal/store/`, `internal/store/postgres/`, `internal/service/`, `internal/geo/`

## Fixed in This PR

### 1. Sentinel errors using `fmt.Errorf` instead of `errors.New`
**Files:** `internal/service/activity_service.go`

All 8 sentinel error variables (ErrSubmitted, ErrMaxActivities, etc.) used `fmt.Errorf("...")` with no format verbs. This is functionally equivalent to `errors.New` but slower, allocates more, and misleads readers into thinking there is formatting involved. Changed to `errors.New`.

### 2. Inconsistent error format constant naming
**Files:** `internal/service/target_service.go`

`errGettingTarget` did not match the `errFmt` prefix convention established by `errFmtGettingActivity` in `activity_service.go` and `errFmtMarshalTargetFields` in `target_repository.go`. Renamed to `errFmtGettingTarget`.

### 3. Inconsistent `whereClause()` output between query builders
**Files:** `internal/store/postgres/activity_repository.go`, `internal/store/postgres/target_repository.go`

`activityQueryBuilder.whereClause()` returned `" WHERE ..."` (leading space) while `targetQueryBuilder.whereClause()` returned `"WHERE ..."` (no leading space). This meant callers had to add or omit trailing spaces inconsistently. Standardized to leading space in both.

### 4. Helper functions scattered across repository files
**Files:** `internal/store/postgres/helpers.go`, `internal/store/postgres/audit_repository.go`, `internal/store/postgres/territory_repository.go`

`nullJSONIfNil` lived in `audit_repository.go` and `marshalJSONField` lived in `territory_repository.go`. Both are general-purpose helpers. Moved to `helpers.go` alongside `nullIfEmpty`.

### 5. Inconsistent `rows.Err()` handling
**Files:** `internal/store/postgres/dashboard_repository.go`, `internal/store/postgres/target_repository.go`, `internal/store/postgres/collection_repository.go`

Five methods returned `rows.Err()` directly without wrapping, while every other repository method wrapped with `fmt.Errorf("iterating X: %w", err)`. Fixed all five to use the consistent wrapped pattern.

### 6. Duplicate `monthsInRange` / `frequencyMonths` functions
**Files:** `internal/service/dashboard_service.go`, `internal/service/target_service.go`

Identical implementations with different names. Removed `frequencyMonths` from `target_service.go` and unified on `monthsInRange` from `dashboard_service.go`.

### 7. Stray blank lines at top of file
**Files:** `internal/service/user_service.go`

Two blank lines before `package service`. Removed.

---

## Noted But NOT Fixed (Requires Larger Refactor)

### A. Three different target scope builders doing the same thing
**Files:** `internal/store/postgres/target_repository.go`, `internal/store/postgres/dashboard_repository.go`

There are THREE implementations of target RBAC scope SQL generation:
- `targetQueryBuilder.applyScope()` -- struct method, used by `List`
- `targetScopeConditionsAliased()` / `targetScopeConditions()` -- free function, used by `CoverageStats`
- `buildTargetScopeConditions()` -- returns a struct, used by `VisitStatus`/`FrequencyStatus`

The same problem exists for activity scopes:
- `activityQueryBuilder.applyScope()` in `activity_repository.go`
- `activityScopeConditionsAliased()` in `dashboard_repository.go`

**Recommendation:** Pick ONE approach (the `buildXScopeConditions` returning a result struct is the most flexible) and use it everywhere. The query builder pattern is fine for `List` methods but the scope logic should be extracted and shared.

### B. `WHERE 1=1` vs structured query builder pattern
**Files:** `internal/store/postgres/audit_repository.go`, `internal/store/postgres/territory_repository.go`

These two repositories use the `WHERE 1=1` + string concatenation pattern, while `activity_repository.go` and `target_repository.go` use a structured query builder. Should pick one approach. The query builder is clearly better for maintainability.

### C. Duplicate `fieldActivityTypes` helper across services
**Files:** `internal/service/activity_service.go`, `internal/service/target_service.go`, `internal/service/dashboard_service.go`

Three near-identical implementations:
- `ActivityService.fieldActivityTypes()` -- method on service
- `TargetService.fieldActivityTypes()` -- method on service  
- `fieldActivityTypeKeys(cfg)` -- free function in dashboard

All iterate `cfg.Activities.Types` filtering `Category == "field"`. Should be a method on `*config.TenantConfig`.

### D. Duplicate test error message constants
**Files:** `internal/service/activity_service_test.go`, `internal/service/target_service_test.go`, `internal/service/dashboard_service_test.go`, `internal/service/collection_service_test.go`

Four different constants all with value `"unexpected error: %v"`:
- `testActErrUnexpected`
- `testErrUnexpected`
- `testDashErrUnexpected`
- `testCollErrUnexpected`

Should be ONE constant in `test_helpers_test.go`.

### E. `containsString` lives in `collection_service.go` but used by `territory_service.go`
**Files:** `internal/service/collection_service.go`, `internal/service/territory_service.go`

The `containsString` utility is defined in `collection_service.go` but imported by proximity in `territory_service.go` (same package). Should live in a dedicated `helpers.go` in the service package, or use `slices.Contains` from the standard library (Go 1.21+).

### F. `rows.Close()` in `dashboard_repository.ActivityStats` is not deferred
**File:** `internal/store/postgres/dashboard_repository.go`

The `ActivityStats` method runs two queries reusing the `rows` variable. The first query's `rows.Close()` is called inline (not deferred), which means if a `Scan` error occurs before the `rows.Close()` call, the rows won't be closed. Consider using separate variables (`statusRows`, `typeRows`) with `defer`.

### G. `strPtr` helper in store test but not shared with service tests
**File:** `internal/store/postgres/test_helpers_test.go`

The `strPtr` function is defined in postgres test helpers. Service tests use `strPtr` too (in `target_service_test.go` constants section) -- these are in different test packages so can't share. Not a bug, but worth noting the duplication.

### H. `testTime()` uses 2025 date while project is active in 2026
**File:** `internal/store/postgres/test_helpers_test.go`

`testTime()` returns `time.Date(2025, 3, 15, ...)` which is in the past relative to the project timeline. Not a bug (tests don't care about "now"), but slightly confusing.
