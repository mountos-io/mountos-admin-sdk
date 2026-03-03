.PHONY: all check build install clean ts-install ts-check ts-build go-check go-build

all: check build

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
