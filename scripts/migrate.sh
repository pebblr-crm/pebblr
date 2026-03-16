#!/usr/bin/env bash
# Run database migrations against local or remote DB.
# The API binary includes a 'migrate' subcommand.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

DB_SECRET_PATH="${DB_SECRET_PATH:-/run/secrets/db-url}"

if [[ ! -f "$DB_SECRET_PATH" ]]; then
  echo "Error: DB secret not found at $DB_SECRET_PATH"
  echo "In local dev, set DB_SECRET_PATH to a file containing the connection string."
  exit 1
fi

echo "==> Running migrations..."
./bin/api migrate --db-secret "$DB_SECRET_PATH"
echo "==> Migrations complete."
