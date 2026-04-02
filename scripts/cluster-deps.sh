#!/usr/bin/env bash
# Install cluster dependencies: cert-manager, External Secrets Operator, Envoy Gateway.
# Idempotent — safe to run multiple times.
#
# Usage: scripts/cluster-deps.sh
set -euo pipefail

# Pinned versions — keep in sync with Makefile variables.
ESO_VERSION="${ESO_VERSION:-0.12.1}"
ENVOY_GW_VERSION="${ENVOY_GW_VERSION:-v1.3.0}"
CERT_MANAGER_VERSION="${CERT_MANAGER_VERSION:-v1.17.1}"

log() { echo "==> $*"; }

log "Adding Helm repositories..."
helm repo add external-secrets https://charts.external-secrets.io 2>/dev/null || true
helm repo add jetstack https://charts.jetstack.io 2>/dev/null || true
helm repo update

log "Installing cert-manager ${CERT_MANAGER_VERSION}..."
helm upgrade --install cert-manager jetstack/cert-manager \
  --version "$CERT_MANAGER_VERSION" \
  --namespace cert-manager --create-namespace \
  --set crds.enabled=true --wait

log "Installing External Secrets Operator ${ESO_VERSION}..."
helm upgrade --install external-secrets external-secrets/external-secrets \
  --version "$ESO_VERSION" \
  --namespace external-secrets-operator --create-namespace --wait

log "Installing Envoy Gateway ${ENVOY_GW_VERSION}..."
helm upgrade --install eg oci://docker.io/envoyproxy/gateway-helm \
  --version "$ENVOY_GW_VERSION" \
  --namespace envoy-gateway-system --create-namespace --wait

log "Applying GatewayClass..."
kubectl apply -f deploy/k8s/gateway/gatewayclass.yaml

log "Cluster dependencies installed."
