# Pebblr CRM — Makefile
# CI/CD pipelines call these targets only.
# Multi-step logic is in scripts/.

.DEFAULT_GOAL := help
.PHONY: help build test lint typecheck dev dev-api dev-web docker-build \
        cluster-up cluster-down secrets-up deploy migrate clean

BINARY        := bin/api
GO_CMD        := cmd/api
WEB_DIR       := web
HELM_CHART    := deploy/helm/pebblr
IMAGE_NAME    ?= pebblr-api
IMAGE_TAG     ?= latest

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort

build: ## Build Go binary and React frontend
	@scripts/build.sh

test: ## Run Go tests and frontend tests
	@go test ./... && cd $(WEB_DIR) && bun test

lint: ## Run golangci-lint and ESLint
	@golangci-lint run ./... && cd $(WEB_DIR) && bun run lint

typecheck: ## Run tsc --noEmit in web/
	@cd $(WEB_DIR) && bun run typecheck

dev: ## Start local dev environment (Kind cluster + port-forwards + Vite dev server)
	@scripts/dev.sh

dev-api: ## Run Go API server locally with hot reload
	@air -c .air.toml || go run ./$(GO_CMD)

dev-web: ## Run Vite dev server
	@cd $(WEB_DIR) && bun run dev

docker-build: ## Build Docker images
	@scripts/docker-build.sh

cluster-up: ## Create Kind cluster
	@scripts/cluster-up.sh

cluster-down: ## Destroy Kind cluster
	@scripts/cluster-down.sh

secrets-up: ## Apply ESO and test secrets to Kind cluster
	@scripts/secrets-up.sh

deploy: ## Helm install to local cluster
	@helm upgrade --install pebblr $(HELM_CHART) --namespace pebblr --create-namespace

migrate: ## Run database migrations
	@scripts/migrate.sh

clean: ## Clean build artifacts
	@rm -rf bin/ web/dist/ web/node_modules/.vite
