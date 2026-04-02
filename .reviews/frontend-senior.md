# Frontend Code Review -- Senior Engineer Audit

Reviewer perspective: "The Senior Engineer Who Has Seen It All"
Scope: All TypeScript/React code in `web/src/`
Date: 2026-04-02

---

## CRITICAL -- Security

### S1. Access token stored in module-level global variable

**File:** `web/src/auth/provider.tsx` (line 8)

```ts
let _currentUser: AuthenticatedUser | null = null
```

The access token lives in a module-scoped mutable variable (`_currentUser.accessToken`).
Any code in the bundle -- including third-party dependencies -- can import or reach this
via the token provider closure. This is no worse than typical SPA memory storage, but it
also gets serialised into `sessionStorage` verbatim (line 22, `saveDemoSession`).

**Risk:** `sessionStorage` is accessible to any script on the same origin. A single XSS
vector (or a compromised npm package) can read the token from storage.

**Recommendation:**
- In demo mode it is acceptable, but add an explicit comment documenting the trade-off.
- In production OIDC mode (not yet wired), use `httpOnly` cookies managed by a BFF
  (backend-for-frontend) proxy so tokens never touch JS at all. Plan for this now.

### S2. No token expiry check before API calls

**File:** `web/src/api/client.ts` (line 14)

```ts
const token = getAccessToken?.()
```

The token provider returns whatever token is in memory. `AuthenticatedUser.expiresAt` is
stored but never checked. If the token expires mid-session, all API calls will start
returning 401s that surface as generic errors.

**Recommendation:** In `buildHeaders`, check `expiresAt` and trigger a silent refresh (or
redirect to sign-in) before attaching a stale token. At minimum, detect 401 responses in
the `request` function and clear the user session instead of showing an opaque error.

### S3. No CSRF protection on state-changing demo endpoints

**File:** `web/src/auth/provider.tsx` (line 71)

```ts
const response = await fetch('/demo/token', { ... })
```

`/demo/token` is a `POST` that returns a bearer token. If the demo proxy is ever exposed
beyond localhost, a cross-origin form post could trigger login on behalf of a victim.
The main API client sets `Content-Type: application/json` (which triggers CORS preflight),
but these raw `fetch` calls in the auth provider do the same. Acceptable for now, but
worth noting if the demo proxy is ever deployed to a shared environment.

### S4. `dangerouslySetInnerHTML` / XSS via user content

No instances of `dangerouslySetInnerHTML` found -- good. All user-supplied content
(target names, feedback text, notes, tags) is rendered as React children, which auto-
escapes HTML. The `str()` helper also prevents non-string values from leaking through.

**Status: PASS.**

---

## HIGH -- Error Handling & Resilience

### E1. No React Error Boundary

**Files:** `web/src/App.tsx`, `web/src/routes/__root.tsx`

There is no `ErrorBoundary` component anywhere in the tree. A single thrown render error
(e.g., `undefined` access on malformed API data, Google Maps SDK crash) will white-screen
the entire application.

**Recommendation:** Add an `ErrorBoundary` at the root (wrapping `<RouterProvider />`) and
ideally per-route. React Router has first-class support for route-level `errorComponent`.

### E2. Full-page Spinner with no error or timeout handling

**Files:** Every route component (`planner.tsx`, `targets.tsx`, `activities.tsx`, etc.)

Pattern:

```tsx
if (isLoading) return <Spinner />
```

If the API is down or the network is unreachable, `isLoading` stays `true` forever. The
user stares at a spinner with no feedback, no retry button, no timeout message.

**Recommendation:** Check `isError` / `error` from the query result and render an error
state. TanStack Query provides `isError`, `error`, and `refetch` -- use them:

```tsx
if (isLoading) return <Spinner />
if (isError) return <ErrorState error={error} onRetry={refetch} />
```

### E3. Mutations swallow errors silently

**Files:** `web/src/hooks/useActivities.ts`, `web/src/hooks/useAudit.ts`, etc.

Most mutation hooks only define `onSuccess` but no `onError`. When a mutation fails
(network error, 403 from RBAC, 422 validation), the user gets no feedback.

**Recommendation:** Add a global `onError` handler to the `MutationCache` on the
`QueryClient`, or add `onError` callbacks to each mutation call site. The toast system
already exists -- wire it up.

### E4. Demo account fetch silently swallows errors

**File:** `web/src/App.tsx` (line 71)

```ts
.catch(() => {})
```

If `/demo/accounts` fails, `accounts` stays empty and the user sees "Loading accounts..."
forever with no way to retry.

### E5. No 401 interception / session expiry handling

**File:** `web/src/api/client.ts`

When the API returns 401 (token expired or revoked), the error propagates as a generic
`ApiError`. There is no mechanism to clear the session, redirect to sign-in, or attempt
token refresh.

**Recommendation:** Add a response interceptor in the `request` function:

```ts
if (response.status === 401) {
  // Clear auth state, redirect to /sign-in
}
```

---

## HIGH -- Accessibility

### A1. Missing `aria-label` on icon-only buttons

**Files:** `web/src/routes/planner.tsx` (lines 494-502), `web/src/routes/reps.$id.tsx`,
`web/src/routes/dashboard.tsx`, `web/src/routes/targets.$id.tsx`

Week navigation chevron buttons, mobile map close button, and breadcrumb back buttons
have no accessible label:

```tsx
<button onClick={prevWeek} className="...">
  <ChevronLeft size={16} />
</button>
```

Screen readers will announce these as "button" with no context.

**Recommendation:** Add `aria-label="Previous week"`, `aria-label="Next week"`, etc.

### A2. Search inputs missing `<label>` associations

**Files:** `web/src/routes/planner.tsx` (line 319), `web/src/routes/targets.tsx` (line 138)

Search inputs use placeholder text but have no `<label>` element (even a visually-hidden
one). The `for`/`id` binding is missing.

### A3. Drag-and-drop has no keyboard alternative

**File:** `web/src/routes/planner.tsx`, `web/src/components/calendar/WeekView.tsx`

The core planning interaction (drag targets onto calendar days) is mouse-only. `draggable`
elements have `role="button"` and `tabIndex={0}`, which is good for click activation, but
the actual drag-to-day workflow has no keyboard equivalent.

**Recommendation:** The "Bulk Schedule" modal partially addresses this. Consider making
it more discoverable as the primary keyboard path, or add keyboard-driven reordering
using arrow keys.

### A4. Focus trap missing in Modal

**File:** `web/src/components/ui/Modal.tsx`

The modal traps scroll (`overflow: hidden`) and listens for Escape, but does not trap
keyboard focus. Tab will cycle through elements behind the modal overlay.

**Recommendation:** Implement focus trap (move focus into the modal on open, cycle
Tab/Shift+Tab within the modal, restore focus on close). Consider a small library
like `focus-trap-react` or a manual implementation.

### A5. Color-only status indicators

Priority dots (red/amber/grey) and status dots convey meaning through color alone, with
no text or icon alternative for colour-blind users. The text labels (`A`, `B`, `C`) are
sometimes present alongside the dots (good) but not always (e.g., map markers).

---

## MEDIUM -- Performance

### P1. Module-scoped `QueryClient` and `router` singletons

**File:** `web/src/App.tsx` (lines 43, 35)

`queryClient` and `router` are created at module scope. This means:
- They survive hot module replacement during development, which can cause stale state.
- In test environments, they leak state between tests.

**Recommendation:** Create them inside the `App` component using `useState` lazy
initializer, or use `useMemo` with an empty deps array. (TanStack docs recommend this
pattern.)

### P2. Large data fetches without pagination in Planner

**File:** `web/src/routes/planner.tsx` (line 81)

```ts
const { data: targetData } = useTargets({ limit: 500 })
```

Loading 500 targets in a single request, then filtering client-side. As the customer's
target portfolio grows, this will cause:
- Slow initial page load
- Large JSON parse blocking the main thread
- Large React reconciliation tree for the target list

**Recommendation:** Add server-side filtering (the backend supports `?q=` and `?type=`).
Consider virtual scrolling for the target list (500+ items).

### P3. `MapContainer` re-creates `APIProvider` on every render

**File:** `web/src/components/map/MapContainer.tsx`

Every `<MapContainer>` instance creates its own `<APIProvider>`. If two maps are on the
same page (e.g., coverage page has 1, target detail has 1), the Google Maps SDK is
initialized twice.

**Recommendation:** Lift `<APIProvider>` to `App.tsx` or the root layout, outside
individual map components.

### P4. Toast `<style>` tag injected on every render

**File:** `web/src/components/ui/Toast.tsx` (line 55)

The keyframe `<style>` block is re-injected into the DOM every time the toast list
changes. Move these keyframes into `global.css`.

### P5. Planner component is ~700 lines with 15+ `useState` calls

**File:** `web/src/routes/planner.tsx`

This is a maintainability concern that will become a performance concern. Each `useState`
setter triggers a re-render of the entire 700-line component. Consider extracting a
`usePlannerState` reducer or splitting into smaller components.

---

## MEDIUM -- Robustness

### R1. `useActivity` called with empty string when no ID selected

**File:** `web/src/components/activities/ActivityDetailModal.tsx` (line 17)

```ts
const { data: activity, isLoading, refetch } = useActivity(activityId ?? '')
```

When `activityId` is `null`, this calls `useActivity('')`. The hook has `enabled: !!id`,
so it won't fire a request for empty string, but it still creates a query cache entry
for key `['activities', 'detail', '']`. Harmless but sloppy.

### R2. `new Date(dateStr)` parsing is timezone-sensitive

**Files:** `web/src/lib/dates.ts`, multiple route files

`new Date('2026-04-02')` is parsed as UTC midnight, but `getDate()`, `getDay()` etc.
operate in local time. If the user is in a timezone behind UTC (e.g., UTC-5), the date
will roll back to April 1. The `formatDate` and `getMonday` functions use local time
consistently, which is correct as long as date strings from the API are always `YYYY-MM-DD`
(without time/zone). Document this assumption and/or normalize dates on receipt.

### R3. Sidebar role visibility logic has a gap

**File:** `web/src/layouts/Sidebar.tsx` (line 44)

```ts
const visibleItems = navItems.filter(
  (item) => item.roles.includes(role ?? '') || (role === 'admin' && !item.roles.includes('rep')),
)
```

When `role` is `null` (unauthenticated or still loading), `role ?? ''` yields `''`, which
matches nothing -- correct. But when `role` is `'admin'`, the second clause
`!item.roles.includes('rep')` excludes items tagged with `['rep']`. This means admins
cannot see Planner, Targets, or Activities -- which may be intentional, but it is surprising
and should be explicitly documented.

### R4. Audit page `columns` memo has stale closure over `handleReview`

**File:** `web/src/routes/audit.tsx` (lines 47-110)

```ts
const columns = useMemo(() => [...], [])
// eslint-disable-next-line react-hooks/exhaustive-deps
```

The columns definition captures `handleReview` in a closure, but the deps array is empty
with a lint suppression. If `updateStatus` ever changes identity (it will on unmount/
remount), the column action buttons will call a stale function. The eslint override
comment confirms this was known but not fixed.

**Recommendation:** Include `handleReview` in the deps array (or use a ref).

---

## LOW -- Code Quality

### Q1. Duplicated helper functions across routes

`getLat`, `getLng`, `getClassification` are copy-pasted in:
- `web/src/routes/planner.tsx`
- `web/src/routes/targets.tsx`
- `web/src/routes/coverage.tsx`
- `web/src/routes/reps.$id.tsx`
- `web/src/routes/targets.$id.tsx`

**Recommendation:** Extract into `web/src/lib/target-helpers.ts`.

### Q2. Hardcoded status strings mixed with config-driven values

Several components hardcode Romanian status names (`'realizat'`, `'anulat'`, `'planificat'`)
alongside English equivalents (`'completed'`, `'cancelled'`). These should come from the
tenant config exclusively, not be hardcoded in the UI.

### Q3. `sign-in.tsx` form elements are all `disabled` with no active auth flow

The sign-in page renders a Microsoft SSO button and email/password fields, all permanently
disabled. This is fine for a placeholder, but if someone navigates here in production
they get a dead page with no explanation.

### Q4. No TypeScript `readonly` on query key arrays

Query key factories use `as const` (good), but the returned arrays from factory functions
are not `readonly`. This is cosmetic -- TanStack Query handles it -- but `readonly` would
prevent accidental mutation.

---

## Summary

| Severity | Count | Key Items |
|----------|-------|-----------|
| Critical | 2 | Token in sessionStorage (S1), no expiry check (S2) |
| High     | 5 | No ErrorBoundary (E1), spinner-forever (E2), silent mutation errors (E3), no 401 handling (E5), missing aria-labels (A1) |
| Medium   | 5 | Module-scoped singletons (P1), 500-item fetch (P2), stale closure in audit (R4), timezone parsing (R2), role gap (R3) |
| Low      | 4 | Duplicated helpers (Q1), hardcoded statuses (Q2), dead sign-in (Q3), readonly keys (Q4) |

### Top 3 priorities for next sprint:

1. **Add ErrorBoundary + error/loading states** -- this is the difference between "it works in demos" and "it works in production." One malformed API response should not white-screen the app.

2. **Handle 401 responses globally** -- detect token expiry, clear session, redirect to sign-in. Without this, production users will get stuck on cryptic error states.

3. **Wire up mutation error feedback** -- the toast system exists and is good. Use it. Every failed save/create/delete should tell the user what happened.
