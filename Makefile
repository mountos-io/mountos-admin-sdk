.PHONY: all gen check build install clean ts-install ts-check ts-build ts-publish go-check go-build bump-minor-version bump-major-version help

all: gen check build

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Generator ───────────────────────────────────────────────

gen: ## Generate Go + TS SDK from api.yaml
	cd gen && go run . --spec ../api.yaml --go-out ../go --ts-out ../ts/src --doc-out ..

gen-browser: ## Generate browser client for admin-client
	cd gen && go run . --spec ../api.yaml --go-out ../go --ts-out ../ts/src --browser-client-out ../../mountos-admin-client/src/lib/core/api --doc-out ..

# ── TypeScript ──────────────────────────────────────────────

ts-install: ## Install TS dependencies
	cd ts && npm install

ts-check: ts-install ## Type-check TS
	cd ts && npx tsc --noEmit

ts-build: ts-install ## Build TS
	cd ts && npx tsc

ts-publish: ts-build ## Publish TS package to npm
	cd ts && npm publish --access public

ts-clean: ## Remove TS build artifacts
	rm -rf ts/dist ts/node_modules

# ── Go ──────────────────────────────────────────────────────

go-check: ## Vet Go code
	cd go && go vet ./...

go-build: ## Build Go code
	cd go && go build ./...

# ── Combined ────────────────────────────────────────────────

install: ts-install ## Install all dependencies

check: ts-check go-check ## Run all checks

build: ts-build go-build ## Build all

clean: ts-clean ## Clean all artifacts

# ── Version ────────────────────────────────────────────────

VERSION := $(shell jq -r .version ts/package.json)
MAJOR   := $(word 1,$(subst ., ,$(VERSION)))
MINOR   := $(word 2,$(subst ., ,$(VERSION)))
PATCH   := $(word 3,$(subst ., ,$(VERSION)))

bump-minor-version: ## Bump minor version, commit and tag
	$(eval NEW := $(MAJOR).$(shell echo $$(($(MINOR)+1))).0)
	@jq '.version = "$(NEW)"' ts/package.json > ts/package.json.tmp && mv ts/package.json.tmp ts/package.json
	@git add .
	@git commit -m "v$(NEW)"
	@git tag "v$(NEW)"
	@git tag "go/v$(NEW)"
	@echo "$(VERSION) → $(NEW)"

bump-major-version: ## Bump major version, commit and tag
	$(eval NEW := $(shell echo $$(($(MAJOR)+1))).0.0)
	@jq '.version = "$(NEW)"' ts/package.json > ts/package.json.tmp && mv ts/package.json.tmp ts/package.json
	@git add .
	@git commit -m "v$(NEW)"
	@git tag "v$(NEW)"
	@git tag "go/v$(NEW)"
	@echo "$(VERSION) → $(NEW)"
