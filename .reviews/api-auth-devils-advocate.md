# Devil's Advocate Review: `internal/api/`, `internal/auth/`, `cmd/pebblr/`

**Reviewer posture:** Challenge every design decision. If the current code works, ask whether it works _for the right reasons_ and whether it will survive the next three feature additions without a rewrite.

---

## 1. The `serve()` Function Is a 265-Line God Function

**File:** `cmd/pebblr/serve.go`

`serve()` does config loading, migration, DB connection, service construction, handler construction, auth provider selection, router assembly, and graceful shutdown. That is at least five distinct responsibilities in one function.

**Challenge:** Why not decompose this? A `buildServices()` and `buildHandlers()` extraction would make `serve()` a readable pipeline instead of a wall of `NewXxxService` / `NewXxxHandler` calls. Right now, adding a new resource means touching `serve()` in three places (service, handler, router config). That is O(3n) modification points per resource -- a maintenance tax that will only grow.

**Alternative:** Introduce a lightweight `app` struct that owns the DB pool, services, and handlers. `serve()` becomes `app.Run()`, and handler registration is declarative.

---

## 2. `RouterConfig` Is a Bag of 11 Handlers -- And Growing

**File:** `internal/api/router.go`, lines 18-32

Every new resource adds a field to `RouterConfig`, a line to `buildRouteSpecs()`, and a line to `serve()`. This is a classic "struct that grows forever" anti-pattern.

**Challenge:** Why is the router config a flat struct instead of a registry? You already have the `routeSpec` abstraction with `buildRouteSpecs()` -- but then `ConfigHandler` gets special-cased outside the loop (line 125-129). The `DemoHandler` also lives outside the spec table. Two different registration mechanisms for the same concept.

**Alternative:** A `RouteRegistrar` interface (`Register(chi.Router)`) per handler, with a slice `[]RouteRegistrar` passed to the router. Each handler knows its own routes; the router just iterates. The special-casing of `/config` and `/demo` disappears.

---

## 3. The `mountOrStub` / `newRouterIfNotNil` Generics -- Solving the Wrong Problem?

**File:** `internal/api/router.go`, lines 78-117

You built generic nil-checking infrastructure (`newRouterIfNotNil[T]`) so that handlers can be optional and stubs are served instead. But _when_ would a handler actually be nil in production? Every handler is constructed unconditionally in `serve()`. The nil path only fires if someone deliberately passes a nil service. This is defensive code for a scenario that cannot occur at runtime.

**Challenge:** Is this complexity justified? If the answer is "we want to enable incremental feature rollout," then feature flags belong at the service layer or middleware level, not at the router-registration level. A 501 stub endpoint that silently appears in production is a support nightmare -- clients get a response that looks like a bug, not a disabled feature.

**Alternative:** Remove the nil-guard machinery. If a feature is not ready, do not register the route at all. Or use feature-flag middleware that returns 404 with a clear message.

---

## 4. Two Separate `contextKey` Types in `auth` and `api` -- Collision Waiting to Happen

**Files:** `internal/auth/middleware.go:9` and `internal/api/middleware.go:12`

Both packages define `type contextKey string` independently. Today they use different key values (`"claims"` vs `"logger"`), so there is no collision. But this is a ticking bomb. Anyone adding a new context key in either package could accidentally shadow the other.

**Challenge:** Why not a single shared `contextkey` package, or at minimum use the package-qualified type (which is already unique due to Go's type system)? The fact that both types are named `contextKey` makes code review harder -- a grep for `contextKey` hits both, and you have to verify which one you are looking at.

**Alternative:** The `auth` package's context key should be an unexported struct type (e.g., `type claimsKeyType struct{}`), not a string. This is the idiomatic Go pattern and eliminates any possibility of collision even across packages.

---

## 5. `ClaimsBridge` Takes Only the First Role -- Silently Dropping Multi-Role Users

**File:** `internal/auth/bridge.go`, lines 21-24

```go
role := domain.RoleRep
if len(claims.Roles) > 0 {
    role = claims.Roles[0]
}
```

Azure AD can assign multiple app roles. This code takes the first one and ignores the rest. If a user has `[manager, admin]`, they get `manager`. If the order changes in Azure AD (which is not guaranteed to be stable), they get a different role.

**Challenge:** Is single-role by design or by accident? The `UserClaims.Roles` field is a slice, suggesting multi-role was intended. But `domain.User.Role` is singular. This is a data model mismatch. Either commit to single-role (change `UserClaims.Roles` to `UserClaims.Role`) or commit to multi-role (change `domain.User.Role` to `domain.User.Roles` and update the RBAC enforcer).

**The current code is a silent bug** if anyone configures multiple roles in Azure AD.

---

## 6. `StaticAuthenticator` Returns Hardcoded Admin Claims -- Security Footgun

**File:** `internal/auth/static.go`, lines 27-33

The static authenticator always returns admin-level claims. Anyone who knows the static token gets full admin access. The comment says "local development and E2E testing," but:

**Challenge:** What prevents this from being deployed to production? The `auth-provider` flag defaults to `"static"`. If someone deploys without explicitly setting `--auth-provider=azuread`, every request with the JWT secret gets admin access. There is no guardrail.

**Alternative:**
1. Log a loud `WARN` at startup when static auth is active (similar to demo's "NOT FOR PRODUCTION" log line).
2. Better: refuse to start with `static` auth unless an explicit `--dev-mode` flag is also passed.
3. Allow the static authenticator to return configurable claims (role, team, user ID) for testing different RBAC scenarios in E2E tests.

---

## 7. The Auth Middleware Returns `http.Error` While Handlers Use `writeError` -- Inconsistent Error Format

**File:** `internal/auth/middleware.go`, lines 21-22, 26-27

```go
http.Error(w, `{"error":{"code":"UNAUTHORIZED",...}}`, http.StatusUnauthorized)
```

And in `internal/auth/bridge.go`, line 17:

```go
http.Error(w, `{"error":{"code":"UNAUTHORIZED",...}}`, http.StatusUnauthorized)
```

These are hand-crafted JSON strings passed to `http.Error`, which appends a newline. Meanwhile, every handler uses the `writeError()` helper that properly sets `Content-Type: application/json` and encodes via `json.Encoder`.

**Challenge:** The auth middleware produces subtly different responses than the rest of the API:
- `http.Error` sets `Content-Type: text/plain; charset=utf-8` (Go's default), not `application/json`.
- `http.Error` appends `\n` to the body.

A client parsing the `Content-Type` header to decide how to decode the response will break on auth errors.

**Fix:** The auth middleware should use the same `writeError()` from `internal/api`, or the auth package needs its own equivalent that sets the correct content type.

---

## 8. `demo/handler.go` Has Its Own `writeError` -- Three Error Writers Total

**Files:** `internal/api/handlers.go:29`, `internal/auth/demo/handler.go:133`, `internal/auth/middleware.go:21`

Three different places write error responses, each with slightly different behavior:
- `api.writeError` -- sets Content-Type, uses `json.Encoder`
- `demo.writeError` -- sets Content-Type, uses `json.Encoder`, but encodes a `map[string]map[string]string` instead of `errorResponse`
- `auth.Middleware` / `ClaimsBridge` -- uses `http.Error` with a raw JSON string

**Challenge:** Why does the demo package duplicate the error response logic? It cannot import `internal/api` (that would create a dependency from auth -> api), but it _could_ use a shared `httperr` package. Or the demo handler could live in `internal/api` since it is an HTTP handler, not an auth primitive.

**Alternative:** Extract `writeError` into a tiny `internal/httputil` package that both `api` and `demo` import. This eliminates three implementations of the same concept.

---

## 9. Handler Boilerplate: Every Single Handler Starts With the Same 4 Lines

Every handler method begins with:

```go
actor, err := rbac.UserFromContext(r.Context())
if err != nil {
    writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
    return
}
```

This is repeated **30+ times** across the handler files.

**Challenge:** This check is redundant given the middleware chain. `auth.Middleware` + `auth.ClaimsBridge` already guarantee that a user is in context by the time a handler runs. If the middleware fails, the request never reaches the handler. The `UserFromContext` check in every handler is belt-and-suspenders that adds 4 lines of noise per endpoint.

**Alternative:**
1. If defense-in-depth is the goal, extract an `actorFromRequest(r) *domain.User` helper that panics (caught by Recoverer middleware) on missing user -- turning 4 lines into 1.
2. Or trust the middleware chain and remove the redundant checks entirely, using `rbac.MustUserFromContext(ctx)` that panics -- the Recoverer will catch it and return 500, which is correct because a missing user at handler level means the middleware chain is broken.

---

## 10. Inconsistent Response Envelope Patterns

Compare these response patterns across handlers:

| Handler | List response | Single response |
|---------|--------------|-----------------|
| Target | `{items, total, page, limit}` via typed struct | `{target: {...}}` via typed struct |
| Activity | `{items, total, page, limit}` via typed struct | `{activity: {...}}` via typed struct |
| Team | `{items, total}` via typed struct | `{team, members}` via typed struct |
| User | `{items, total}` via typed struct | `{user: {...}}` via typed struct |
| Collection | `{items}` via `map[string]any` | raw object (no envelope) |
| Territory | `{items, total}` via `map[string]any` | raw object (no envelope) |
| Audit | `{items, total, page, limit}` via `map[string]any` | N/A |
| Dashboard | raw object (no envelope) | N/A |

**Challenge:** Collection and Territory use `map[string]any` for response encoding instead of typed structs. Collection's List omits `total`. Collection's Get returns a raw object without a `collection` wrapper. Territory's Get also returns unwrapped. Dashboard endpoints return unwrapped stats. This is inconsistent -- a frontend developer has to learn different envelope shapes per resource.

**Alternative:** Define a standard `ListResponse[T]` generic and a `SingleResponse[T]` generic. Enforce the convention: lists always return `{items, total, page, limit}`, singles always return `{<resource_name>: {...}}`. The typed struct approach used by Target/Activity is correct; the `map[string]any` shortcuts in Collection/Territory/Audit should be replaced.

---

## 11. `os.Getenv("WEB_DIST_PATH")` and `os.Getenv("SECRET_MOUNT_PATH")` -- Env Vars for Non-Secrets?

**File:** `cmd/pebblr/serve.go`, lines 109-113

The project's CLAUDE.md is emphatic: "Secrets are never in env vars." These are not secrets, but the inconsistency is confusing. `configPath` is a CLI flag. `authProvider` is a CLI flag. But `WEB_DIST_PATH` and `SECRET_MOUNT_PATH` are env vars. Why the mixed approach?

**Challenge:** If `--config` is a flag, why is `--web-dist` not also a flag? The env var approach makes the configuration surface area invisible -- you cannot discover these by running `pebblr serve --help`.

**Alternative:** Make them flags with env-var fallback (using `fs.StringVar` with a default from `os.Getenv`). At minimum, document them in the usage string.

---

## 12. Migrations Run Inside `serve()` -- Blocking Startup, No Rollback Strategy

**File:** `cmd/pebblr/serve.go`, lines 68-69, 170-195

`serve()` runs migrations before starting the HTTP server. This means:
- If a migration is slow, the readiness probe fails, and Kubernetes kills the pod.
- If a migration fails, the pod crashes. In a multi-replica deployment, all replicas race to migrate.
- There is no rollback path -- a bad migration bricks the deployment.

**Challenge:** Why are migrations coupled to the application server? The `cmd/migrate/main.go` already exists as a separate binary. The standard Kubernetes pattern is to run migrations as an init container or a Job, not inside the application server.

**Alternative:** Remove `runMigrations()` from `serve()`. Run migrations exclusively via `cmd/migrate` in a Kubernetes Job or init container. The application server should _verify_ the schema version matches its expectations, not _apply_ migrations.

---

## 13. The Azure AD Authenticator Does Not Rate-Limit JWKS Refreshes

**File:** `internal/auth/azuread/azuread.go`, lines 144-164

When a JWT presents an unknown `kid`, the authenticator refreshes the entire JWKS from Azure AD. An attacker can send JWTs with random `kid` values to force unlimited JWKS fetches, causing:
- Rate limiting from Azure AD (they do throttle the JWKS endpoint)
- Latency spikes for legitimate requests (all blocked on the mutex)

**Challenge:** There is no cooldown, no rate limit, no cache TTL. The `getKey()` method will happily call `refreshKeys()` on every single request with an unknown kid.

**Alternative:** Add a minimum refresh interval (e.g., 5 minutes). Cache the last refresh time. If a refresh was performed recently and the kid is still unknown, return an error immediately. This is what libraries like `go-jose` and `coreos/go-oidc` do.

---

## 14. `buildAuthenticator()` Lives in `cmd/pebblr/` -- Untestable

**File:** `cmd/pebblr/serve.go`, lines 219-264

`buildAuthenticator()` is a factory function with real I/O (reads secret files, makes HTTP calls to Azure AD). It lives in `package main`, which means it cannot be imported or unit-tested from another package.

**Challenge:** The auth provider selection logic (which provider, which secrets to read, how to configure each) is business logic that deserves tests. But because it is in `main`, the only way to test it is integration-level.

**Alternative:** Move `buildAuthenticator()` to `internal/auth` as a `NewFromConfig(ctx, cfg AuthConfig)` function. The `AuthConfig` struct can hold provider name, secret paths, etc. This makes it testable and reusable (e.g., for a future CLI tool that needs auth).

---

## 15. `os.Exit(1)` Inside a Goroutine -- Unclean Shutdown

**File:** `cmd/pebblr/serve.go`, lines 148-154

```go
go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        logger.Error("server error", "err", err)
        os.Exit(1)
    }
}()
```

`os.Exit(1)` bypasses all deferred functions -- the database pool `defer pool.Close()` on line 77 never runs. Open connections are leaked to the OS.

**Challenge:** This is a hard exit that skips cleanup. In a container context, this _might_ be acceptable (the container dies and connections are force-closed), but it is sloppy and makes local development debugging harder (connection pool exhaustion on repeated restarts).

**Alternative:** Send the error through a channel, or call `cancel()` on the signal context and let the main goroutine handle the exit.

---

## 16. Package Boundary Question: Should `auth/demo` Even Exist?

**Files:** `internal/auth/demo/`

The demo package contains two things:
1. An `Authenticator` (implements `auth.Authenticator`) -- this belongs in `auth/`.
2. An HTTP `Handler` (serves `/demo/accounts` and `/demo/token`) -- this is an HTTP handler, which conceptually belongs in `api/`.

**Challenge:** The demo handler imports `domain` and queries users -- it is closer to an API handler than an auth primitive. By placing it in `auth/demo`, you force the router to know about `auth/demo` specifically (the `DemoHandler *demo.Handler` field in `RouterConfig`). If the handler lived in `api/`, it would follow the same pattern as every other handler.

**Alternative:** Split: keep `demo.Authenticator` in `internal/auth/demo/`, move `demo.Handler` to `internal/api/demo_handler.go`. The handler becomes a regular API handler that depends on the demo authenticator.

---

## 17. No Request Body Size Limits

**All handler files**

Every handler uses `json.NewDecoder(r.Body).Decode(...)` without limiting the body size. An attacker can send a multi-gigabyte POST body and exhaust server memory.

**Challenge:** The JWKS fetch in `azuread.go` correctly uses `io.LimitReader(resp.Body, 1<<20)` -- so the pattern is known. Why is it not applied to API request bodies?

**Alternative:** Add `http.MaxBytesReader` middleware or apply it per-handler. A reasonable default for a CRM API is 1-10 MB.

---

## Summary: Priority Ranking

| # | Issue | Severity | Effort |
|---|-------|----------|--------|
| 7 | Auth middleware returns wrong Content-Type | **Bug** | Low |
| 5 | ClaimsBridge drops multi-role users silently | **Bug** | Medium |
| 15 | `os.Exit(1)` in goroutine skips cleanup | **Bug** | Low |
| 13 | No JWKS refresh rate limit | **Security** | Medium |
| 17 | No request body size limits | **Security** | Low |
| 6 | Static auth defaults to admin, no guardrail | **Security** | Low |
| 12 | Migrations in serve() block startup | **Architecture** | Medium |
| 14 | buildAuthenticator untestable in main | **Testability** | Medium |
| 8 | Three duplicate writeError implementations | **Consistency** | Low |
| 10 | Inconsistent response envelopes | **Consistency** | Medium |
| 9 | 30+ copies of actor-from-context boilerplate | **Maintainability** | Low |
| 11 | Mixed flags/env-vars for config | **Discoverability** | Low |
| 1 | serve() god function | **Maintainability** | Medium |
| 2 | RouterConfig struct grows forever | **Maintainability** | Medium |
| 3 | mountOrStub nil-guard machinery | **Over-engineering** | Low |
| 4 | Duplicate contextKey types | **Style** | Low |
| 16 | demo.Handler in wrong package | **Package design** | Low |
