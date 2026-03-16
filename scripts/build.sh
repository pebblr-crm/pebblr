#!/usr/bin/env bash
# Build Go binary and React frontend.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

echo "==> Building Go binary..."
mkdir -p bin
go build -o bin/api ./cmd/api

echo "==> Building React frontend..."
cd web
bun install --frozen-lockfile
bun run build

echo "==> Build complete."
