Audit data access code for RBAC enforcement gaps.

Pebblr enforces per-row RBAC: users may only access leads assigned to them. This must be enforced at the data layer, not just the API layer.

Scan Go files for database queries that select leads without a user filter:

```bash
grep -rn "SELECT\|\.Find\|\.Query\|\.Get" --include="*.go" internal/
```

For each query touching the `leads` table (or any row-gated resource):
1. Check whether a user ID or role filter is applied in the WHERE clause or query builder
2. Flag any query that returns all rows without scoping to the requesting user
3. Note whether the filter is applied at the repository layer (good) or only at the handler layer (bad — too late)

Report violations. A query that returns unscoped lead data is a security defect.
