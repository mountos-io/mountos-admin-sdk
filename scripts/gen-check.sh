#!/usr/bin/env bash
# Regenerates every generated file into a scratch tree and fails when a committed copy differs.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

mkdir -p "$tmp/go" "$tmp/ts" "$tmp/rust" "$tmp/docs"
(cd "$root/gen" && go run . --spec ../api.yaml --go-out "$tmp/go" --ts-out "$tmp/ts" --rust-out "$tmp/rust" \
  --doc-out "$tmp" --docs-out "$tmp/docs" --contract-out "$tmp/contract-snapshot.json" >/dev/null)

stale=0
compare() { # compare <generated> <committed>
  if ! cmp -s "$1" "$2"; then
    echo "gen-check: stale ${2#"$root"/}"
    stale=1
  fi
}

for f in "$tmp"/go/*; do compare "$f" "$root/go/$(basename "$f")"; done
for f in "$tmp"/ts/*; do compare "$f" "$root/ts/src/$(basename "$f")"; done
for f in "$tmp"/rust/*; do compare "$f" "$root/rust/src/$(basename "$f")"; done
for f in "$tmp"/docs/*; do compare "$f" "$root/docs/$(basename "$f")"; done
compare "$tmp/api.md" "$root/api.md"
compare "$tmp/contract-snapshot.json" "$root/contract-snapshot.json"

# Copies that make gen keeps next to the packages.
compare "$root/api.yaml" "$root/ts/api.yaml"
compare "$root/api.md" "$root/ts/api.md"
compare "$root/SKILL.md" "$root/ts/SKILL.md"
compare "$root/api.md" "$root/rust/api.md"
compare "$root/LICENSE" "$root/rust/LICENSE"
compare "$root/NOTICE" "$root/rust/NOTICE"

if [ "$stale" -ne 0 ]; then
  echo "gen-check: run make gen and commit the result"
  exit 1
fi
