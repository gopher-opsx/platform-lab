#!/usr/bin/env sh
# Purpose: Checks that the local services expose Prometheus-format metrics after traffic has been generated.
# Workflow: Queries metric endpoints and verifies expected application telemetry is available for scraping.

set -eu

for port in 8080 8081 8082 8083 8084 8085 8086; do
  curl -fsS "http://localhost:${port}/healthz" >/dev/null
  curl -fsS "http://localhost:${port}/metrics" | grep -q "http_server_requests_total"
done

curl -fsS "http://localhost:9090/-/ready" >/dev/null
curl -fsS "http://localhost:9090/api/v1/targets" | grep -q '"health":"up"'

echo "Platform Lab's Prometheus metrics smoke passed"
