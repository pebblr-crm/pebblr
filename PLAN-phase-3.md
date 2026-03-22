# Phase 3 — Reporting & Dashboard ❌

> Back to [overview](PLAN.md)

## Checklist

15. ❌ **Dashboard stats API** — planned vs realized, coverage, field vs non-field, per user/team/period
16. 🔧 **Frontend: Dashboard** — basic dashboard with StatCard, UnassignedLeadCard, TeamPerformanceCard exists; needs DrMax KPIs
17. ❌ **Joint visit** — co-visitor association, activity visible to both users
18. ❌ **Frequency tracking** — visits per account vs target from config rules

---

## 1. Dashboard Stats API ❌

New endpoints (or extended existing `/api/v1/dashboard`) to provide DrMax-specific KPIs:

- **Planned vs Realized** — count of activities by status (`planificat` vs `realizat`) for a given period
- **Coverage** — percentage of assigned accounts visited at least once in the period
- **Field vs Non-field** — split of activities by `category` (field/non_field) from config
- **Per user/team/period** — filterable by `creator_id`, `team_id`, date range
- **Target compliance** — visits per classification vs `rules.frequency` targets

### Possible Endpoints

```
GET /api/v1/dashboard/activities?period=2026-03&team_id=...
GET /api/v1/dashboard/coverage?period=2026-03&user_id=...
GET /api/v1/dashboard/frequency?period=2026-03&user_id=...
```

---

## 2. Frontend Dashboard ❌

The existing dashboard (`web/src/components/Dashboard.tsx`) has generic stat cards. For DrMax:

- **Replace/extend** StatCard data sources to use activity-based stats
- **Planned vs Realized chart** — bar or donut chart per user/team
- **Coverage heatmap** — accounts visited vs not visited
- **Frequency compliance table** — per account classification, target vs actual
- **Period selector** — month/week picker to filter all KPIs

---

## 3. Joint Visit ❌

When a visit has `joint_visit_user_id` set:

- The activity appears in both users' planners
- Both users can view the activity detail
- Only the creator can edit/submit
- RLS policy already accounts for this (see migration 007: `OR joint_visit_user_id = current_setting('app.user_id')::uuid`)
- Dashboard stats count the visit for the creator, not the co-visitor

---

## 4. Frequency Tracking ❌

Config defines minimum visit frequency per account classification:

```yaml
rules:
  frequency:
    a: 4   # 4 visits/month for A-class accounts
    b: 2
    c: 1
```

Implementation:

- Query: count realized visits per account per month, join with account classification
- Compare against target from config
- Surface in dashboard as compliance percentage
- Optionally: highlight under-visited accounts in account list
