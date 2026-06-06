TS_RUNTIME ?= node
TS_INSTALL_bun  := bun install
TS_INSTALL_deno := deno install
TS_INSTALL_node := npm install
TS_INSTALL      := $(or $(TS_INSTALL_$(TS_RUNTIME)),$(TS_RUNTIME) install)

TS_RUN_bun  := bun run
TS_RUN_deno := deno task
TS_RUN_node := npm run
TS_RUN      := $(or $(TS_RUN_$(TS_RUNTIME)),$(TS_RUNTIME) run)

.PHONY: all gen docs check build install clean ts-install ts-check ts-build ts-publish go-check go-build rust-check rust-build rust-test rust-publish rust-clean tag tag-minor tag-major help

all: gen check build

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Generator ───────────────────────────────────────────────

gen: ## Generate Go + TS + Rust SDK and docs from api.yaml
	cd gen && go run . --spec ../api.yaml --go-out ../go --ts-out ../ts/src --rust-out ../rust/src --doc-out .. --docs-out ../docs
	cp api.yaml ts/api.yaml
	cp api.md ts/api.md
	cp SKILL.md ts/SKILL.md
	cp api.md LICENSE NOTICE rust/

docs: ## Regenerate docs/ts.md, docs/go.md and docs/rust.md from api.yaml
	cd gen && go run . --spec ../api.yaml --go-out ../go --ts-out ../ts/src --rust-out ../rust/src --doc-out .. --docs-out ../docs

# ── TypeScript ──────────────────────────────────────────────

ts-install: ## Install TS dependencies
	cd ts && $(TS_INSTALL)

ts-check: ts-install ## Type-check TS
	cd ts && $(TS_RUN) check

ts-build: ts-install ## Build TS
	cd ts && $(TS_RUN) build

ts-publish: ts-build ## Publish TS package to npm
	@npm whoami >/dev/null 2>&1 || npm login
	cd ts && npm publish --access public

ts-clean: ## Remove TS build artifacts
	rm -rf ts/dist ts/node_modules

# ── Go ──────────────────────────────────────────────────────

go-check: ## Vet Go code
	cd go && go vet ./...

go-build: ## Build Go code
	cd go && go build ./...

# ── Rust ────────────────────────────────────────────────────

rust-check: ## Clippy-lint Rust code (warnings are errors)
	cd rust && cargo clippy --all-targets --locked -- -D warnings

rust-build: ## Build Rust crate
	cd rust && cargo build --locked

rust-test: ## Run Rust unit + doc tests
	cd rust && cargo test --locked

rust-publish: rust-build ## Publish Rust crate to crates.io
	cd rust && cargo publish

rust-clean: ## Remove Rust build artifacts
	cd rust && cargo clean

# ── Combined ────────────────────────────────────────────────

install: ts-install ## Install all dependencies

check: ts-check go-check rust-check ## Run all checks

build: ts-build go-build rust-build ## Build all

clean: ts-clean rust-clean ## Clean all artifacts

# ── Version ────────────────────────────────────────────────

VERSION := $(shell jq -r .version ts/package.json)
MAJOR   := $(word 1,$(subst ., ,$(VERSION)))
MINOR   := $(word 2,$(subst ., ,$(VERSION)))
PATCH   := $(word 3,$(subst ., ,$(VERSION)))

tag: tag-minor ## Alias for tag-minor

tag-minor: ## Bump minor version, commit and tag
	$(eval NEW := $(MAJOR).$(shell echo $$(($(MINOR)+1))).0)
	@jq '.version = "$(NEW)"' ts/package.json > ts/package.json.tmp && mv ts/package.json.tmp ts/package.json
	@git add .
	@git commit -m "v$(NEW)"
	@git tag "v$(NEW)"
	@git tag "go/v$(NEW)"
	@git push origin --tags $(shell git branch --show-current)
	@echo "$(VERSION) → $(NEW)"

tag-major: ## Bump major version, commit and tag
	$(eval NEW := $(shell echo $$(($(MAJOR)+1))).0.0)
	@jq '.version = "$(NEW)"' ts/package.json > ts/package.json.tmp && mv ts/package.json.tmp ts/package.json
	@git add .
	@git commit -m "v$(NEW)"
	@git tag "v$(NEW)"
	@git tag "go/v$(NEW)"
	@git push origin --tags $(shell git branch --show-current)
	@echo "$(VERSION) → $(NEW)"
