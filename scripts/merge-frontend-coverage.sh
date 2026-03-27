#!/usr/bin/env bash
# Merge unit-test (vitest) and E2E (Playwright/Istanbul) LCOV reports
# into a single web/coverage/lcov.info for SonarCloud.
set -euo pipefail

WEB_DIR="$(cd "$(dirname "$0")/../web" && pwd)"
UNIT_LCOV="$WEB_DIR/coverage/lcov.info"
E2E_LCOV="$WEB_DIR/coverage-e2e/lcov.info"
MERGED="$WEB_DIR/coverage/lcov.info"

# If only unit coverage exists, nothing to merge
if [ ! -f "$E2E_LCOV" ]; then
  echo "No E2E coverage found; using unit coverage only."
  exit 0
fi

# If only E2E coverage exists, copy it
if [ ! -f "$UNIT_LCOV" ]; then
  echo "No unit coverage found; using E2E coverage only."
  mkdir -p "$WEB_DIR/coverage"
  cp "$E2E_LCOV" "$MERGED"
  exit 0
fi

# Both exist — concatenate (SonarCloud merges overlapping files automatically)
echo "Merging unit + E2E coverage reports..."
cat "$UNIT_LCOV" "$E2E_LCOV" > "${MERGED}.tmp"
mv "${MERGED}.tmp" "$MERGED"
echo "Merged coverage written to $MERGED"
