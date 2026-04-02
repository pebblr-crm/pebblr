# Infrastructure Review: The Idiomatic Obsessive

Reviewer persona: idiomatic correctness for Helm 4, Kubernetes labels,
migration naming, Makefile conventions, and shell scripts.

Scope: `deploy/`, `migrations/`, `Makefile`, `scripts/`, `skaffold.yaml`

---

## Changes Applied (this PR)

### Helm Chart

| File | Change | Why |
|---|---|---|
| `_helpers.tpl` | Added `app.kubernetes.io/part-of: pebblr` to common labels | Kubernetes recommended label schema requires `part-of` to group components of a distributed application |
| `deployment.yaml` | Pod template now uses `pebblr.labels` (not just `selectorLabels`) + adds `app.kubernetes.io/component: api` | Pods should carry the full label set for observability; `component` distinguishes workloads within the chart |
| `deployment.yaml` | Wrapped `podAnnotations` in `{{- with }}` guard | Previously rendered an empty `annotations: {}` block when no annotations were set |
| `externalsecret.yaml` | `v1beta1` -> `v1` | ESO graduated the `ExternalSecret` CRD to `v1` in 0.9.x; the chart pins ESO 0.12.1 |
| `externalsecret-origin-tls.yaml` | `v1beta1` -> `v1` | Same rationale |
| `ingress.yaml` -> `httproute.yaml` | Renamed file | The file contains a Gateway API `HTTPRoute`, not a `networking.k8s.io/v1` Ingress. Naming the file `ingress.yaml` is misleading |
| `Chart.yaml` | Added `kubeVersion: ">= 1.28.0-0"` | Documents minimum K8s version; `helm install` will fail-fast on unsupported clusters |
| `migration-job.yaml` | Added `trunc 63` to Job name | The name pattern `{fullname}-migrate-{rev}-{timestamp}` can exceed the 63-char DNS label limit on long release names |
| `values-aks.yaml` | Added TODO comment about `externalSecrets.enabled: false` | The file header claims ESO/KeyVault integration, but ESO is disabled -- likely a WIP state that should be flagged |

### Kubernetes Manifests

| File | Change | Why |
|---|---|---|
| `deploy/k8s/postgres.yaml` | Replaced bare `app: postgres` labels with `app.kubernetes.io/name`, `app.kubernetes.io/component`, `app.kubernetes.io/part-of` | Follows the Kubernetes recommended label schema; bare `app` labels are a legacy pattern |
| `deploy/k8s/postgres.yaml` | Added `name: postgresql` to container port, used named port in Service `targetPort` | Named ports are the idiomatic K8s pattern; they decouple port numbers from Service definitions |

### Scripts

| File | Change | Why |
|---|---|---|
| `scripts/cluster-db.sh` | Updated pod selector from `app=postgres` to `app.kubernetes.io/name=postgres` | Matches the updated postgres.yaml labels |
| `scripts/seed.sh` | Same selector update | Same rationale |
| `scripts/helm-ci-install.sh` | Set executable permission (`chmod +x`) | Script is called from Makefile but lacked execute bit |

### Makefile

| Change | Why |
|---|---|
| Split monolithic `.PHONY` into grouped declarations | Improves readability; groups related targets (build, dev, e2e, validation) |

---

## Observations (not changed, for future consideration)

### Helm

1. **No `HorizontalPodAutoscaler` template.** `values.yaml` defines `autoscaling.enabled/minReplicas/maxReplicas/targetCPU` but there is no corresponding HPA template. The `{{- if not .Values.autoscaling.enabled }}` guard on `replicas:` is dead code until an HPA template exists.

2. **Istio and Gateway API coexist.** The chart ships templates for both Istio (`istio-gateway.yaml`, `istio-virtualservice.yaml`) and Gateway API (`gateway.yaml`, `httproute.yaml`). Both are guarded by their own feature flags, which is fine, but there is no mutual exclusion. A user could enable both `gateway.enabled` and `istio.enabled` and get conflicting routing. Consider adding a validation check in `_helpers.tpl` or NOTES.txt.

3. **No `.helmignore` file.** The chart directory should have a `.helmignore` to exclude `values-*.yaml` environment overlays, `README.md`, etc. from the packaged chart artifact.

4. **NOTES.txt does not handle the Istio path.** It only shows Gateway API and ClusterIP access instructions. When `istio.enabled` is true, the user gets no instructions.

### Migrations

5. **Naming convention is consistent and correct.** The `NNN_description.{up,down}.sql` pattern matches golang-migrate's expected format. Sequential numbering (001-009) is clean. No gaps, no timestamp-based naming conflicts.

6. **`IF NOT EXISTS` / `IF EXISTS` discipline is good.** Up migrations use `CREATE TABLE IF NOT EXISTS`; down migrations use `DROP TABLE IF EXISTS`. This makes migrations re-entrant in edge cases.

7. **No transaction wrapping.** golang-migrate runs each file in its own transaction by default for PostgreSQL, so this is fine. Just documenting: the migrations rely on the tool's implicit transaction boundary.

### Makefile

8. **`AKS_GUARD` is a recipe-level variable expanded via `$(AKS_GUARD)`.** This works but is unconventional. The more idiomatic Makefile pattern would be a guard target that other targets depend on (e.g., `_guard-not-aks:`). However, the current approach is simpler for a flat Makefile and works correctly as-is.

9. **No `check` or `ci` meta-target.** The quality gates documented in CLAUDE.md (`make test && make lint && make typecheck`) could be rolled into a single `make check` target. This avoids human error in CI pipelines and matches the CLAUDE.md convention.

### Skaffold

10. **`VITE_STATIC_TOKEN` build arg in skaffold.yaml.** The default profile passes `VITE_STATIC_TOKEN: "local-jwt-secret-not-for-production"` as a Docker build arg. This bakes a token into the image at build time. If the image is ever pushed to a registry, the token goes with it. For local-only builds this is acceptable, but worth noting.

11. **The `kind` profile duplicates the full Helm release spec.** Skaffold profiles support strategic merge patching -- only the changed fields need to appear. The duplication means `chartPath`, `namespace`, `createNamespace`, and `setValueTemplates` are repeated verbatim. If the base release changes, the profile could drift.

### Shell Scripts

12. **`scripts/helm-ci-install.sh` uses `helm install` not `helm upgrade --install`.** If the CI job is re-run against the same cluster without cleanup, the install will fail because the release already exists. The `--dry-run` flag prevents actual creation, so in practice this is fine, but `upgrade --install` is the defensive choice.

---

## Summary

The infrastructure is well-structured for an early-stage project. The Helm chart
follows most best practices: proper use of `_helpers.tpl`, named templates for
labels/selectors/names, security contexts, resource limits, and probes. The
migration naming is clean and consistent. The Makefile correctly delegates
complex logic to scripts.

The changes in this PR address concrete idiomatic issues: deprecated API
versions, missing standard labels, misleading file names, and DNS label length
safety. The observations section flags architectural items for future
consideration.
