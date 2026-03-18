#!/usr/bin/env bash
# Deploy on-cluster PostgreSQL, run migrations, seed data, and create app secrets.
# Works for both local dev (pebblr namespace) and e2e (pebblr-e2e namespace).
#
# Usage: scripts/cluster-db.sh <namespace> [up|stop|reset]
set -euo pipefail

NAMESPACE="${1:?Usage: $0 <namespace> [up|stop|reset]}"
ACTION="${2:-up}"

POSTGRES_MANIFEST="deploy/k8s/postgres.yaml"
DB_PASSWORD="pebblr-local"
DB_HOST="postgres.${NAMESPACE}.svc.cluster.local"
DB_DSN="postgres://pebblr:${DB_PASSWORD}@${DB_HOST}:5432/pebblr?sslmode=disable"
TIMEOUT=120

log() { echo "==> $*"; }

do_up() {
  # ── Namespace ───────────────────────────────────────────────────────────
  kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

  # ── Deploy PostgreSQL ───────────────────────────────────────────────────
  log "Deploying PostgreSQL into ${NAMESPACE}..."
  kubectl apply -f "$POSTGRES_MANIFEST" -n "$NAMESPACE"
  kubectl rollout status deployment/postgres -n "$NAMESPACE" --timeout="${TIMEOUT}s"
  kubectl wait pod -l app=postgres -n "$NAMESPACE" --for=condition=Ready --timeout="${TIMEOUT}s"

  # ── Run migrations and seed data ────────────────────────────────────────
  log "Running migrations and seeding data..."

  kubectl create configmap cluster-db-migrations \
    --from-file=migrations/ \
    --from-file=scripts/seed-data.sql \
    -n "$NAMESPACE" \
    --dry-run=client -o yaml | kubectl apply -f -

  kubectl delete job cluster-db-migrate-seed -n "$NAMESPACE" --ignore-not-found

  cat <<JOBEOF | kubectl apply -n "$NAMESPACE" -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: cluster-db-migrate-seed
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
              export PGPASSWORD="${DB_PASSWORD}"
              PSQL="psql -h ${DB_HOST} -U pebblr -d pebblr"

              echo "Applying migrations..."
              for f in /data/*.up.sql; do
                echo "  -> \$(basename "\$f")"
                \$PSQL -f "\$f"
              done

              echo "Seeding data..."
              \$PSQL -f /data/seed-data.sql

              echo "Migration and seeding complete."
          volumeMounts:
            - name: data
              mountPath: /data
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 200m
              memory: 128Mi
      volumes:
        - name: data
          configMap:
            name: cluster-db-migrations
JOBEOF

  log "Waiting for migration job to complete..."
  kubectl wait job/cluster-db-migrate-seed -n "$NAMESPACE" \
    --for=condition=Complete --timeout="${TIMEOUT}s"

  # ── Create app secrets ──────────────────────────────────────────────────
  log "Creating app secrets..."
  # Secret name must match Helm's {{ fullname }}-secrets.
  # For release "pebblr" in ns "pebblr" → "pebblr-secrets".
  # For release "pebblr-e2e" in ns "pebblr-e2e" → "pebblr-e2e-secrets".
  local secret_name
  if [ "$NAMESPACE" = "pebblr" ]; then
    secret_name="pebblr-secrets"
  else
    secret_name="${NAMESPACE}-secrets"
  fi

  kubectl create secret generic "$secret_name" \
    --from-literal=db-dsn="$DB_DSN" \
    --from-literal=db-url="$DB_DSN" \
    --from-literal=db-password="$DB_PASSWORD" \
    --from-literal=jwt-secret="local-jwt-secret-not-for-production" \
    -n "$NAMESPACE" \
    --dry-run=client -o yaml | kubectl apply -f -

  log "Database ready in namespace ${NAMESPACE}."
}

do_stop() {
  log "Removing PostgreSQL from ${NAMESPACE}..."
  kubectl delete -f "$POSTGRES_MANIFEST" -n "$NAMESPACE" --ignore-not-found
  kubectl delete job cluster-db-migrate-seed -n "$NAMESPACE" --ignore-not-found
  kubectl delete configmap cluster-db-migrations -n "$NAMESPACE" --ignore-not-found
  log "Done."
}

case "$ACTION" in
  up)    do_up ;;
  stop)  do_stop ;;
  reset) do_stop; do_up ;;
  *)     echo "Usage: $0 <namespace> [up|stop|reset]" >&2; exit 1 ;;
esac
