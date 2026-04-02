#!/usr/bin/env bash
# Validates the Helm chart against the running Kind cluster using --dry-run.
# Renders all templates and validates them against the live Kubernetes API without
# creating any actual resources.
set -euo pipefail

NAMESPACE="pebblr-e2e"
RELEASE="pebblr-ci"
CHART="deploy/helm/pebblr"
CI_VALUES="deploy/helm/pebblr/values-ci.yaml"

echo "Validating Helm chart install (dry-run against Kind cluster)..."
helm install "$RELEASE" "$CHART" \
  --namespace "$NAMESPACE" \
  --create-namespace \
  --values "$CI_VALUES" \
  --dry-run \
  --debug

echo "Helm chart validation passed."
