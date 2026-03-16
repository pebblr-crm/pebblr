# Contributing to Pebblr

## Quick Start

```bash
make cluster-up      # spin up local Kind cluster
make helm-validate   # validate Helm chart (dry-run)
make test            # run all tests
make lint            # run linters
make typecheck       # TypeScript type check
```

## Development Workflow

1. Write tests first (TDD). Do not submit untested code.
2. Run all quality gates before committing:
   ```bash
   make test && make lint && make typecheck
   ```
3. Keep PRs focused — one concern per PR.

## Makefile Conventions

CI/CD calls **Makefile targets only**. When adding automation:

| Scenario | Where it lives |
|---|---|
| Single command | Inline in the `Makefile` target |
| Multi-step / branching logic | `scripts/<name>.sh`, called from Makefile |

**Do not** create a shell script just to wrap one command. Inline it in the Makefile.

**Do not** put multi-step logic directly in a Makefile recipe — extract it to `scripts/`.

Example — correct single-command target:
```makefile
e2e: ## Run E2E tests
	@go test -v -tags=e2e -count=1 -timeout=10m ./e2e/...
```

Example — correct multi-step target delegating to a script:
```makefile
helm-validate: ## Validate Helm chart (dry-run)
	@scripts/helm-ci-install.sh
```

## Secrets

- **Never** use environment variables for secret values.
- Read secrets from mounted files (e.g., `/run/secrets/db-password`).
- Local dev uses Kind + federated Azure credentials — no static secrets.

## Backend (Go)

- Return errors; do not panic.
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`.
- No global state — pass dependencies via constructors.
- Keep HTTP handlers thin; business logic lives in the service layer.
- Always pass `context.Context` as the first argument.

## Frontend (TypeScript / React)

- Functional components and hooks only — no class components.
- `"strict": true` in tsconfig; no implicit `any`.
- Use TanStack Query for all server state — no ad-hoc `fetch` outside query/mutation hooks.
- Use TanStack Table for tabular data.

## RBAC

Per-row RBAC is a core invariant. Every data access path must enforce it. PostgreSQL RLS is an additional safety net, not the primary control.

## Kubernetes / Helm

- One Helm 4 chart per service under `deploy/helm/`.
- All secrets via `ExternalSecret` CRD (External Secrets Operator → Azure KeyVault).
- Every container must have resource limits, liveness probes, and readiness probes.
