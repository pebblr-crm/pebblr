# Phase 3 — Reporting & Map Planner 🔧

> Back to [overview](PLAN.md)

## Checklist

16. ✅ **Dashboard stats API** — planned vs realized, coverage, field vs non-field, per user/team/period
17. ✅ **Frontend: Dashboard** — activity KPIs, coverage, frequency compliance, period selector, 15 tests
18. ✅ **Joint visit** — co-visitor validation, planner indicator, detail badge, 7 service tests
19. ✅ **Frequency tracking** — visits per target vs frequency from config rules
20. ❌ **Target collections** — user-created saved groups of targets for reuse across planning cycles
21. ❌ **Clone week** — duplicate a planned week's activities 3 weeks forward
22. ❌ **Activity display name** — auto-generated label: type + target name + date
23. ❌ **Planner density** — ensure 15+ activities visible per day

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

## 4. Frequency Tracking ✅

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
- **Map planner integration:** show frequency compliance as marker color intensity (red = behind, green = on track)

---

## 5. Target Collections ❌

Reps visit ~165 targets on a 3-week rotation. They naturally group targets by geography (e.g., "Bucharest Nord", "Ploiesti route") or day (e.g., "Monday targets"). Target collections let them save these groupings and reuse them across planning cycles.

### Data Model

- New `target_collections` table: `id`, `name`, `creator_id`, `team_id`, `created_at`, `updated_at`
- Join table `target_collection_items`: `collection_id`, `target_id`
- RBAC: rep sees own collections, manager sees team collections
- RLS as defense-in-depth

### API

```
POST   /api/v1/collections              — create collection
GET    /api/v1/collections              — list user's collections
GET    /api/v1/collections/:id          — get collection with target IDs
PUT    /api/v1/collections/:id          — rename or update targets
DELETE /api/v1/collections/:id          — delete collection
POST   /api/v1/collections/:id/targets  — add targets to collection
DELETE /api/v1/collections/:id/targets  — remove targets from collection
```

### Map Planner Integration

This is the primary consumer. In the map planner:

- **"Save as collection" button** — after selecting targets on the map, save the selection as a named collection
- **Collection picker dropdown** — load a saved collection to pre-select those targets on the map
- Loading a collection highlights all its targets and centers the map on their centroid
- Collections appear as a filterable list in the right panel above individual targets
- Drag an entire collection to a day slot to assign all its targets at once

### UX Flow

1. Rep opens map planner for Week 1
2. Selects 8 targets in the Bucharest area by clicking markers
3. Drags them to Monday → batch creates 8 activities
4. Clicks "Save as collection" → names it "Bucharest Mon"
5. Next cycle (3 weeks later), opens map planner, picks "Bucharest Mon" from dropdown
6. All 8 targets pre-selected, drags to Monday → done in seconds

---

## 6. Clone Week ❌

Reps plan on a 3-week rotation: Week 1, Week 2, Week 3, then repeat. After planning Week 1, they should be able to clone it to the next Week 1 occurrence (3 weeks later) in one action.

### Implementation

- **Backend:** `POST /api/v1/activities/clone-week` — accepts `source_week_start` (ISO date) and `target_week_start` (ISO date). Clones all activities from the source week to the target week with status reset to `planned`, new UUIDs, and updated `due_date` (same weekday offset).
- **Validation:** target week must be in the future, source week must have activities, no duplicate creation if target week already has activities (warn or skip conflicts)
- **Frontend:** "Clone week" button in the planner week view header. Clicking opens a date picker defaulting to +3 weeks. Confirmation shows count of activities to be cloned.
- Works with both the map planner and the week view planner

### Why not copy-paste individual activities?

The 3-week cycle is the natural unit. Copying individual activities is slow and error-prone with 15 activities/day × 5 days = 75 activities per week. One-click clone is the right abstraction.

---

## 7. Activity Display Name ❌

Activities currently show the activity type label in planner cards. For quick scanning with 15+ activities per day, a richer display name helps.

### Implementation

- **Format:** `{activity type} — {target name} — {date}` (e.g., "Visit — Dr. Popescu — Mar 24")
- **Generated client-side** from existing data (activity type from config, target name denormalized on activity, due date)
- **Used in:** planner ActivityCard, activity list, daily view, week/month grid cells
- **Non-field activities** (no target): just `{type} — {date}` (e.g., "Training — Mar 24")
- No backend changes needed — purely a frontend display improvement

---

## 8. Planner Density ❌

DrMax reps have up to 15 field activities per day. The week and daily views must accommodate this without scrolling.

### Implementation

- **Compact activity cards** — reduce padding, use single-line layout with truncation
- **Week view:** each day column scrollable independently if needed, but aim for all 15 visible
- **Daily view:** vertical list with minimal spacing
- **Map planner day slots:** show count badge + scrollable list when > 5 activities assigned
- **Hide empty fields** in planner cards — only show fields that have values (addresses DrMax feedback #9)
- **Responsive:** on smaller screens, switch to count-only badges that expand on click
