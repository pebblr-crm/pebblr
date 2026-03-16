#!/usr/bin/env bash
# E2E test runner — executes tests in e2e/ against a live Kind cluster.
# Requires a running cluster with the pebblr chart installed.
set -euo pipefail

echo "Running E2E tests..."
go test -v -tags=e2e -count=1 -timeout=10m ./e2e/...
echo "E2E tests complete."
