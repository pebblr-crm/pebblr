# Frontend Clean Code Review

Reviewer persona: **The Clean Code Geek** -- obsessed with SOLID, meaningful naming, small components, single responsibility, DRY, clean abstractions.

Scope: Every `.ts` / `.tsx` file under `web/src/`.

---

## Executive Summary

The codebase is well-structured at the macro level: hooks wrap TanStack Query correctly, types are cleanly separated, UI primitives are small and reusable, and the `lib/` utilities are focused. However, several route files have grown into God Components that violate Single Responsibility, and there are DRY violations across multiple files where identical helper functions are copy-pasted rather than shared.

Overall health: **B-** -- solid foundation with focused areas that need extraction and deduplication.

---

## Critical Findings

### 1. `routes/planner.tsx` is a God Component (700+ lines, ~15 pieces of state)

**Severity: High**
**Principle violated: Single Responsibility, Open/Closed**

`PlannerPage` manages: week navigation, target selection, drag-and-drop state, day assignments, batch creation, bulk scheduling modal, mobile map overlay, search/filter state, toast notifications, map rendering, and calendar rendering. This is at least 5 distinct responsibilities in one function.

**Recommendation -- extract into focused custom hooks:**

```
usePlannerWeek()          -- weekStart, prevWeek, nextWeek, goToday, dateFrom, dateTo
usePlannerDragDrop()      -- dragTargetId, dragActivityId, dragPending, isDragging, handleDrop
usePlannerAssignments()   -- dayAssignments, addToDay, removeFromDay, totalAssigned, handleCreateActivities
usePlannerSelection()     -- selectedTargetIds, toggleTarget, clearSelection
```

Extract the **Bulk Schedule Modal** into its own file `components/planner/BulkScheduleModal.tsx`.
Extract the **Mobile Map Overlay** into `components/planner/MobileMapOverlay.tsx`.
Extract the **Target Sidebar Panel** into `components/planner/TargetSidebar.tsx`.

### 2. `routes/activities.tsx` is another God Component (705 lines, two components in one file)

**Severity: High**
**Principle violated: Single Responsibility, file colocation**

This file contains both `CreateActivityModal` (a complex 400-line form component) and `ActivitiesPage` (a 300-line timeline page). `CreateActivityModal` alone has enough state and logic to warrant its own file and possibly its own hook.

**Recommendation:**

- Extract `CreateActivityModal` to `components/activities/CreateActivityModal.tsx`.
- Extract `renderActivityCard` to `components/activities/ActivityCard.tsx` (it is a full component masquerading as a render function).
- Extract timeline grouping helpers (`getWeekKey`, `groupByWeekThenDay`, `formatWeekLabel`, `getDayLabel`) to `lib/timeline.ts`.

### 3. DRY Violation: `getLat`, `getLng`, `getClassification` duplicated 4 times

**Severity: High**
**Principle violated: DRY**

These three field-accessor functions are copy-pasted identically across:

- `routes/planner.tsx` (lines 40-52)
- `routes/targets.tsx` (lines 24-40)
- `routes/coverage.tsx` (lines 21-33)
- `routes/reps.$id.tsx` (lines 23-35)

**Recommendation:** Move them to `lib/target-fields.ts`:

```ts
export function getLat(fields: Record<string, unknown>): number | null { ... }
export function getLng(fields: Record<string, unknown>): number | null { ... }
export function getClassification(fields: Record<string, unknown>): string { ... }
export function getCity(fields: Record<string, unknown>): string { ... }
```

### 4. DRY Violation: `buildQuery` pattern repeated across hooks

**Severity: Medium**
**Principle violated: DRY**

Every hook file (`useActivities.ts`, `useAudit.ts`, `useDashboard.ts`, `useTargets.ts`) has its own `buildQuery` function that does the same thing: iterate over params, set them on `URLSearchParams`, join with a base path.

**Recommendation:** Extract a generic utility to `lib/query.ts`:

```ts
export function buildQueryString(base: string, params: Record<string, string | number | undefined>): string {
  const qs = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') qs.set(key, String(value))
  }
  const q = qs.toString()
  return q ? `${base}?${q}` : base
}
```

---

## Moderate Findings

### 5. `routes/dashboard.tsx` -- `MonthGrid` is 120+ lines co-located in the route file

**Severity: Medium**
**Principle violated: Single Responsibility, file size**

`MonthGrid` is a self-contained calendar grid component with its own props interface. It should be extracted to `components/calendar/MonthGrid.tsx` alongside `WeekView.tsx`.

### 6. `targets.$id.tsx` -- `ScheduleVisitModal` co-located in route file

**Severity: Medium**
**Principle violated: Single Responsibility**

Similar to #5. This is a modal with its own state. Extract to `components/targets/ScheduleVisitModal.tsx`.

### 7. `visit_type` badge styling is inlined in 3 places despite `visitTypeBadge()` existing in `lib/styles.ts`

**Severity: Medium**
**Principle violated: DRY**

`lib/styles.ts` exports `visitTypeBadge()` but it is not used anywhere. Instead, `ActivityDetailModal.tsx` (line 194-199), `targets.$id.tsx` (line 219-224), and `activities.tsx` (line 499-502) all inline the same conditional class logic for `f2f` vs `remote` badges.

**Recommendation:** Use the existing `visitTypeBadge()` utility in all three places.

### 8. `WeekView` has too many props (19 props) -- a "prop explosion" smell

**Severity: Medium**
**Principle violated: Interface Segregation**

`WeekViewProps` takes 19 optional callback/state props. This is a sign the component is trying to serve too many callers with too many optional features. When used from `dashboard.tsx` and `reps.$id.tsx`, most of these props are unused.

**Recommendation:** Consider a compound component pattern or split into `ReadOnlyWeekView` (for dashboard/rep drill-down) and `EditableWeekView` (for planner) that extends it. Alternatively, group related props into configuration objects:

```ts
interface DragConfig {
  isDragging: boolean
  draggingActivityId?: string | null
  draggingPending?: { sourceDate: string; targetId: string } | null
  onDrop?: (dateStr: string) => void
  onActivityDragStart?: (activityId: string) => void
  onActivityDragEnd?: () => void
  onPendingDragStart?: (sourceDate: string, targetId: string) => void
  onPendingDragEnd?: () => void
}
```

### 9. `DayColumn` inside `WeekView.tsx` is 160+ lines with 18 props

**Severity: Medium**
**Principle violated: Single Responsibility, readability**

`DayColumn` is complex enough to be its own file. The blocker rendering, visit card rendering, and pending assignment rendering are three distinct sub-responsibilities that could be extracted into smaller components: `VisitCard`, `PendingAssignmentCard`, `BlockerCard`.

### 10. Audit page `columns` memoization has suppressed exhaustive-deps warning

**Severity: Medium**
**Principle violated: Correctness**

`routes/audit.tsx` line 108-109:
```ts
// eslint-disable-next-line react-hooks/exhaustive-deps
[], 
```

The `columns` memo depends on `handleReview` which captures `updateStatus`. By suppressing the warning, the review buttons may use stale closures. Either include `handleReview` in deps (making it a `useCallback`), or use a ref pattern.

---

## Minor Findings

### 11. `App.tsx` -- `DemoGate` uses raw `fetch()` instead of the `api` client or a TanStack Query hook

**Severity: Low**
**Principle violated: Consistency**

The `useEffect` in `DemoGate` calls `fetch('/demo/accounts')` directly. The project convention (per CLAUDE.md) is: "Use TanStack Query for server state; avoid ad-hoc fetch calls outside of query/mutation hooks."

This is a one-off for demo mode, so it is low severity, but it would be cleaner as a `useDemoAccounts()` hook.

### 12. `auth/provider.tsx` -- module-level mutable state (`_currentUser`)

**Severity: Low**
**Principle violated: Encapsulation, testability**

`_currentUser` is a module-level `let` variable mutated from multiple functions. This makes the auth module hard to test in isolation and couples the token provider to a global singleton. Consider using a ref inside the provider or a small class instance.

### 13. `api/client.ts` -- module-level mutable state (`getAccessToken`)

**Severity: Low**
**Principle violated: Same as #12**

`getAccessToken` is a module-level `let` set via `setTokenProvider()`. Same coupling concern.

### 14. `components/ui/Toast.tsx` -- uses `setTimeout` without cleanup

**Severity: Low**
**Principle violated: Resource management**

`useToast` fires `setTimeout` calls that are never cleaned up. If the component unmounts, the callbacks will try to update state on an unmounted component. Use `useRef` to track timers and clear them on unmount.

### 15. `Sidebar.tsx` -- role filtering logic is subtly broken for admins

**Severity: Low**
**Principle violated: Clarity**

Line 44-46:
```ts
const visibleItems = navItems.filter(
  (item) => item.roles.includes(role ?? '') || (role === 'admin' && !item.roles.includes('rep')),
)
```

An admin will see: dashboard, coverage, console, audit (via the second condition), but NOT planner, targets, or activities (since those are `['rep']` and the second condition excludes `rep` items). This may be intentional, but the logic is hard to read. A clearer approach:

```ts
const ROLE_HIERARCHY: Record<Role, Role[]> = {
  admin: ['admin', 'manager'],
  manager: ['manager'],
  rep: ['rep'],
}
```

### 16. Naming: `str()` helper is too terse

**Severity: Low**
**Principle violated: Meaningful naming**

`str()` in `lib/helpers.ts` should be named something like `toStringOrEmpty()` or `safeString()` to convey its null-coalescing behavior. Callers use `str(activity.fields?.feedback)` -- without context, `str` reads like a type cast.

### 17. `priorityBorder`, `priorityColors`, `priorityDot`, `priorityStyle`, `priorityLabel` are scattered

**Severity: Low**
**Principle violated: Cohesion**

Priority-related style mappings live in `lib/styles.ts`, `WeekView.tsx` (`priorityBorder`), and `TargetMarker.tsx` (`priorityColors`). Consolidate all priority style mappings into `lib/styles.ts`.

---

## Positive Observations

These are worth calling out -- the code does many things right:

1. **Hook pattern is excellent.** Every hook file follows the same structure: query key factory, `buildQuery` helper, exported hooks. This is textbook TanStack Query usage.

2. **UI primitives are small and focused.** `Badge`, `Button`, `Card`, `EmptyState`, `Spinner`, `Modal` are all under 60 lines. Perfect single-responsibility.

3. **Type definitions are clean and well-separated.** Each domain concept has its own type file. No `any` types in the type layer.

4. **`lib/styles.ts` centralizes style mappings.** This is the right approach -- just needs the remaining scattered duplicates to be consolidated here.

5. **`lib/dates.ts` and `lib/helpers.ts` are pure, tested utilities.** Small, focused, no side effects.

6. **i18n is set up cleanly** with proper language detection and fallback.

---

## Recommended Refactoring Priority

| Priority | Item | Effort |
|----------|------|--------|
| 1 | Extract `getLat`/`getLng`/`getClassification` to `lib/target-fields.ts` | Small |
| 2 | Extract `CreateActivityModal` to its own file | Small |
| 3 | Use existing `visitTypeBadge()` everywhere | Small |
| 4 | Extract `MonthGrid` to `components/calendar/MonthGrid.tsx` | Small |
| 5 | Extract `ScheduleVisitModal` to its own file | Small |
| 6 | Split `PlannerPage` into custom hooks and sub-components | Medium |
| 7 | Consolidate scattered priority style maps | Small |
| 8 | Extract generic `buildQueryString` to `lib/query.ts` | Small |
| 9 | Reduce `WeekView` prop count via grouping or compound components | Medium |
| 10 | Fix audit page stale closure lint suppression | Small |
