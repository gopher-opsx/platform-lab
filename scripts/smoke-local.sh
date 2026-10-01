#!/usr/bin/env sh
# Purpose: Validates the complete local Platform Lab business flow through the Web BFF.
# Workflow: Verifies one successful order, one inventory-rejected order, and one payment-rejected order.

set -eu

base_url="${BASE_URL:-http://localhost:8080}"
customer="smoke-$(date +%s)"

curl -fsS "$base_url/healthz" >/dev/null
curl -fsS "$base_url/readyz" >/dev/null
curl -fsS "$base_url/api/products" | grep -q 'prod-001'

wait_for_status() {
  order_id="$1"
  expected="$2"
  attempt=0
  status=""

  while [ "$attempt" -lt 20 ]; do
    status=$(curl -fsS "$base_url/api/orders/$order_id" | sed -n 's/.*"status":"\([^"]*\)".*/\1/p')
    [ "$status" = "$expected" ] && return 0
    attempt=$((attempt+1))
    sleep 1
  done

  echo "FAIL order $order_id expected '$expected' but ended at '$status'" >&2
  return 1
}

# 1. Happy path: inventory is available and payment is within the authorization limit.
success_json=$(curl -fsS \
  -X POST "$base_url/api/orders" \
  -H 'Content-Type: application/json' \
  -H "X-Customer-ID: $customer" \
  -d '{"currency":"USD","items":[{"productId":"prod-002","quantity":1,"unitPriceCents":89900}]}')

success_id=$(printf '%s' "$success_json" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
test -n "$success_id"
wait_for_status "$success_id" "confirmed"

# 2. Inventory rejection: prod-001 has seeded stock of 25, so requesting 100
#    must be rejected by inventory and the order must be cancelled.
inventory_json=$(curl -fsS \
  -X POST "$base_url/api/orders" \
  -H 'Content-Type: application/json' \
  -H "X-Customer-ID: $customer" \
  -d '{"currency":"USD","items":[{"productId":"prod-001","quantity":100,"unitPriceCents":169900}]}')

inventory_id=$(printf '%s' "$inventory_json" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
test -n "$inventory_id"
wait_for_status "$inventory_id" "cancelled"

# 3. Payment rejection: inventory is available, but the order value exceeds
#    PAYMENT_MAX_AUTH_CENTS and must be cancelled after payment processing.
payment_json=$(curl -fsS \
  -X POST "$base_url/api/orders" \
  -H 'Content-Type: application/json' \
  -H "X-Customer-ID: $customer" \
  -d '{"currency":"USD","items":[{"productId":"prod-003","quantity":1,"unitPriceCents":600000}]}')

payment_id=$(printf '%s' "$payment_json" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
test -n "$payment_id"
wait_for_status "$payment_id" "cancelled"

echo "Platform Lab smoke passed:"
echo "  happy path:          $success_id confirmed"
echo "  inventory rejection: $inventory_id cancelled"
echo "  payment rejection:   $payment_id cancelled"
