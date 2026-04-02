# Nitpicky Ninny Review: internal/api, internal/auth, cmd/pebblr

Reviewed: 2026-04-02
Scope: All Go files in `internal/api/`, `internal/auth/`, `cmd/pebblr/`

---

## FIXED in this PR

### 1. Hardcoded string literals vs constants (team, user, config handlers)

`team_handler.go`, `user_handler.go`, and `config_handler.go` used inline
`"Content-Type"`, `"application/json"`, `"missing authenticated user"`, and
`"an unexpected error occurred"` instead of the constants defined in
`constants.go` (`headerContentType`, `contentTypeJSON`, `errMissingUser`,
`errUnexpected`). Every other handler used the constants.

**Fixed:** All three files now use the shared constants.

### 2. Inconsistent JSON encoding: `writeJSON` vs bare `json.NewEncoder`

The codebase provides `writeJSON(w, r, v)` which logs encoding errors via the
request-scoped logger. Yet several handlers bypassed it with
`_ = json.NewEncoder(w).Encode(...)`, silently swallowing errors:

- `collection_handler.go`: Create, List, Get, Update (4 occurrences)
- `territory_handler.go`: List, Get, Create, Update (4 occurrences)
- `audit_handler.go`: List (1 occurrence)
- `target_handler.go`: FrequencyStatus (1 occurrence)
- `activity_handler.go`: CloneWeek (1 occurrence)

**Fixed:** All replaced with `writeJSON(w, r, ...)`.

### 3. Missing `w.WriteHeader(http.StatusOK)` on successful responses

Several handlers that used bare `json.NewEncoder(w).Encode(...)` also omitted
the explicit `w.WriteHeader(http.StatusOK)` call. While `net/http` implicitly
writes 200 on first `Write`, omitting it breaks the consistent pattern used by
target, activity, and dashboard handlers.

**Fixed:** Explicit `w.WriteHeader(http.StatusOK)` added to all affected paths.

### 4. Auth middleware: `http.Error` sets `Content-Type: text/plain`

`auth/middleware.go` and `auth/bridge.go` used `http.Error()` to return JSON
error bodies. `http.Error` sets `Content-Type: text/plain; charset=utf-8`, so
clients parsing `Content-Type` would not recognize these as JSON responses.

**Fixed:** Replaced with `writeJSONError()` helper that sets
`Content-Type: application/json` and writes the structured error envelope.

### 5. Struct field alignment in `RouterConfig`

`RouterConfig` had inconsistent field alignment: the first group used
single-space padding while the second group (ConfigHandler through WebDistPath)
used double-space padding for alignment. Same issue mirrored in `serve.go`.

**Fixed:** Normalized to standard `gofmt` alignment.

### 6. `fmt.Println("server stopped")` in serve.go

All other log messages in `serve()` use the structured `slog.Logger`. The
shutdown message at the end used bare `fmt.Println`, which bypasses structured
logging and would not appear in JSON log output.

**Fixed:** Changed to `logger.Info("server stopped")`.

---

## NOT FIXED -- Documented for future action

### 7. Inconsistent response envelopes across handlers

Single-item responses:
- Target, Activity: wrapped in `{"target": ...}` / `{"activity": ...}` (good)
- User: wrapped in `{"user": ...}` (good)
- Team detail: `{"team": ..., "members": [...]}` (good)
- Collection Get: returns raw object with NO envelope key (inconsistent)
- Territory Get: returns raw object with NO envelope key (inconsistent)

List responses:
- Target, Activity, Audit: `{"items":[], "total":N, "page":N, "limit":N}` (good, full pagination)
- Team, User: `{"items":[], "total":N}` (missing page/limit -- no pagination support)
- Collection List: `{"items":[]}` (missing total entirely)
- Territory List: `{"items":[], "total":N}` (missing page/limit)

**Recommendation:** Standardize all single-item responses to `{"<resource>": ...}`
and all list responses to `{"items":[], "total":N, "page":N, "limit":N}`.
The CLAUDE.md project spec requires pagination on all collection routes.

### 8. `writeError` cannot use request-scoped logger

`writeError(w, status, code, message)` falls back to `slog.Default()` for
encoding errors because it doesn't receive `*http.Request`. Meanwhile,
`writeJSON(w, r, v)` does receive the request and uses `LoggerFromContext`.
If `writeError` ever fails to encode, the error goes to the default logger
without request context (request_id, method, path).

**Recommendation:** Add `*http.Request` parameter to `writeError` for parity,
or accept the current trade-off since encoding a small error struct rarely fails.

### 9. `demo/handler.go` has its own `writeError` and constants

The demo handler package defines its own `headerContentType`, `contentTypeJSON`
constants and its own `writeError` function (lines 13-14, 133-139). The
`writeError` in demo uses `map[string]map[string]string` encoding instead of
the `errorResponse`/`errorDetail` struct used in the main API package.

While these are in separate packages (so no compilation conflict), the error
envelope structure differs: demo encodes `map[string]map[string]string` which
produces `{"error":{"code":"X","message":"Y"}}` -- structurally identical but
generated differently, so any future additions to the error envelope (e.g. a
`details` field) would need to be duplicated.

**Recommendation:** Consider extracting a shared `httperr` package or importing
the API error types.

### 10. Duplicate `contextKey` type definitions

Both `internal/api/middleware.go` and `internal/auth/middleware.go` define their
own `contextKey string` type. This is technically correct (separate packages,
separate key spaces), but could be confusing during maintenance.

**Status:** Acceptable -- no action needed.

### 11. `BatchCreate` in activity_handler.go leaks internal error messages

Line 576: `batchErrors = append(batchErrors, map[string]string{"targetId": item.TargetID, "error": err.Error()})`

This exposes raw `err.Error()` strings from the service layer to the client.
Internal errors (database timeouts, constraint violations) could leak
implementation details.

**Recommendation:** Map errors through `mapActivityServiceError`-style logic,
or return only the error code without the raw message.

### 12. Test coverage gaps

- `collection_handler.go`: No test file exists. Zero handler coverage.
- `meHandler` in `router.go`: Not tested directly (only indirectly via router test that doesn't inject a user).
- `target_handler.go`: `Assign` handler has no test for missing `teamId` (optional field behavior untested).
- `activity_handler.go`: `BatchCreate` has no tests at all.
- `activity_handler.go`: `CloneWeek` has no handler-level tests.
- `team_handler.go`: No test for service error propagation.
- `user_handler.go`: No test for service error propagation.
- `config_handler.go`: No test that verifies response Content-Type header.
- `audit_handler.go`: No test for `UpdateStatus` with missing/empty status field.
- Auth middleware: No test for malformed Authorization header (e.g., "Basic token" instead of "Bearer token").

### 13. Router comment says `/{id}` in stubs but activity stubs are incomplete

In `buildRouteSpecs`, the activity stubs list only
`[]string{"/", "/{id}", "/{id}/submit", "/{id}/status"}` but the actual
`NewActivityRouter` also registers `POST /batch`, `POST /clone-week`,
`PATCH /{id}`, `DELETE /{id}`, and `PUT /{id}`. The stubs only fire when the
handler is nil (feature not wired up), so this is low-impact, but the stub
paths don't match the real route surface.

### 14. `audit_handler.go` default limit differs from other handlers

Target and Activity handlers default to `limit = 20`. The Audit handler defaults
to `limit = 50`. No comment explains why. The Audit handler also caps limit at
200 while others have no cap.

**Recommendation:** Extract default/max pagination constants or document the
reasoning for per-handler differences.

### 15. Import ordering inconsistency in `activity_handler.go`

The import block includes `"github.com/pebblr/pebblr/internal/config"` which is
only used for the `config.FieldError` type in `validationErrorResponse`. This
couples the HTTP handler layer to the config package for a single type
reference. Consider whether `FieldError` belongs in a shared types package or
the domain layer.
