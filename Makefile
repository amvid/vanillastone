.PHONY: dev dev-build down logs build build-web prod prod-image web web-install desktop desktop-all desktop-assets hooks clean help

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

# --- Web client (Vite via Deno). Normally run via `make dev` (dockerized). ---
web-install: ## Install client deps locally (deno reads npm deps from package.json)
	cd web && deno install

web: ## Run Vite dev server locally (proxies /ws,/register,/login -> :8080)
	cd web && deno task dev

# --- Desktop app (Deno Desktop). Bakes SERVER_URL into the UI, packages a native
#     binary. Requires deno 2.9+. See README.desktop.md. ---
# DESKTOP_OS: macos | linux | windows.  DESKTOP_ARCH: arm64 | x64 (windows: x64 only).
DESKTOP_OS   ?= macos
DESKTOP_ARCH ?= arm64

desktop: ## Build desktop app. DESKTOP_OS=macos|linux|windows DESKTOP_ARCH=arm64|x64 SERVER_URL=https://...
	@test -n "$(SERVER_URL)" || { echo "SERVER_URL required, e.g. make desktop SERVER_URL=https://game.example.com DESKTOP_OS=macos"; exit 1; }
	cd web && VITE_SERVER_URL=$(SERVER_URL) deno task build
	$(MAKE) desktop-assets
	@set -e; case "$(DESKTOP_OS)-$(DESKTOP_ARCH)" in \
	  macos-arm64)  t=aarch64-apple-darwin;      o=Vanillastone.dmg;;      \
	  macos-x64)    t=x86_64-apple-darwin;       o=Vanillastone.dmg;;      \
	  linux-arm64)  t=aarch64-unknown-linux-gnu; o=Vanillastone.AppImage;; \
	  linux-x64)    t=x86_64-unknown-linux-gnu;  o=Vanillastone.AppImage;; \
	  windows-x64)  t=x86_64-pc-windows-msvc;    o=Vanillastone.msi;;      \
	  *) echo "unsupported: DESKTOP_OS=$(DESKTOP_OS) DESKTOP_ARCH=$(DESKTOP_ARCH) (windows is x64 only)" >&2; exit 1;; \
	esac; \
	echo "building $$o for $$t"; \
	cd desktop && deno desktop --target $$t --icon build/icon.png --include static --output build/$$o main.ts

desktop-all: ## Build desktop app for every supported platform (SERVER_URL=https://...)
	@test -n "$(SERVER_URL)" || { echo "SERVER_URL required, e.g. make desktop-all SERVER_URL=https://game.example.com"; exit 1; }
	cd web && VITE_SERVER_URL=$(SERVER_URL) deno task build
	$(MAKE) desktop-assets
	cd desktop && deno desktop --all-targets --icon build/icon.png --include static --output build/Vanillastone main.ts

# desktop-assets stages the UI bundle next to main.ts (for --include) and renders
# the app icon from the favicon. Internal helper; `desktop`/`desktop-all` call it.
desktop-assets:
	rm -rf desktop/static && cp -r web/static desktop/static
	mkdir -p desktop/build
	sips -s format png web/public/favicon.svg --out desktop/build/icon.png --resampleHeightWidth 1024 1024 >/dev/null

# --- Build / prod ---
build: ## Build dev docker image
	docker compose build

build-web: ## Build the React client into web/static (run before committing)
	cd web && deno install && deno task build

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
