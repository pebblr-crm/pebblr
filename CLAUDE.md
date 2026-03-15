# Pebblr — Claude Code Project Context

Pebblr is a self-hosted CRM for field sales lead management. Customers deploy it per-tenant in Kubernetes. Reps visit customer locations; the system tracks leads, assignments, and outcomes.

## Product

- **Domain:** Field sales / lead tracking CRM
- **Multi-tenant:** Each customer gets their own isolated deployment
- **Hosting:** Self-hosted, Kubernetes (AKS)
- **Users:** Reps who visit customer sites; managers who assign and track leads
- **Access control:** Per-row RBAC — users see only their assigned leads
- **Feature flags:** Planned integration with a centralized licensing server

## Architecture

### Backend

- **Language:** Go
- **Style:** Clean REST API (designed to support future mobile clients)
- **Auth:** Azure AD (Entra ID) via OIDC
- **Secrets:** File mounts only — never environment variables
  - External Secrets Operator → Azure KeyVault → mounted files
- **RBAC:** Row/document-level; enforced in the data layer

### Frontend

- **Language:** TypeScript (strict mode)
- **Framework:** React + TypeScript (.tsx)
- **Bundler:** Vite
- **Server state:** TanStack Query
- **Data grids:** TanStack Table
- **UI inspiration:** [Twenty CRM](https://github.com/twentyhq/twenty) — clean, dense, data-forward
- **Live updates:** SSE or WebSocket (nice-to-have, not MVP-blocking)

### Infrastructure

- **Production:** AKS (Azure Kubernetes Service)
- **Packaging:** Helm 4
- **Local dev:** Kind cluster on Linux/WSL
  - Federated credentials from a test Azure tenant
  - Managed identity (no static credentials in local env)
- **No native Windows support required**

## Development Practices

### TDD

Write tests before or alongside implementation. Do not merge untested code.

### Makefile convention

CI/CD pipelines call **Makefile targets only**. Every target is either a one-liner or delegates to a shell script in `scripts/`. Never put multi-step logic directly in a `Makefile` recipe — extract it.

### Languages

- **Go** for all backend services
- **TypeScript** for all frontend code
- No other languages without explicit decision.

### Quality gates (run before every commit)

```bash
make test       # all tests
make lint       # zero errors
make typecheck  # no TypeScript errors (tsc --noEmit)
```

### Secrets handling

- Read secrets from mounted files (e.g., `/run/secrets/db-password`)
- Never read from env vars for secret values
- In local dev, use Kind + federated credentials; do not use static secrets

### Auth

- Azure AD (Entra ID) is the identity provider
- Use OIDC tokens; validate audience and issuer in middleware
- No local username/password auth

## Conventions

### Go

- Standard library preferred; minimize third-party deps
- Errors returned, not panicked
- `internal/` for non-exported packages
- HTTP handlers thin — business logic in service layer
- Context threading: always pass `context.Context` as first arg
- **Error wrapping:** Use `fmt.Errorf("doing X: %w", err)` — always wrap with context describing the operation
- **Dependency injection:** No global state; pass all dependencies via constructors (e.g., `NewLeadService(db, rbac)`)

### API

- **Versioning:** All endpoints under `/api/v1/...` — increment version on breaking changes
- **Error responses:** Structured JSON: `{"error": {"code": "NOT_FOUND", "message": "lead not found"}}`
- **List endpoints:** Support pagination (`?page=`, `?limit=`) and filtering (`?status=`, `?assignee=`) on all collection routes

### TypeScript / React

- **Framework:** React with functional components and hooks only — no class components
- **Bundler:** Vite
- **Strict TypeScript:** `"strict": true` in tsconfig; no implicit `any`
- **Server state:** TanStack Query (`@tanstack/react-query`) for all API data fetching and caching
- **Data grids:** TanStack Table (`@tanstack/react-table`) for tabular data
- **Component files:** `.tsx` extension for all React components
- **No global state:** Prefer React context or TanStack Query cache; avoid Redux or other state managers

### Kubernetes / Helm

- Helm 4 chart per service
- Use `ExternalSecret` CRD for all secrets
- Resource limits required on all containers
- Liveness and readiness probes required

## Key Decisions

| Decision | Choice | Reason |
|---|---|---|
| Backend language | Go | Performance, simplicity, strong stdlib |
| Frontend model | React + TypeScript | Team velocity, ecosystem maturity, TanStack Query/Table integration |
| Auth provider | Azure AD (Entra ID) | Customer requirement, enterprise SSO |
| Secret delivery | File mounts via ESO | Security posture; avoids env var leakage |
| Deployment | AKS + Helm 4 | Customer's existing Azure investment |
| Local dev | Kind + federated creds | Mirrors prod auth without static secrets |

## Project Layout (planned)

```
pebblr/
├── cmd/               # Go binaries (main packages)
│   └── api/           # REST API server
├── internal/          # Go internal packages
│   ├── auth/          # Azure AD OIDC middleware
│   ├── leads/         # Lead domain logic
│   └── rbac/          # Row-level access control
├── web/               # React + TypeScript frontend
│   ├── src/
│   │   └── components/  # React .tsx components
│   ├── vite.config.ts
│   └── tsconfig.json
├── deploy/
│   └── helm/          # Helm 4 charts
│       └── pebblr/
├── scripts/           # Shell scripts called by Makefile
├── Makefile
└── README.md
```

## What Claude Should Know

- This is early-stage — many files don't exist yet. Scaffold thoughtfully.
- Always check `Makefile` targets before suggesting build/test commands.
- Secrets are **never** in env vars. If you see `os.Getenv` for a secret value, flag it.
- Per-row RBAC is a core invariant. Any data access must respect it.
- Frontend is React + TypeScript — do not introduce other frameworks or Web Components.
- Use TanStack Query for server state; avoid ad-hoc fetch calls outside of query/mutation hooks.
- AKS + Helm 4 is the deployment target; local dev uses Kind.
