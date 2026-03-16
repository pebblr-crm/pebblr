#!/usr/bin/env bash
# Start full local dev stack: postgres in cluster, API server, Vite dev server.
# Requires Kind cluster to be running (make cluster-up).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLUSTER_NAME="${CLUSTER_NAME:-pebblr-local}"
API_PORT="${API_PORT:-8080}"
WEB_PORT="${WEB_PORT:-5173}"

# Check cluster is up
if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  echo "Error: Kind cluster '$CLUSTER_NAME' not found. Run 'make cluster-up' first."
  exit 1
fi

echo "==> Starting port-forward for API (localhost:$API_PORT)..."
kubectl port-forward svc/pebblr -n pebblr "$API_PORT:8080" &
PF_PID=$!

cleanup() {
  echo "==> Stopping port-forward..."
  kill "$PF_PID" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

echo "==> Starting Vite dev server on port $WEB_PORT..."
cd "$REPO_ROOT/web"
bun run dev --port "$WEB_PORT"
