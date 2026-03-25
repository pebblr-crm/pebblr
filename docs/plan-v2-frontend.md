# Pebblr v2 Frontend Rebuild — Implementation Plan

> Checked in so we can resume after interruptions.
> Last updated: 2026-03-26

## Overview

Rebuild the entire frontend based on new UX mockups (in `uiv2/`) and design prompts (in `docs/ux-prompts/`). The new frontend lives in `web-v2/` and coexists with the current `web/` frontend. A cookie-based toggle (`?ui=v2` / `?ui=v1`) lets users switch between them.

## Decisions

| Decision | Choice |
|---|---|
| Attachments on activities | Scrapped (revisit later if needed) |
| Route optimization | Maybe — revisit at the end |
| Territories | New `territories` table (not just a field on teams) |
| UI toggle | `?ui=v2` query param sets `pebblr_ui` cookie; backend serves correct SPA |
| Route prefixes | No `/rep/`, `/manager/`, `/admin/` prefixes — role-based redirect at `/` |
| Map library | MapLibre GL JS (`maplibre-gl` + `react-map-gl`) — free, GeoJSON-native |

## Route Structure (v2)

```
/                  → redirect based on role (planner | dashboard | console)
/planner           → Rep Planning Workspace (map + calendar)
/targets           → Rep Target Portfolio (table + map sidebar)
/targets/$id       → Rep Visit Details (pre-visit context)
/activities        → Rep Activity Log (timeline)
/activities/new    → Mobile Activity Submission (2-step form)
/dashboard         → Manager Team Dashboard (KPIs)
/reps/$id          → Manager Rep Drill-Down (read-only)
/coverage          → Manager Coverage Map (territories + heatmap)
/console           → Admin Console (users, teams, rules)
/audit             → Admin Audit Logs (review workflow)
/sign-in           → Sign-In
```

Sidebar shows different menu items based on `user.role`.

---

## Phase 1: Infrastructure — Dual SPA Serving

**Status:** Not started

### 1.1 Go router changes

**File:** `internal/api/router.go`

- Add `WebV2DistPath string` to `RouterConfig`
- Replace `mountSPA(r, path)` with `mountDualSPA(r, v1Path, v2Path)`:
  - Check `?ui=v2` or `?ui=v1` query param → set `pebblr_ui` cookie (30 days, `Path=/`, `SameSite=Lax`), redirect without query param
  - Read `pebblr_ui` cookie to decide which dist directory to serve from
  - Default to v1 if no cookie
  - If `v2DistPath` is empty, ignore cookie entirely (backward compatible)
  - Static file lookup + SPA fallback per the chosen dist directory

**File:** `cmd/pebblr/serve.go`

- Read `WEB_V2_DIST_PATH` env var (line ~105, same pattern as `WEB_DIST_PATH`)
- Pass as `WebV2DistPath` in `RouterConfig`

### 1.2 Build system

**File:** `Makefile`

- Add `WEB_V2_DIR := web-v2`
- Update `build` target to also build web-v2
- Update `test`, `lint`, `typecheck` to include web-v2
- Add `dev-web-v2` target (Vite on port 5174)

**File:** `Dockerfile`

- Add `web-v2-builder` stage (same pattern as `web-builder`)
- Copy `web-v2/dist` to `/app/web-v2/dist` in runtime stage

### 1.3 Deployment

**File:** `deploy/helm/pebblr/values.yaml` — add `webV2DistPath: "/app/web-v2/dist"`
**File:** `deploy/helm/pebblr/templates/configmap.yaml` — add `WEB_V2_DIST_PATH`

### 1.4 Tests

- `internal/api/router_test.go` — test cookie set/read, SPA fallback, static assets from correct dist, v2 path empty fallback

### Milestone

`?ui=v2` shows a "Hello v2" page. `?ui=v1` switches back. Both SPAs can call `/api/v1/*`.

---

## Phase 2: Backend Additions

**Status:** Not started

### 2.1 Migration 007: Territories

**Files:** `migrations/007_territories.{up,down}.sql`

```sql
CREATE TABLE territories (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    team_id    UUID        NOT NULL REFERENCES teams(id),
    region     TEXT,
    boundary   JSONB,      -- GeoJSON polygon
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- indexes on team_id, region
-- RLS: manager/admin see all, rep sees own team's territories
```

**Full stack:**
- `internal/domain/territory.go` — `Territory` struct
- `internal/store/territory_store.go` — `TerritoryRepository` interface (`Get`, `List`, `ListByTeam`, `Create`, `Update`, `Delete`)
- `internal/store/postgres/territory_repository.go` — pgx implementation
- `internal/store/store.go` — add `Territories()` to `Store` interface
- `internal/store/postgres/store_impl.go` — implement `Territories()`
- `internal/service/territory_service.go` — CRUD + RBAC
- `internal/api/territory_handler.go` — `GET/POST /`, `GET/PUT/DELETE /{id}`
- `internal/api/router.go` — add `TerritoryHandler` to config, mount at `/territories`
- `cmd/pebblr/serve.go` — wire territory service + handler

### 2.2 Migration 008: Audit Log Review Status

**Files:** `migrations/008_audit_log_status.{up,down}.sql`

```sql
ALTER TABLE audit_log ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE audit_log ADD COLUMN reviewed_by UUID REFERENCES users(id);
ALTER TABLE audit_log ADD COLUMN reviewed_at TIMESTAMPTZ;
CREATE INDEX idx_audit_status ON audit_log(status);
```

**Changes:**
- `internal/domain/audit.go` — add `Status`, `ReviewedBy`, `ReviewedAt` fields
- `internal/store/audit_store.go` — add `List(ctx, filter)` and `UpdateStatus(ctx, id, status, reviewerID)` to `AuditRepository`
- `internal/store/postgres/audit_repository.go` — implement
- New `internal/api/audit_handler.go` — `GET /audit` (list+filter, admin only), `PATCH /audit/{id}/status` (admin only)
- Wire in router + serve

### 2.3 Config-Only Changes (no migrations)

These use existing `fields` JSONB — just tenant config additions in `config/tenant.json`:
- **Tags** on activities → field with `"type": "multi_select"`, `"key": "tags"`
- **Agenda/checklist** → field with `"type": "checklist"`, `"key": "agenda"`
- **Geo coordinates** on targets → ensure `latitude`/`longitude` fields defined

### Milestone

All new endpoints functional. `make test` passes. v1 frontend unaffected.

---

## Phase 3: Frontend v2 Scaffold

**Status:** Not started

### 3.1 Project setup

Initialize `web-v2/` with same stack as `web/`:
- React 19, TypeScript strict, Vite, Bun
- Tailwind CSS v4, lucide-react, motion
- TanStack Query v5, TanStack Table, TanStack Router
- react-i18next (EN/RO)
- MapLibre GL JS (`maplibre-gl` + `react-map-gl`)

Vite dev server on port **5174** (v1 is 5173), same `/api` and `/demo` proxy config.

### 3.2 Directory structure

```
web-v2/src/
├── main.tsx
├── App.tsx
├── api/            # Fetch wrapper + per-resource clients
│   ├── client.ts
│   ├── targets.ts, activities.ts, teams.ts, territories.ts
│   ├── audit.ts, dashboard.ts, collections.ts, config.ts, me.ts
├── hooks/          # TanStack Query hooks
│   ├── useTargets.ts, useActivities.ts, useTeams.ts, useTerritories.ts
│   ├── useAudit.ts, useDashboard.ts, useCollections.ts, useConfig.ts, useMe.ts
├── types/          # TS interfaces (mirror web/src/types/ + territory, audit status)
├── auth/           # Auth context/provider (port from web/src/services/auth.ts)
├── i18n/           # en.ts, ro.ts
├── layouts/        # AppShell (sidebar + topbar + content), MobileShell
├── components/
│   ├── ui/         # Button, Badge, Card, Dialog, Input, Select, Spinner, EmptyState
│   ├── data/       # DataTable, StatCard, KPIBar
│   ├── map/        # MapContainer, TargetMarker, TerritoryPolygon, RouteLayer
│   ├── fields/     # FieldRenderer, MultiSelectField, ChecklistField, SelectField, TextField
│   └── calendar/   # WeekView, DayColumn
├── routes/         # TanStack Router file-based
│   ├── __root.tsx
│   ├── index.tsx         → role-based redirect
│   ├── sign-in.tsx
│   ├── planner.tsx
│   ├── targets.tsx
│   ├── targets.$id.tsx
│   ├── activities.tsx
│   ├── activities.new.tsx
│   ├── dashboard.tsx
│   ├── reps.$id.tsx
│   ├── coverage.tsx
│   ├── console.tsx
│   └── audit.tsx
└── styles/
    └── global.css
```

### 3.3 Shared patterns to establish

- **API client** — port `web/src/services/api.ts` pattern (fetch wrapper, Bearer token, structured errors)
- **Auth provider** — React context wrapping existing auth flow, provides `useAuth()` hook
- **Query hooks** — each exports `useXList(filter)`, `useX(id)`, `useCreateX()`, `useUpdateX()`, `useDeleteX()`
- **Role guard** — `RequireRole` component + `createRoleGuard(roles)` for route `beforeLoad`
- **i18n** — copy structure from `web/src/i18n/`, extend with v2-specific keys

### Milestone

`web-v2/` builds, lints, typechecks. Shell with sidebar renders. Auth works. Can fetch `/api/v1/me`. `make build` builds both frontends.

---

## Phase 4: View Implementation

### P0 — Build First

| View | Route | Key Components | Data Hooks |
|---|---|---|---|
| Rep Planning Workspace | `/planner` | MapContainer + TargetMarker, WeekView + DayColumn, nudge banners, PlannerToolbar | useTargets, useActivities, useTerritories, useDashboard |
| Mobile Activity Submission | `/activities/new` | 2-step form (outcome → notes), quick tags (MultiSelectField), outcome radio grid | useCreateActivity, useConfig |
| Rep Target Portfolio | `/targets` | DataTable, map sidebar, bulk action footer, filter bar | useTargets, useTargetVisitStatus, useTargetFrequencyStatus |

### P1 — Build Second

| View | Route | Key Components |
|---|---|---|
| Rep Visit Details | `/targets/$id` | Target header, activity timeline, field grid, route context map |
| Rep Activity Log | `/activities` | DataTable with timeline grouping, recovery balance card, compliance nudge |
| Manager Team Dashboard | `/dashboard` | KPI cards (StatCard), rep performance table (DataTable), activity breakdown chart |
| Manager Rep Drill-Down | `/reps/$id` | Read-only banner, stats cards, map + schedule (reuse planner components) |

### P2 — Build Third

| View | Route | Key Components |
|---|---|---|
| Manager Coverage Map | `/coverage` | Full-screen MapContainer, TerritoryPolygon, heatmap layer, team/rep filter panel |
| Admin Console | `/console` | User/team/rule CRUD forms, sub-navigation tabs |
| Admin Audit Logs | `/audit` | DataTable with diff view (old/new values), status filter, review action buttons |
| Sign-In | `/sign-in` | SSO buttons, demo picker (restyle existing auth) |

### Maybe — Revisit at End

- Route optimization button in planner (external service integration)

---

## Phase 5: Quality & Testing

- **TDD** — every file gets a companion `.test.ts(x)`
- **Backend tests:** repository integration tests, service unit tests, handler HTTP tests
- **Frontend tests:** component tests (`@testing-library/react`), hook tests (`renderHook`)
- **Quality gates at every phase:** `make test && make lint && make typecheck`
- **E2E:** extend `e2e/` for territory CRUD, audit status, SPA toggle

---

## Sequencing

```
Phase 1 (Infrastructure)  ─┬─> Phase 2 (Backend)  ──> Phase 4a (P0 Views)
                            └─> Phase 3 (Scaffold)  ──┘       │
                                                         Phase 4b (P1)
                                                               │
                                                         Phase 4c (P2)
```

Phases 2 and 3 can run in parallel.

---

## Notes

- v1 frontend stays untouched — both must coexist
- Both consume the same `/api/v1/*` endpoints
- v2 can add NEW endpoints but must not break v1's API contract
- No shared `packages/types/` workspace yet — duplicate types intentionally to keep frontends decoupled
- MapLibre GL JS is ~200KB gzipped — use `React.lazy` for map components
