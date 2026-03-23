# Pebblr — DrMax MVP Implementation Plan

> **Last updated:** 2026-03-23
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
| **Database (migrations 001–007)** | ✅ Done | Users, teams, leads, calendar_events, targets, external_id |
| **Database (migrations 008–011)** | ✅ Done | Drop leads (008), drop calendar_events (009), activities table with RLS (010), audit_log (011) |
| **Activity domain + store** | ✅ Done | Entity, AuditEntry, ActivityRepository, AuditRepository, PostgreSQL impls, RBAC |
| **Activity API** | ✅ Done | CRUD handlers, status transitions, submit flow, business rules (7 endpoints, 38 tests) |
| **Business rules** | ✅ Done | Max activities/day, blocked days, target required, status transitions, submit lock |
| **Frontend foundation** | ✅ Done | React + TypeScript strict, Vite, TanStack Router/Query/Table, Tailwind |
| **Frontend pages (existing)** | ✅ Done | Dashboard, team — with tests |
| **Frontend pages (DrMax)** | 🔧 Partial | Targets list/detail done; activity form/detail done; planner not started |
| **Helm / K8s / CI** | ✅ Done | Helm chart, Kind cluster, ExternalSecret, migration job, Makefile targets |
| **Target import** | ✅ Done | `POST /api/v1/targets/import` — admin-only bulk upsert by external ID |
| **Dead code removal** | ✅ Done | Leads, lead_events, CalendarEvent code all removed; replaced by Target + Activity domains |
| **Next step** | | Frontend: Planner (item 14) — weekly/monthly calendar view |

## Context

**Client:** DrMax Romania — pharmaceutical field sales CRM for Medical Division Team (18 reps, 3 managers).

**Current state:** DrMax runs on Twenty CRM with a fragile per-user object duplication hack (54 custom objects, 126 workflows, PowerShell webhook) to work around Twenty's lack of row-level security. Pebblr replaces this with a proper multi-tenant CRM with native RBAC.

**Current Pebblr codebase:** Core infrastructure is complete — auth, RBAC, tenant config, targets (doctors/pharmacies). Activity domain, API, and business rules are fully implemented with 38 backend tests. Frontend activity form, detail, and edit pages are done with config-driven dynamic fields and 12 component tests. All dead code (Lead, CalendarEvent, lead_events) has been removed. Next: Planner (weekly/monthly calendar view).

**Key design constraint:** Nothing client-specific is hardcoded. Enums (statuses, activity types, specialties, products, etc.) and field-level requirements are driven by a JSON tenant configuration file. Validation happens at the API layer against this config, not via DB constraints on enum values.

---

## Phase Overview

### Phase 1 — Foundation (config + targets) ✅

1. ✅ **Tenant config system** — `internal/config/` package, JSON loading, validation, tests
2. ✅ **Config API endpoint** — `GET /api/v1/config`
3. ✅ **Target domain + store** — `Target` entity, PostgreSQL repo with JSONB fields, migration 006
4. ✅ **Target API** — CRUD handlers, RBAC (rep sees own, manager sees team)
5. ✅ **Target import endpoint** — bulk upsert for admin/scripts
6. ✅ **Frontend: Target list + detail** — config-driven list with type filter, dynamic columns, detail page with resolved option labels
7. ✅ **Remove dead code** — Lead domain removed (migration 008), CalendarEvent removed with Activity domain landing
8. 🔧 **Seed script** — exists with sample users/teams; needs DrMax-specific doctor/pharmacy target data

→ [Full details](PLAN-phase-1.md)

### Phase 2 — Activities (core workflow) 🔧

9. ✅ **Activity domain + store** — `Activity` + `AuditEntry` entities, PostgreSQL repos (migrations 009–011), RBAC enforcer with activity methods
10. ✅ **Activity API** — CRUD + status transitions + submit, all validated against config (7 endpoints, 22 service + 16 handler tests)
11. ✅ **Audit log** — migration 011, `AuditRepository` interface + PostgreSQL impl
12. ✅ **Business rules enforcement** — max activities/day, blocked days, target required, status transitions, submit lock
13. ✅ **Frontend: Activity form** — config-driven dynamic form, create/edit routes, target search, multi-select fields, 12 tests
14. ❌ **Frontend: Planner** — weekly/monthly calendar view with activities
15. ✅ **Frontend: Activity detail** — view + report/submit flow, status transitions, edit link

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

## Dead Code Removal ✅

All generic CRM scaffold code has been removed:

- **Lead domain** — removed in commit `e260728` (PR #51): domain, service, store, handler, events, RBAC methods, frontend pages/services/types
- **CalendarEvent domain** — removed when Activity domain landed: domain, service, store, handler, tests, frontend calendar page/service/types still present (to be cleaned in frontend activity work)
- **Database** — migration 008 drops `leads` + `lead_events`; migration 009 drops `calendar_events`
- **Frontend leads/my-leads pages** — removed in PR #51
- **Frontend calendar components** — `web/src/routes/calendar/`, `web/src/services/calendar.ts`, `web/src/types/calendar.ts`, `web/src/components/calendar/` still present; will be removed when Planner lands (item 14)

---

## Open Questions

1. **Territory assignment** — For MVP: territory = set of targets with `assignee_id` = user. No separate territory table.

2. **Data import frequency** — Start with one-time + manual re-import. Automated sync is Phase 4+.

3. **CLM (Closed Loop Marketing)** — Deprioritize for MVP. Can be added as another dynamic field type later.

4. **IQVIA integration** — Out of scope for MVP. Potential (A/B/C) is manually set or imported.

5. **Plan generation algorithm** — Phase 4. For MVP, reps create activities manually.

6. **Retrospective edit restrictions** — Configurable in `rules` section of tenant config. Default: allow edits up to N days back.
