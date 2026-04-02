# Frontend Devil's Advocate Review

Reviewer perspective: challenge every abstraction, question every boundary, poke at every "obvious" decision.

---

## 1. Do We Actually Need TanStack Router?

`App.tsx` manually imports every route, builds the tree by hand, and registers the router type. There is no code-generation, no route-level data loading, no type-safe search params, no pending/error states at the route level. The only thing TanStack Router is doing here that React Router v7 or even a flat `switch` could not do is... nothing, really.

**The cost:**
- Manual route tree wiring in `App.tsx` (11 imports, manual `addChildren`)
- Every route file has boilerplate: `createRoute({ getParentRoute: () => rootRoute, path: '...', component: ... })`
- The `declare module` augmentation for type-safe links is declared but barely used -- most navigation uses `useNavigate()` with string paths, and `reps.$id.tsx` even uses a raw `<a href="/dashboard">` tag (line 83)
- No file-based routing is configured despite TanStack Router supporting it

**Recommendation:** Either lean into TanStack Router's strengths (file-based routing with codegen, route-level loaders/pendingComponent, type-safe `<Link>` everywhere) or drop it for something lighter. Right now it is ceremony without payoff.

---

## 2. The `planner.tsx` God Component

`planner.tsx` is **703 lines** in a single file. It manages:
- Week navigation state
- Target selection state (checkbox-like multi-select)
- Drag-and-drop state for 3 different drag sources (targets, activities, pending assignments)
- Day assignments (pending creation buffer)
- Target search + priority filter
- Map display (desktop + mobile overlay)
- Batch activity creation
- Bulk schedule modal
- Activity detail modal
- Clone week action
- Overdue A-priority target computation
- Toast notifications
- Stats/coverage display

That is at least 7 distinct concerns in one function. The `handleDrop` callback alone is 65 lines with 3 different code paths.

**Recommendation:** Extract at minimum:
- `usePlannerDrag` hook (drag state + drop handler)
- `usePlannerAssignments` hook (day assignments CRUD + batch create)
- `usePlannerWeek` hook (week navigation + date range computation)
- `TargetListPanel` component (search, filter, selection, the entire left panel)
- `BulkScheduleModal` component (it is already mostly self-contained in the JSX)

The planner is the core of this app. When bugs appear here, nobody will want to touch a 700-line function with 16 `useState` calls.

---

## 3. The `activities.tsx` Is Also a God Component

`activities.tsx` is **706 lines** with an inline `CreateActivityModal` (320 lines on its own) that handles dynamic form rendering, target search autocomplete, recovery balance display, and config-driven field generation.

`CreateActivityModal` deserves its own file under `components/activities/`. It is complex enough to need its own tests -- and currently has zero tests.

**The activities page itself** duplicates the week-grouping/day-grouping logic (`groupByWeekThenDay`, `getWeekKey`, `formatWeekLabel`, `getDayLabel`) that could live in `lib/dates.ts` and be tested independently.

---

## 4. Duplicated Helper Functions Everywhere

These functions appear identically in 4+ route files:

```ts
function getLat(fields: Record<string, unknown>): number | null { ... }
function getLng(fields: Record<string, unknown>): number | null { ... }
function getClassification(fields: Record<string, unknown>): string { ... }
```

Found in: `planner.tsx`, `targets.tsx`, `coverage.tsx`, `reps.$id.tsx`, `targets.$id.tsx` (inline variants).

This is a clear signal that `Target` field access needs a small utility module -- perhaps `lib/target-fields.ts`. The `Record<string, unknown>` bag is the root cause: there is no typed accessor layer for the dynamic `fields` object, so every consumer re-invents the same unsafe access pattern.

---

## 5. The `types/` Directory -- Co-locate or Commit

There are 9 separate type files in `types/`. Some are tightly coupled to a single hook (`dashboard.ts` maps 1:1 to `useDashboard.ts`; `audit.ts` maps 1:1 to `useAudit.ts`). Others are genuinely shared (`user.ts`, `api.ts`).

**The question:** Is the separation actually helping? When you change the dashboard API, you edit `types/dashboard.ts`, then `hooks/useDashboard.ts`, then `routes/dashboard.tsx`. That is 3 files for what is conceptually one feature.

Consider co-locating types with their hooks when they are not shared. `useActivities.ts` already defines `BatchCreateItem` and `BatchCreateResult` inline -- so the convention is already inconsistent.

**If you keep separate type files**, at least make it consistent: move `BatchCreateItem`/`BatchCreateResult` to `types/activity.ts`.

---

## 6. The Hook Abstraction: Is It Helping?

Every hook file follows the exact same pattern:
1. Define query keys
2. Define a `buildQuery` function for URL params
3. Export one `useX(params)` hook per endpoint

This is reasonable, but `buildQuery` is copy-pasted verbatim 4 times with slight field differences. A generic `buildQueryString(params: Record<string, string | number | undefined>)` utility would eliminate all of them.

Also: `useMe.ts` does not follow the key factory pattern that every other hook uses. It has a bare `['me']` key with no factory object. Minor inconsistency, but in a codebase that is otherwise quite disciplined, it stands out.

---

## 7. Auth Architecture: Module-Level Mutable State

`auth/provider.tsx` stores the current user in a **module-level mutable variable** (`let _currentUser`). This is not React state -- it is a global. The React state (`useState`) is synchronized manually with this global, creating two sources of truth.

The reason is clear: `setTokenProvider` in `api/client.ts` needs synchronous access to the token, and React state is not synchronously readable from outside React. But this creates a class of bugs where the global and the React state can diverge (e.g., if `setUser` throws, or during concurrent rendering).

**Recommendation:** Consider making the API client itself a dependency of `AuthProvider` via context, or use a ref for the token instead of a module-level `let`. At minimum, document why the global exists and what invariants must hold.

---

## 8. Hardcoded Romanian Status Keys

`lib/styles.ts` has:
```ts
export const statusVariant = { planificat: 'primary', realizat: 'success', anulat: 'danger' }
export const statusDot = { realizat: 'bg-emerald-500', planificat: 'bg-blue-500', anulat: 'bg-red-500', ... }
```

These are Romanian-language status keys baked into the frontend code. The backend config endpoint returns `statuses` with keys and labels -- the frontend should map by key from config, not hardcode a bilingual lookup table.

What happens when a second customer uses English status keys? Or when DrMax renames `planificat` to `programat`? This file becomes a lie.

The `audit.tsx` route also defines its own local `statusVariant` that shadows the one from `lib/styles.ts`. Two different status-to-variant maps in the same codebase.

---

## 9. `DemoGate` Uses Raw `fetch` Outside TanStack Query

`App.tsx` line 66-72: `DemoGate` fetches `/demo/accounts` with a raw `fetch` call inside a `useEffect`. The CLAUDE.md says "use TanStack Query for all API data fetching." This bypasses the cache, has no error retry, no loading state management, and the error handler is `catch(() => {})`.

This is a demo feature, but demo code has a way of becoming production code. Either use a query hook or at minimum use the `api` client.

---

## 10. `sign-in.tsx`: Dead Code That Violates Auth Policy

The sign-in page renders a username/password form (lines 68-89) that is entirely `disabled`. CLAUDE.md says "No local username/password auth." This form trains users to expect password-based login and will confuse anyone looking at the code.

The Microsoft sign-in button is also `disabled`. This page is a static mockup that does nothing. If Azure AD OIDC is the auth path, the sign-in flow should redirect to the IdP, not render a fake form.

---

## 11. i18n Is Only Used in the Sidebar

`useTranslation()` is imported in exactly one place: `Sidebar.tsx`. Every other component uses hardcoded English strings ("Activity Log", "Target Portfolio", "Bulk Schedule", "Clone Week", etc.).

This is not i18n -- it is the appearance of i18n. The `en.ts` and `ro.ts` translation files define keys for nav items and a few common actions, but none of the actual page content is translated.

**Either commit to i18n** (wrap all user-visible strings) **or remove the dependency** and save 3 dependencies (`i18next`, `i18next-browser-languagedetector`, `react-i18next`) plus the translation files. Half-implemented i18n is worse than no i18n because it creates the false impression that the app supports multiple languages.

---

## 12. No Error Boundaries

There is no `ErrorBoundary` component anywhere. Every route renders `<Spinner />` for loading and has no error handling UI. If any query fails, the page either shows a spinner forever or crashes.

TanStack Router supports `errorComponent` per route. TanStack Query supports `error` state. Neither is used.

---

## 13. The `DataTable` Pagination Is Client-Side Only

`DataTable` uses `getPaginationRowModel()` for client-side pagination. But the API supports server-side pagination (`?page=&limit=`). Most hooks pass `limit: 200` or `limit: 500`, fetching everything and paginating in the browser.

For the current scale (DrMax Romania, likely <1000 targets), this works. But the architecture is set up for server-side pagination (the API returns `total`, `page`, `limit`) and then ignores it. The `audit.tsx` route is the only one that actually implements server-side pagination with a `page` state variable -- and it still passes the data to `DataTable` which does its own client-side pagination on top.

---

## 14. Accessibility Gaps

- `Modal.tsx`: The overlay `div` has `onClick={onClose}` but no `role`, no `aria-label`. The modal itself has no `role="dialog"` or `aria-modal="true"`.
- `WeekView.tsx` / `DayColumn`: Drop zones have no ARIA attributes for drag-and-drop operations. Screen readers get nothing.
- `Sidebar.tsx`: Active nav item styling is visual-only, no `aria-current="page"`.
- Several `<button>` elements have only icon children with no accessible label (e.g., the priority filter buttons in planner, the pagination chevrons in dashboard).
- `Toast.tsx`: Toasts have no `role="alert"` or `aria-live` region.

---

## 15. `MapContainer` Creates a New `APIProvider` Per Instance

Every `MapContainer` wraps its children in `<APIProvider apiKey={...}>`. If a page renders 2 maps (e.g., `planner.tsx` renders both a desktop and mobile map), that is 2 `APIProvider` instances. The Google Maps JS API should be loaded once at the app level.

---

## 16. Tests Are Sparse for the Complexity

Coverage by area:
- `lib/` -- well tested (dates, helpers, styles all have tests)
- `components/ui/` -- have test files but I only saw Badge, Button, Card, EmptyState, Modal, Spinner, Toast test files (not all read but they exist)
- `components/data/DataTable` -- good test coverage
- `components/calendar/WeekView` -- good test coverage
- **Routes** -- only `planner.test.tsx` exists. Zero tests for activities, dashboard, coverage, console, audit, targets, target detail, rep drill-down.
- **Hooks** -- zero tests for any hook.
- `auth/` -- one test (useAuth throws outside provider). No tests for the provider logic, demo login flow, session persistence.

The most complex code (the routes, especially `activities.tsx` with its `CreateActivityModal`) has the least test coverage.

---

## 17. `motion` (Framer Motion) Is a Dependency With Zero Usage

`package.json` lists `motion: ^12.23.24` (Framer Motion). A search of all source files shows zero imports of `motion` or `framer-motion`. This is 65KB+ of dead weight in the bundle.

---

## Summary of Recommended Actions

| Priority | Issue | Effort |
|----------|-------|--------|
| High | Extract planner.tsx into sub-components + hooks | L |
| High | Extract CreateActivityModal from activities.tsx | M |
| High | Deduplicate getLat/getLng/getClassification into lib/target-fields.ts | S |
| High | Add error boundaries to routes | M |
| Medium | Decide on TanStack Router: use file-based routing or downgrade | M |
| Medium | Remove hardcoded Romanian status keys; derive from config | M |
| Medium | Commit to i18n or remove it | S |
| Medium | Fix raw fetch in DemoGate | S |
| Medium | Add accessibility attributes to Modal, Toast, nav | M |
| Medium | Remove dead sign-in form / motion dependency | S |
| Low | Co-locate or consistently separate types | S |
| Low | Generic buildQueryString utility | S |
| Low | Hoist APIProvider to app level | S |
| Low | Add tests for hooks and remaining routes | L |
