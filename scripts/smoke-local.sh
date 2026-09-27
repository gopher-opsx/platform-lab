#!/usr/bin/env sh
# Purpose: Validates the complete local Platform-lab business flow through the Web BFF.
# Workflow: Creates one successful and one compensated order, then waits for confirmed and cancelled Saga outcomes.

set -eu

base_url="${BASE_URL:-http://localhost:8080}"
customer="smoke-$(date +%s)"

curl -fsS "$base_url/healthz" >/dev/null
curl -fsS "$base_url/readyz" >/dev/null
curl -fsS "$base_url/api/products" | grep -q 'prod-001'

success_json=$(curl -fsS -X POST "$base_url/api/orders" -H 'Content-Type: application/json' -H "X-Customer-ID: $customer" -d '{"currency":"USD","items":[{"productId":"prod-002","quantity":1,"unitPriceCents":89900}]}')
success_id=$(printf '%s' "$success_json" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
test -n "$success_id"

attempt=0
while [ "$attempt" -lt 20 ]; do
  status=$(curl -fsS "$base_url/api/orders/$success_id" | sed -n 's/.*"status":"\([^"]*\)".*/\1/p')
  [ "$status" = "confirmed" ] && break
  attempt=$((attempt+1)); sleep 1
done
[ "$status" = "confirmed" ]

failure_json=$(curl -fsS -X POST "$base_url/api/orders" -H 'Content-Type: application/json' -H "X-Customer-ID: $customer" -d '{"currency":"USD","items":[{"productId":"prod-003","quantity":1,"unitPriceCents":600000}]}')
failure_id=$(printf '%s' "$failure_json" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
test -n "$failure_id"

attempt=0
while [ "$attempt" -lt 20 ]; do
  status=$(curl -fsS "$base_url/api/orders/$failure_id" | sed -n 's/.*"status":"\([^"]*\)".*/\1/p')
  [ "$status" = "cancelled" ] && break
  attempt=$((attempt+1)); sleep 1
done
[ "$status" = "cancelled" ]

echo "Platform Lab smoke passed: $success_id confirmed; $failure_id cancelled"
