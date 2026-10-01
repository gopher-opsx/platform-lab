#!/usr/bin/env bash
# Purpose: Generates controlled HTTP traffic against the local Platform Lab storefront for performance and observability labs.
# Workflow: Sends repeated requests with configurable concurrency/count so students can watch platform behavior under load.

set -euo pipefail

URL="${1:-${STOREFRONT_URL:-http://localhost:4200}}"
REQUESTS="${REQUESTS:-500}"
CONCURRENCY="${CONCURRENCY:-25}"

printf 'Generating %s requests with concurrency %s against %s\n' "$REQUESTS" "$CONCURRENCY" "$URL"

seq "$REQUESTS" | xargs -P "$CONCURRENCY" -I{}   curl -fsS -o /dev/null "${URL%/}/api/products"

printf 'PASS HTTP load completed\n'
