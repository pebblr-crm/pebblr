# Pebblr — DrMax MVP Implementation Plan

> **Last updated:** 2026-03-22
>
> **Legend:** ✅ Done | 🔧 Partial | ❌ Not started
>
> **Phase details:** [Phase 1](PLAN-phase-1.md) | [Phase 2](PLAN-phase-2.md) | [Phase 3](PLAN-phase-3.md) | [Phase 4](PLAN-phase-4.md)

## Implementation Summary

| Area | Status | Details |
|------|--------|---------|
| **Tenant config system** | ✅ Done | `internal/config/` — structs, JSON loader, field-level validator, full test coverage |
| **Config API endpoint** | ✅ Done | `GET /api/v1/config` — returns tenant config for frontend |
| **Target domain** | ✅ Done | Entity, repository, service, handler, RBAC, migration 006 (replaces old Customer) |
| **Auth & RBAC** | ✅ Done | Azure AD config, OIDC middleware, static test auth, per-row RBAC with PostgreSQL RLS |
| **Database (migrations 001–006)** | ✅ Done | Users, teams, leads (soft delete, JSONB, priority), calendar_events, targets (dropped customers) |
| **Database (migrations 007–008)** | ❌ | Activities, Audit log tables |
| **Activity domain** | ❌ | Entity, repository, service, handler, submit flow, business rules |
| **Frontend foundation** | ✅ Done | React + TypeScript strict, Vite, TanStack Router/Query/Table, Tailwind |
| **Frontend pages (existing)** | ✅ Done | Dashboard, leads, calendar, team, my-leads — all with tests |
| **Frontend pages (DrMax)** | ❌ | Targets list/detail, planner, activity form/detail |
| **Helm / K8s / CI** | ✅ Done | Helm chart, Kind cluster, ExternalSecret, migration job, Makefile targets |
| **Next step** | | Phase 1 cleanup (remove dead code) → Phase 2 (activities) |

## Context

**Client:** DrMax Romania — pharmaceutical field sales CRM for Medical Division Team (18 reps, 3 managers).

**Current state:** DrMax runs on Twenty CRM with a fragile per-user object duplication hack (54 custom objects, 126 workflows, PowerShell webhook) to work around Twenty's lack of row-level security. Pebblr replaces this with a proper multi-tenant CRM with native RBAC.

**Current Pebblr codebase:** Core infrastructure is complete — auth, RBAC, tenant config, targets (doctors/pharmacies). The Lead/CalendarEvent/lead_events code from the generic CRM scaffold is unused by DrMax and scheduled for removal. Next: activities domain + planner UI.

**Key design constraint:** Nothing client-specific is hardcoded. Enums (statuses, activity types, specialties, products, etc.) and field-level requirements are driven by a JSON tenant configuration file. Validation happens at the API layer against this config, not via DB constraints on enum values.

---

## Phase Overview

### Phase 1 — Foundation (config + targets) ✅

1. ✅ **Tenant config system** — `internal/config/` package, JSON loading, validation, tests
2. ✅ **Config API endpoint** — `GET /api/v1/config`
3. ✅ **Target domain + store** — `Target` entity, PostgreSQL repo with JSONB fields, migration 006
4. ✅ **Target API** — CRUD handlers, RBAC (rep sees own, manager sees team)
5. ❌ **Target import endpoint** — bulk upsert for admin/scripts
6. ❌ **Frontend: Target list + detail** — reuse `DataTable`, status badge, pagination from leads pages
7. ❌ **Remove dead code** — drop Lead, CalendarEvent, lead_events, Customer code + migrations; remove frontend lead/calendar/my-leads pages
8. 🔧 **Seed script** — exists with sample users/teams; needs DrMax-specific doctor/pharmacy target data

→ [Full details](PLAN-phase-1.md)

### Phase 2 — Activities (core workflow) ❌

9. ❌ **Activity domain + store** — `Activity` entity, PostgreSQL repo, migration 007
10. ❌ **Activity API** — CRUD + status transitions + submit, all validated against config
11. ❌ **Audit log** — migration 008, generic audit recording on activity changes
12. ❌ **Business rules enforcement** — max activities/day, blocked days (vacation/holiday), status transitions
13. ❌ **Frontend: Activity form** — dynamic form from config, per-type field rendering
14. ❌ **Frontend: Planner** — weekly/monthly calendar view with activities (reuse CalendarGrid patterns)
15. ❌ **Frontend: Activity detail** — view + report/submit flow

→ [Full details](PLAN-phase-2.md)

### Phase 3 — Reporting & Dashboard ❌

16. ❌ **Dashboard stats API** — planned vs realized, coverage, field vs non-field, per user/team/period
17. 🔧 **Frontend: Dashboard** — basic dashboard exists; needs DrMax KPIs (replace lead-based stats)
18. ❌ **Joint visit** — co-visitor association, activity visible to both users
19. ❌ **Frequency tracking** — visits per target vs frequency from config rules

→ [Full details](PLAN-phase-3.md)

### Phase 4 — Post Go-Live Optimizations ❌

20. ❌ Weekend activity + recovery days
21. ❌ Drag & drop calendar
22. ❌ Copy-paste activities
23. ❌ Advanced filtering with saved filters
24. ❌ Target group management (quarterly)
25. ❌ Plan generation (rule-based monthly plan proposal)

→ [Full details](PLAN-phase-4.md)

---

## Dead Code Removal Plan

The generic CRM scaffold included Lead, Customer, and CalendarEvent domains. DrMax uses **Targets** (doctors/pharmacies) and **Activities** (visits, time-off) instead. Customer was already dropped in migration 006. The rest should be removed to reduce maintenance burden and confusion.

| Code to Remove | Replacement | Notes |
| --- | --- | --- |
| `internal/domain/lead.go`, `lead_status.go` | Target domain | Target covers the "entity reps visit" concept |
| `internal/domain/calendar_event.go` | Activity domain (Phase 2) | Activities replace calendar events |
| `internal/service/lead_service.go` | Target service (done) | — |
| `internal/service/calendar_event_service.go` | Activity service (Phase 2) | — |
| `internal/store/lead_store.go` + postgres impl | Target store (done) | — |
| `internal/store/calendar_event_store.go` + postgres impl | Activity store (Phase 2) | — |
| `internal/api/lead_handler.go` | Target handler (done) | — |
| `internal/api/calendar_event_handler.go` | Activity handler (Phase 2) | — |
| `internal/events/` (lead events) | Audit log (Phase 2) | Generalized audit replaces lead-specific events |
| `internal/rbac/` lead methods | Already has target methods | Remove `CanViewLead`, `ScopeLeadQuery`, etc. |
| Frontend: leads pages, my-leads, calendar | Target list/detail + Planner | Reuse `DataTable`, pagination, status badge patterns |
| `web/src/services/leads.ts` | `targets.ts` (done) | — |
| `web/src/services/calendar.ts` | Activity service (Phase 2) | — |
| `web/src/types/lead.ts`, `calendar.ts` | `target.ts` (done) + activity types | — |
| Dashboard lead-based stats | Activity-based KPIs (Phase 3) | — |

**Strategy:** Remove lead code in Phase 1 (backend is already unused). Remove calendar_event code when Activity domain lands in Phase 2. Reuse frontend patterns (DataTable, status badges, pagination) — don't rewrite from scratch.

---

## Open Questions

1. **Territory assignment** — For MVP: territory = set of targets with `assignee_id` = user. No separate territory table.

2. **Data import frequency** — Start with one-time + manual re-import. Automated sync is Phase 4+.

3. **CLM (Closed Loop Marketing)** — Deprioritize for MVP. Can be added as another dynamic field type later.

4. **IQVIA integration** — Out of scope for MVP. Potential (A/B/C) is manually set or imported.

5. **Plan generation algorithm** — Phase 4. For MVP, reps create activities manually.

6. **Retrospective edit restrictions** — Configurable in `rules` section of tenant config. Default: allow edits up to N days back.
