#!/usr/bin/env bash
# Deploy to local Kind cluster, passing build-time env vars from web/.env.development.
set -euo pipefail

ENV_FILE="web/.env.development"

# Read VITE_GOOGLE_MAPS_API_KEY from the env file (empty string if missing).
MAPS_KEY=""
if [ -f "$ENV_FILE" ]; then
  MAPS_KEY=$(grep -s '^VITE_GOOGLE_MAPS_API_KEY=' "$ENV_FILE" | cut -d= -f2- || true)
fi

export VITE_GOOGLE_MAPS_API_KEY="${MAPS_KEY}"

exec skaffold run --default-repo=""
