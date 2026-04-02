# Idiomatic Go Review: `internal/api/`, `internal/auth/`, `cmd/pebblr/`

Reviewer persona: **The Idiomatic Obsessive**

## Changes Applied in This PR

### 1. Consistent use of package constants (team, user, config handlers)

Several handlers used hardcoded `"Content-Type"`, `"application/json"`, `"missing
authenticated user"`, and `"an unexpected error occurred"` strings instead of the
constants already defined in `constants.go` (`headerContentType`, `contentTypeJSON`,
`errMissingUser`, `errUnexpected`). Fixed across `team_handler.go`,
`user_handler.go`, and `config_handler.go`.

When you define constants, use them everywhere or delete them. Half-adopted
constants are worse than no constants -- they create the illusion of a single
source of truth that doesn't actually exist.

### 2. `writeJSON` used consistently instead of bare `json.NewEncoder`

Multiple handlers (`collection_handler.go`, `territory_handler.go`,
`audit_handler.go`, `target_handler.go` FrequencyStatus, `activity_handler.go`
CloneWeek) were calling `_ = json.NewEncoder(w).Encode(...)` directly, silently
discarding encode errors. The codebase already has `writeJSON` which logs encode
failures via the request-scoped logger. Every successful JSON response should go
through `writeJSON`.

### 3. Auth middleware: JSON Content-Type on error responses

`auth.Middleware` and `auth.ClaimsBridge` used `http.Error` which sets
`Content-Type: text/plain; charset=utf-8`. For a JSON API with a documented error
contract (`{"error":{"code":"...","message":"..."}}`), every response -- including
auth failures -- must have `Content-Type: application/json`. Replaced with a local
`writeJSONError` helper.

### 4. `cmd/pebblr/serve.go`: `errors.Is` for sentinel comparison

`err != http.ErrServerClosed` was a direct comparison. Idiomatic Go since 1.13
uses `errors.Is` for sentinel errors. Fixed to `!errors.Is(err, http.ErrServerClosed)`.

### 5. `cmd/pebblr/serve.go`: No `os.Exit` inside a goroutine

The ListenAndServe goroutine called `os.Exit(1)` on error. This bypasses deferred
cleanup, skips the graceful shutdown path, and is untestable. Replaced with a
channel-based signal back to the main goroutine.

### 6. `cmd/pebblr/serve.go`: Logger instead of `fmt.Println`

`fmt.Println("server stopped")` was the only unstructured output in the entire
server lifecycle. Replaced with `logger.Info("server stopped")` for consistency.

### 7. `writeError` in `handlers.go`: Use constant for Content-Type

The `writeError` helper itself used the hardcoded string `"Content-Type"` instead of
`headerContentType`. Fixed.

---

## Observations (Not Changed -- For Future Consideration)

### A. `contextKey` type collision across packages

Both `internal/api/middleware.go` and `internal/auth/middleware.go` define their own
unexported `type contextKey string`. This is fine -- each package's type is distinct
because Go's type identity includes the package path. But it's worth noting that if
these packages were ever merged, you'd get a silent collision. The current design is
correct.

### B. Handler boilerplate: actor extraction

Every single handler method starts with:
```go
actor, err := rbac.UserFromContext(r.Context())
if err != nil {
    writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
    return
}
```

This is a 5-line tax on every endpoint. Consider a `requireUser` middleware that
runs after `ClaimsBridge` and stores the `*domain.User` so handlers can call
`rbac.MustUserFromContext(r.Context())` (which panics if absent, safe because the
middleware guarantees it). This would eliminate the boilerplate and make the auth
contract explicit in the middleware chain rather than scattered across handlers.

### C. `ClaimsBridge` picks only `Roles[0]`

```go
role := domain.RoleRep
if len(claims.Roles) > 0 {
    role = claims.Roles[0]
}
```

This discards all roles after the first. If a user has `[admin, manager]`, they
get `admin` and lose `manager`. If the domain truly supports only one role per
user, `UserClaims.Roles` should be `Role` (singular). If multiple roles are
meaningful, the bridge should carry them all.

### D. `demo.Handler.writeError` duplicates `api.writeError`

The `demo` package has its own `writeError` with a slightly different envelope
shape (`map[string]map[string]string` vs `errorResponse` struct). This means demo
endpoints return a subtly different error shape than API endpoints. Consider
extracting a shared `httputil` or `apiutil` package for the error envelope.

### E. Leaky abstraction: `service` error types in handlers

The `mapActivityServiceError`, `mapTargetServiceError`, etc. functions use
`errors.Is` against sentinel errors from the `service` package. This is correct and
idiomatic. However, the error mapping is duplicated across handlers -- the
`ErrForbidden -> 403`, `ErrNotFound -> 404`, `ErrInvalidInput -> 400` mapping is
repeated verbatim. A shared `mapCommonServiceError` helper with a resource-name
parameter would reduce duplication while keeping the mapping explicit.

### F. `RouterConfig` is a grab bag

`RouterConfig` has 11 handler fields plus a logger, authenticator, and web path.
This is a construction-time concern -- it's essentially a manual DI container. It
works, but as the service grows it will become unwieldy. Consider grouping related
handlers or using a `HandlerSet` struct.

### G. `mountSPA` uses `os.DirFS` + `http.Dir` redundantly

`mountSPA` creates both `http.Dir(distPath)` (for the file server) and
`os.DirFS(distPath)` (for the existence check). These are two different
abstractions over the same directory. Using `http.Dir` for both or `os.DirFS` for
both would be more consistent.

### H. No request body size limits

`json.NewDecoder(r.Body).Decode(...)` is used throughout without an
`http.MaxBytesReader` wrapper. A malicious client can send an arbitrarily large
body. Consider a middleware that wraps `r.Body` with `http.MaxBytesReader(w, r.Body, maxBytes)`.

---

**Overall assessment:** The codebase is quite clean. Handler patterns are consistent
and thin. Error handling is explicit. Interface segregation is well done (each
handler defines exactly the service interface it needs). Context threading is
correct throughout. The issues found are mostly about consistency (constants,
`writeJSON` usage, Content-Type headers) rather than fundamental design problems.
