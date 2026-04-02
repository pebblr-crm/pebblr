# Pebblr CRM — Makefile
# CI/CD pipelines call these targets only.

.DEFAULT_GOAL := help
.PHONY: help build test lint typecheck dev-api dev-web dev-db dev-db-stop dev-db-reset seed cluster-up cluster-deps deploy migrate validate-config clean helm-validate e2e e2e-teardown e2e-cluster e2e-db e2e-deploy e2e-web e2e-web-integration sonar

# ── Pinned versions ───────────────────────────────────────────────────────────
ESO_VERSION           := 0.12.1
ENVOY_GW_VERSION      := v1.3.0
CERT_MANAGER_VERSION  := v1.17.1

# ── Paths ─────────────────────────────────────────────────────────────────────
GO_CMD   := cmd/pebblr
WEB_DIR  := web
CLUSTER  := pebblr-local
KIND_CFG := deploy/kind/kind-config.yaml

# ── AKS safety guard ─────────────────────────────────────────────────────────
# Blocks destructive/local-only targets from running against an AKS cluster.
AKS_GUARD := @if kubectl get nodes -o jsonpath='{.items[*].metadata.name}' 2>/dev/null | grep -q 'aks-'; then echo 'ERROR: Refusing to run against AKS cluster. This target is for local Kind only.'; exit 1; fi

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort

build: ## Build Go binary and React frontend
	@go build -o bin/pebblr ./$(GO_CMD)
	@cd $(WEB_DIR) && bun install --frozen-lockfile && bun run build

test: ## Run Go tests and frontend tests
	@go test ./...
	@cd $(WEB_DIR) && bun run test

lint: ## Run golangci-lint and ESLint
	@golangci-lint run ./...
	@cd $(WEB_DIR) && bun run lint

typecheck: ## Run tsc --noEmit
	@cd $(WEB_DIR) && bun run typecheck

dev-api: ## Run Go API server locally with hot reload
	@air -c .air.toml || go run ./$(GO_CMD) serve

dev-web: ## Run Vite dev server
	@cd $(WEB_DIR) && bun run dev

dev-db: ## Deploy on-cluster PostgreSQL, run migrations, and seed data (pebblr namespace)
	$(AKS_GUARD)
	@scripts/cluster-db.sh pebblr

dev-db-stop: ## Remove PostgreSQL from the pebblr namespace
	$(AKS_GUARD)
	@scripts/cluster-db.sh pebblr stop

dev-db-reset: ## Destroy and recreate on-cluster PostgreSQL with fresh seed data
	$(AKS_GUARD)
	@scripts/cluster-db.sh pebblr reset

seed: ## Load sample data (users, teams, customers, leads, calendar events) into on-cluster PostgreSQL
	$(AKS_GUARD)
	@scripts/seed.sh

cluster-deps: ## Install cert-manager, ESO, and Envoy Gateway into the current cluster (idempotent)
	$(AKS_GUARD)
	@ESO_VERSION=$(ESO_VERSION) ENVOY_GW_VERSION=$(ENVOY_GW_VERSION) CERT_MANAGER_VERSION=$(CERT_MANAGER_VERSION) scripts/cluster-deps.sh

cluster-up: ## Recreate local Kind cluster and install all dependencies
	$(AKS_GUARD)
	@kind delete cluster --name $(CLUSTER) 2>/dev/null || true
	@kind create cluster --name $(CLUSTER) --config $(KIND_CFG)
	@$(MAKE) cluster-deps

deploy: ## Build and deploy to local Kind cluster via Skaffold
	$(AKS_GUARD)
	@scripts/deploy-local.sh

migrate: ## Run database migrations
	$(AKS_GUARD)
	@go run ./cmd/migrate

helm-validate: ## Validate Helm chart against a running Kind cluster (dry-run)
	@scripts/helm-ci-install.sh

e2e-teardown: ## Delete the Kind cluster used for E2E testing
	@kind delete cluster --name $(CLUSTER)

e2e: ## Run E2E tests against a running Kind cluster
	@go test -v -tags=e2e -count=1 -timeout=10m ./e2e/...

e2e-web: ## Run Playwright E2E tests (starts Vite dev server automatically)
	@cd $(WEB_DIR) && bun run test:e2e

e2e-web-integration: ## Run Playwright integration tests against the Kind cluster
	@scripts/e2e-web.sh

e2e-cluster: ## Create a lightweight Kind cluster for E2E (no cert-manager/ESO/Envoy)
	$(AKS_GUARD)
	@kind create cluster --name $(CLUSTER) --config $(KIND_CFG) --wait 120s

e2e-db: ## Deploy PostgreSQL, run migrations, seed data, and create secrets (pebblr-e2e namespace)
	$(AKS_GUARD)
	@scripts/cluster-db.sh pebblr-e2e

e2e-deploy: ## Build, load, and deploy the app into pebblr-e2e namespace via Skaffold
	$(AKS_GUARD)
	@skaffold run -p e2e --default-repo="" --status-check=true

validate-config: ## Validate tenant config file
	@go run ./$(GO_CMD) config validate --config config/tenant.json

sonar: ## Run SonarCloud analysis locally
	@docker run --rm --network=host -v $(CURDIR):/usr/src -w /usr/src \
		sonarsource/sonar-scanner-cli \
		-Dsonar.host.url=https://sonarcloud.io \
		-Dsonar.token=$${SONAR_TOKEN:?Set SONAR_TOKEN}

clean: ## Clean build artifacts
	@rm -rf bin/ web/dist/ web/node_modules/.vite
