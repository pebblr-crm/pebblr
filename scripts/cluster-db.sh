#!/usr/bin/env bash
# Deploy on-cluster PostgreSQL and create app secrets.
# Migrations are handled by the Helm migration job (golang-migrate).
# Works for both local dev (pebblr namespace) and e2e (pebblr-e2e namespace).
#
# PostgreSQL runs in a shared "pebblr-db" namespace so both dev and e2e
# namespaces share a single instance. Each app namespace gets its own
# Kubernetes Secret pointing at the shared database.
#
# Usage: scripts/cluster-db.sh <namespace> [up|stop|reset]
set -euo pipefail

NAMESPACE="${1:?Usage: $0 <namespace> [up|stop|reset]}"
ACTION="${2:-up}"

DB_NAMESPACE="pebblr-db"
POSTGRES_MANIFEST="deploy/k8s/postgres.yaml"
DB_PASSWORD="pebblr-local"
DB_HOST="postgres.${DB_NAMESPACE}.svc.cluster.local"
DB_DSN="postgres://pebblr:${DB_PASSWORD}@${DB_HOST}:5432/pebblr?sslmode=disable"
TIMEOUT=120

log() { echo "==> $*"; }

ensure_postgres() {
  # Deploy PostgreSQL into the shared namespace (idempotent).
  kubectl create namespace "$DB_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

  if kubectl get deployment/postgres -n "$DB_NAMESPACE" &>/dev/null; then
    log "PostgreSQL already running in ${DB_NAMESPACE}."
  else
    log "Deploying PostgreSQL into ${DB_NAMESPACE}..."
    kubectl apply -f "$POSTGRES_MANIFEST" -n "$DB_NAMESPACE"
  fi

  kubectl rollout status deployment/postgres -n "$DB_NAMESPACE" --timeout="${TIMEOUT}s"
  kubectl wait pod -l app=postgres -n "$DB_NAMESPACE" --for=condition=Ready --timeout="${TIMEOUT}s"
}

create_secret() {
  # Secret name must match Helm's {{ fullname }}-secrets.
  # For release "pebblr" in ns "pebblr" → "pebblr-secrets".
  # For release "pebblr-e2e" in ns "pebblr-e2e" → "pebblr-e2e-secrets".
  local secret_name
  if [ "$NAMESPACE" = "pebblr" ]; then
    secret_name="pebblr-secrets"
  else
    secret_name="${NAMESPACE}-secrets"
  fi

  kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

  kubectl create secret generic "$secret_name" \
    --from-literal=db-dsn="$DB_DSN" \
    --from-literal=db-url="$DB_DSN" \
    --from-literal=db-password="$DB_PASSWORD" \
    --from-literal=jwt-secret="local-jwt-secret-not-for-production" \
    -n "$NAMESPACE" \
    --dry-run=client -o yaml | kubectl apply -f -
}

do_up() {
  ensure_postgres
  create_secret
  log "Database ready in namespace ${NAMESPACE} (postgres in ${DB_NAMESPACE})."
}

do_stop() {
  log "Removing PostgreSQL from ${DB_NAMESPACE}..."
  kubectl delete -f "$POSTGRES_MANIFEST" -n "$DB_NAMESPACE" --ignore-not-found
  log "Done."
}

case "$ACTION" in
  up)    do_up ;;
  stop)  do_stop ;;
  reset) do_stop; do_up ;;
  *)     echo "Usage: $0 <namespace> [up|stop|reset]" >&2; exit 1 ;;
esac
