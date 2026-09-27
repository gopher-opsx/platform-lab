#!/usr/bin/env bash
# Purpose: Generates controlled Kafka backlog for worker-scaling demonstrations.
# Workflow: Publishes synthetic messages to create measurable consumer lag without requiring real checkout traffic.

set -euo pipefail

: "${KAFKA_BROKERS:?set KAFKA_BROKERS}"
: "${KAFKA_SASL_PASSWORD:?set KAFKA_SASL_PASSWORD}"

TOPIC="${KAFKA_TOPIC:-orders}"
COUNT="${EVENT_COUNT:-300}"
KCAT_IMAGE="${KCAT_IMAGE:-edenhill/kcat:1.7.1}"

echo "Publishing ${COUNT} events to ${TOPIC}"

for i in $(seq 1 "${COUNT}"); do
  printf '{"id":"load-%s-%s","type":"order.created","data":{"orderId":"load-%s"}}\n' \
    "$(date +%s)" "${i}" "${i}"
done |
  MSYS_NO_PATHCONV=1 docker run --rm -i \
    "${KCAT_IMAGE}" \
    -P \
    -b "${KAFKA_BROKERS}" \
    -t "${TOPIC}" \
    -X security.protocol=SASL_SSL \
    -X sasl.mechanisms=PLAIN \
    -X 'sasl.username=$ConnectionString' \
    -X "sasl.password=${KAFKA_SASL_PASSWORD}"

printf 'PASS published %s events to %s\n' "${COUNT}" "${TOPIC}"
