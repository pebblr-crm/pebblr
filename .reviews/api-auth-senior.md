# API & Auth Security Review -- Senior Engineer

Reviewer: The Senior Engineer Who Has Seen It All
Date: 2026-04-02
Scope: `internal/api/`, `internal/auth/`, `cmd/pebblr/`

---

## Executive Summary

The codebase is reasonably well-structured for its maturity. Auth middleware is
correctly positioned, RBAC checks happen in every handler, and secrets are read
from files (not env vars) as required. That said, I found several issues that
range from "will bite you at 3 AM" to "someone will find this in a pentest."

Findings are ordered by severity. Inline code fixes accompany each finding
where the fix is straightforward.

---

## CRITICAL

### C1. No request body size limit -- OOM denial of service

**Files:** Every handler that calls `json.NewDecoder(r.Body).Decode()`
**Impact:** An attacker sends a 10 GB POST body and your pod gets OOM-killed.

Every JSON decode reads from `r.Body` without any size cap. The `http.Server`
has `ReadTimeout: 15s` which limits *time*, but over a fast network 15 seconds
is enough to push hundreds of MB.

**Fix:** Add a `maxBodySize` middleware or wrap the body at the handler level:

```go
// In middleware.go or at the top of each POST/PUT/PATCH handler:
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB default
```

Better yet, add it as chi middleware on the `/api/v1` group so every mutating
endpoint inherits it. The import handler may need a higher limit -- use a
per-route override there.

**Status:** FIX INCLUDED in this PR (`internal/api/middleware.go`)

---

### C2. No CORS configuration at all

**Files:** `internal/api/router.go`
**Impact:** The SPA frontend will work in dev (same-origin), but any
cross-origin deployment or CDN-hosted frontend will break. Worse, with no
explicit CORS policy, a future misconfiguration could allow credential-bearing
cross-origin requests from arbitrary origins.

**Fix:** Add an explicit CORS middleware with an allowlist. The `go-chi/cors`
package is already a chi ecosystem dependency and trivial to add. At minimum:

```go
r.Use(cors.Handler(cors.Options{
    AllowedOrigins:   cfg.AllowedOrigins, // from tenant config
    AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Authorization", "Content-Type"},
    AllowCredentials: true,
    MaxAge:           300,
}))
```

**Status:** NOTED -- requires a config change (add `AllowedOrigins` to
`TenantConfig` or server flags). Not fixed in this PR but should be done
before any non-same-origin deployment.

---

## HIGH

### H1. ClaimsBridge picks only the first role -- privilege confusion

**File:** `internal/auth/bridge.go:21-24`

```go
role := domain.RoleRep
if len(claims.Roles) > 0 {
    role = claims.Roles[0]
}
```

If Azure AD assigns `["rep", "admin"]`, the user gets `rep` because it is
first. If the order flips (which Azure AD does not guarantee), the user
silently gets elevated to `admin`. The highest-privilege role should be
selected deterministically, or better: propagate all roles and let the RBAC
enforcer evaluate them.

**Fix:** Use the highest-privilege role:

```go
role := rbac.HighestRole(claims.Roles)
```

Or propagate all roles if the RBAC enforcer supports it (preferred long-term).

**Status:** FIX INCLUDED in this PR (`internal/auth/bridge.go`)

---

### H2. Demo endpoints have no rate limiting

**File:** `internal/auth/demo/handler.go`, `internal/api/router.go:70-73`

The `/demo/token` endpoint issues signed JWTs with no rate limit, no
CAPTCHA, nothing. An attacker can brute-force user IDs (they are UUIDs, but
the `/demo/accounts` endpoint lists them all) and generate unlimited tokens.

In a demo environment this is "acceptable," but the demo handler should never
be reachable in production. There is no guard besides the `--auth-provider`
flag.

**Fix (defense-in-depth):**
1. Log a loud WARNING at startup when demo mode is active (already done, good).
2. Add rate limiting to `/demo/token` -- even a simple in-memory token bucket.
3. Consider adding a startup-time env guard: refuse to start in demo mode if
   a `PRODUCTION=true` env var is set or if the Kubernetes namespace matches
   a production pattern.

**Status:** NOTED -- recommend implementing before any internet-facing demo.

---

### H3. Batch create leaks internal error messages to clients

**File:** `internal/api/activity_handler.go:574`

```go
batchErrors = append(batchErrors, map[string]string{"targetId": item.TargetID, "error": err.Error()})
```

`err.Error()` from the service layer may contain internal details (SQL errors,
constraint names, stack info). Every other handler in the codebase correctly
maps errors to safe codes. This one bypasses that.

**Fix:** Use the same `mapActivityServiceError` pattern, or at minimum
sanitize the error string.

**Status:** FIX INCLUDED in this PR (`internal/api/activity_handler.go`)

---

### H4. Import endpoint has no item count limit

**File:** `internal/api/target_handler.go:301-353`

The import endpoint checks `len(req.Targets) == 0` but has no upper bound.
Someone can POST 100,000 targets in one request and stall the server for
minutes (or trigger a transaction timeout).

**Fix:** Add a cap:

```go
const maxImportItems = 1000
if len(req.Targets) > maxImportItems {
    writeError(w, http.StatusBadRequest, "BAD_REQUEST",
        fmt.Sprintf("import limited to %d items per request", maxImportItems))
    return
}
```

**Status:** FIX INCLUDED in this PR (`internal/api/target_handler.go`)

---

### H5. Batch create endpoint has no item count limit

**File:** `internal/api/activity_handler.go:527-597`

Same issue as H4. No upper bound on `req.Items`. And each item triggers an
individual `svc.Create` call -- so 10,000 items means 10,000 DB round trips
in a single HTTP request.

**Status:** FIX INCLUDED in this PR (`internal/api/activity_handler.go`)

---

## MEDIUM

### M1. Pagination limit has no upper bound

**Files:** `internal/api/target_handler.go:98-101`,
`internal/api/activity_handler.go:197-200`

```go
limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
if limit < 1 {
    limit = 20
}
```

A client can request `?limit=999999` and dump the entire table. Add a max:

```go
if limit > 200 {
    limit = 200
}
```

The audit handler already does this correctly (line 82: `l <= 200`). Apply the
same pattern everywhere.

**Status:** FIX INCLUDED in this PR

---

### M2. JWKS refresh has no rate limiting -- DoS amplification

**File:** `internal/auth/azuread/azuread.go:144-164`

When a token arrives with an unknown `kid`, the authenticator immediately calls
Azure AD to refresh the JWKS. An attacker can send tokens with random `kid`
values and force the server to hammer the Azure AD JWKS endpoint on every
request. Azure AD will eventually rate-limit you, and then legitimate key
rotations will fail.

**Fix:** Add a cooldown (e.g., refresh at most once per 60 seconds):

```go
type Authenticator struct {
    // ...existing fields...
    lastRefresh time.Time
    refreshInterval time.Duration // e.g., 60s
}
```

**Status:** FIX INCLUDED in this PR (`internal/auth/azuread/azuread.go`)

---

### M3. `Content-Type` header not set before `WriteHeader` in some paths

Scattered across handlers. When `writeError` is called, it correctly sets
Content-Type before WriteHeader. But in some success paths, `w.WriteHeader`
comes before `w.Header().Set()` in the implicit ordering of `json.NewEncoder`.
This is actually fine because `Encode` calls `Write` which triggers an
implicit 200 -- but it is fragile and inconsistent.

**Status:** NOTED -- low risk, cosmetic.

---

### M4. SPA handler path traversal is mitigated but fragile

**File:** `internal/api/router.go:143-144`

```go
path := filepath.Clean(req.URL.Path)
if _, err := fs.Stat(os.DirFS(distPath), path[1:]); err == nil {
```

`filepath.Clean` handles `..` sequences, and `os.DirFS` is rooted, so this is
safe today. But `filepath.Clean` on Windows can behave differently with
backslashes. Since the project explicitly does not support native Windows,
this is acceptable -- but add a comment explaining the security invariant.

**Status:** NOTED

---

### M5. Auth middleware returns `http.Error` not structured JSON

**File:** `internal/auth/middleware.go:20-21`

```go
http.Error(w, `{"error":...}`, http.StatusUnauthorized)
```

This sets `Content-Type: text/plain; charset=utf-8`, not `application/json`.
Clients parsing the response as JSON may work (the body is JSON) but the
content type is wrong. Pedantic, but it breaks API contracts.

**Status:** FIX INCLUDED in this PR (`internal/auth/middleware.go`)

---

### M6. Static authenticator always returns admin role

**File:** `internal/auth/static.go:27-33`

In local dev and E2E testing, the static authenticator always returns
`domain.RoleAdmin`. This means tests never exercise the `rep` or `manager`
RBAC paths unless the test explicitly overrides the context.

**Status:** NOTED -- acceptable for now, but consider making the role
configurable for E2E tests.

---

## LOW

### L1. Server listen address is hardcoded

**File:** `cmd/pebblr/serve.go:138`

```go
Addr: ":8080",
```

Should be configurable via flag or config for multi-service dev environments.

**Status:** NOTED

### L2. Graceful shutdown exits with `fmt.Println`

**File:** `cmd/pebblr/serve.go:167`

```go
fmt.Println("server stopped")
```

Should use the structured logger for consistency. Also, `os.Exit(1)` inside
the goroutine (line 153) bypasses the graceful shutdown entirely -- the defer
on `pool.Close()` never runs.

**Status:** FIX INCLUDED in this PR (`cmd/pebblr/serve.go`)

### L3. `ReadHeaderTimeout` is not set

**File:** `cmd/pebblr/serve.go:137-143`

Go's HTTP server without `ReadHeaderTimeout` is vulnerable to slowloris-style
attacks. `ReadTimeout` covers the full read (headers + body), but
`ReadHeaderTimeout` lets you be stricter on headers while allowing longer
body uploads.

**Status:** FIX INCLUDED in this PR (`cmd/pebblr/serve.go`)

---

## Summary of Changes in This PR

| Finding | File | Change |
|---------|------|--------|
| C1 | `internal/api/middleware.go` | Add `maxBodySize` middleware |
| H1 | `internal/auth/bridge.go` | Pick highest-privilege role deterministically |
| H3 | `internal/api/activity_handler.go` | Sanitize batch error messages |
| H4 | `internal/api/target_handler.go` | Cap import batch size |
| H5 | `internal/api/activity_handler.go` | Cap batch create size |
| M1 | `internal/api/target_handler.go`, `activity_handler.go` | Cap pagination limit |
| M2 | `internal/auth/azuread/azuread.go` | Add JWKS refresh cooldown |
| M5 | `internal/auth/middleware.go` | Use structured JSON for auth errors |
| L2 | `cmd/pebblr/serve.go` | Fix shutdown logging and goroutine exit |
| L3 | `cmd/pebblr/serve.go` | Add `ReadHeaderTimeout` |
