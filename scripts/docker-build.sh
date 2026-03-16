#!/usr/bin/env bash
# Build and tag Docker images.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

IMAGE_NAME="${IMAGE_NAME:-pebblr-api}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
FULL_TAG="${IMAGE_NAME}:${IMAGE_TAG}"

echo "==> Building Docker image: $FULL_TAG"
docker build -t "$FULL_TAG" .

echo "==> Image built: $FULL_TAG"

if [[ "${LOAD_KIND:-false}" == "true" ]]; then
  CLUSTER_NAME="${CLUSTER_NAME:-pebblr-local}"
  echo "==> Loading image into Kind cluster: $CLUSTER_NAME"
  kind load docker-image "$FULL_TAG" --name "$CLUSTER_NAME"
fi
