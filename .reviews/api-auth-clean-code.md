# Clean Code Review: `internal/api/`, `internal/auth/`, `cmd/pebblr/`

Reviewer persona: **The Clean Code Geek**
Scope: naming, function length, SRP, handler complexity, middleware clarity, router organization, DRY violations

---

## Changes Applied in This PR

These are concrete code fixes committed alongside this review.

### 1. DRY: Hardcoded strings replaced with shared constants

**Files:** `team_handler.go`, `user_handler.go`, `config_handler.go`

These handlers used inline `"Content-Type"`, `"application/json"`, `"missing authenticated user"`, and `"an unexpected error occurred"` while the rest of the codebase consistently uses `headerContentType`, `contentTypeJSON`, `errMissingUser`, and `errUnexpected` from `constants.go`. Every handler should speak the same vocabulary.

### 2. DRY: `writeError` in `handlers.go` used hardcoded header

**File:** `handlers.go`

The very function responsible for writing error responses was not using its own package's constants. Fixed to use `headerContentType` and `contentTypeJSON`.

### 3. Consistency: Mixed JSON encoding strategies unified

**Files:** `collection_handler.go`, `territory_handler.go`, `audit_handler.go`, `activity_handler.go` (`CloneWeek`), `target_handler.go` (`FrequencyStatus`)

Some handlers used `writeJSON(w, r, v)` (which logs encoding errors), while others used `_ = json.NewEncoder(w).Encode(v)` (which silently swallows them). All success-path responses now use `writeJSON` for uniform error handling.

### 4. Logging: `fmt.Println` replaced with structured logger

**File:** `cmd/pebblr/serve.go`

`fmt.Println("server stopped")` was the only unstructured log line in the entire server lifecycle. Changed to `logger.Info("server stopped")` for consistency with the rest of the shutdown sequence.

---

## Findings Not Changed (Recommendations for Future Work)

### 5. SRP: `serve()` in `cmd/pebblr/serve.go` is a 130-line God Function

`serve()` does config loading, validation, migrations, DB connection, service construction, handler construction, auth setup, router assembly, server lifecycle, and graceful shutdown. That is at least five distinct responsibilities.

**Recommendation:** Extract into focused builder functions:
- `buildServices(pool, enforcer, tenantCfg, logger) -> services struct`
- `buildHandlers(services) -> handlers struct`
- `startServer(router, logger) -> error`

This makes each step independently testable and readable.

### 6. DRY: Repetitive actor-extraction boilerplate in every handler

Every single handler method begins with:
```go
actor, err := rbac.UserFromContext(r.Context())
if err != nil {
    writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
    return
}
```

This is 4 lines repeated 20+ times. Two options:

**Option A:** A middleware that rejects requests without a valid user before they reach handlers, then a simple `rbac.MustUserFromContext(ctx)` that panics (caught by Recoverer) on the impossible case.

**Option B:** A helper that returns `(actor, ok)` and writes the error:
```go
func requireActor(w http.ResponseWriter, r *http.Request) (*domain.User, bool) {
    actor, err := rbac.UserFromContext(r.Context())
    if err != nil {
        writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
        return nil, false
    }
    return actor, true
}
```
This cuts the boilerplate in half and ensures the error response is never accidentally inconsistent.

### 7. SRP: `mapActivityServiceError` is a 30-line switch doing error translation

This function maps 12 different sentinel errors. As new error types are added, this will grow unboundedly. Consider an error-to-HTTP-response mapping table:

```go
var activityErrorMappings = []struct {
    target error
    status int
    code   string
    msg    string
}{
    {service.ErrForbidden, http.StatusForbidden, "FORBIDDEN", "access denied"},
    {store.ErrNotFound, http.StatusNotFound, "NOT_FOUND", "activity not found"},
    // ...
}
```
A single generic `mapServiceError(w, err, mappings)` function could serve all handlers, eliminating the per-resource `mapXxxServiceError` functions.

### 8. Naming: `mapXxxServiceError` functions do not "map" -- they write

Functions named `mapTargetServiceError`, `mapActivityServiceError`, etc. don't return a mapped value -- they write directly to the ResponseWriter. A more intention-revealing name would be `writeServiceError` or `handleServiceError`. "Map" implies a pure transformation.

### 9. Auth middleware uses `http.Error` with raw JSON strings

**File:** `internal/auth/middleware.go`

```go
http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"..."}}`, http.StatusUnauthorized)
```

`http.Error` sets `Content-Type: text/plain`, which contradicts the JSON body. The correct approach is to set the JSON content type explicitly. Since `auth` cannot import `api` (circular dependency), the auth package should have its own minimal `writeJSONError` helper, or the error response writing should be extracted to a shared `httputil` package.

### 10. `demo/handler.go` duplicates constants and `writeError`

**File:** `internal/auth/demo/handler.go`

Declares its own `headerContentType` and `contentTypeJSON` constants, and its own `writeError` function. The `writeError` implementation differs from `api.writeError` -- it uses `map[string]map[string]string` instead of the typed `errorResponse` struct, producing a slightly different JSON shape.

This is a consistency risk. If the demo endpoints are part of the same API surface, they should produce identical error envelopes. Consider a shared `httputil` or `apiutil` package for response helpers.

### 11. `parseDashboardFilter` and `parseDashboardDateRange` are in `dashboard_handler.go` but used by `target_handler.go`

`target_handler.go`'s `FrequencyStatus` calls `parseDashboardFilter(r)`. This cross-handler coupling means dashboard parsing logic is not where you'd expect to find it when reading `target_handler.go`. Consider moving shared query-parsing helpers (date ranges, pagination) to a dedicated `request.go` or `parse.go` file.

### 12. Pagination parsing is duplicated

`target_handler.go` and `activity_handler.go` both inline:
```go
page, _ := strconv.Atoi(r.URL.Query().Get("page"))
if page < 1 { page = 1 }
limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
if limit < 1 { limit = 20 }
```

This should be a single `parsePagination(r *http.Request) (page, limit int)` function.

### 13. `BatchCreate` leaks internal error messages to the client

**File:** `activity_handler.go`, line 558:
```go
batchErrors = append(batchErrors, map[string]string{"targetId": item.TargetID, "error": err.Error()})
```

`err.Error()` may contain internal details (DB errors, stack info). Batch error responses should use the same sanitized error mapping as single-item responses.

### 14. Code organization praise

Things that are already clean and well-structured:

- **Thin handlers:** Business logic lives in the service layer; handlers only decode, delegate, encode. Textbook.
- **Interface-based DI:** Every handler depends on a `XxxServicer` interface, not a concrete type. Excellent testability.
- **Generic `newRouterIfNotNil`:** Elegant use of Go generics to avoid nil-check boilerplate in router setup.
- **`routeSpec` table-driven routing:** The `buildRouteSpecs` + `mountOrStub` pattern is a clean way to handle optional handlers.
- **Auth package design:** Clean `Authenticator` interface with three implementations (static, azuread, demo). The `ClaimsBridge` middleware is a thoughtful separation of token validation from domain user construction.
- **`azuread` package:** Proper JWKS refresh on unknown kid, proper error wrapping, body size limits on HTTP reads. Solid.

---

## Priority

| # | Finding | Severity | Effort |
|---|---------|----------|--------|
| 1-4 | DRY / consistency fixes | Low | Done |
| 6 | Actor-extraction boilerplate | Medium | Small |
| 12 | Pagination duplication | Medium | Small |
| 9-10 | Auth error response inconsistency | Medium | Medium |
| 5 | `serve()` SRP violation | Medium | Medium |
| 7-8 | Error mapping design | Low | Medium |
| 13 | Leaked error messages in batch | High | Small |
| 11 | Cross-handler coupling | Low | Small |
