#!/usr/bin/env bash
# Seed the on-cluster PostgreSQL with sample data.
# Runs seed-data.sql via kubectl exec against the postgres pod in pebblr-db namespace.
#
# Usage: scripts/seed.sh
set -euo pipefail

DB_NAMESPACE="pebblr-db"
SEED_FILE="scripts/seed-data.sql"

log() { echo "==> $*"; }

POD=$(kubectl get pod -n "$DB_NAMESPACE" -l app.kubernetes.io/name=postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
if [ -z "$POD" ]; then
  echo "ERROR: No postgres pod found in namespace ${DB_NAMESPACE}. Run 'make dev-db' first." >&2
  exit 1
fi

log "Seeding database via ${POD} in ${DB_NAMESPACE}..."
kubectl exec -n "$DB_NAMESPACE" "$POD" -i -- psql -U pebblr -d pebblr < "$SEED_FILE"
log "Seed data loaded."
