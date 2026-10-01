#!/usr/bin/env bash
# Purpose: Generates controlled Kafka traffic against the local Platform Lab broker.
# Workflow: Creates a dedicated local training topic if needed and publishes synthetic messages
#           without requiring Azure Event Hubs, SASL credentials, or external Kafka tooling.

set -euo pipefail

# Prevent Git Bash/MSYS on Windows from converting Linux container paths
# such as /opt/kafka/... into Windows host paths.
export MSYS_NO_PATHCONV=1

KAFKA_CONTAINER="${KAFKA_CONTAINER:-platform-lab-kafka}"
KAFKA_BROKER="${KAFKA_BROKER:-localhost:9092}"
TOPIC="${KAFKA_TOPIC:-course-load}"
COUNT="${EVENT_COUNT:-300}"

printf 'Kafka container: %s\n' "$KAFKA_CONTAINER"
printf 'Broker:          %s\n' "$KAFKA_BROKER"
printf 'Topic:           %s\n' "$TOPIC"
printf 'Events:          %s\n' "$COUNT"

# Ensure the local training topic exists.
docker exec "$KAFKA_CONTAINER" \
  /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server "$KAFKA_BROKER" \
  --create \
  --if-not-exists \
  --topic "$TOPIC" \
  --partitions 3 \
  --replication-factor 1 >/dev/null

# Publish synthetic training events.
for i in $(seq 1 "$COUNT"); do
  printf '{"id":"load-%s-%s","type":"course.load","sequence":%s,"createdAt":"%s"}\n' \
    "$(date +%s)" \
    "$i" \
    "$i" \
    "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
done |
  docker exec -i "$KAFKA_CONTAINER" \
    /opt/kafka/bin/kafka-console-producer.sh \
    --bootstrap-server "$KAFKA_BROKER" \
    --topic "$TOPIC" >/dev/null

printf 'PASS published %s events to %s\n' "$COUNT" "$TOPIC"
