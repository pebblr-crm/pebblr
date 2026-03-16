#!/usr/bin/env bash
# Destroy Kind cluster.
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-pebblr-local}"

echo "==> Deleting Kind cluster: $CLUSTER_NAME"
kind delete cluster --name "$CLUSTER_NAME"
echo "==> Cluster deleted."
