# Infrastructure Review -- Clean Code Geek

Reviewed: `deploy/`, `migrations/`, `scripts/`, `Makefile`, `skaffold.yaml`, `Dockerfile`

## Changes Made (this PR)

### 1. DRY: Extracted shared Helm template helpers

**Files:** `_helpers.tpl`, `deployment.yaml`, `migration-job.yaml`

The Deployment and migration Job duplicated five blocks of identical YAML:
- Pod security context (`runAsNonRoot`, `runAsUser`, `fsGroup`)
- Container security context (`allowPrivilegeEscalation`, `readOnlyRootFilesystem`, `capabilities`)
- Secrets volume definition
- Secrets volume mount
- Image spec (`image` + `imagePullPolicy`)

Extracted into named templates: `pebblr.podSecurityContext`, `pebblr.containerSecurityContext`, `pebblr.secretsVolume`, `pebblr.secretsVolumeMount`, `pebblr.image`.

Now a single change to security policy or mount config propagates automatically to both the app container and the migration job.

### 2. Naming: Renamed `ingress.yaml` to `httproute.yaml`

The file contained a `gateway.networking.k8s.io/v1 HTTPRoute`, not a `networking.k8s.io/v1 Ingress`. The old name was misleading -- anyone scanning the template directory would expect a classic Ingress resource. The filename now matches the resource kind it contains.

### 3. Convention: Extracted `cluster-deps` Makefile target to script

**Files:** `Makefile`, `scripts/cluster-deps.sh`

The `cluster-deps` target inlined 8 commands with helm repo management, three separate `helm upgrade --install` invocations, and a `kubectl apply`. This violates the project convention:

> Multi-step or complex logic (more than ~2 commands) must be extracted into a script under `scripts/`, then called from the Makefile target.

Extracted to `scripts/cluster-deps.sh`. The Makefile passes pinned versions via environment variables so the script and Makefile stay in sync without duplicating version constants.

### 4. Fix: Type inconsistency in `values-e2e.yaml`

`database.port` was `"5432"` (string) in `values-e2e.yaml` but `5432` (integer) everywhere else. Fixed to integer for consistency.

### 5. Fix: Guard empty `podAnnotations` in deployment.yaml

When overlay values set `podAnnotations: {}` (as `values-ci.yaml` and `values-e2e.yaml` do), the template rendered an empty `annotations:` block. Wrapped with `{{- with .Values.podAnnotations }}` so the annotations key is omitted entirely when empty.

---

## Observations (no changes, for discussion)

### A. Skaffold profile repetition

The `kind` and `e2e` profiles in `skaffold.yaml` must redeclare the full Helm release block (chartPath, namespace, setValueTemplates) even when only adding valuesFiles. This is a Skaffold limitation, not a code smell -- there is no inheritance mechanism for profile overrides in Skaffold v2.

### B. Values overlay duplication (CI, E2E, local)

`values-ci.yaml`, `values-e2e.yaml`, and `values-local.yaml` all repeat the same three-line stanza:

```yaml
externalSecrets:
  enabled: false
serviceAccount:
  annotations: {}
podAnnotations: {}
```

A shared `values-non-prod.yaml` base could eliminate this, with each overlay importing it. However, Helm does not support values file inheritance natively -- you would need to chain `-f` flags in the Skaffold/Makefile invocations. Worth considering if overlays continue to grow.

### C. Migration naming is clean

The `NNN_description.{up,down}.sql` convention is followed consistently across all 9 migrations. Each migration has a clear, descriptive name. The down migrations correctly reverse operations in dependency order. No issues here.

### D. Seed data organization

`scripts/seed-data.sql` is well-structured with section headers and deterministic UUIDs. The `ON CONFLICT DO NOTHING` pattern makes it idempotent. The only minor concern is the file's length (400+ lines) -- if more entity types are added, consider splitting into per-table files loaded by `seed.sh`.

### E. Dockerfile is solid

The multi-stage build is clean: Go builder, web builder, minimal runtime image. Non-root user, read-only FS comment about secrets. One minor observation: the web-builder stage installs bun via npm (`npm install -g bun`) because the base is `node:25-alpine`. An `oven/bun` base image would be marginally cleaner, though the current approach works fine.

### F. Migration Job hardcodes resource limits

The migration job in `migration-job.yaml` hardcodes `cpu: 100m/200m` and `memory: 64Mi/128Mi` rather than using values from `values.yaml`. This is likely intentional (migrations are lightweight), but if a migration ever needs more memory (e.g., a large data backfill like 009), the only way to increase it is to edit the template directly. Consider adding `migrationResources` to values.yaml.
