# Pebblr

A self-hosted CRM for field sales lead management. Built for teams where reps visit customer locations and need tight, per-rep lead visibility with full audit trails.

## What it does

- **Lead management:** Create, assign, and track leads through their lifecycle
- **Field-first:** Designed for reps on the road visiting customer sites
- **Multi-customer-type:** Supports different customer categories with distinct workflows
- **Per-row access control:** Reps see only their assigned leads; managers see their teams
- **Self-hosted:** Each customer runs their own isolated instance in Kubernetes

## Tech stack

| Layer | Technology |
|---|---|
| Backend | Go — clean REST API |
| Frontend | React + TypeScript (Vite, TanStack Query, TanStack Table) |
| Auth | Azure AD (Entra ID) via OIDC |
| Secrets | External Secrets Operator → Azure KeyVault |
| Deployment | AKS + Helm 4 |
| Local dev | Kind (Kubernetes in Docker) |

**UI inspiration:** [Twenty CRM](https://github.com/twentyhq/twenty) — clean, data-forward design.

## Architecture

```
┌─────────────────────────────────────┐
│  Browser (React + TypeScript)       │
└────────────┬────────────────────────┘
             │ REST / SSE
┌────────────▼────────────────────────┐
│  Go API server                      │
│  - Azure AD OIDC middleware         │
│  - Per-row RBAC enforcement         │
│  - Lead / rep / customer domains    │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│  Database (TBD)                     │
└─────────────────────────────────────┘

Secrets: ESO → Azure KeyVault → file mounts (never env vars)
```

## Local development

### Prerequisites

- Docker
- [Kind](https://kind.sigs.k8s.io/)
- `kubectl`, `helm`
- Go 1.22+
- Node.js 20+
- Access to a test Azure tenant with federated credentials configured

### Setup

```bash
# Create local Kind cluster
make cluster-up

# Apply ESO + secret configuration (test tenant)
make secrets-up

# Run API server (watches for changes)
make dev-api

# Run frontend (watches for changes)
make dev-web
```

### Quality gates

```bash
make test        # all tests (Go + TS)
make lint        # zero lint errors
make typecheck   # TypeScript strict check
```

All CI/CD pipelines call `make` targets only. Do not run `go test` or `tsc` directly in CI.

## Deployment

Pebblr ships as a Helm 4 chart. Each customer gets a dedicated namespace and values file.

```bash
helm install pebblr ./deploy/helm/pebblr \
  --namespace <customer-ns> \
  --values ./deploy/<customer>.values.yaml
```

Secrets are delivered via `ExternalSecret` resources pointing to Azure KeyVault. No static secrets in the cluster.

## Security model

- **Auth:** Azure AD (Entra ID) — OIDC, no local passwords
- **Secrets:** File mounts via External Secrets Operator; environment variables are never used for secret values
- **RBAC:** Row-level — enforced in the data access layer, not just the API
- **Network:** AKS + standard Kubernetes network policies

## Development principles

- **TDD:** Write tests before or alongside code
- **Makefile convention:** CI calls make targets; targets delegate to `scripts/` for anything non-trivial
- **Secrets hygiene:** `os.Getenv` for secret values is a bug; use file reads
- **React + TypeScript:** Functional components and hooks; TanStack Query for server state

## Project status

Early bootstrap. Core scaffolding is being laid. See `.seeds/` for active tasks.
