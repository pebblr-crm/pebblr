#!/usr/bin/env bash
# Sets up the full E2E environment in a Kind cluster:
#   1. Creates the Kind cluster (via make cluster-up)
#   2. Deploys PostgreSQL into the cluster
#   3. Waits for PostgreSQL to be ready
#   4. Runs migrations and seeds data via a Job
#   5. Creates the app secrets
#   6. Builds and loads the Docker image into Kind
#   7. Deploys the app via Helm
#   8. Waits for the app to be ready
#
# Usage: scripts/e2e-setup.sh
set -euo pipefail

CLUSTER="pebblr-local"
NAMESPACE="pebblr-e2e"
CHART="deploy/helm/pebblr"
E2E_VALUES="deploy/helm/pebblr/values-e2e.yaml"
POSTGRES_MANIFEST="deploy/k8s/e2e/postgres.yaml"
DB_DSN="postgres://pebblr:pebblr-e2e-password@postgres.pebblr-e2e.svc.cluster.local:5432/pebblr?sslmode=disable"
TIMEOUT=120

log() { echo "==> $*"; }
die() { echo "ERROR: $*" >&2; exit 1; }

wait_for_rollout() {
  local resource="$1" ns="$2" timeout="$3"
  log "Waiting for $resource in $ns (timeout ${timeout}s)..."
  kubectl rollout status "$resource" -n "$ns" --timeout="${timeout}s"
}

wait_for_pod_ready() {
  local label="$1" ns="$2" timeout="$3"
  log "Waiting for pod with label $label in $ns..."
  kubectl wait pod -l "$label" -n "$ns" --for=condition=Ready --timeout="${timeout}s"
}

# ── Step 1: Kind cluster ────────────────────────────────────────────────────
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER}$"; then
  log "Kind cluster '$CLUSTER' already exists, reusing."
else
  log "Creating Kind cluster via 'make cluster-up'..."
  make cluster-up
fi

# ── Step 2: Deploy PostgreSQL ───────────────────────────────────────────────
log "Deploying PostgreSQL..."
kubectl apply -f "$POSTGRES_MANIFEST"
wait_for_rollout "deployment/postgres" "$NAMESPACE" "$TIMEOUT"
wait_for_pod_ready "app=postgres" "$NAMESPACE" "$TIMEOUT"

# ── Step 3: Run migrations and seed data ────────────────────────────────────
log "Running migrations and seeding data..."

# Create a ConfigMap from migrations and seed SQL
kubectl create configmap e2e-migrations \
  --from-file=migrations/ \
  --from-file=scripts/seed-data.sql \
  -n "$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -

# Delete previous migration job if it exists (idempotent re-runs).
kubectl delete job e2e-migrate-seed -n "$NAMESPACE" --ignore-not-found

# Run a one-shot Job to apply migrations and seed data
cat <<'JOBEOF' | kubectl apply -n "$NAMESPACE" -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: e2e-migrate-seed
  namespace: pebblr-e2e
spec:
  backoffLimit: 3
  ttlSecondsAfterFinished: 300
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: migrate
          image: postgres:16-alpine
          command:
            - sh
            - -c
            - |
              set -e
              export PGPASSWORD="pebblr-e2e-password"
              PSQL="psql -h postgres.pebblr-e2e.svc.cluster.local -U pebblr -d pebblr"

              echo "Applying migrations..."
              for f in /migrations/*.up.sql; do
                echo "  -> $(basename "$f")"
                $PSQL -f "$f"
              done

              echo "Seeding data..."
              $PSQL -f /seed/seed-data.sql

              echo "Migration and seeding complete."
          volumeMounts:
            - name: migrations
              mountPath: /migrations
            - name: seed
              mountPath: /seed
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 200m
              memory: 128Mi
      volumes:
        - name: migrations
          configMap:
            name: e2e-migrations
            items:
              - key: 001_initial_schema.up.sql
                path: 001_initial_schema.up.sql
              - key: 002_lead_soft_delete.up.sql
                path: 002_lead_soft_delete.up.sql
        - name: seed
          configMap:
            name: e2e-migrations
            items:
              - key: seed-data.sql
                path: seed-data.sql
JOBEOF

log "Waiting for migration job to complete..."
kubectl wait job/e2e-migrate-seed -n "$NAMESPACE" --for=condition=Complete --timeout="${TIMEOUT}s"

# ── Step 4: Create app secrets ──────────────────────────────────────────────
log "Creating app secrets..."
kubectl create secret generic pebblr-e2e-pebblr-secrets \
  --from-literal=db-dsn="$DB_DSN" \
  --from-literal=db-url="$DB_DSN" \
  --from-literal=db-password="pebblr-e2e-password" \
  --from-literal=jwt-secret="e2e-jwt-secret-not-for-production" \
  -n "$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -

# ── Step 5: Build Docker image and load into Kind ──────────────────────────
log "Building Docker image..."
docker build -t pebblr-api:e2e .

log "Loading image into Kind cluster..."
kind load docker-image pebblr-api:e2e --name "$CLUSTER"

# ── Step 6: Deploy the app via Helm ─────────────────────────────────────────
log "Installing pebblr via Helm..."
helm upgrade --install pebblr-e2e "$CHART" \
  --namespace "$NAMESPACE" \
  --values "$E2E_VALUES" \
  --set image.tag=e2e \
  --wait \
  --timeout "${TIMEOUT}s"

wait_for_rollout "deployment/pebblr-e2e-pebblr" "$NAMESPACE" "$TIMEOUT"

# ── Step 7: Verify ──────────────────────────────────────────────────────────
log "E2E environment is ready."
log "Namespace: $NAMESPACE"
kubectl get pods -n "$NAMESPACE"
echo ""
log "Run 'make e2e' to execute the test suite."
