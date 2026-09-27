#!/usr/bin/env bash
# Purpose: Builds or validates all Platform Lab container images as a CI quality gate.
# Workflow: Catches Dockerfile and build-context failures across the eight deployable components before release.

set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
source "$ROOT/scripts/lib/course-common.sh"
require_cmd docker
files=(
  apps/storefront/Dockerfile
  services/web-bff/Dockerfile
  services/catalog-service/Dockerfile
  services/cart-service/Dockerfile
  services/order-service/Dockerfile
  services/inventory-service/Dockerfile
  services/payment-service/Dockerfile
  services/notification-service/Dockerfile
)
for file in "${files[@]}"; do
  info "validating $file"
  docker buildx build --platform linux/amd64 --file "$ROOT/$file" --load --tag "platform-lab-ci:$(basename "$(dirname "$file")")" "$ROOT" >/dev/null
  pass "$file"
done
