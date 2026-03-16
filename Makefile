# Pebblr CRM — Makefile
# CI/CD pipelines call these targets only.

.DEFAULT_GOAL := help
.PHONY: help build test lint typecheck dev-api dev-web cluster-up deploy migrate clean helm-validate e2e

# ── Pinned versions ───────────────────────────────────────────────────────────
ESO_VERSION           := 0.12.1
INGRESS_NGINX_VERSION := 4.12.1

# ── Paths ─────────────────────────────────────────────────────────────────────
GO_CMD  := cmd/api
WEB_DIR := web
CLUSTER := pebblr-local
KIND_CFG := deploy/kind/kind-config.yaml

# ── AKS safety guard ─────────────────────────────────────────────────────────
# Blocks cluster-up and deploy from running against an AKS cluster.
AKS_GUARD := @if kubectl get nodes -o jsonpath='{.items[*].metadata.name}' 2>/dev/null | grep -q 'aks-'; then echo 'ERROR: Refusing to run against AKS cluster. This target is for local Kind only.'; exit 1; fi

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort

build: ## Build Go binary and React frontend
	@go build -o bin/api ./$(GO_CMD)
	@cd $(WEB_DIR) && bun install --frozen-lockfile && bun run build

test: ## Run Go tests and frontend tests
	@go test ./... && cd $(WEB_DIR) && bun run test

lint: ## Run golangci-lint and ESLint
	@golangci-lint run ./... && cd $(WEB_DIR) && bun run lint

typecheck: ## Run tsc --noEmit in web/
	@cd $(WEB_DIR) && bun run typecheck

dev-api: ## Run Go API server locally with hot reload
	@air -c .air.toml || go run ./$(GO_CMD)

dev-web: ## Run Vite dev server
	@cd $(WEB_DIR) && bun run dev

cluster-up: ## Recreate local Kind cluster; install ESO and ingress-nginx (pinned versions)
	$(AKS_GUARD)
	@kind delete cluster --name $(CLUSTER) 2>/dev/null || true
	@kind create cluster --name $(CLUSTER) --config $(KIND_CFG)
	@helm repo add external-secrets https://charts.external-secrets.io 2>/dev/null || true
	@helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx 2>/dev/null || true
	@helm repo update
	@helm upgrade --install external-secrets external-secrets/external-secrets \
		--version $(ESO_VERSION) \
		--namespace external-secrets-operator --create-namespace --wait
	@helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
		--version $(INGRESS_NGINX_VERSION) \
		--namespace ingress-nginx --create-namespace --wait

deploy: ## Build and deploy to local Kind cluster via Skaffold
	$(AKS_GUARD)
	@skaffold run --default-repo=""

migrate: ## Run database migrations
	@go run ./$(GO_CMD) migrate

helm-validate: ## Validate Helm chart against a running Kind cluster (dry-run)
	@scripts/helm-ci-install.sh

e2e: ## Run E2E tests against a running Kind cluster
	@scripts/e2e.sh

clean: ## Clean build artifacts
	@rm -rf bin/ web/dist/ web/node_modules/.vite
