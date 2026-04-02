# Senior Engineer Review: Store, Service & Geo Layers

Reviewer: "The Senior Engineer Who Has Seen It All"
Date: 2026-04-02
Scope: `internal/store/`, `internal/store/postgres/`, `internal/service/`, `internal/geo/`, `migrations/`

---

## Overall Assessment

The codebase is well-structured for an early-stage project. Parameterized queries
throughout (no SQL injection). RBAC scope enforcement is consistently applied at
the store layer for list queries. Error wrapping is disciplined. Secrets are read
from files, not env vars. The bones are solid.

That said, there are production-time-bomb patterns I have seen bring down systems
at 2 AM. Here they are, ordered by blast radius.

---

## CRITICAL: Fix Before Production

### C1. Target Upsert is an Unbounded Single Transaction

**File:** `internal/store/postgres/target_repository.go` lines 220-271

The `Upsert` method opens a single transaction and issues one INSERT per target
in a loop. With an import of 10,000 targets (routine for pharma CRM), this will:

- Hold a single long-running transaction, blocking autovacuum and escalating locks
- Accumulate WAL without checkpoint, risking OOM on the replica
- Risk statement-level timeout if any individual row is slow

**Recommendation:** Batch into chunks of 100-500 rows. Use `pgx.CopyFrom` or
multi-row INSERT with `unnest`. If chunking, commit per-chunk and report partial
progress. At minimum, add a size guard:

```go
if len(targets) > 5000 {
    return nil, fmt.Errorf("import batch too large (%d targets, max 5000)", len(targets))
}
```

### C2. Dashboard Issues Two Separate Queries Without Shared Snapshot

**File:** `internal/store/postgres/dashboard_repository.go` `ActivityStats` method

The `ActivityStats` method runs a `GROUP BY status` query and then a separate
`GROUP BY activity_type` query against the same table with the same WHERE clause.
Between the two queries, rows can be inserted/updated, producing inconsistent
counts (status totals != category totals). At scale this will confuse users and
make the dashboard appear broken.

**Recommendation:** Combine into a single query:

```sql
SELECT status, activity_type, COUNT(*)
FROM activities WHERE ...
GROUP BY status, activity_type
```

Then aggregate in Go. Or wrap both queries in a single READ ONLY transaction with
REPEATABLE READ isolation.

### C3. `autoCompleteNonFieldActivities` Writes During a Read Path

**File:** `internal/service/activity_service.go` lines 204-239

The `List` method calls `autoCompleteNonFieldActivities`, which issues UPDATE
statements for each activity whose due date has passed. This means:

- A read-only API call (`GET /activities`) has write side effects
- If 50 activities need auto-completion, that is 50 individual UPDATEs
- Under load this creates write contention on a read endpoint
- The errors are silently swallowed (`slog.Warn`) -- partial state transitions

**Recommendation:** Move auto-completion to a periodic background job or a
database trigger. If it must stay in the read path, do it as a single bulk UPDATE
with RETURNING and use a transaction:

```sql
UPDATE activities SET status = 'completed', updated_at = NOW()
WHERE status = $1 AND activity_type = ANY($2) AND due_date < $3
  AND submitted_at IS NULL AND deleted_at IS NULL
RETURNING id
```

### C4. `rows.Err()` Check After `rows.Close()` in Dashboard

**File:** `internal/store/postgres/dashboard_repository.go` lines 57-60, 79-81

Pattern:
```go
rows.Close()
if err := rows.Err(); err != nil {
    return nil, ...
}
```

Per `pgx` semantics, `rows.Close()` consumes remaining rows and sets the error
state. Calling `rows.Err()` after `rows.Close()` is correct in pgx (unlike
`database/sql` where it is undefined). However, the `defer rows.Close()` is
missing -- if the scan loop panics, the rows leak. The current code calls
`rows.Close()` explicitly but not via defer.

**Recommendation:** Use `defer rows.Close()` immediately after `Query()` and
check `rows.Err()` after the iteration loop, before returning. This is already
done correctly in most other repositories -- make it consistent.

---

## HIGH: Fix Before First Customer

### H1. Missing Indexes for Activity Soft-Delete Queries

Almost every activity query includes `WHERE deleted_at IS NULL`. There is no
partial index on this condition.

**File:** `migrations/003_activities.up.sql`

**Recommendation:** Add a partial index:

```sql
CREATE INDEX idx_activities_not_deleted ON activities(id)
    WHERE deleted_at IS NULL;
```

Better yet, a composite partial index for the most common list query:

```sql
CREATE INDEX idx_activities_creator_date_active
    ON activities(creator_id, due_date)
    WHERE deleted_at IS NULL;
```

### H2. Target `Get` Has No RBAC Enforcement at Store Level

**File:** `internal/store/postgres/target_repository.go` line 52-58

`target_repository.Get()` returns any target by ID with no scope check. The
service layer (`TargetService.Get`) does call `CanViewTarget`, but other internal
callers (like `ActivityService.checkTargetAccess`) fetch the target via
`s.targets.Get()` and then check access separately. If any new caller forgets
the RBAC check, data leaks.

**Recommendation:** This is acceptable if the team is disciplined, but document
the contract clearly. Consider adding a `GetScoped(ctx, id, scope)` method that
bakes in the RBAC check for defense-in-depth.

### H3. Audit Record Failures are Silently Discarded

**Files:** `internal/service/activity_service.go` (lines 118, 264, 330, 374, 466)
and `internal/service/target_service.go` (line 151)

Pattern: `_ = s.audit.Record(ctx, ...)`. If the audit insert fails (connection
error, constraint violation, disk full), nobody knows. For a CRM that handles
pharmaceutical field sales, audit is likely a compliance requirement.

**Recommendation:** At minimum, log the error:

```go
if err := s.audit.Record(ctx, entry); err != nil {
    slog.Error("audit record failed", "entity", entry.EntityType, "id", entry.EntityID, "err", err)
}
```

Better: return the error and let the caller decide. For compliance-sensitive
operations, the business operation should fail if audit fails.

### H4. `UserService.List` Returns All Users Without Pagination

**File:** `internal/service/user_service.go` and `internal/store/postgres/user_repository.go`

`List` returns ALL users with no pagination. For a multi-team deployment with
hundreds of users, this is wasteful. The query also does a LEFT JOIN + array_agg
which has O(users * team_memberships) cost.

**Recommendation:** Add pagination parameters or at minimum a LIMIT guard.

### H5. Collection `replaceItems` Does Not Validate Target IDs

**File:** `internal/store/postgres/collection_repository.go` lines 168-190

The `replaceItems` method inserts target IDs into `target_collection_items`
without verifying they exist. The FK constraint will catch invalid UUIDs, but:

- The error message from a FK violation is cryptic to end users
- A bulk insert with one bad ID fails the entire batch with no indication of which ID

**Recommendation:** Either validate upfront or catch the FK violation error and
return a meaningful `ErrInvalidInput` with the bad target ID.

### H6. `CloneWeek` Hardcoded to 200-Activity Page Limit

**File:** `internal/service/activity_service.go` lines 493-498

`CloneWeek` fetches source week activities with `limit=200`. If a power user has
more than 200 activities in a week (unlikely but possible with a generous
`max_activities_per_day` config), the clone silently drops the rest.

**Recommendation:** Either paginate until exhausted, or return an error if the
page is full:

```go
if len(sourcePage.Activities) >= 200 {
    return nil, fmt.Errorf("source week has too many activities to clone")
}
```

---

## MEDIUM: Hardening

### M1. No Connection Pool Limits Configured

**File:** `internal/store/postgres/postgres.go` lines 34-51

`Connect` parses the DSN and creates a pool but never configures:

- `MaxConns` (defaults to 4 in pgx, which is low for production)
- `MinConns` (no warm connections at startup)
- `MaxConnLifetime` (connections live forever, stale after PG restart)
- `MaxConnIdleTime` (idle connections never reclaimed)
- `HealthCheckPeriod`

**Recommendation:** After `pgxpool.ParseConfig`, set sensible defaults:

```go
cfg.MaxConns = 20
cfg.MinConns = 2
cfg.MaxConnLifetime = 30 * time.Minute
cfg.MaxConnIdleTime = 5 * time.Minute
cfg.HealthCheckPeriod = 30 * time.Second
```

Or better: make these configurable via the DSN or a separate config file.

### M2. Missing `rows.Err()` Check After Loop

**File:** `internal/store/postgres/target_repository.go` line 355

`VisitStatus` returns `result, rows.Err()` -- this is correct. But worth noting
that `FrequencyStatus` (line 416) does the same. Both are fine. However, compare
with `dashboard_repository.go` where the pattern is inconsistent (see C4 above).

### M3. Google Geocoder Uses `http.DefaultClient`

**File:** `internal/geo/google.go` line 22

`http.DefaultClient` has no timeout. A slow/hung Google API response will hold a
goroutine and a DB connection (if called during an import transaction) forever.

**Recommendation:** Set a timeout:

```go
httpClient: &http.Client{Timeout: 10 * time.Second},
```

### M4. Google Geocoder Response Body Not Size-Limited

**File:** `internal/geo/google.go` line 70

`json.NewDecoder(resp.Body).Decode(...)` reads the entire body. A malicious or
buggy upstream could send gigabytes. Use `io.LimitReader`:

```go
limited := io.LimitReader(resp.Body, 1<<20) // 1 MB
if err := json.NewDecoder(limited).Decode(&gResp); err != nil { ... }
```

### M5. Target ILIKE Query Needs Escaping

**File:** `internal/store/postgres/target_repository.go` line 131

```go
b.addCondition("name ILIKE $%d", "%"+*filter.Query+"%")
```

This is parameterized (safe from injection), but the `%` and `_` characters in
the user's query are not escaped. A search for `%` matches everything. A search
for `_` matches any single character.

**Recommendation:** Escape special LIKE characters:

```go
escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(*filter.Query)
b.addCondition("name ILIKE $%d", "%"+escaped+"%")
```

### M6. RLS Policy for Targets Allows All Managers to See All Targets

**File:** `migrations/002_targets.up.sql` lines 30-35

The RLS policy grants all managers access to all targets regardless of team:

```sql
current_setting('app.user_role', true) IN ('manager', 'admin')
```

But the application layer scopes managers to their team's targets. This means RLS
is a weaker guard than intended -- a manager who bypasses the app layer (direct
DB access, SQL injection in a different service) sees everything.

**Recommendation:** Tighten the RLS policy for managers to include team_id check:

```sql
USING (
    current_setting('app.user_role', true) = 'admin'
    OR assignee_id::TEXT = current_setting('app.user_id', true)
    OR (current_setting('app.user_role', true) = 'manager'
        AND team_id IN (
            SELECT team_id FROM team_members
            WHERE user_id = current_setting('app.user_id', true)::uuid
        ))
)
```

### M7. Activities RLS Policy Does Not Check `deleted_at`

**File:** `migrations/003_activities.up.sql` lines 33-38

The RLS policy does not filter out soft-deleted activities. A direct query (e.g.
from a data analyst or a future reporting service) will see deleted rows.

**Recommendation:** Add `AND deleted_at IS NULL` to the RLS policy, or create
separate policies for SELECT and UPDATE/DELETE operations.

### M8. `team_repository.go` is a Stub That Lies

**File:** `internal/store/postgres/team_repository.go`

`Get` returns `ErrNotFound` for every input. `List` returns an empty slice.
These are not "not implemented" -- they are silently wrong. Any code relying on
team data will behave as if no teams exist.

**Recommendation:** Either implement the repository or have it return
`errNotImplemented` for ALL methods (including Get and List) so callers fail
loudly. Currently `List` silently returns empty, which will cause RBAC scoping
to produce wrong results for managers.

---

## LOW: Polish

### L1. Store Creates New Repository Instance On Every Call

**File:** `internal/store/postgres/store_impl.go`

Every call to `db.Users()`, `db.Targets()`, etc. allocates a new struct. These
are tiny (just a pool reference), so it is not a performance issue, but it means
you cannot cache state in a repository instance. Consider caching them as fields
on DB if you ever need repository-level state.

### L2. `nullIfEmpty` Helper Only Used in a Few Places

The helper converts `""` to `nil` for nullable columns. Consider using
`pgtype.Text` or a custom scanner instead to make the intent clearer at the type
level.

### L3. Audit Log Table Will Grow Without Bound

**File:** `migrations/004_audit_log.up.sql`

No partitioning strategy. For a CRM where every activity create/update/delete
generates an audit entry, this table will grow to millions of rows in the first
year. The `created_at DESC` ordering means recent queries are fast initially but
degrade as the table grows.

**Recommendation:** Plan for table partitioning by `created_at` (monthly or
quarterly) or implement a retention policy. At minimum, document the expected
growth rate.

### L4. `ErrConflict` Defined But Never Returned

**File:** `internal/store/errors.go`

`ErrConflict` is defined but no repository method maps a unique constraint
violation to it. The `ON CONFLICT` clause in upsert silently resolves conflicts.
The unique index in migration 009 would cause a raw Postgres error if violated.

**Recommendation:** Map `pgconn.PgError` with code `23505` to `store.ErrConflict`
in the activity repository's Create method.

### L5. Missing `ON DELETE` Behavior for Activity Foreign Keys

**File:** `migrations/003_activities.up.sql`

The `target_id`, `creator_id`, `joint_visit_user_id`, and `team_id` FKs have
no explicit `ON DELETE` behavior (defaults to `RESTRICT`). This means you cannot
delete a user who has created activities, which is correct for data integrity but
should be documented as intentional.

---

## Migration-Specific Findings

### Migration 009: Data-Modifying Migration

The migration soft-deletes duplicate rows before adding a unique index. This is
necessary but dangerous:

- No transaction wrapping (golang-migrate runs each file as a single transaction
  by default, but verify this)
- The ROW_NUMBER keeps the **oldest** (ORDER BY created_at ASC, rn > 1 means
  newer duplicates are soft-deleted). Confirm this is the desired behavior --
  typically you want to keep the **most recent**.
- No audit trail for the auto-soft-deleted rows

### Missing Migration: `app.user_id` / `app.user_role` Session Settings

The RLS policies reference `current_setting('app.user_id', true)` and
`current_setting('app.user_role', true)`, but I see no code that sets these
session variables before queries. If they are never set, the RLS policies
effectively deny all access for non-superuser connections (the `true` parameter
returns NULL on missing setting, which will not match anything).

**Recommendation:** Verify that the application sets these via
`SET LOCAL app.user_id = '...'` within each request's transaction, or document
that RLS is purely defense-in-depth and the app connection uses a superuser/
`BYPASSRLS` role.

---

## Summary Scorecard

| Category | Grade | Notes |
|---|---|---|
| SQL Injection Safety | A | All queries parameterized |
| RBAC Enforcement | B+ | Consistent at service layer; store layer trusts callers for Get |
| Transaction Boundaries | B- | Upsert is unbounded; dashboard has snapshot inconsistency |
| Error Handling | B | Good wrapping; audit errors silently swallowed |
| Connection Management | C+ | No pool tuning; no HTTP client timeouts |
| Migration Quality | B | Good indexes; RLS policies could be tighter |
| N+1 Patterns | A- | Generally good; auto-complete is O(n) updates |
| Data Integrity | B+ | Good constraints; some gaps in validation |

Total: **Solid foundation, needs production hardening before first customer deploy.**
