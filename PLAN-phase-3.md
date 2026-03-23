# Phase 3 — Reporting & Dashboard ❌

> Back to [overview](PLAN.md)

## Checklist

16. ✅ **Dashboard stats API** — planned vs realized, coverage, field vs non-field, per user/team/period
17. ✅ **Frontend: Dashboard** — activity KPIs, coverage, frequency compliance, period selector, 15 tests
18. ✅ **Joint visit** — co-visitor validation, planner indicator, detail badge, 7 service tests
19. ❌ **Frequency tracking** — visits per target vs frequency from config rules

---

## 1. Dashboard Stats API ✅

New endpoints (or rework existing `/api/v1/dashboard`) to provide DrMax-specific KPIs:

- **Planned vs Realized** — count of activities by status (`planificat` vs `realizat`) for a given period
- **Coverage** — percentage of assigned targets visited at least once in the period
- **Field vs Non-field** — split of activities by `category` (field/non_field) from config
- **Per user/team/period** — filterable by `creator_id`, `team_id`, date range
- **Target compliance** — visits per classification vs `rules.frequency` targets

### Possible Endpoints

```
GET /api/v1/dashboard/activities?period=2026-03&team_id=...
GET /api/v1/dashboard/coverage?period=2026-03&user_id=...
GET /api/v1/dashboard/frequency?period=2026-03&user_id=...
```

Remove existing lead-based dashboard stats (`DashboardService` currently aggregates lead data — replace entirely with activity-based metrics).

---

## 2. Frontend Dashboard ✅

The existing dashboard (`web/src/routes/index.tsx`) has generic stat cards. For DrMax:

- **Remove** `UnassignedLeadCard` (already planned for Phase 1 cleanup)
- **Replace** lead-based `StatCard` data with activity-based KPIs
- **Add** planned vs realized chart — bar or donut chart per user/team
- **Add** coverage heatmap — targets visited vs not visited
- **Add** frequency compliance table — per target classification, actual vs required
- **Add** period selector — month/week picker to filter all KPIs

Reuse existing `StatCard` component structure, `TeamPerformanceCard` patterns.

---

## 3. Joint Visit ✅

When a visit has `joint_visit_user_id` set:

- The activity appears in both users' planners (repo List query includes `OR joint_visit_user_id`)
- Both users can view the activity detail (RBAC `CanViewActivity` checks JointVisitUID)
- Only the creator can edit/submit (RBAC `CanUpdateActivity` checks CreatorID only)
- RLS policy as defense-in-depth (`OR joint_visit_user_id = current_setting('app.user_id')::uuid`)
- Dashboard stats count the visit for the creator, not the co-visitor (`DashboardFilter.UserID` → `creator_id`)
- **Service validation:** self-reference rejected, non-existent user rejected (7 tests)
- **Frontend:** Users icon on planner ActivityCard, "Joint visit with {name}" badge on detail page

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
