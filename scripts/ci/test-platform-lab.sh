#!/usr/bin/env bash
# Purpose: Runs the Platform Lab application test suite used by CI before images are built or deployed.
# Workflow: Executes the Go and Angular validation commands from a single repeatable entry point.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

info() {
  printf 'INFO  %s\n' "$*"
}

pass() {
  printf 'PASS  %s\n' "$*"
}

fail() {
  printf 'FAIL  %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

require_cmd go
require_cmd npm

for mod in "$ROOT"/services/*; do
  [[ -f "$mod/go.mod" ]] || continue
  info "go test ${mod#$ROOT/}"
  (cd "$mod" && go test ./...)
  pass "${mod#$ROOT/}"
done

info "storefront tests/build"
(
  cd "$ROOT/apps/storefront"
  npm ci
  npm test -- --watch=false
  npm run build
)
pass "storefront validation"
