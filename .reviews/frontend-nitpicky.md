# Frontend Nitpicky Review

Reviewer: The Nitpicky Ninny
Scope: `web/src/` -- every .ts, .tsx, .css file
Date: 2026-04-02

---

## 1. Duplicated Utility Functions Across Files

**Severity: High (DRY violation)**

The functions `getLat`, `getLng`, and `getClassification` are copy-pasted across FOUR files with identical implementations:

| Function | File |
|---|---|
| `getLat` | `routes/planner.tsx`, `routes/coverage.tsx`, `routes/targets.tsx`, `routes/reps.$id.tsx` |
| `getLng` | `routes/planner.tsx`, `routes/coverage.tsx`, `routes/targets.tsx`, `routes/reps.$id.tsx` |
| `getClassification` | `routes/planner.tsx`, `routes/coverage.tsx`, `routes/targets.tsx`, `routes/reps.$id.tsx` |
| `getCity` | `routes/targets.tsx` (only place, fine) |

Similarly, `getTargetPriority` exists in both `components/calendar/WeekView.tsx` and `routes/activities.tsx` with slightly different signatures.

**Recommendation:** Extract these to `lib/helpers.ts` or a new `lib/target-fields.ts`.

---

## 2. Inconsistent Query Key Factory Patterns

**Severity: Medium (Inconsistent API across hooks)**

Most hook files define a `*Keys` factory object at the top, but they follow two different structures:

**Pattern A (with `lists()` intermediary):**
- `useActivities.ts`: `activityKeys.all -> lists() -> list(params) -> detail(id)`
- `useTargets.ts`: `targetKeys.all -> lists() -> list(params) -> detail(id) -> visitStatus() -> frequencyStatus()`

**Pattern B (flat, no `lists()` intermediary):**
- `useAudit.ts`: `auditKeys.all -> list(params)` -- no `lists()` level
- `useTeams.ts`: `teamKeys.all -> detail(id)` -- no `lists()` level
- `useTerritories.ts`: `territoryKeys.all -> detail(id)` -- no `lists()` level
- `useUsers.ts`: `userKeys.all -> detail(id)` -- no `lists()` level

**Pattern C (no shared `all` ancestor):**
- `useDashboard.ts`: `dashboardKeys.activities(f) / coverage(f) / frequency(f) / recovery(f)` -- each is independent, no shared `all` key

**Pattern D (inline literal, no factory):**
- `useMe.ts`: `queryKey: ['me']` -- no factory at all, just an inline array

**Recommendation:** Pick ONE pattern and apply it everywhere. For hooks that list AND detail, use Pattern A. For hooks that only list, use Pattern A without `detail`. For `useMe`, at least define `const meKeys = { all: ['me'] as const }` for consistency. For `useDashboard`, add an `all: ['dashboard'] as const` ancestor so invalidation can target the whole domain.

---

## 3. Inconsistent `buildQuery` Function Naming

**Severity: Low**

Three hooks define a local `buildQuery` function: `useActivities.ts`, `useAudit.ts`, `useTargets.ts`. Meanwhile, `useDashboard.ts` also defines `buildQuery` but with a *different* signature -- it takes `(base: string, filter: DashboardFilter)` instead of just params.

These are not exported so there is no collision, but the inconsistent signatures are confusing. Consider either:
- Extracting a shared `buildQueryString(params: Record<string, string | undefined>): string` in `lib/helpers.ts`
- Or at minimum documenting the pattern

---

## 4. Inconsistent `React.ReactNode` vs Imported `ReactNode`

**Severity: Low (Style inconsistency)**

Some files use `React.ReactNode` (implicit global React namespace):
- `components/ui/Badge.tsx` -- `children: React.ReactNode`
- `components/ui/Card.tsx` -- `children: React.ReactNode`
- `components/ui/EmptyState.tsx` -- `icon?: React.ReactNode`, `action?: React.ReactNode`
- `layouts/Sidebar.tsx` -- `icon: React.ReactNode`
- `routes/console.tsx` -- `icon: React.ReactNode`

Other files explicitly import `ReactNode`:
- `components/ui/Modal.tsx` -- `import { useEffect, type ReactNode } from 'react'`
- `layouts/AppShell.tsx` -- `import { useState, type ReactNode } from 'react'`
- `auth/provider.tsx` -- `import { useState, useEffect, useCallback, type ReactNode } from 'react'`

**Recommendation:** Pick ONE style. The explicit import (`type ReactNode`) is more precise and doesn't rely on the global JSX namespace. Use it everywhere.

---

## 5. Inconsistent Import Ordering

**Severity: Low (Style)**

There is no enforced import order convention. Files mix these groups inconsistently:

Expected order should be:
1. React / React types
2. Third-party libraries (`@tanstack/*`, `lucide-react`, `react-i18next`)
3. Internal aliases (`@/hooks/*`, `@/components/*`, `@/lib/*`, `@/types/*`, `@/auth/*`, `@/api/*`)
4. Relative imports (`./`, `../`)
5. Type-only imports (using `import type`)

**Violations:**
- `main.tsx` uses relative imports (`./App`, `./styles/global.css`) while the rest of the codebase uses `@/` aliases. Should be `@/App` and `@/styles/global.css`.
- `auth/context.test.ts` uses relative import `./context` while other test files (like `api/client.test.ts`) use `@/types/api`. Mixed conventions.
- Some files have type imports mixed with value imports on the same line (fine via `import type` or `import { type X }`), others separate them entirely.

---

## 6. `statusVariant` Name Collision

**Severity: Medium (Confusing)**

There is a `statusVariant` object exported from `lib/styles.ts`:
```ts
export const statusVariant: Record<string, 'primary' | 'success' | 'danger' | 'default'>
```

And ANOTHER `statusVariant` declared locally inside `routes/audit.tsx`:
```ts
const statusVariant: Record<string, 'warning' | 'success' | 'default'>
```

These have **different type signatures** and **different key-value mappings**. The audit page's version maps `pending -> warning`, while the global one maps Romanian status names. This is a naming collision waiting to happen if someone refactors imports.

**Recommendation:** Rename the audit-local one to `auditStatusVariant`.

---

## 7. Inconsistent Component Export Styles

**Severity: Low**

All components use named exports (`export function Badge(...)`) -- this is consistent and good.

However, the `useToast` hook in `components/ui/Toast.tsx` is a hook file placed in the UI components directory. It returns both a hook function and a component (`ToastContainer`). This is the only "hook masquerading as a component file" in the `components/` directory.

**Recommendation:** Either move it to `hooks/useToast.ts` and re-export a `ToastContainer` component, or rename the file to `useToast.tsx` to signal it is a hook.

---

## 8. Missing Test Files

**Severity: Medium (TDD violation per CLAUDE.md)**

The following files have **no corresponding test files**:

| File | Missing test |
|---|---|
| `hooks/useActivities.ts` | No test |
| `hooks/useAudit.ts` | No test |
| `hooks/useConfig.ts` | No test |
| `hooks/useDashboard.ts` | No test |
| `hooks/useMe.ts` | No test |
| `hooks/useTargets.ts` | No test |
| `hooks/useTeams.ts` | No test |
| `hooks/useTerritories.ts` | No test |
| `hooks/useUsers.ts` | No test |
| `layouts/AppShell.tsx` | No test |
| `layouts/Sidebar.tsx` | No test |
| `components/activities/ActivityDetailModal.tsx` | No test |
| `components/map/MapContainer.tsx` | No test |
| `components/map/TargetMarker.tsx` | No test |
| `routes/activities.tsx` | No test |
| `routes/audit.tsx` | No test |
| `routes/console.tsx` | No test |
| `routes/coverage.tsx` | No test |
| `routes/dashboard.tsx` | No test |
| `routes/index.tsx` | No test |
| `routes/reps.$id.tsx` | No test |
| `routes/sign-in.tsx` | No test |
| `routes/targets.tsx` | No test |
| `routes/targets.$id.tsx` | No test |

That is 24 untested files vs 13 tested files. CLAUDE.md says "Write tests before or alongside implementation. Do not merge untested code."

---

## 9. Interface Props Naming Inconsistency

**Severity: Low**

Some components define their props interface with a `*Props` suffix, some do not:

| File | Pattern |
|---|---|
| `Badge.tsx` | `BadgeProps` |
| `Button.tsx` | `ButtonProps` |
| `Card.tsx` | `CardProps` |
| `EmptyState.tsx` | `EmptyStateProps` |
| `Modal.tsx` | `ModalProps` |
| `Spinner.tsx` | Inline `{ label?: string }` -- no named interface at all |
| `StatCard.tsx` | `StatCardProps` |
| `WeekView.tsx` | `WeekViewProps` |
| `DayColumn` (in WeekView) | `DayColumnProps` |
| `ActivityDetailModal.tsx` | `ActivityDetailModalProps` |
| `MapContainer.tsx` | Inline (would need to verify) |
| `TargetMarker.tsx` | Inline (would need to verify) |
| `AppShell.tsx` | `AppShellProps` |
| `Sidebar.tsx` | Inline `{ currentPath: string; onNavigate?: () => void }` |

**Recommendation:** `Spinner` and `Sidebar` should have named props interfaces for consistency.

---

## 10. CSS Class String Construction Inconsistency

**Severity: Low**

The codebase uses three different patterns for conditional CSS classes:

**Pattern A: Template literals with ternary**
```ts
className={`base-classes ${condition ? 'active' : 'inactive'}`}
```
Used by: Most files

**Pattern B: Array join**
Not used (good -- one less pattern)

**Pattern C: Inline object record + lookup**
```ts
const variants = { primary: '...', secondary: '...' } as const
className={`${variants[variant]} ${sizes[size]} ${className}`}
```
Used by: `Badge.tsx`, `Button.tsx`

This is fine, but there is no `cn()` or `clsx()` utility. Empty strings from `className = ''` defaults will produce double spaces in class strings. Not a bug, but sloppy. Consider adopting `clsx` or a `cn()` utility.

---

## 11. `inputCls` / `selectCls` Inconsistency

**Severity: Low**

In `routes/activities.tsx`, there are TWO different class constant strings for form inputs:

```ts
const inputCls = 'w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500'
const selectCls = 'w-full text-sm border border-slate-300 rounded-md py-2 px-3 bg-white focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none'
```

Note the differences:
- `inputCls` uses `rounded-lg`, `selectCls` uses `rounded-md`
- Property ordering differs (`border` before `text-sm` vs after)
- `selectCls` adds `bg-white`, `inputCls` does not

Meanwhile, `routes/sign-in.tsx` and `routes/targets.$id.tsx` inline their own input styles that differ again (`py-2.5` instead of `py-2`).

**Recommendation:** Extract a shared input/select style constant to `lib/styles.ts` or create proper `Input`/`Select` UI components.

---

## 12. Hardcoded Locale Strings

**Severity: Medium (i18n gap)**

The i18n setup exists (`i18n/en.ts`, `i18n/ro.ts`) and is used in `Sidebar.tsx` via `useTranslation()`. However, MOST strings in the application are hardcoded English:

- `routes/activities.tsx`: "Activity Log", "Review your submitted visits...", "Log Activity", "Action Required", "Load more", "All Types", "All", etc.
- `routes/dashboard.tsx`: "Team Dashboard", "Classification", "Compliance", etc.
- `routes/audit.tsx`: "Audit Logs", "Immutable change history...", "Export Logs", etc.
- `routes/console.tsx`: "Configuration", "Users & Roles", "Business Rules", etc.
- `routes/sign-in.tsx`: "Welcome back", "Sign in with your organization account"
- `routes/planner.tsx`: "Search targets...", "Clone Week", "Bulk Schedule", etc.
- `routes/targets.tsx`: "Target Portfolio", "Search targets...", "More filters"
- `routes/targets.$id.tsx`: "Target Details", "Visit History", "Schedule Visit"
- `components/activities/ActivityDetailModal.tsx`: "Activity", "Promoted Products", "Feedback", "Submit"
- `components/ui/Spinner.tsx`: "Loading..."

Only `Sidebar.tsx` actually calls `t()`. The i18n system is set up but barely utilized.

---

## 13. Type File Convention Inconsistency

**Severity: Low**

Most type files export only interfaces and type aliases. But `types/api.ts` also exports a **class** (`ApiError`). This is an implementation detail living in a types file.

**Recommendation:** Move `ApiError` to `api/errors.ts` or `lib/errors.ts` and re-export the type from `types/api.ts` if needed.

---

## 14. `collection.ts` Type File Is Unused

**Severity: Low**

`types/collection.ts` defines a `Collection` interface but it is not imported anywhere in the codebase. Dead code.

---

## 15. Inconsistent Error Handling in `App.tsx`

**Severity: Medium**

The `DemoGate` component fetches demo accounts with raw `fetch()` instead of using TanStack Query:

```ts
fetch('/demo/accounts')
  .then((r) => { ... })
  .catch(() => {})
```

This violates the project convention in CLAUDE.md: "Use TanStack Query for server state; avoid ad-hoc fetch calls outside of query/mutation hooks." The `auth/provider.tsx` also uses raw `fetch()` for `/demo/token`, but that one is in the auth layer before QueryClientProvider is available, so it gets a pass.

However, `DemoGate` renders *inside* `QueryClientProvider`, so it could and should use a query hook.

---

## 16. Unused Imports

**Severity: Low**

- `routes/targets.$id.tsx`: Imports `useCreateActivity` from `@/hooks/useActivities` but the local `ScheduleVisitModal` already calls `useCreateActivity` within itself. The outer component does NOT use it directly. (Actually, the import is used inside `ScheduleVisitModal` which is in the same file -- so this is fine. Withdrawing this item.)

---

## 17. Blank Line Inconsistency

**Severity: Pedantic (but I am the Nitpicky Ninny)**

- `routes/activities.tsx` line 47: Double blank line between route export and `getWeekKey`
- `routes/planner.tsx` line 54: Double blank line between `getClassification` and `PlannerPage`
- `routes/planner.tsx` line 296: Double blank line before the loading check

Most files use single blank lines for separation. Double blank lines should be removed.

---

## 18. `eslint-disable` Comment

**Severity: Low**

`routes/audit.tsx` line 834 contains:
```ts
// eslint-disable-next-line react-hooks/exhaustive-deps
```

This suppresses the exhaustive-deps rule for `useMemo` with an empty dependency array, even though `handleReview` is used inside the memo. This is a legitimate concern -- `handleReview` captures `updateStatus` which changes on each render. The disable is intentional but should have a comment explaining *why*.

---

## 19. `v2` Badge in AppShell

**Severity: Observation**

`layouts/AppShell.tsx` renders a "v2" badge next to the Pebblr logo in the mobile header. This is presumably a holdover from a migration. Is this intentional for production?

---

## Summary

| Category | Count |
|---|---|
| DRY violations (duplicated code) | 1 major (3 functions x 4 files) |
| Inconsistent patterns | 8 (query keys, imports, ReactNode, props, CSS, inputCls, buildQuery, statusVariant) |
| Missing tests | 24 files |
| i18n gaps | Nearly all user-facing strings hardcoded |
| Dead code | 1 file (`types/collection.ts`) |
| Convention violations | 2 (raw fetch in QueryClient scope, ApiError class in types file) |
| Style nits | 3 (double blank lines, eslint-disable without comment, v2 badge) |

Total findings: 39 individual items across 12 categories.
