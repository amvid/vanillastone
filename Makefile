.PHONY: dev dev-build down logs build build-web prod prod-image web web-install hooks clean help

# Host port for `make prod` (host:PORT -> container:8080). Override: make prod PORT=80
PORT ?= 8080

# --- Dev (server :8080 + vite :5173, both dockerized, hot reload) ---
dev: ## Run server + web (docker compose up). Open http://localhost:5173
	docker compose up

dev-build: ## Rebuild images + run (after go.mod / Dockerfile / deps change)
	docker compose up --build

down: ## Stop containers
	docker compose down

logs: ## Tail all logs
	docker compose logs -f

# --- Web client (Vite + pnpm). Normally run via `make dev` (dockerized). ---
web-install: ## Install client deps locally (uses corepack pnpm)
	cd web && corepack pnpm install

web: ## Run Vite dev server locally (proxies /ws,/register,/login -> :8080)
	cd web && corepack pnpm dev

# --- Build / prod ---
build: ## Build dev docker image
	docker compose build

build-web: ## Build the React client into web/static (run before committing)
	cd web && corepack pnpm install --frozen-lockfile && corepack pnpm build

hooks: ## Enable the committed git hooks (pre-commit builds the frontend)
	git config core.hooksPath .githooks
	@echo "git hooks enabled (.githooks)"

prod: ## Deploy: git pull + build & run prod compose. PORT=8080 (host port)
	git pull --ff-only
	PORT=$(PORT) docker compose -f docker-compose.prod.yml up -d --build

prod-image: ## Build the prod static binary image only (no run)
	docker build --target prod -t vanillastone:prod .

clean: ## Stop + remove volumes (wipes caches)
	docker compose down -v

help: ## List targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
