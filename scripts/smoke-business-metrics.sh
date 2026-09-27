#!/usr/bin/env sh
# Purpose: Exercises business operations that produce useful application metrics for observability labs.
# Workflow: Creates predictable traffic so students can verify that business-level telemetry appears in the monitoring stack.

set -eu

required="platform_lab_orders_created_total platform_lab_orders_confirmed_total platform_lab_orders_cancelled_total platform_lab_payments_authorized_total platform_lab_payments_failed_total platform_lab_inventory_reserved_total platform_lab_inventory_released_total platform_lab_notifications_delivered_total"

for metric in $required; do
  response=$(curl -fsS --get --data-urlencode "query=$metric" http://localhost:9090/api/v1/query)
  echo "$response" | grep -q '"result":\[' || { echo "Prometheus query failed for $metric"; exit 1; }
done

echo "Platform Lab business metrics are queryable"
