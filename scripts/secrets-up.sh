#!/usr/bin/env bash
# Apply ExternalSecret resources for test tenant.
# Requires the cluster to be running and ESO installed.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NAMESPACE="${NAMESPACE:-pebblr}"

echo "==> Creating namespace: $NAMESPACE"
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

echo "==> Applying ExternalSecret resources..."
kubectl apply -f "$REPO_ROOT/deploy/helm/pebblr/templates/externalsecret.yaml" -n "$NAMESPACE" || \
  echo "Note: ExternalSecret CRD resources are managed by Helm. Run 'make deploy' to install."

echo "==> Secrets configuration applied."
