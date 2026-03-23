# Phase 3 — Reporting & Dashboard 🔧

> Back to [overview](PLAN.md)

## Checklist

16. ✅ **Dashboard stats API** — 3 RBAC-scoped endpoints (stats, coverage, user-stats), 17 backend tests
17. ✅ **Frontend: Dashboard** — activity-based KPIs, stat cards, status/category breakdowns, user stats table, period selector, 16 tests
18. ❌ **Joint visit** — co-visitor association, activity visible to both users
19. ❌ **Frequency tracking** — visits per target vs frequency from config rules

---

## 1. Dashboard Stats API ✅

Three RBAC-scoped endpoints under `/api/v1/dashboard/`:

```
GET /api/v1/dashboard/stats?period=2026-03&teamId=...&creatorId=...
GET /api/v1/dashboard/coverage?period=2026-03&teamId=...&creatorId=...
GET /api/v1/dashboard/user-stats?period=2026-03&teamId=...
```

### Backend Package Structure

```
internal/
├── store/dashboard_store.go              # DashboardRepository interface, response types
├── store/postgres/dashboard_repository.go # PostgreSQL impl (3 aggregate queries)
├── service/dashboard_service.go          # DashboardService (RBAC, period parsing, category breakdown)
├── service/dashboard_service_test.go     # 8 service tests
├── api/dashboard_handler.go             # DashboardHandler (3 endpoints)
├── api/dashboard_handler_test.go        # 9 handler tests
└── api/router.go                        # Dashboard routes wired
```

**Stats endpoint** returns: total, submitted count, by-status breakdown, by-type breakdown, by-category (field/non_field) computed from config.

**Coverage endpoint** returns: total targets, visited targets (distinct targets with realized field activities), coverage percentage. Uses both ActivityScope (for activities) and TargetScope (for targets) RBAC.

**User-stats endpoint** returns: per-user breakdown with name (from JOIN), total count, and by-status breakdown. Sorted by total descending.

---

## 2. Frontend Dashboard ✅

Replaced the generic dashboard with activity-based KPIs.

### Components

- `web/src/components/dashboard/PeriodSelector.tsx` — Month navigator with prev/next/today
- `web/src/components/dashboard/UserStatsTable.tsx` — Per-user stats table with config-driven status labels
- `web/src/components/dashboard/StatCard.tsx` — Reused from existing (unchanged)
- `web/src/routes/index.tsx` — Dashboard page with stat cards, status/category breakdowns, user table

### TanStack Query Hooks

```typescript
useDashboardStats(params)   // GET /dashboard/stats
useCoverageStats(params)    // GET /dashboard/coverage
useUserStats(params)        // GET /dashboard/user-stats
```

### KPI Layout

1. **Stat cards row** — Total Activities, Submitted (with progress bar), Target Coverage (%), Targets Visited
2. **Breakdowns** — Status breakdown bars (config-driven labels), Field vs Non-field category bars
3. **User table** — Per-rep totals + per-status breakdown columns

### Tests

- 6 PeriodSelector tests (render, prev/next, year wrapping, today)
- 5 UserStatsTable tests (render, labels, totals, empty state, fallback)
- 3 StatCard tests (render, primary variant, change indicator)
- 2 dashboard mock setup tests (TanStack Query mocks wired)

---

## 3. Joint Visit ❌

When a visit has `joint_visit_user_id` set:

- The activity appears in both users' planners
- Both users can view the activity detail
- Only the creator can edit/submit
- RLS policy already accounts for this (see activities migration: `OR joint_visit_user_id = current_setting('app.user_id')::uuid`)
- Dashboard stats count the visit for the creator, not the co-visitor

---

## 4. Frequency Tracking ❌

Config defines minimum visit frequency per target classification:

```json
"rules": {
  "frequency": { "a": 4, "b": 2, "c": 1 }
}
```

Implementation:

- Query: count realized visits per target per month, join with target classification (from `fields->'potential'`)
- Compare against target from config
- Surface in dashboard as compliance percentage
- Highlight under-visited targets in target list
