#!/usr/bin/env bash
# scripts/e2e-web.sh — Run Playwright integration tests against the Kind cluster.
#
# Prerequisites:
#   make e2e-cluster && make e2e-db && make e2e-deploy
#
# This script:
#   1. Port-forwards the pebblr-e2e service to a free local port
#   2. Waits for /healthz to return 200
#   3. Runs Playwright with playwright-integration.config.ts
#   4. Cleans up the port-forward on exit

set -euo pipefail

NAMESPACE="${E2E_NAMESPACE:-pebblr-e2e}"
SERVICE="${E2E_SERVICE:-svc/pebblr-e2e}"
SERVICE_PORT="${E2E_SERVICE_PORT:-8080}"
TIMEOUT_SECS="${E2E_WAIT_TIMEOUT:-60}"

cleanup() {
  if [[ -n "${PF_PID:-}" ]]; then
    kill "$PF_PID" 2>/dev/null || true
    wait "$PF_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

# Find a free local port.
LOCAL_PORT=$(python3 -c 'import socket; s=socket.socket(); s.bind(("",0)); print(s.getsockname()[1]); s.close()')

echo "==> Port-forwarding ${SERVICE} (${NAMESPACE}) → 127.0.0.1:${LOCAL_PORT}"
kubectl port-forward "${SERVICE}" "${LOCAL_PORT}:${SERVICE_PORT}" -n "${NAMESPACE}" &>/dev/null &
PF_PID=$!

# Wait for the app to respond.
echo "==> Waiting for /healthz (timeout ${TIMEOUT_SECS}s)..."
deadline=$((SECONDS + TIMEOUT_SECS))
while (( SECONDS < deadline )); do
  if curl -sf "http://127.0.0.1:${LOCAL_PORT}/healthz" >/dev/null 2>&1; then
    echo "==> App is ready at http://127.0.0.1:${LOCAL_PORT}"
    break
  fi
  sleep 0.5
done

if ! curl -sf "http://127.0.0.1:${LOCAL_PORT}/healthz" >/dev/null 2>&1; then
  echo "ERROR: App did not become ready within ${TIMEOUT_SECS}s" >&2
  kubectl get pods -n "${NAMESPACE}" >&2
  exit 1
fi

echo "==> Verifying SPA is served..."
HTTP_STATUS=$(curl -s -o /dev/null -w '%{http_code}' -L "http://127.0.0.1:${LOCAL_PORT}/")
if [[ "$HTTP_STATUS" != "200" ]]; then
  echo "WARNING: / returned HTTP ${HTTP_STATUS} (expected 200)"
fi

# Run Playwright.
echo "==> Running Playwright integration tests..."
cd "$(dirname "$0")/../web"

E2E_BASE_URL="http://127.0.0.1:${LOCAL_PORT}" \
  bunx playwright test --config=playwright-integration.config.ts "$@"
