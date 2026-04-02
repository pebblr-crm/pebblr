# Devil's Advocate Review: `internal/domain/`, `internal/rbac/`, `internal/config/`

This review challenges every assumption and design decision I found. I am not being negative -- I am stress-testing the design so it earns its complexity. Every challenge comes with a concrete alternative.

---

## 1. `domain/role.go` -- The Permission System Nobody Uses

### Challenge

`Role.Permissions()` returns a `[]Permission` with 10 named constants. But **nothing in the entire codebase calls `Permissions()`** except the test that asserts admin > manager > rep. The RBAC enforcer (`policy.go`) never checks permissions -- it switches directly on the Role enum. So we have two parallel authorization models:

1. Role-based switch statements in `policy.go` (the one actually enforced)
2. A permission list in `role.go` (decorative)

These *will* drift. They already encode subtly different semantics: `PermAssignLeads` exists as a permission but no enforcer method checks for it. `PermAddNote` exists but there is no `CanAddNote` method.

### Why this matters

A future developer sees `Permission` constants and reasonably assumes they are authoritative. They write `if hasPermission(actor, PermAssignLeads)` and believe they have RBAC coverage. Meanwhile, the real access control lives in `canAccessTarget()` switch statements. This is a trap.

### Alternatives

**Option A (recommended): Delete `Permission` entirely.** The switch-on-role pattern in `policy.go` is simpler, correct, and already tested. If you need permission-based access later, add it *when you need it*, not as speculative infrastructure.

**Option B: Make permissions the single source of truth.** Rewrite the enforcer to actually check `actor.Role.Permissions()` against required permissions. But this adds indirection for no current benefit.

---

## 2. `rbac/rbac.go` -- The Enforcer Interface is Premature

### Challenge

`Enforcer` is an interface with 7 methods. There is exactly one implementation: `policyEnforcer`. No tests use a mock enforcer. No alternate implementation is planned (the PostgreSQL RLS layer is a separate defense-in-depth mechanism, not a Go `Enforcer` implementation).

Why is this an interface? The classic Go answer: "accept interfaces, return structs." But `NewEnforcer()` returns `Enforcer` (the interface), not `*policyEnforcer`. And the only consumer that needs to be swapped in tests is... nothing, because the enforcer itself is purely in-memory logic with no side effects.

### Why this matters

The interface forces every new entity type to update 3 places: the interface definition, the implementation, and the tests. Compare this to just having exported functions: `rbac.CanViewTarget(actor, target)`. Same testability, zero ceremony.

### Alternatives

**Option A (recommended): Export the policy functions directly.** `rbac.CanViewTarget(ctx, actor, target) bool`. No interface, no constructor, no struct. When (if!) you need a second implementation, introduce the interface then. Go interfaces are cheap to retrofit.

**Option B: Keep the interface but return the concrete type from `NewEnforcer()`.** At least follow "accept interfaces, return structs" properly.

---

## 3. `rbac/policy.go` -- The `context.Context` Argument is Ceremonial

### Challenge

Every enforcer method takes `context.Context` as its first argument. Every implementation ignores it (`_ context.Context`). There are no database calls, no timeouts, no tracing spans, no cancellation checks. The context is pure cargo cult.

"But we might need it later!" -- Then add it later. Go interfaces are satisfied implicitly; adding `context.Context` to a function signature is a non-breaking change in practice (callers already have a context available in HTTP handlers).

### Counter-argument I preemptively reject

"The context carries the user via `rbac.WithUser()`." True, but the enforcer methods *don't read it* -- they receive the actor as an explicit parameter. The context-user is read by handlers, not by the enforcer. So the context argument on enforcer methods serves no purpose.

---

## 4. `domain/activity.go` -- `PrepareForResponse()` is a Presentation Concern in a Domain Object

### Challenge

`PrepareForResponse()` mutates the struct to inject `JointVisitUID` back into the `Fields` map "so the frontend sees them as dynamic fields." This is a view-layer concern baked into the domain entity. The domain struct now knows about frontend rendering requirements.

### Why this matters

- The method mutates state, making the struct unreliable after the call (is `JointVisitUID` the source of truth, or `Fields["joint_visit_user_id"]`?).
- If you add more hoisted fields, you must remember to update this method. The `hoistedFields` map in `config/validator.go` and this method are coupled but live in different packages with no compile-time connection.

### Alternatives

**Option A (recommended): Move this to the API/handler layer.** Create a `toActivityResponse()` function in `internal/api/` that maps `domain.Activity` to a response DTO. The domain struct stays clean. The mapping lives where it belongs -- next to the HTTP serialization.

**Option B: Use a separate `ActivityResponse` struct.** Compose it from `Activity` and add the re-injected fields there.

---

## 5. `domain/activity.go` -- `ActivityPatch` and `ApplyTo()` -- Is the Domain Layer a PATCH Engine?

### Challenge

`ActivityPatch` with its `*string` optional fields and `FieldsPresent bool` flag is implementing HTTP PATCH semantics (RFC 7396 JSON Merge Patch) inside the domain layer. The domain package now knows about HTTP request parsing conventions.

`ApplyTo()` is 30 lines of nil-checking boilerplate that will grow linearly with every new field. This is mechanical code that a code generator could produce, but instead it is hand-maintained with a real risk of forgetting to add a new field.

### Why this matters

This pattern couples the domain to the API's choice of patch semantics. If you switch to JSON Patch (RFC 6902) or add GraphQL mutations, the domain must change for API-layer reasons.

### Alternatives

**Option A: Move `ActivityPatch` and `ApplyTo` to `internal/api/` or `internal/service/`.** The domain defines what an Activity *is*; the service layer defines how to partially update it.

**Option B: Use a generic map-based approach.** Let the service layer accept `map[string]any` and apply validated changes to the domain struct. Less type-safe but eliminates the growing boilerplate.

---

## 6. `domain/territory.go` -- `Boundary map[string]any` is a Time Bomb

### Challenge

GeoJSON stored as `map[string]any`. No validation, no type safety, no guarantee it is actually GeoJSON. Any caller can stuff `{"boundary": {"evil": true}}` in there and the system will happily persist it.

### Why this matters

GeoJSON has a well-defined structure. A `map[string]any` field says "I have no idea what goes here." At minimum, this should have a validation method. Better yet, use a proper GeoJSON type or at least `json.RawMessage` to defer parsing without pretending it is typed.

### Alternative

Use `json.RawMessage` for the boundary field. It is still flexible, but it clearly signals "this is opaque JSON" rather than "this is a Go map I will interact with programmatically." Add a `ValidateBoundary()` method when you need geographic operations.

---

## 7. `config/validator.go` -- `hoistedFields` is a Hidden Coupling

### Challenge

`hoistedFields` in `config/validator.go` is a package-level `map[string]bool` that lists field keys (`duration`, `account_id`, `routing`, `joint_visit_user_id`) that correspond to real Activity struct columns. But:

- The Activity struct in `domain/activity.go` defines these as Go struct fields.
- The database layer in `store/postgres/` maps them to SQL columns.
- The validator in `config/validator.go` hardcodes them as strings.

Three independent truth sources for the same concept. Add a new hoisted field and you must update all three by hand. There is no compile-time or test-time check that they agree.

### Alternative

Define the hoisted field keys as constants in `domain/` and reference them from both `config/` and `store/`. Or better: add a test that asserts the hoisted fields set matches a well-known list derived from the Activity struct's json tags.

---

## 8. `config/tenant.go` -- Linear Scans Everywhere

### Challenge

`AccountType()`, `ActivityType()`, `InitialStatus()`, `IsValidStatus()`, `IsValidTransition()`, `ResolveOptions()` -- every lookup is a linear scan through a slice. The config is loaded once at startup and never changes. Why not build index maps during `Load()`?

### Why this matters

With 11 activity types (the real config has 11), this is fine. But the pattern is setting a precedent. More importantly, every caller of `IsValidStatus()` in a hot path (e.g., activity creation) pays O(n) for something that should be O(1).

### Alternative

Add a `func (c *TenantConfig) buildIndexes()` called from `Load()` that creates `map[string]*ActivityTypeConfig`, `map[string]*AccountTypeConfig`, `map[string]*StatusDef` lookups. The exported methods become constant-time. The config stays immutable after load.

---

## 9. `config/schema.go` -- Why Two Validation Phases?

### Challenge

`LoadAndValidate()` runs JSON Schema validation first, then semantic validation. The comment says "if schema validation found structural issues, semantic validation will likely panic or produce confusing errors." This is defensive coding against your own validator code.

If your semantic validator panics on bad input, *fix the semantic validator*. A well-written `validateConfig()` should handle any syntactically valid JSON gracefully. The JSON Schema phase is doing the semantic validator's job of input validation.

### Counter-position

JSON Schema provides good user-facing error messages and is language-agnostic (the frontend could use the same schema). That is a legitimate reason. But the comment reveals the real motivation: the Go code is fragile. Fix the fragility; keep the schema for UX reasons.

---

## 10. `domain/` -- Every Struct Has JSON Tags but No Validation

### Challenge

Every domain struct is annotated with `json:"..."` tags. This is convenient for serialization, but it means domain structs are directly serialized as API responses. No DTOs, no response mapping, no ability to change the API contract independently of the domain model.

Today: rename `Activity.CreatorID` -> domain change = API breaking change.

### Why this matters

The domain package doc says "no dependencies on infrastructure (HTTP, database, etc.)." But JSON serialization *is* an infrastructure concern. The json tags make every domain struct an implicit API contract.

### Alternative

Remove json tags from domain structs. Create response types in `internal/api/` with their own tags. Yes, this is more code. But it cleanly separates the API contract from the domain model. When you add mobile clients or versioned APIs, you will need this separation anyway.

---

## 11. `rbac/policy.go` -- The "Deny All" Default is Fragile

### Challenge

```go
// Default: deny all.
return TargetScope{AssigneeIDs: []string{""}}
```

Denying by scoping to the empty string `""` is clever but fragile. It depends on no real user ever having an empty ID. If the database allows empty strings in `assignee_id` (it shouldn't, but defenses should not assume other defenses hold), this leaks data.

### Alternative

Add an explicit `DenyAll bool` field to `TargetScope` and `ActivityScope`. The store layer checks `if scope.DenyAll { return nil, nil }` before executing any query. The intent is clear and the safety does not depend on database constraints.

---

## 12. `domain/user.go` -- `OnlineStatus` is Premature for an Early-Stage CRM

### Challenge

You are building field sales CRM. Your users are reps visiting pharmacies. Is "online status" (online/away/offline) a core domain concept at this stage? This looks like UI chrome borrowed from the Twenty CRM inspiration, not a business requirement.

### Why this matters

Every type you add to the domain package becomes load-bearing. Other packages import it, database migrations create columns for it, the API serves it. Removing it later is harder than adding it later.

### Alternative

Remove `OnlineStatus` from the domain until there is a concrete feature (e.g., live dispatch, real-time manager dashboard) that requires it. If you need presence, consider it as a separate ephemeral concern (Redis/WebSocket), not a domain entity field.

---

## Summary: The Pattern I See

The codebase shows signs of **premature formalization** -- building infrastructure for flexibility that is not yet needed:

- Permission constants nobody checks
- An interface with one implementation
- Context arguments nobody reads
- JSON tags that couple domain to API
- Patch semantics in the domain layer
- Online status for a field sales CRM

The code is well-written, well-tested, and consistent. That makes these concerns more important, not less: well-executed premature abstraction is *harder* to remove than sloppy code, because it looks intentional and nobody wants to delete clean code.

**My overarching recommendation:** Strip the domain layer to the minimum that supports current features. Add abstractions when the second use case arrives, not when the first one *might* justify them.
