# Idiomatic Obsessive Review: `internal/store/`, `internal/service/`, `internal/geo/`

Reviewer persona: **The Idiomatic Obsessive** -- pgx/v5 patterns, proper scanning,
context propagation, error wrapping, interface placement, repository pattern the Go way.

---

## Verdict

The codebase is well-structured. Interfaces live in the right package (`store/`),
implementations live in `store/postgres/`, services depend on interfaces not
implementations, and context flows through every call. That said, there are concrete
idiom violations and missed pgx/v5 opportunities that should be addressed before this
hardens further.

---

## 1. pgx/v5: Use `pgx.CollectRows` / `pgx.CollectOneRow` instead of manual scan loops

**Severity: Medium -- Missed idiomatic pgx/v5 API**

Every repository manually iterates `rows.Next()` + scan + `rows.Err()`. pgx/v5
introduced `pgx.CollectRows` and `pgx.RowToStructByPos` / `pgx.RowToAddrOfStructByPos`
specifically to eliminate this boilerplate. The project is already on pgx/v5 -- use it.

**Current pattern (repeated ~10 times):**
```go
rows, err := r.pool.Query(ctx, query, args...)
if err != nil {
    return nil, fmt.Errorf("querying X: %w", err)
}
defer rows.Close()

var results []*T
for rows.Next() {
    var item T
    if err := rows.Scan(...); err != nil {
        return nil, fmt.Errorf("scanning X: %w", err)
    }
    results = append(results, &item)
}
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("iterating X: %w", err)
}
```

**Idiomatic replacement:**
```go
rows, err := r.pool.Query(ctx, query, args...)
if err != nil {
    return nil, fmt.Errorf("querying X: %w", err)
}
results, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*T, error) {
    // scan logic here
})
if err != nil {
    return nil, fmt.Errorf("scanning X: %w", err)
}
```

This is safe against forgetting `rows.Close()` (CollectRows does it), safe against
forgetting `rows.Err()` (CollectRows checks it), and 40% less code.

Note: This requires adding `CollectRows` to the `dbPool` interface or changing the
scan functions to accept `pgx.CollectableRow`. The mock interface (`pgxmock`) supports
this as of v4.

**Affected files:**
- `postgres/user_repository.go` (List)
- `postgres/target_repository.go` (List, VisitStatus, FrequencyStatus)
- `postgres/activity_repository.go` (List)
- `postgres/audit_repository.go` (ListByEntity, List, scanAuditEntries)
- `postgres/dashboard_repository.go` (ActivityStats -- 2 loops, FrequencyStats,
  WeekendFieldActivities, RecoveryActivities)
- `postgres/territory_repository.go` (List)
- `postgres/collection_repository.go` (List)

---

## 2. `rows.Err()` missing after scan loop

**Severity: High -- Silent data loss**

Two places return `result, rows.Err()` without wrapping the error:

- `target_repository.go:355` -- `VisitStatus` returns `result, rows.Err()` bare.
- `target_repository.go:416` -- `FrequencyStatus` returns `result, rows.Err()` bare.
- `dashboard_repository.go:287` -- `WeekendFieldActivities` returns bare.
- `dashboard_repository.go:326` -- `RecoveryActivities` returns bare.
- `collection_repository.go:94` -- `List` returns bare.

Bare `rows.Err()` loses the operation context. Every other scan loop in the codebase
wraps it (good), but these five do not. They should be:
```go
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("iterating X: %w", err)
}
return result, nil
```

---

## 3. `scanUser` reuse is inconsistent -- `List` duplicates scan logic

**Severity: Low -- DRY violation**

`user_repository.go` defines `scanUser(row pgx.Row)` that scans 7 columns. But `List`
scans 8 columns (adds `TeamIDs`) with inline scan logic. This means the column list is
duplicated and will drift.

**Fix:** Either extend `scanUser` to accept an optional `TeamIDs` destination, or use a
`scanUserRow` function for the `List` variant. The target/activity repos already use
`scanTarget`/`scanActivity` for both single-row and multi-row cases via
`pgx.Row` (which `pgx.Rows` satisfies). Follow the same pattern here.

---

## 4. `scanTarget` and `scanTargetWithFlag` violate DRY

**Severity: Low**

`scanTarget` and `scanTargetWithFlag` are nearly identical -- 30 lines each, differing
only by an extra `flag` column. Extract common field scanning into a shared helper:

```go
func scanTargetCore(row pgx.Row, extras ...any) (*domain.Target, error) {
    var t domain.Target
    var fieldsJSON []byte
    dests := []any{
        &t.ID, &t.ExternalID, &t.TargetType, &t.Name, &fieldsJSON,
        &t.AssigneeID, &t.TeamID,
        &t.ImportedAt, &t.CreatedAt, &t.UpdatedAt,
    }
    dests = append(dests, extras...)
    if err := row.Scan(dests...); err != nil {
        // ... handle ErrNoRows only when extras is empty
    }
    // unmarshal fieldsJSON
    return &t, nil
}
```

---

## 5. `dbPool` interface is too narrow -- consider `pgxpool.Pool` method set

**Severity: Medium -- Blocks idiomatic pgx patterns**

The `dbPool` interface in `postgres.go` only exposes `Query`, `QueryRow`, `Exec`, and
`Begin`. This is fine for basic operations, but it prevents using:

- `pgx.CollectRows` (needs `Query` return, which it already has -- OK)
- `pool.SendBatch` for batched inserts (used idiomatically in `replaceItems`)
- `pool.CopyFrom` for bulk loads (future import optimization)

For now, the interface is adequate, but consider adding `SendBatch` when
`replaceItems` is refactored (see finding #8).

---

## 6. `AuditFilter` embeds pagination in the filter struct -- breaks convention

**Severity: Medium -- Inconsistency**

Every other repository separates pagination from filters:
```go
List(ctx, scope, filter, page, limit int)
```

But `AuditRepository.List` buries `Page` and `Limit` inside `AuditFilter`:
```go
type AuditFilter struct {
    EntityType *string
    ActorID    *string
    Status     *string
    Page       int
    Limit      int
}
```

This breaks the clean separation used everywhere else. Extract `Page` and `Limit` into
separate parameters to match the convention:

```go
List(ctx context.Context, filter AuditFilter, page, limit int) ([]*domain.AuditEntry, int, error)
```

---

## 7. `_ = argIdx` silencing in audit/territory repos

**Severity: Low -- Code smell**

In `audit_repository.go:106` and `territory_repository.go:56`:
```go
_ = argIdx
```

This silences the "unused variable" warning after the last filter is applied. It signals
that the query builder is manually tracking arg indices -- further evidence that a
shared query-builder should be extracted (see finding #10). In the meantime, this is
acceptable but warrants a `// last use` comment like the dashboard repo does.

---

## 8. `replaceItems` should use `pgx.CopyFrom` or `SendBatch`

**Severity: Medium -- Performance / idiomatic pgx**

`collection_repository.go:replaceItems` builds a dynamic multi-value INSERT:
```go
vals[i] = fmt.Sprintf("($1::UUID, $%d::UUID)", i+2)
```

For bulk inserts, pgx/v5 provides `pgx.CopyFrom` which is dramatically faster and
avoids the `Sprintf` parameter-index juggling. For transactional inserts (which this
is), use `tx.CopyFrom`:

```go
_, err := tx.CopyFrom(ctx,
    pgx.Identifier{"target_collection_items"},
    []string{"collection_id", "target_id"},
    pgx.CopyFromSlice(len(targetIDs), func(i int) ([]any, error) {
        return []any{collectionID, targetIDs[i]}, nil
    }),
)
```

---

## 9. Error wrapping in `scanAuditEntries` -- incorrect `pgx.ErrNoRows` check

**Severity: Medium -- Incorrect error handling**

`audit_repository.go:145`:
```go
if err := rows.Err(); err != nil && !errors.Is(err, pgx.ErrNoRows) {
```

`rows.Err()` never returns `pgx.ErrNoRows`. That error only comes from `row.Scan()`
on a `QueryRow` call. The `!errors.Is(err, pgx.ErrNoRows)` guard is dead code and
misleading -- it suggests the author was unsure about the pgx contract. Remove it:

```go
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("iterating audit log: %w", err)
}
```

---

## 10. Duplicated query-builder pattern -- extract a shared `queryBuilder`

**Severity: Medium -- DRY / maintenance risk**

Three separate query builder types exist:
- `targetQueryBuilder` in `target_repository.go`
- `activityQueryBuilder` in `activity_repository.go`
- Inline arg/condition tracking in `audit_repository.go`, `dashboard_repository.go`,
  `territory_repository.go`, `collection_repository.go`

They all do the same thing: accumulate `conditions []string`, `args []any`, `argIdx int`.
Extract a single shared type:

```go
// queryBuilder accumulates SQL WHERE conditions with positional pgx arguments.
type queryBuilder struct {
    conditions []string
    args       []any
    argIdx     int
}

func newQueryBuilder() *queryBuilder {
    return &queryBuilder{argIdx: 1}
}

func (b *queryBuilder) add(sqlFmt string, val any) {
    b.conditions = append(b.conditions, fmt.Sprintf(sqlFmt, b.argIdx))
    b.args = append(b.args, val)
    b.argIdx++
}

func (b *queryBuilder) where() string {
    if len(b.conditions) == 0 {
        return ""
    }
    return "WHERE " + strings.Join(b.conditions, " AND ")
}
```

This eliminates ~100 lines of duplicated logic and makes scope-condition helpers
composable.

---

## 11. `dashboard_repository.go` -- `rows.Close()` called before `rows.Err()`

**Severity: High -- Silently swallows errors**

In `ActivityStats`, the status-query loop:
```go
rows.Close()
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("iterating status rows: %w", err)
}
```

This is correct -- `rows.Err()` is valid after `Close()`. However, the error-path
inside the loop calls `rows.Close()` then returns immediately without checking
`rows.Err()`:

```go
if err := rows.Scan(&status, &count); err != nil {
    rows.Close()
    return nil, fmt.Errorf("scanning status row: %w", err)
}
```

This is technically fine (the scan error is returned), but inconsistent with `defer
rows.Close()` used everywhere else. Use `defer` and let it clean up:

```go
rows, err := r.pool.Query(ctx, statusQuery, args...)
if err != nil { ... }
defer rows.Close()
```

The manual `rows.Close()` call exists because the same `rows` variable is reused for
the type query. Fix by using separate variable names (`statusRows`, `typeRows`) with
separate defers.

---

## 12. `containsString` in `collection_service.go` -- use `slices.Contains`

**Severity: Low -- stdlib available since Go 1.21**

```go
func containsString(slice []string, s string) bool {
    for _, v := range slice {
        if v == s { return true }
    }
    return false
}
```

Replace with `slices.Contains(slice, s)` from the standard library. This utility is
also used in `territory_service.go` and could be removed entirely.

---

## 13. `GoogleGeocoder` stores API key in struct field -- verify sourcing

**Severity: Info -- Secrets audit**

`geo/google.go`:
```go
type GoogleGeocoder struct {
    apiKey string
    ...
}

func NewGoogleGeocoder(apiKey string) *GoogleGeocoder {
```

The key is passed as a string. Per project policy, secrets must come from mounted files,
not environment variables. Verify that the caller reads from a file mount. The geocoder
itself is fine -- it's the constructor callsite that matters.

---

## 14. `GoogleGeocoder.httpClient` defaults to `http.DefaultClient`

**Severity: Medium -- No timeout, shared global**

```go
httpClient: http.DefaultClient,
```

`http.DefaultClient` has no timeout. A hung geocoding request blocks forever. Inject a
client with a timeout, or at minimum set one:

```go
httpClient: &http.Client{Timeout: 10 * time.Second},
```

Better: accept `*http.Client` as a constructor parameter for testability, with a
sensible default.

---

## 15. Service layer: audit record errors silently discarded

**Severity: Medium -- Observability gap**

Throughout `activity_service.go`:
```go
_ = s.audit.Record(ctx, &domain.AuditEntry{...})
```

Audit failures are silently discarded. This is a deliberate fire-and-forget design
(audit should not block the happy path), but it should at minimum be logged:

```go
if err := s.audit.Record(ctx, &domain.AuditEntry{...}); err != nil {
    slog.Warn("audit record failed", "entity", id, "err", err)
}
```

The `autoCompleteNonFieldActivities` method already logs failures with `slog.Warn` --
follow the same pattern for audit.

---

## 16. Scope condition builders are duplicated between repos

**Severity: Medium -- DRY**

RBAC scope SQL generation appears in three forms:
1. `targetQueryBuilder.applyScope` in `target_repository.go`
2. `buildTargetScopeConditions` in `target_repository.go`
3. `targetScopeConditions` / `targetScopeConditionsAliased` in `dashboard_repository.go`

These do the same thing with slightly different signatures. Consolidate into a single
set of scope-condition builders in a shared file (e.g., `scope_conditions.go`):

```go
func appendTargetScope(b *queryBuilder, scope rbac.TargetScope, prefix string) bool
func appendActivityScope(b *queryBuilder, scope rbac.ActivityScope, prefix string) bool
```

---

## 17. `store.Store` accessor methods allocate on every call

**Severity: Low -- Allocation pressure**

`store_impl.go`:
```go
func (db *DB) Users() store.UserRepository {
    return &userRepository{pool: db.pool}
}
```

Every call to `db.Users()` allocates a new `userRepository`. These are stateless (just
hold a pool pointer), so cache them on the `DB` struct:

```go
type DB struct {
    pool  dbPool
    users *userRepository
    // ...
}

func New(pool *pgxpool.Pool) *DB {
    db := &DB{pool: pool}
    db.users = &userRepository{pool: pool}
    // ...
    return db
}

func (db *DB) Users() store.UserRepository { return db.users }
```

---

## 18. `territory_repository.go` -- boundary unmarshal duplicated 4 times

**Severity: Low -- Extract helper**

The pattern:
```go
if len(boundaryJSON) > 0 {
    t.Boundary = make(map[string]any)
    if err := json.Unmarshal(boundaryJSON, &t.Boundary); err != nil {
        return nil, fmt.Errorf("unmarshalling territory boundary: %w", err)
    }
}
```

appears in `Get`, `List`, `Create`, and `Update`. Extract:
```go
func unmarshalBoundary(data []byte) (map[string]any, error) {
    if len(data) == 0 { return nil, nil }
    m := make(map[string]any)
    if err := json.Unmarshal(data, &m); err != nil {
        return nil, fmt.Errorf("unmarshalling territory boundary: %w", err)
    }
    return m, nil
}
```

---

## 19. `geo.Geocoder` interface placement is correct

**Severity: None -- Positive note**

The `Geocoder` interface is defined in `geo/geocoder.go` alongside the `Result` type,
and consumed by `service/target_service.go`. Since `geo` is a shared package with
multiple potential implementations (Google, Nominatim, mock), defining the interface at
the declaration site is the right call. This follows the Go proverb: "Accept interfaces,
return structs" -- the service accepts `geo.Geocoder`, the google package returns
`*GoogleGeocoder`.

---

## 20. `store.Store` aggregate interface -- correct placement

**Severity: None -- Positive note**

The `Store` interface in `store/store.go` aggregates all repository interfaces. This is
the right pattern for a unit-of-work or aggregate root in Go -- it lives in the package
that defines the repository contracts, and the postgres package implements it.

---

## Summary of Actionable Items

| # | Severity | Item |
|---|----------|------|
| 1 | Medium | Adopt `pgx.CollectRows` across all scan loops |
| 2 | High | Wrap bare `rows.Err()` returns in 5 locations |
| 3 | Low | Unify `scanUser` for single-row and List variants |
| 4 | Low | Deduplicate `scanTarget` / `scanTargetWithFlag` |
| 5 | Medium | Consider extending `dbPool` for `SendBatch`/`CopyFrom` |
| 6 | Medium | Extract `Page`/`Limit` from `AuditFilter` |
| 7 | Low | Replace `_ = argIdx` with comment or remove |
| 8 | Medium | Use `tx.CopyFrom` in `replaceItems` |
| 9 | Medium | Remove dead `pgx.ErrNoRows` check in `scanAuditEntries` |
| 10 | Medium | Extract shared `queryBuilder` type |
| 11 | High | Use `defer rows.Close()` with separate vars in dashboard |
| 12 | Low | Replace `containsString` with `slices.Contains` |
| 13 | Info | Verify geocoder API key sourced from file mount |
| 14 | Medium | Set HTTP client timeout for geocoder |
| 15 | Medium | Log audit record failures instead of discarding |
| 16 | Medium | Consolidate scope-condition builders |
| 17 | Low | Cache repository instances on `DB` struct |
| 18 | Low | Extract `unmarshalBoundary` helper |
