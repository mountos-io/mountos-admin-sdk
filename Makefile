GOBIN_DIR ?= $(HOME)/go/bin
GO_TOOLS = PATH="$(GOBIN_DIR):$$PATH"
CARGO_BIN_DIR ?= $(HOME)/.cargo/bin

TS_RUNTIME ?= node
TS_INSTALL_bun  := bun install
TS_INSTALL_deno := deno install
TS_INSTALL_node := npm install
TS_INSTALL      := $(or $(TS_INSTALL_$(TS_RUNTIME)),$(TS_RUNTIME) install)

TS_RUN_bun  := bun run
TS_RUN_deno := deno task
TS_RUN_node := npm run
TS_RUN      := $(or $(TS_RUN_$(TS_RUNTIME)),$(TS_RUNTIME) run)

.PHONY: all gen gen-check gen-test docs check vet test build install clean audit hooks lint vuln \
  example-go example-rust example-ts \
  ts-install ts-check ts-lint ts-build ts-test ts-audit ts-publish ts-clean \
  go-check go-build go-test \
  rust-check rust-fmt rust-audit rust-build rust-test rust-publish rust-clean \
  tag tag-minor tag-major bump-version help

all: gen check build

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

hooks: ## Install the git hooks (lefthook)
	lefthook install

# ── Generator ───────────────────────────────────────────────

gen: ## Generate Go + TS + Rust SDK and docs from api.yaml
	cd gen && go run . --spec ../api.yaml --go-out ../go --ts-out ../ts/src --rust-out ../rust/src --doc-out .. --docs-out ../docs
	cp api.yaml ts/api.yaml
	cp api.md ts/api.md
	cp SKILL.md ts/SKILL.md
	cp api.md LICENSE NOTICE rust/

gen-check: ## Fail when a generated file or a copy kept by make gen is stale
	./scripts/gen-check.sh

gen-test: ## Run the generator's own unit tests (dispatch matrix, etc.)
	cd gen && go test ./...

docs: ## Regenerate docs/ts.md, docs/go.md and docs/rust.md from api.yaml
	cd gen && go run . --spec ../api.yaml --go-out ../go --ts-out ../ts/src --rust-out ../rust/src --doc-out .. --docs-out ../docs

# ── TypeScript ──────────────────────────────────────────────

ts-install: ## Install TS dependencies
	cd ts && $(TS_INSTALL)

ts-check: ts-install ## Type-check TS
	cd ts && $(TS_RUN) check

ts-lint: ts-install ## Lint TS with ESLint (type-aware rules)
	cd ts && $(TS_RUN) lint

ts-audit: ts-install ## Audit runtime npm dependencies (needs network, fails on known advisories)
	cd ts && npm audit --omit=dev

ts-build: ts-install ## Build TS
	cd ts && $(TS_RUN) build

ts-test: ts-install ## Run TS fixture/contract tests (node:test against dist/)
	cd ts && $(TS_RUN) test

ts-publish: ts-build ## Publish TS package to npm
	@npm whoami >/dev/null 2>&1 || npm login
	cd ts && npm publish --access public

ts-clean: ## Remove TS build artifacts
	rm -rf ts/dist ts/node_modules

# ── Go ──────────────────────────────────────────────────────

go-check: ## Vet Go code (client module, generator and example)
	cd go && go vet ./...
	cd gen && go vet ./...
	cd examples/go && go vet ./...

lint: ## Run golangci-lint on the Go client module and the generator
	@$(GO_TOOLS) command -v golangci-lint >/dev/null 2>&1 || { echo "lint: golangci-lint not found in PATH or $(GOBIN_DIR)"; exit 1; }
	cd go && $(GO_TOOLS) golangci-lint run --config ../.golangci.yml ./...
	cd gen && $(GO_TOOLS) golangci-lint run --config ../.golangci.yml ./...

vuln: ## Fail on known vulnerabilities in Go (govulncheck), Rust (cargo audit) and TypeScript (npm audit)
	@$(GO_TOOLS) command -v govulncheck >/dev/null 2>&1 || { echo "vuln: govulncheck not found in PATH or $(GOBIN_DIR)"; exit 1; }
	cd go && $(GO_TOOLS) govulncheck ./...
	cd gen && $(GO_TOOLS) govulncheck ./...
	@PATH="$(CARGO_BIN_DIR):$$PATH" cargo audit --version >/dev/null 2>&1 || { echo "vuln: cargo-audit not found in $(CARGO_BIN_DIR) (install: cargo install cargo-audit)"; exit 1; }
	cd rust && PATH="$(CARGO_BIN_DIR):$$PATH" cargo audit --deny warnings
	cd ts && npm audit

go-build: ## Build Go code
	cd go && go build ./...

go-test: ## Run Go fixture/contract tests
	cd go && go test -count=1 ./...

# ── Rust ────────────────────────────────────────────────────

rust-check: ## Clippy-lint Rust code (warnings are errors, public types must be non_exhaustive)
	cd rust && cargo clippy --all-targets --locked -- -D warnings

rust-fmt: ## Check Rust formatting (generated modules are skipped)
	cd rust && cargo fmt --check

rust-audit: ## Audit crate dependencies with cargo audit (fails on advisories, skips when not installed)
	@if PATH="$(CARGO_BIN_DIR):$$PATH" cargo audit --version >/dev/null 2>&1; then cd rust && PATH="$(CARGO_BIN_DIR):$$PATH" cargo audit --deny warnings; else echo "rust-audit: cargo-audit is not installed, skipping (install: cargo install cargo-audit)"; fi

rust-build: ## Build Rust crate
	cd rust && cargo build --locked

rust-test: ## Run Rust unit + doc tests
	cd rust && cargo test --locked

rust-publish: rust-build ## Publish Rust crate to crates.io
	cd rust && cargo publish

rust-clean: ## Remove Rust build artifacts
	cd rust && cargo clean

# ── Examples ────────────────────────────────────────────────
# The examples are separate modules that depend on the SDK by path. They are
# built on every check so a changed API cannot leave a stale example behind.

example-go: ## Build and vet the Go example
	cd examples/go && go build -o /dev/null ./... && go vet ./...

example-rust: ## Lint the Rust example (the lock file follows the SDK version, so no --locked)
	cd examples/rust && cargo clippy -- -D warnings

example-ts: ts-build ## Type-check the TypeScript example against the built package
	cd examples/ts && npm install --no-audit --no-fund --prefer-offline && npx tsc -p .

# ── Combined ────────────────────────────────────────────────

install: ts-install ## Install all dependencies

vet: go-check lint rust-fmt rust-check ts-check ts-lint ## go vet, golangci-lint, cargo fmt --check, clippy -D warnings, tsc --noEmit, eslint

check: vet vuln rust-audit gen-check example-go example-rust example-ts ## Vet, govulncheck, cargo audit, generated-file drift, examples

audit: ts-audit rust-audit ## Dependency audits (npm audit needs network, cargo audit also runs in check)

# Not folded into `check`/`all`. Run explicitly (`make test`) or per-language
# (`make go-test`, etc.).
test: gen-test ts-test go-test rust-test ## Run all fixture/contract tests

build: ts-build go-build rust-build ## Build all

clean: ts-clean rust-clean ## Clean all artifacts

# ── Version ────────────────────────────────────────────────

VERSION := $(shell jq -r .version ts/package.json)
MAJOR   := $(word 1,$(subst ., ,$(VERSION)))
MINOR   := $(word 2,$(subst ., ,$(VERSION)))
PATCH   := $(word 3,$(subst ., ,$(VERSION)))

tag: tag-minor ## Alias for tag-minor

# bump-version writes $(NEW) into every per-language manifest, keeping TS (npm)
# and Rust (crates.io) in lockstep. Go versions purely via the go/v* git tag, so
# it needs no file edit. Cargo.lock is refreshed (--workspace touches only our
# package, not deps) so `--locked` builds keep working after the bump.
bump-version:
	@jq '.version = "$(NEW)"' ts/package.json > ts/package.json.tmp && mv ts/package.json.tmp ts/package.json
	@sed 's/^version = .*/version = "$(NEW)"/' rust/Cargo.toml > rust/Cargo.toml.tmp && mv rust/Cargo.toml.tmp rust/Cargo.toml
	@cd rust && cargo update --workspace --offline
	@git add .
	@git commit -m "v$(NEW)"
	@git tag "v$(NEW)"
	@git tag "go/v$(NEW)"
	@git push origin --tags $(shell git branch --show-current)
	@echo "$(VERSION) → $(NEW)"

tag-minor: ## Bump minor version, commit and tag
	@$(MAKE) bump-version NEW=$(MAJOR).$(shell echo $$(($(MINOR)+1))).0

tag-major: ## Bump major version, commit and tag
	@$(MAKE) bump-version NEW=$(shell echo $$(($(MAJOR)+1))).0.0
