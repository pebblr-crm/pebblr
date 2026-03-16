#!/usr/bin/env bash
# Create Kind cluster, install ESO, apply base config.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
KIND_CONFIG="$REPO_ROOT/deploy/kind/kind-config.yaml"
CLUSTER_NAME="${CLUSTER_NAME:-pebblr-local}"

echo "==> Creating Kind cluster: $CLUSTER_NAME"
kind create cluster --name "$CLUSTER_NAME" --config "$KIND_CONFIG"

echo "==> Installing External Secrets Operator..."
helm repo add external-secrets https://charts.external-secrets.io
helm repo update
helm install external-secrets external-secrets/external-secrets \
  --namespace external-secrets-operator \
  --create-namespace \
  --wait

echo "==> Cluster ready: $CLUSTER_NAME"
