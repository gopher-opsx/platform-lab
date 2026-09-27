# Contracts

This directory contains source-controlled examples of Platform Lab's synchronous and asynchronous interfaces.

- `openapi` describes browser-facing HTTP APIs such as the Web BFF.
- `asyncapi` contains representative Kafka envelopes for orders, inventory, payments, and notifications.

Every event envelope uses `eventId`, `eventType`, `eventVersion`, `occurredAt`, `aggregateId`, optional correlation/causation/tracing metadata, and a typed `payload`.

Contract examples are documentation and test fixtures; the authoritative runtime behavior remains the service code and automated tests.
