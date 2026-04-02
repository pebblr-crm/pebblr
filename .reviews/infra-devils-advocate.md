# Infrastructure Review: Devil's Advocate

**Scope:** `deploy/`, `migrations/`, `Makefile`, `scripts/`, `skaffold.yaml`, `Dockerfile`, `.github/workflows/`
**Date:** 2026-04-02

---

## 1. Do You Really Need Both Istio AND Envoy Gateway?

**Files:** `deploy/helm/pebblr/templates/istio-gateway.yaml`, `istio-virtualservice.yaml`, `gateway.yaml`, `ingress.yaml` (which is actually an HTTPRoute), `values.yaml`

The chart ships four ingress-adjacent template files and two independent toggle systems:

- `gateway.enabled` -- activates Gateway API resources (Envoy Gateway)
- `istio.enabled` -- activates Istio VirtualService + optional Istio Gateway

The local dev stack (`values-local.yaml`) uses Envoy Gateway. The AKS deploy (`values-aks.yaml`) uses Istio. These are two completely different networking stacks with different CRDs, different operational models, and different failure modes. The Helm chart has to carry templates for both, and both need testing.

**The question:** Is the plan to converge on one, or is this permanent? If Istio is the production choice, why not use Istio everywhere including local dev? If Gateway API is the future, why ship Istio templates at all?

**Risk:** Nobody tests the Istio path in CI. The e2e workflow creates a lightweight Kind cluster with *neither* Envoy Gateway nor Istio. The AKS deploy uses Istio but there is no automated test for the Istio template rendering with `values-aks.yaml`. A broken Istio VirtualService template would only be caught when deploying to production.

**Recommendation:** Pick one networking model. Gateway API is the standard going forward -- Istio itself supports it. Drop the legacy `networking.istio.io` VirtualService/Gateway resources and use Gateway API everywhere. If Istio must stay for the AKS cluster, at minimum add a CI job that runs `helm template` with `values-aks.yaml` to validate those templates.

---

## 2. Skaffold: Adding Value or Adding Complexity?

**Files:** `skaffold.yaml`, `scripts/deploy-local.sh`, `Makefile`

Skaffold is used for exactly two things:
1. `make deploy` -- calls `skaffold run` via `scripts/deploy-local.sh`
2. `make e2e-deploy` -- calls `skaffold run -p e2e`

That is it. No `skaffold dev` for hot-reload. No file watching. No multi-service orchestration. The project has a single artifact (`pebblr-api`), a single Helm chart, and a single Dockerfile.

What Skaffold does here: builds a Docker image, loads it into Kind, and runs `helm upgrade --install`. You could replace the entire `skaffold.yaml` with:

```bash
docker build -t pebblr-api:local .
kind load docker-image pebblr-api:local --name pebblr-local
helm upgrade --install pebblr deploy/helm/pebblr \
  -f deploy/helm/pebblr/values-local.yaml \
  --set image.tag=local -n pebblr --create-namespace
```

That is three commands, no extra binary to install, no YAML schema to debug, no Skaffold version pinning. The E2E workflow already installs Skaffold from `latest` (!) which is a reproducibility risk in itself.

**The question:** What is Skaffold buying you that three shell commands would not? If the answer is "image tag coordination," that is solved by hardcoding a local tag. If the answer is "future multi-service support," you have one service.

**Recommendation:** Either use Skaffold's dev-loop features (file sync, hot reload) to justify its presence, or remove it and use plain `docker build` + `kind load` + `helm upgrade`. If you keep it, pin the version in CI instead of pulling `latest`.

---

## 3. The Migration Strategy Has a Scaling Problem

**Files:** `migrations/001-009`, `deploy/helm/pebblr/templates/migration-job.yaml`

### 3a. Helm Pre-Install Hook Means Migrations Block Every Deploy

The migration job is annotated with `helm.sh/hook: pre-upgrade,pre-install`. Every single Helm upgrade -- even a config-only change -- waits for the migration job to complete. With 9 migrations today this is fast, but golang-migrate runs them sequentially, and the job has `backoffLimit: 3`. A flaky network connection to Postgres means three retries before the entire deploy fails.

**The question:** When you have 50+ migrations and a migration takes 30 seconds to acquire a lock on a busy table, are you comfortable with every config change being blocked by this?

**Alternative:** Run migrations as a separate CI step or a dedicated Argo/Flux workflow, not as a Helm hook. The deploy should be decoupled from schema changes.

### 3b. Migration 009 Mutates Data -- In a DDL Migration

Migration `009_activity_unique_target_date.up.sql` does a `UPDATE activities SET deleted_at = NOW()` before creating a unique index. This is a data migration baked into a schema migration. If it fails partway through, the `UPDATE` has already committed (golang-migrate does not wrap multi-statement files in a transaction by default on Postgres for DDL). You could end up with soft-deleted records but no unique index.

**Recommendation:** Split data migrations from schema migrations. Consider wrapping the UPDATE + CREATE INDEX in an explicit transaction, or use a two-phase approach (migration 009a: data fix, 009b: add constraint).

### 3c. No RLS on the Audit Log

Tables `targets`, `activities`, `target_collections`, `target_collection_items`, and `territories` all have RLS policies. The `audit_log` table does not. The audit log contains `old_value` and `new_value` JSONB columns that could contain sensitive lead data. If any code path exposes audit log queries without application-layer filtering, RLS will not save you.

**The question:** Is this intentional because only admins query the audit log? If so, document it. If not, add an RLS policy.

---

## 4. Four Values Files: Why Not Overlays?

**Files:** `values.yaml`, `values-local.yaml`, `values-aks.yaml`, `values-e2e.yaml`, `values-ci.yaml`

There are four environment-specific values files plus the base. They share a lot of repeated structure (every single one disables `externalSecrets`, sets `serviceAccount.annotations: {}`, etc.). When you add a new feature toggle, you need to remember to set it in four places.

**The question:** Have you considered Kustomize overlays or Helmfile to reduce duplication? With Helmfile, the base values stay in `values.yaml` and each environment only overrides what differs. With your current approach, adding a new secret key means editing `values.yaml` + `values-local.yaml` + `values-e2e.yaml` + `values-ci.yaml` + `values-aks.yaml`.

**Counterpoint:** Helmfile adds another tool. But the current approach does not scale past 5 environments without becoming a maintenance burden.

---

## 5. The Dockerfile Installs npm to Install Bun

**File:** `Dockerfile`, line 19-20

```dockerfile
FROM node:25-alpine AS web-builder
RUN npm install -g bun && bun install --frozen-lockfile
```

The project mandate is "use Bun, never npm or yarn." Yet the Dockerfile uses a Node image and runs `npm install -g bun`. This is slower than using the official `oven/bun` image, adds npm as a transitive dependency, and contradicts the project's own conventions.

**Recommendation:** Use `FROM oven/bun:1.2.5-alpine AS web-builder` instead.

---

## 6. Secret Handling Contradictions

### 6a. VITE_STATIC_TOKEN as a Build Arg

**File:** `skaffold.yaml` line 12, `Dockerfile` lines 26-28, `.github/workflows/deploy.yml` line 57

`VITE_STATIC_TOKEN` is passed as a Docker build argument. Build args are baked into image layers and visible via `docker history`. The value in `skaffold.yaml` is a hardcoded string `"local-jwt-secret-not-for-production"`, but in CI it comes from `${{ secrets.VITE_STATIC_TOKEN }}`. If this is truly a secret, it should not be a build arg -- it will be visible in the image metadata. If it is not a secret, why is it in GitHub Secrets?

### 6b. Postgres Password Hardcoded in Script AND Manifest

**Files:** `scripts/cluster-db.sh` (line 20), `deploy/k8s/postgres.yaml` (line 11)

The local dev password `pebblr-local` appears in both places. This is fine for local dev, but the pattern trains developers to think hardcoded credentials in scripts are acceptable. The `CLAUDE.md` says "secrets are never in env vars" but `deploy/k8s/postgres.yaml` passes the password via `envFrom` + a Secret with `stringData`. That is technically an env var on the Postgres container.

**The question:** Is the "no env vars for secrets" rule only for the application container, or is it a blanket policy? If blanket, the Postgres dev manifest violates it.

---

## 7. The AKS Deploy Workflow Uses Static Credentials

**File:** `.github/workflows/deploy.yml`

The TODO on line 7 says it all: "Replace ACR admin credentials + kubeconfig with Azure Workload Identity." The workflow currently:
- Uses `ACR_USERNAME` / `ACR_PASSWORD` (ACR admin credentials, which Microsoft discourages)
- Base64-decodes an entire kubeconfig from a GitHub Secret

This is the highest-risk item in the review. A leaked `KUBE_CONFIG` secret gives full cluster access. ACR admin credentials cannot be scoped to push-only.

**The question:** When is this TODO getting done? It should be the next infrastructure task, not a "nice-to-have."

---

## 8. Helm Chart Says v2, Project Says Helm 4

**File:** `deploy/helm/pebblr/Chart.yaml` line 1

```yaml
apiVersion: v2
```

`CLAUDE.md` says "Helm 4 chart per service." CI installs Helm `v3.14.0`. The Chart apiVersion is `v2` (which is Helm 3). There is no Helm 4 anything here. Is Helm 4 an aspiration, or should the docs be updated to match reality?

---

## 9. CI Does Not Validate the Full Deploy Path

**File:** `.github/workflows/ci.yml`

The CI workflow runs `helm lint` but does *not* run `helm template` with any values file. `helm lint` catches syntax errors but not template logic bugs (e.g., a typo in `{{ if .Values.istio.enbled }}`). The `helm-ci-install.sh` script does a `--dry-run` install, but it is called via `make helm-validate` which is not invoked in any CI workflow.

**Recommendation:** Add `make helm-validate` to the CI pipeline, or at minimum run `helm template` with each values file to catch rendering errors.

---

## 10. Kind Cluster Has 3 Nodes but E2E Uses 1

**File:** `deploy/kind/kind-config.yaml`

The Kind config creates 1 control-plane + 2 workers (3 nodes total). The E2E workflow uses the same config. For E2E tests that just need to deploy a single pod and run API tests, 3 nodes means:
- Longer cluster creation time in CI
- More memory consumption on the runner
- No actual benefit (you are not testing multi-node scheduling)

**Recommendation:** Create a separate `kind-config-ci.yaml` with a single node for E2E, or parametrize the config.

---

## 11. The `ingress.yaml` Is Not an Ingress

**File:** `deploy/helm/pebblr/templates/ingress.yaml`

The file is named `ingress.yaml` but contains a Gateway API `HTTPRoute` resource. This is confusing. The Kubernetes `Ingress` resource and Gateway API `HTTPRoute` are different APIs. Name the file `httproute.yaml` to match what it actually contains.

---

## 12. Migration Job Name Will Collide

**File:** `deploy/helm/pebblr/templates/migration-job.yaml` line 4

```yaml
name: {{ include "pebblr.fullname" . }}-migrate-{{ .Release.Revision }}-{{ now | date "20060102150405" }}
```

The `now` function generates a timestamp at template-render time. If Helm renders the template twice in quick succession (e.g., a retry), the names could collide. More importantly, the `hook-delete-policy: before-hook-creation` means the old job is deleted before creating the new one. If the old job is still running (backoffLimit not exhausted), you could have a race condition where the delete races the running pod.

**Recommendation:** Use `helm.sh/hook-delete-policy: hook-succeeded` only, and handle failures explicitly. Or add `ttlSecondsAfterFinished` to auto-clean completed jobs.

---

## Summary: Priority Actions

| # | Issue | Severity | Effort |
|---|-------|----------|--------|
| 7 | Static AKS credentials in CI | **Critical** | Medium |
| 6a | Secret baked into Docker image layer | **High** | Low |
| 1 | Untested Istio templates in prod path | **High** | Medium |
| 9 | CI does not validate full deploy path | **High** | Low |
| 3b | Data mutation in DDL migration | **Medium** | Low |
| 3c | No RLS on audit_log | **Medium** | Low |
| 5 | npm-to-install-bun in Dockerfile | **Low** | Low |
| 2 | Skaffold overhead for single-service | **Low** | Medium |
| 8 | Docs say Helm 4, reality is Helm 3 | **Low** | Low |
| 11 | ingress.yaml is actually an HTTPRoute | **Low** | Low |
| 10 | 3-node Kind cluster for CI | **Low** | Low |
