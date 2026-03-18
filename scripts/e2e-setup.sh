#!/usr/bin/env bash
# Sets up the full E2E environment in a Kind cluster.
# Orchestrates the composable Makefile targets:
#   e2e-cluster → e2e-db → e2e-image → e2e-deploy
#
# Usage: scripts/e2e-setup.sh
set -euo pipefail

CLUSTER="pebblr-local"

log() { echo "==> $*"; }

# ── Step 1: Kind cluster ────────────────────────────────────────────────────
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER}$"; then
  log "Kind cluster '$CLUSTER' already exists, reusing."
else
  log "Creating Kind cluster..."
  make e2e-cluster
fi

# ── Step 2: Database setup ──────────────────────────────────────────────────
log "Setting up database..."
make e2e-db

# ── Step 3: Docker image ───────────────────────────────────────────────────
log "Building and loading Docker image..."
make e2e-image

# ── Step 4: Deploy the app ──────────────────────────────────────────────────
log "Deploying app via Helm..."
make e2e-deploy

# ── Done ────────────────────────────────────────────────────────────────────
log "E2E environment is ready."
kubectl get pods -n pebblr-e2e
echo ""
log "Run 'make e2e' to execute the test suite."
