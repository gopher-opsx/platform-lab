#!/usr/bin/env bash
# Purpose: Runs the application test suite used by CI before images are built or deployed.
# Workflow: Executes the Go and Angular validation commands from a single repeatable entry point.

set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
source "$ROOT/scripts/lib/course-common.sh"
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
