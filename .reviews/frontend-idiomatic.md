# Frontend Idiomatic Review -- The Idiomatic Obsessive

Reviewed: all files under `web/src/` (types, hooks, api, auth, routes, components, layouts, lib, i18n).

---

## Changes applied in this PR

### 1. Query key factories -- consistent hierarchy (FIXED)

Every TanStack Query key factory should follow the `all > lists/details > list(params)/detail(id)` convention from the [TanStack Query docs](https://tanstack.com/query/latest/docs/framework/react/guides/query-keys). This enables surgical invalidation -- you can invalidate all lists without touching cached detail pages, or nuke the entire entity scope.

**Before (multiple hooks):** flat keys like `['teams']` served double duty for both list fetches and the root invalidation scope, and detail keys like `['teams', id]` sat at the same level. `useMe` used a bare `['me']` literal with no factory at all.

**After:** all key factories now have `all`, `lists()`, `details()`, and `detail(id)` levels. `useMe` exports a `meKeys` factory. `useDashboard` now roots keys under a shared `all: ['dashboard']`. `useAudit` gained a `lists()` level.

Files changed:
- `hooks/useMe.ts`
- `hooks/useAudit.ts`
- `hooks/useTeams.ts`
- `hooks/useUsers.ts`
- `hooks/useTerritories.ts`
- `hooks/useDashboard.ts`

### 2. Duplicated target-field helpers extracted (FIXED)

`getLat`, `getLng`, `getClassification`, and `getCity` were copy-pasted across four route files. Extracted to `lib/target-fields.ts` and imported.

Files changed:
- `lib/target-fields.ts` (new)
- `routes/planner.tsx`
- `routes/coverage.tsx`
- `routes/targets.tsx`
- `routes/reps.$id.tsx`

### 3. Unsafe `as Role` cast replaced with type guard (FIXED)

In `auth/provider.tsx`, the demo login handler cast `data.account.role as Role` with no validation. If the server ever returns an unexpected string, this silently produces an invalid role that passes through the type system. Replaced with a `parseRole()` guard that validates and falls back to `'rep'` with a console warning.

Files changed:
- `auth/provider.tsx`

### 4. `eslint-disable react-hooks/exhaustive-deps` removed (FIXED)

In `routes/audit.tsx`, the `columns` memo suppressed the exhaustive-deps rule because it closed over `handleReview`. The fix was to inline the mutation call directly in the column cell renderer, then properly declare `updateStatus` in the dependency array.

Files changed:
- `routes/audit.tsx`

### 5. Unsound `undefined as T` in API client documented (FIXED)

In `api/client.ts`, the 204 handler did `return undefined as T`, which lies to the type system whenever `T` is not `void`. Changed to `undefined as unknown as T` with a doc comment explaining the contract: callers that expect data from a 204 have a logic bug.

Files changed:
- `api/client.ts`

### 6. `buildQueryString` utility added (ADDED)

A shared `lib/query-string.ts` helper that encapsulates the "build URLSearchParams, skip nullish values, append to base path" pattern duplicated across `useActivities`, `useAudit`, `useDashboard`, and `useTargets`. The hooks were not migrated in this PR to keep the diff focused, but the utility is ready for adoption.

Files added:
- `lib/query-string.ts`

---

## Remaining findings (not changed -- review items for follow-up)

### HIGH -- should be addressed soon

#### H1. Module-level mutable state in `api/client.ts`

```ts
let getAccessToken: (() => string | null) | null = null
```

This is module-scoped mutable state that is set via `setTokenProvider()`. It works but is invisible to React -- if the provider changes after first render, stale closures in in-flight requests will use the old function. Consider moving the token provider into React context and passing it to the API client via a query client default context, or at minimum, making the reference always point to a stable wrapper that reads from a ref.

#### H2. `DemoGate` in `App.tsx` uses bare `fetch` instead of TanStack Query

The demo account list is fetched with a raw `useEffect` + `fetch` + `useState` pattern. This means no caching, no error state, no retry, and a missing loading indicator. Wrap this in a `useQuery` hook (e.g., `useDemoAccounts`) to match the rest of the app's data-fetching discipline.

#### H3. `Toast.tsx` -- `ToastContainer` is a component created inside `useCallback`

```ts
const ToastContainer = useCallback(() => { ... }, [toasts])
```

This returns a new function reference every time `toasts` changes, which means React treats it as a new component type on every toast. This defeats reconciliation -- the entire toast DOM is unmounted and remounted, breaking CSS animations. Extract `ToastContainer` into a proper named component that receives `toasts` as a prop, or use a portal-based approach.

#### H4. `DataTable.tsx` column type uses `any`

```ts
columns: ColumnDef<T, any>[]
```

This is annotated with `eslint-disable @typescript-eslint/no-explicit-any`. TanStack Table's `ColumnDef` requires a second generic for the cell value type. When columns use `columnHelper.accessor(...)` the type is inferred, but the container prop must accept heterogeneous columns. The idiomatic escape hatch is:

```ts
// eslint-disable-next-line @typescript-eslint/no-explicit-any -- TanStack Table requires `any` for heterogeneous column defs
columns: ColumnDef<T, any>[]
```

This is actually the correct pattern for TanStack Table (they document it). But the disable comment should be on the specific line, not the interface.

#### H5. No error boundaries or error states on query hooks

Every route renders `if (isLoading) return <Spinner />` but never checks `isError` or `error`. A failed API call leaves the user staring at a spinner forever. Each route should handle the error case -- at minimum display the `ApiError.message`.

#### H6. `buildQuery` functions should use the new `buildQueryString` utility

The `buildQuery` function is copy-pasted across `useActivities.ts`, `useAudit.ts`, `useDashboard.ts`, and `useTargets.ts` with the same pattern. Now that `lib/query-string.ts` exists, these should be migrated in a follow-up.

### MEDIUM -- improve when convenient

#### M1. Discriminated union types for activity status

`Activity.status` is typed as `string`, but the domain has a finite set of statuses (`planificat`, `realizat`, `anulat`, etc.). These should be a string literal union:

```ts
export type ActivityStatus = 'planificat' | 'realizat' | 'anulat'
```

This would enable exhaustive switch checks and eliminate the `statusVariant[activity.status] ?? 'default'` fallback pattern.

Same applies to `Activity.activityType` -- should be a union, not `string`.

#### M2. `WeekView` prop drilling

`WeekView` passes 15+ props through to `DayColumn`. This is a classic case for a context provider or a compound component pattern. The drag state alone has 6 related props. Consider a `PlannerContext` that provides drag handlers and the target map.

#### M3. `MapContainer` creates a new `APIProvider` on every mount

Each `MapContainer` instance wraps its children in `<APIProvider apiKey={...}>`. If multiple maps are rendered (e.g., coverage + detail), each one initializes the Google Maps JS API independently. Lift the `APIProvider` to a layout-level provider (e.g., in `__root.tsx` or `AppShell`).

#### M4. Missing `onError` callbacks on mutations

Mutations like `useCreateActivity`, `useCloneWeek`, `useBatchCreateActivities` only define `onSuccess`. A failed mutation is silently swallowed. Add `onError` handlers that surface the error (e.g., via the toast system).

#### M5. `activityIcon` in `activities.tsx` uses a component type as value in a Record

```ts
const activityIcon: Record<string, typeof Stethoscope> = { ... }
```

This works but the type annotation is imprecise -- `typeof Stethoscope` is the specific Lucide component type. Use `React.ComponentType<LucideProps>` or `typeof LucideIcon` for clarity.

### LOW -- nice-to-have

#### L1. Inconsistent page size constants

`activities.tsx` uses `PAGE_SIZE = 20` and `limit: 200`, `audit.tsx` uses `limit: 50`, `targets.tsx` uses `limit: 200`, `planner.tsx` uses `limit: 500`. These should be co-located with the hooks or in a shared constants file.

#### L2. `i18n` keys only used in sidebar

The `useTranslation` hook is imported only in `Sidebar.tsx`. All other routes use hardcoded English strings. Either commit to i18n across the app or remove the i18n dependency to reduce bundle size.

#### L3. `Sidebar.tsx` role filtering logic is fragile

```ts
const visibleItems = navItems.filter(
  (item) => item.roles.includes(role ?? '') || (role === 'admin' && !item.roles.includes('rep')),
)
```

This special-cases admin to see non-rep items, but an admin cannot see rep-only items (planner, targets, activities). If admins should see everything, the logic should be `role === 'admin' || item.roles.includes(role)`. If the current behavior is intentional, add a comment.

#### L4. `sign-in.tsx` has a disabled email/password form

The sign-in page renders a non-functional email/password form alongside the Microsoft SSO button. This is confusing for users. Either remove it or gate it behind a feature flag.

#### L5. No `Suspense` boundaries

The app uses no React Suspense. With TanStack Query's `suspense: true` option and React 19's improved Suspense support, this would simplify the loading-state boilerplate across all routes.
