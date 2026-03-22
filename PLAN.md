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
| **Config API endpoint** | ❌ | `GET /api/v1/config` handler not yet wired |
| **Existing backend** | ✅ Done | Lead, Customer, CalendarEvent, User, Team, Dashboard — full CRUD + RBAC + tests |
| **Auth & RBAC** | ✅ Done | Azure AD config, OIDC middleware, static test auth, per-row RBAC with PostgreSQL RLS |
| **Database (migrations 001–005)** | ✅ Done | Users, teams, customers, leads (soft delete, JSONB fields, priority), calendar_events |
| **Database (migrations 006–008)** | ❌ | Accounts, Activities, Audit log tables |
| **Account domain** | ❌ | Entity, repository, service, handler, import endpoint |
| **Activity domain** | ❌ | Entity, repository, service, handler, submit flow, business rules |
| **Frontend foundation** | ✅ Done | React + TypeScript strict, Vite, TanStack Router/Query/Table, Tailwind |
| **Frontend pages (existing)** | ✅ Done | Dashboard, leads, customers, calendar, team, my-leads — all with tests |
| **Frontend pages (DrMax)** | ❌ | Accounts list/detail, planner, activity form/detail |
| **Helm / K8s / CI** | ✅ Done | Helm chart, Kind cluster, ExternalSecret, migration job, Makefile targets |
| **Next step** | | Phase 1 items 2–6: config endpoint, account domain + API + frontend |

## Context

**Client:** DrMax Romania — pharmaceutical field sales CRM for Medical Division Team (18 reps, 3 managers).

**Current state:** DrMax runs on Twenty CRM with a fragile per-user object duplication hack (54 custom objects, 126 workflows, PowerShell webhook) to work around Twenty's lack of row-level security. Pebblr replaces this with a proper multi-tenant CRM with native RBAC.

**Current Pebblr codebase:** Has generic domain entities (Lead, Customer, User, Team, CalendarEvent) with RBAC, event audit trail, PostgreSQL RLS, React+TanStack frontend. Needs domain evolution to support pharmaceutical field sales workflows.

**Key design constraint:** Nothing client-specific is hardcoded. Enums (statuses, activity types, specialties, products, etc.) and field-level requirements are driven by a JSON tenant configuration file. Validation happens at the API layer against this config, not via DB constraints on enum values.

---

## Phase Overview

### Phase 1 — Foundation (config + accounts) 🔧

1. ✅ **Tenant config system** — `internal/config/` package, JSON loading, validation, tests
2. ❌ **Config API endpoint** — `GET /api/v1/config`
3. ❌ **Account domain + store** — `Account` entity, PostgreSQL repo with JSONB fields, migration 006
4. ❌ **Account API** — CRUD handlers, RBAC (rep sees own, manager sees team)
5. ❌ **Account import endpoint** — bulk upsert for admin/scripts
6. ❌ **Frontend: Account list + detail** — dynamic field rendering from config
7. 🔧 **Seed script** — `scripts/seed.sh` and `scripts/seed-data.sql` exist with sample users/teams/customers/leads; needs DrMax-specific doctor/pharmacy account data

→ [Full details](PLAN-phase-1.md)

### Phase 2 — Activities (core workflow) ❌

8. ❌ **Activity domain + store** — `Activity` entity, PostgreSQL repo, migration 007
9. ❌ **Activity API** — CRUD + status transitions + submit, all validated against config
10. ❌ **Audit log** — migration 008, generic audit recording on activity changes
11. ❌ **Business rules enforcement** — max activities/day, blocked days (vacation/holiday), status transitions
12. ❌ **Frontend: Activity form** — dynamic form from config, per-type field rendering
13. ❌ **Frontend: Planner** — weekly/monthly calendar view with activities
14. ❌ **Frontend: Activity detail** — view + report/submit flow

→ [Full details](PLAN-phase-2.md)

### Phase 3 — Reporting & Dashboard ❌

15. ❌ **Dashboard stats API** — planned vs realized, coverage, field vs non-field, per user/team/period
16. 🔧 **Frontend: Dashboard** — basic dashboard exists; needs DrMax KPIs
17. ❌ **Joint visit** — co-visitor association, activity visible to both users
18. ❌ **Frequency tracking** — visits per account vs target from config rules

→ [Full details](PLAN-phase-3.md)

### Phase 4 — Post Go-Live Optimizations ❌

19. ❌ Weekend activity + recovery days
20. ❌ Drag & drop calendar
21. ❌ Copy-paste activities
22. ❌ Advanced filtering with saved filters
23. ❌ Target group management (quarterly)
24. ❌ Plan generation (rule-based monthly plan proposal)

→ [Full details](PLAN-phase-4.md)

---

## What Stays, What Goes

| Current Code                | Decision     | Status | Rationale                                                    |
| --------------------------- | ------------ | ------ | ------------------------------------------------------------ |
| `internal/domain/lead.go`   | **Keep**     | ✅ Kept | May be useful later; not in the way                          |
| `internal/domain/customer.go`| **Keep**    | ✅ Kept | Account replaces it for DrMax, but no need to delete          |
| `internal/api/lead_handler` | **Keep**     | ✅ Kept | Still functional, just not used by DrMax frontend             |
| `internal/api/calendar_*`   | **Keep**     | ✅ Kept | Activity replaces it, but keep for backward compat            |
| `internal/events/`          | **Keep**     | ✅ Kept | Event types still useful; audit_log extends this concept      |
| `internal/rbac/`            | **Extend**   | 🔧 Existing | Currently has lead-scoped methods. Needs `CanViewAccount`, `CanUpdateActivity`, `ScopeAccountQuery` |
| `migrations/001-005`        | **Keep**     | ✅ Kept | Don't touch existing schema; add new tables alongside         |
| Frontend routes             | **Extend**   | 🔧 Existing | Current routes stay; DrMax-specific routes to be added |

---

## Open Questions

1. **Account territory assignment** — Is territory = "the accounts assigned to this user"? Or is there a geographic territory entity? → For MVP: territory = set of accounts with `assignee_id` = user. No separate territory table.

2. **Data import frequency** — One-time import or periodic sync? → Start with one-time + manual re-import. Automated sync is Phase 4+.

3. **CLM (Closed Loop Marketing)** — Material tracking during visits. → Deprioritize for MVP. Can be added as another dynamic field type later.

4. **IQVIA integration** — External market data for doctor potential scoring. → Out of scope for MVP. Potential (A/B/C) is manually set or imported.

5. **Plan generation algorithm** — Auto-propose monthly plan based on frequency rules + priority. → Phase 4. For MVP, reps create activities manually.

6. **Retrospective edit restrictions** — Can reps edit past activities? → Configurable in `rules` section of tenant config. Default: allow edits up to N days back.
