.PHONY: all gen check build install clean ts-install ts-check ts-build go-check go-build

all: gen check build

# ── Generator ───────────────────────────────────────────────

gen:
	cd gen && go run . --spec ../api.yaml --go-out ../go --ts-out ../ts/src --doc-out ..

# ── TypeScript ──────────────────────────────────────────────

ts-install:
	cd ts && npm install

ts-check: ts-install
	cd ts && npx tsc --noEmit

ts-build: ts-install
	cd ts && npx tsc

ts-clean:
	rm -rf ts/dist ts/node_modules

# ── Go ──────────────────────────────────────────────────────

go-check:
	cd go && go vet ./...

go-build:
	cd go && go build ./...

# ── Combined ────────────────────────────────────────────────

install: ts-install

check: ts-check go-check

build: ts-build go-build

clean: ts-clean
