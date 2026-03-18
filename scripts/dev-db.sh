#!/usr/bin/env bash
# dev-db.sh — start a local PostgreSQL 16 dev container, write DSN secret, run migrations, seed data.
set -euo pipefail

CONTAINER_NAME="pebblr-postgres"
PG_VERSION="16"
PG_USER="pebblr"
PG_PASSWORD="pebblr_dev"
PG_DB="pebblr"
PG_PORT="${PG_PORT:-5433}"
DSN_FILE=".local/secrets/db-dsn"

ACTION="${1:-up}"

case "$ACTION" in
  up)
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
      echo "Container '${CONTAINER_NAME}' already exists. Use 'make dev-db-reset' to recreate."
      exit 1
    fi

    echo "Starting PostgreSQL ${PG_VERSION} container..."
    docker run -d \
      --name "$CONTAINER_NAME" \
      -e POSTGRES_USER="$PG_USER" \
      -e POSTGRES_PASSWORD="$PG_PASSWORD" \
      -e POSTGRES_DB="$PG_DB" \
      -p "${PG_PORT}:5432" \
      "postgres:${PG_VERSION}"

    echo "Waiting for PostgreSQL to be ready..."
    for i in $(seq 1 30); do
      if docker exec "$CONTAINER_NAME" pg_isready -U "$PG_USER" -d "$PG_DB" >/dev/null 2>&1; then
        echo "PostgreSQL is ready."
        break
      fi
      sleep 1
      if [ "$i" -eq 30 ]; then
        echo "ERROR: PostgreSQL did not become ready within 30 seconds."
        exit 1
      fi
    done

    mkdir -p "$(dirname "$DSN_FILE")"
    printf 'postgres://%s:%s@localhost:%s/%s?sslmode=disable' \
      "$PG_USER" "$PG_PASSWORD" "$PG_PORT" "$PG_DB" > "$DSN_FILE"
    echo "DSN written to $DSN_FILE"

    echo "Running migrations..."
    go run ./cmd/migrate -dsn-file "$DSN_FILE"

    echo "Seeding data..."
    docker exec -i "$CONTAINER_NAME" \
      psql -U "$PG_USER" -d "$PG_DB" < scripts/seed-data.sql

    echo ""
    echo "Dev database ready."
    echo "  Host:     localhost:${PG_PORT}"
    echo "  Database: ${PG_DB}"
    echo "  DSN file: ${DSN_FILE}"
    ;;

  stop)
    echo "Stopping container '${CONTAINER_NAME}'..."
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm   "$CONTAINER_NAME" 2>/dev/null || true
    rm -f "$DSN_FILE"
    echo "Done."
    ;;

  reset)
    echo "Resetting dev database..."
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm   "$CONTAINER_NAME" 2>/dev/null || true
    exec "$0" up
    ;;

  *)
    echo "Usage: $0 [up|stop|reset]" >&2
    exit 1
    ;;
esac
