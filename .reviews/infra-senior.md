# Infrastructure Review -- Senior Engineer

**Reviewer:** The Senior Engineer Who Has Seen It All
**Date:** 2026-04-02
**Scope:** `deploy/`, `migrations/`, `Makefile`, `scripts/`, `skaffold.yaml`, `Dockerfile`

---

## Executive Summary

The infrastructure is in surprisingly good shape for an early-stage project. Someone here has been burned before -- the security context on pods is locked down, secrets are file-mounted not env-injected, and the Dockerfile runs as a non-root user. That said, I found several issues that range from "will bite you during an incident" to "will wake you up at 3 AM." Grouped by severity below.

---

## CRITICAL -- Fix Before Production

### 1. Image Tag `latest` in Default Values

**File:** `deploy/helm/pebblr/values.yaml:9`

```yaml
image:
  tag: "latest"
```

I have attended the funerals of three deployments caused by `latest` tags. Kubernetes does not re-pull `latest` when `imagePullPolicy: IfNotPresent` is set (which it is, line 8). You will deploy v2 thinking you got v2, but the node already cached v1, and you will debug this for four hours at midnight.

**Recommendation:** Default to the Chart's `appVersion`. The template already falls back to `.Chart.AppVersion` when tag is empty. Change the default to `""` so the fallback always triggers in production. `latest` should never appear in a values file.

**Severity:** CRITICAL -- silent deployment of wrong version in production.

---

### 2. Skaffold Build Arg Leaks a JWT Secret Into the Docker Image Layer Cache

**File:** `skaffold.yaml:12`

```yaml
buildArgs:
  VITE_STATIC_TOKEN: "local-jwt-secret-not-for-production"
```

This is baked into the Docker image build layer. Even though the value says "not-for-production," any image built with Skaffold's default profile carries this token in its layer history. If someone pushes this image to a registry (and they will -- someone always does), the token is extractable with `docker history` or `crane`.

The Dockerfile propagates it to a `VITE_STATIC_TOKEN` env var during the frontend build, which means it gets compiled into the JavaScript bundle. Anyone with browser DevTools can extract it.

**Recommendation:** Remove the `VITE_STATIC_TOKEN` build arg from skaffold.yaml. For local dev, inject it via a `.env` file that Vite reads at dev-server time (not at build time). For production, the frontend should not have a static token at all -- it should use the OIDC flow.

**Severity:** CRITICAL -- credential leakage via Docker layers and JS bundles.

---

### 3. Down Migration 009 Does Not Restore Soft-Deleted Rows

**File:** `migrations/009_activity_unique_target_date.down.sql`

The up migration soft-deletes duplicate rows (sets `deleted_at = NOW()`), then adds a unique index. The down migration only drops the index -- it does not restore the rows that were soft-deleted. If you roll back this migration, you have silently lost data.

```sql
-- up.sql: Soft-deletes duplicates, then adds constraint
-- down.sql: Only drops the index. The soft-deleted rows stay dead.
DROP INDEX IF EXISTS idx_activities_unique_target_date;
```

**Recommendation:** The down migration should restore the soft-deleted rows. At minimum, add a comment acknowledging the data loss is intentional. Better: store the affected IDs in a migration metadata table, or use a reversible marker.

**Severity:** CRITICAL -- silent data loss on rollback.

---

### 4. Down Migration 001 Is a Full Data Wipe With No Safety Net

**File:** `migrations/001_initial_schema.down.sql`

```sql
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "pgcrypto";
```

Running the down migration for 001 drops the `users` table, which cascades to basically everything in the system via foreign keys. If someone accidentally runs `migrate down` one step too many in production, they lose all data.

**Recommendation:** Down migrations that drop core tables should be gated or at minimum wrapped in a transaction with a safety check (e.g., refuse to run if the table has more than 0 rows, or require a `PEBBLR_ALLOW_DESTRUCTIVE=true` env flag). Consider removing destructive down migrations entirely for production -- they exist for dev convenience only.

**Severity:** CRITICAL -- accidental `migrate down` destroys all data.

---

## HIGH -- Fix Before GA

### 5. No `activeDeadlineSeconds` on the Migration Job

**File:** `deploy/helm/pebblr/templates/migration-job.yaml`

The migration Job has `backoffLimit: 3` but no `activeDeadlineSeconds`. If a migration hangs (e.g., waiting on a lock held by an open transaction), it will retry 3 times, each hanging indefinitely. Your Helm upgrade will be stuck, and your release pipeline will time out after whatever CI's global timeout is.

**Recommendation:** Add `activeDeadlineSeconds: 300` (5 minutes is generous for any migration). Also add `ttlSecondsAfterFinished: 600` to clean up completed Job pods.

---

### 6. No `startupProbe` -- Slow Starts Will Get Killed

**File:** `deploy/helm/pebblr/templates/deployment.yaml`

The liveness probe has `initialDelaySeconds: 10` and `periodSeconds: 30`. If the app takes longer than 10 seconds to start (cold JIT, slow DB connection, large config load), the liveness probe will fail and Kubernetes will kill it. Then it restarts, takes 10+ seconds again, gets killed again. Crash loop.

**Recommendation:** Add a `startupProbe` with generous timing (e.g., `failureThreshold: 30, periodSeconds: 2`). This lets the app take up to 60 seconds to start before the liveness probe kicks in.

---

### 7. No `PodDisruptionBudget`

There is no PDB defined. When AKS does a node upgrade (and it will, on its own schedule, without asking), it will drain nodes and your single-replica pod gets evicted with zero availability. Even with `replicaCount: 1`, a PDB with `minAvailable: 0` documents the intent. With `replicaCount > 1`, you need `minAvailable: 1` or the node drain will take everything down.

**Recommendation:** Add a PDB template, enabled when `replicaCount > 1`.

---

### 8. No `NetworkPolicy` Anywhere

There are zero NetworkPolicies in the chart or the k8s manifests. Every pod in the namespace can talk to every other pod in the cluster. If someone deploys a debug pod, or if a dependency gets compromised, lateral movement is unrestricted.

**Recommendation:** Add a default-deny ingress/egress NetworkPolicy, then allow only the specific traffic the app needs (ingress from gateway, egress to PostgreSQL and Azure AD endpoints).

---

### 9. ExternalSecret API Version Is `v1beta1`

**File:** `deploy/helm/pebblr/templates/externalsecret.yaml:2`

```yaml
apiVersion: external-secrets.io/v1beta1
```

ESO v1beta1 has been deprecated since ESO 0.9.x. The chart pins ESO 0.12.1 in the Makefile. You should be using `v1` by now. When ESO drops v1beta1 support, your chart will break with zero warning.

**Recommendation:** Update to `external-secrets.io/v1` in both ExternalSecret templates.

---

### 10. PostgreSQL Dev Manifest Has No Liveness Probe

**File:** `deploy/k8s/postgres.yaml`

The dev PostgreSQL has a readinessProbe (pg_isready) but no livenessProbe. If PostgreSQL hangs (and it does -- I have the scars), it will sit there appearing "ready" to Kubernetes but not actually accepting connections. The app will get connection errors and nobody will understand why.

**Recommendation:** Add a livenessProbe using `pg_isready`. It is cheap and catches the most common hang scenarios.

---

### 11. `database.port` Type Mismatch Will Bite You

**File:** `deploy/helm/pebblr/values.yaml:73` vs `deploy/helm/pebblr/values-e2e.yaml:12`

In `values.yaml`, `database.port` is an integer: `5432`. In `values-e2e.yaml`, it is a string: `"5432"`. The configmap template quotes it anyway (`{{ .Values.database.port | quote }}`), so it works today. But if someone adds a template that does arithmetic or comparison on the port, the type inconsistency will cause a helm template error that only appears in the e2e profile.

**Recommendation:** Make it consistently an integer (or consistently a string) across all values files.

---

## MEDIUM -- Technical Debt

### 12. Gateway Listener on Port 80 Without HTTP-to-HTTPS Redirect

**File:** `deploy/helm/pebblr/templates/gateway.yaml`

When TLS is enabled, the Gateway has both port 80 (HTTP) and port 443 (HTTPS) listeners. There is no redirect from HTTP to HTTPS. Users hitting port 80 get plaintext responses. The Istio gateway template *does* include `httpsRedirect: true`, but the Envoy Gateway template does not.

**Recommendation:** Add an HTTPRoute that redirects port 80 traffic to HTTPS when TLS is enabled, matching the Istio behavior.

---

### 13. Migration Job Shares the Same Image as the App

**File:** `deploy/helm/pebblr/templates/migration-job.yaml:34`

The migration job uses the full app image (which includes the frontend bundle, static assets, etc.) just to run `/app/migrate`. This means every migration downloads the entire app image. Not critical, but wasteful -- especially if images grow.

**Recommendation:** Consider a multi-target Dockerfile that produces a slim `migrate` image with just the binary and migration SQL files. Not urgent, but good hygiene.

---

### 14. `helm.sh/hook-delete-policy` Drops Failed Jobs

**File:** `deploy/helm/pebblr/templates/migration-job.yaml:11`

```yaml
"helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
```

This keeps failed jobs around (good for debugging) but deletes succeeded ones. That is actually the right default. However, there is no `hook-failed` policy, which means failed migration jobs will accumulate. After enough failed attempts (say, during development), you will have a pile of dead Job objects.

**Recommendation:** Add `ttlSecondsAfterFinished` to auto-clean, or periodically purge. The current policy is acceptable for now.

---

### 15. `automountServiceAccountToken: false` Is Good -- But Double-Check ESO

Both the Deployment and the migration Job set `automountServiceAccountToken: false`. This is excellent security hygiene. However, if Azure Workload Identity needs the projected service account token (which it does for federated credentials), this will silently break ESO's ability to authenticate to KeyVault.

**Recommendation:** Verify that the ServiceAccount token is not needed by the app pod itself. If the app uses Workload Identity directly (e.g., for KeyVault SDK calls), you need `automountServiceAccountToken: true` or a projected volume. If only ESO uses it, you are fine.

---

### 16. `helm-ci-install.sh` Uses `helm install` Instead of `helm template`

**File:** `scripts/helm-ci-install.sh:13`

```bash
helm install "$RELEASE" "$CHART" \
  --namespace "$NAMESPACE" \
  --create-namespace \
  --values "$CI_VALUES" \
  --dry-run \
  --debug
```

`helm install --dry-run` requires a running cluster to validate against the Kubernetes API. `helm template` does not. If the CI job does not have cluster access (or the cluster is down), this validation step fails for the wrong reason.

**Recommendation:** Use `helm template` for pure syntax/rendering validation in CI, and `helm install --dry-run` only when a cluster is available. Or keep both and make the CI target resilient.

---

## LOW -- Nits and Observations

### 17. Dockerfile Uses `node:25-alpine` to Install Bun

**File:** `Dockerfile:19`

```dockerfile
FROM node:25-alpine AS web-builder
RUN npm install -g bun && bun install --frozen-lockfile
```

You pull a Node.js image just to install Bun into it via npm. Use the official `oven/bun:*-alpine` image instead. Smaller layer, faster build, fewer moving parts.

---

### 18. `config/` Directory Copied Into Image but Not in `.dockerignore` Context

**File:** `Dockerfile:45`

```dockerfile
COPY config/ ./config/
```

If `config/` contains tenant-specific configuration, this gets baked into every image. Consider mounting it at runtime instead, or validate that it only contains schema/defaults.

---

### 19. No `.dockerignore` Audit

I did not find a `.dockerignore` file in scope. If one does not exist, the Docker build context includes `web/node_modules/`, `.git/`, and everything else. This slows builds and risks leaking sensitive files into build layers.

---

### 20. Seed Data Contains Hardcoded Passwords in Scripts

**File:** `scripts/cluster-db.sh:19`

```bash
DB_PASSWORD="pebblr-local"
```

And `deploy/k8s/postgres.yaml:10`:

```yaml
POSTGRES_PASSWORD: pebblr-local
```

These are marked as local-only, which is fine. But if the Kind cluster is ever exposed (e.g., via `kubectl proxy` or a misconfigured port-forward), this is a known credential. Add a comment in the postgres.yaml that this is intentionally insecure for local dev only. Already partially documented but could be more prominent.

---

### 21. RLS Policy on `targets` Uses String Comparison for UUID

**File:** `migrations/002_targets.up.sql:31`

```sql
OR assignee_id::TEXT = current_setting('app.user_id', true)
```

This casts `assignee_id` to TEXT for comparison. The `activities` migration (003) does it correctly with `::uuid` cast on the setting side. The inconsistency is not a bug (both work), but the TEXT cast on a UUID column defeats index usage. Fix for consistency and performance.

---

## Summary Table

| # | Severity | Area | Issue |
|---|----------|------|-------|
| 1 | CRITICAL | Helm | Default image tag `latest` with `IfNotPresent` pull policy |
| 2 | CRITICAL | Skaffold | JWT token baked into Docker layers and JS bundle |
| 3 | CRITICAL | Migration | Down migration 009 silently loses soft-deleted rows |
| 4 | CRITICAL | Migration | Down migration 001 can wipe all data with no guard |
| 5 | HIGH | Helm | Migration Job has no `activeDeadlineSeconds` |
| 6 | HIGH | Helm | No `startupProbe` -- slow starts cause crash loops |
| 7 | HIGH | Helm | No `PodDisruptionBudget` |
| 8 | HIGH | K8s | No `NetworkPolicy` anywhere |
| 9 | HIGH | Helm | ExternalSecret uses deprecated `v1beta1` API |
| 10 | HIGH | K8s | Dev PostgreSQL has no liveness probe |
| 11 | HIGH | Helm | `database.port` type inconsistency across values files |
| 12 | MEDIUM | Helm | No HTTP-to-HTTPS redirect on Envoy Gateway |
| 13 | MEDIUM | Docker | Migration job uses full app image |
| 14 | MEDIUM | Helm | No TTL on migration Job cleanup |
| 15 | MEDIUM | K8s | Verify `automountServiceAccountToken: false` vs Workload Identity |
| 16 | MEDIUM | CI | `helm install --dry-run` requires cluster access |
| 17 | LOW | Docker | Using Node image to install Bun |
| 18 | LOW | Docker | `config/` baked into image |
| 19 | LOW | Docker | No `.dockerignore` audit |
| 20 | LOW | Scripts | Hardcoded dev passwords (acceptable for local) |
| 21 | LOW | Migration | RLS UUID comparison inconsistency |

---

## What Is Done Well

Credit where it is due:

- **Security context is locked down.** `runAsNonRoot`, `readOnlyRootFilesystem`, `allowPrivilegeEscalation: false`, `drop: ALL`. This is better than 90% of the Helm charts I review.
- **Secrets as file mounts.** Consistent across values, Dockerfile comments, and deployment template. No env var leakage for secrets.
- **AKS safety guard in Makefile.** The `AKS_GUARD` macro that prevents running local-only targets against AKS is a nice touch. Someone has been burned before.
- **Migration Job as Helm hook.** Pre-upgrade, pre-install, correct hook weight. The migration runs before the app deploys. This is the right pattern.
- **RLS on every table that needs it.** targets, activities, target_collections, target_collection_items, territories. Consistent and correct.
- **Down migrations exist for every up migration.** Many teams skip these. Having them (even imperfect ones) is better than not.
- **`set -euo pipefail` in all scripts.** Basic but so often forgotten.

---

*I have seen clusters die from less. Fix the CRITICAL items before any production traffic touches this.*
