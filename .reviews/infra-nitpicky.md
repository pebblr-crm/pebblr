# Infrastructure Review: The Nitpicky Ninny

Reviewer: The Nitpicky Ninny
Scope: `deploy/`, `migrations/`, `Makefile`, `scripts/`, `skaffold.yaml`
Date: 2026-04-02

---

## Fixed in this PR

### Makefile

| File | Issue | Fix |
|---|---|---|
| `Makefile` L13-16 | `:=` alignment inconsistent in Paths section (`GO_CMD` padded to 5 spaces, `KIND_CFG` to 1) | Aligned all four variables to consistent spacing |

### Helm values files - YAML type and quoting consistency

| File | Issue | Fix |
|---|---|---|
| `values-e2e.yaml` L14 | `database.port: "5432"` is a quoted string; `values.yaml` defines it as integer `5432` | Removed quotes to match `values.yaml` type |
| `values-local.yaml` L17-19 | Auth placeholder values unquoted (`local-placeholder`); `values-ci.yaml` quotes them (`"ci-placeholder"`) | Added quotes for consistency across all values files |
| `values-e2e.yaml` L20-22 | Same unquoted auth placeholders | Added quotes |

### Helm templates

| File | Issue | Fix |
|---|---|---|
| `templates/ingress.yaml` | File is named "ingress" but contains a `gateway.networking.k8s.io/v1 HTTPRoute` resource, not an Ingress | Renamed to `templates/httproute.yaml` |
| `templates/externalsecret-origin-tls.yaml` L14 | Uses `ClusterSecretStore` while `externalsecret.yaml` uses `SecretStore` with no comment explaining why | Added comment explaining the cross-namespace requirement |

### Scripts

| File | Issue | Fix |
|---|---|---|
| `scripts/helm-ci-install.sh` | Missing executable permission (`-rw-r--r--` vs `-rwxr-xr-x` for all other scripts) | `chmod +x` |

### Migrations

| File | Issue | Fix |
|---|---|---|
| `migrations/005_activity_label.up.sql` | Only `.up.sql` file without a leading `--` comment describing the migration | Added descriptive comment |

---

## Advisory (not fixed - requires team discussion)

### Migrations: inconsistent IF NOT EXISTS guards

Migrations `001` and `002` use `CREATE TABLE IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS` throughout. Migrations `003` through `009` use plain `CREATE TABLE` and `CREATE INDEX` without the idempotency guard. Since golang-migrate tracks applied versions and will not re-run a migration, the `IF NOT EXISTS` is technically unnecessary -- but the inconsistency signals that the convention was never decided. Pick one and stick with it.

Files affected:
- `001_initial_schema.up.sql` -- uses `IF NOT EXISTS`
- `002_targets.up.sql` -- uses `IF NOT EXISTS`
- `003_activities.up.sql` -- does NOT use `IF NOT EXISTS`
- `004_audit_log.up.sql` -- does NOT
- `006_target_collections.up.sql` -- does NOT
- `007_territories.up.sql` -- does NOT

### skaffold.yaml: profile structure inconsistency

The `kind` profile (line 32) has an `activation` block for automatic context-based activation. The `e2e` profile (line 54) does not. This is likely intentional (e2e is explicitly invoked via `skaffold run -p e2e`) but the asymmetry could confuse new contributors. A comment on the `e2e` profile explaining the lack of auto-activation would help.

### Helm templates: deployment.yaml always emits annotations key

`deployment.yaml` line 17-18 always renders the `annotations:` key on the pod template metadata, even when `podAnnotations` is an empty map `{}`. This produces valid YAML but emits a no-op `annotations: {}` line in the rendered manifest. Consider guarding with `{{- if .Values.podAnnotations }}`.

### Helm templates: ExternalSecret API version

Both `externalsecret.yaml` and `externalsecret-origin-tls.yaml` use `external-secrets.io/v1beta1`. The stable `v1` API has been available since ESO 0.10.0 and this project pins ESO 0.12.1. Consider upgrading to `external-secrets.io/v1`.

### deploy/k8s/postgres.yaml: label convention mismatch

The standalone PostgreSQL manifest uses bare `app: postgres` labels. The Helm chart templates use the standard `app.kubernetes.io/*` label set. While the postgres manifest is not part of the Helm release, using `app.kubernetes.io/name: postgres` and `app.kubernetes.io/component: database` would align the label convention across all Kubernetes resources in the project.

### Scripts: inconsistent logging style

`cluster-db.sh` and `seed.sh` define a `log()` helper (`echo "==> $*"`). `deploy-local.sh` has no logging. `e2e-web.sh` uses inline `echo "==> ..."` without a helper. `helm-ci-install.sh` uses bare `echo`. Consider extracting a shared logging function or at least using the same `echo "==> ..."` prefix consistently.

### Makefile: monolithic .PHONY line

Line 5 declares all phony targets in a single 300+ character line. This is hard to scan and will cause noisy diffs when targets are added. Consider splitting into per-section `.PHONY` declarations adjacent to their targets, or at minimum one target per line with backslash continuation.
