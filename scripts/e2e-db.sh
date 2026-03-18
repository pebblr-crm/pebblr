#!/usr/bin/env bash
# Deploys PostgreSQL, runs migrations, seeds data, and creates app secrets
# in the pebblr-e2e namespace. Assumes a Kind cluster already exists.
#
# Usage: scripts/e2e-db.sh
set -euo pipefail

NAMESPACE="pebblr-e2e"
POSTGRES_MANIFEST="deploy/k8s/e2e/postgres.yaml"
DB_DSN="postgres://pebblr:pebblr-e2e-password@postgres.pebblr-e2e.svc.cluster.local:5432/pebblr?sslmode=disable"
TIMEOUT=120

log() { echo "==> $*"; }

# ── Deploy PostgreSQL ─────────────────────────────────────────────────────
log "Deploying PostgreSQL..."
kubectl apply -f "$POSTGRES_MANIFEST"
kubectl rollout status deployment/postgres -n "$NAMESPACE" --timeout="${TIMEOUT}s"
kubectl wait pod -l app=postgres -n "$NAMESPACE" --for=condition=Ready --timeout="${TIMEOUT}s"

# ── Run migrations and seed data ──────────────────────────────────────────
log "Running migrations and seeding data..."

kubectl create configmap e2e-migrations \
  --from-file=migrations/ \
  --from-file=scripts/seed-data.sql \
  -n "$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -

# Delete previous migration job if it exists (idempotent re-runs).
kubectl delete job e2e-migrate-seed -n "$NAMESPACE" --ignore-not-found

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

# ── Create app secrets ────────────────────────────────────────────────────
log "Creating app secrets..."
kubectl create secret generic pebblr-e2e-pebblr-secrets \
  --from-literal=db-dsn="$DB_DSN" \
  --from-literal=db-url="$DB_DSN" \
  --from-literal=db-password="pebblr-e2e-password" \
  --from-literal=jwt-secret="e2e-jwt-secret-not-for-production" \
  -n "$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -

log "Database setup complete."
