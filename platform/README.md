# Local Platform

The `platform/` directory contains the local runtime environment for Platform Lab.

The default environment is defined in:

```text
platform/docker/compose.yaml
```

It provides the core application stack used throughout the initial Platform Lab course:

- PostgreSQL
- Redis
- Kafka
- Kafka topic initialization
- PostgreSQL bootstrap
- Catalog Service
- Cart Service
- Order Service
- Inventory Service
- Payment Service
- Notification Service
- Web BFF
- Storefront

The core environment does **not** require the observability stack.

---

## Start the Core Platform

From the repository root:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  up -d --build --wait
```

Check container status:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  ps
```

Open the storefront:

```text
http://localhost:4200
```

---

## Run the Smoke Test

After the environment becomes healthy:

```bash
bash scripts/smoke-local.sh
```

The smoke test verifies the main application path and confirms that the local environment is functioning.

---

## View Logs

View logs for a specific service:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  logs --tail=100 SERVICE_NAME
```

For example:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  logs --tail=100 catalog-service
```

Follow logs continuously:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  logs -f --tail=100 catalog-service
```

---

## Stop the Platform

Stop and remove the application containers:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  down
```

Named volumes preserve PostgreSQL and Redis data across normal shutdowns.

Running `down` does not remove those volumes.

---

## Remove Orphan Containers

If a different Compose configuration was previously used, Docker may report orphan containers.

Clean them up with:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  down --remove-orphans
```

This is useful when switching from an environment that previously included observability services.

---

## Rebuild Without Cache

To rebuild the application images completely:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  build --no-cache --pull
```

Then start the environment:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  up -d --wait
```

---

## Full Local Reset

To remove containers, orphan containers, and persistent local volumes:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  down --remove-orphans -v
```

Then rebuild:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  build --no-cache --pull
```

And start again:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  up -d --wait
```

> **Warning:** `-v` deletes the PostgreSQL and Redis named volumes. Use it only when you intentionally want to reset the training environment.

---

## PostgreSQL Initialization

Database initialization scripts run when PostgreSQL creates a fresh volume.

Because initialization scripts run only against a new PostgreSQL data volume, changes to initialization SQL are not automatically applied to an existing local database.

When a completely fresh database is required, reset the environment with:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  down -v
```

Migrations for existing databases are currently applied explicitly.

Automated migration handling remains a future platform improvement.

---

## Kafka Initialization

The `kafka-init` service prepares the Kafka topics required by the application.

It is an initialization container and is expected to complete successfully and exit.

Similarly, `postgres-bootstrap` performs database bootstrap work and may appear as an exited container after successful initialization.

These initialization containers are different from the long-running application services.

---

## Optional Observability Environment

Platform Lab also includes:

```text
platform/docker/compose.observability.yaml
```

This optional Compose overlay adds observability components such as:

- Prometheus
- Grafana
- OpenTelemetry Collector
- Tempo

The observability environment is **not required for the core Course 1 platform**.

When observability is intentionally needed, start both Compose files:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  -f platform/docker/compose.observability.yaml \
  up -d --build --wait
```

Check the combined environment:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  -f platform/docker/compose.observability.yaml \
  ps
```

Stop the combined environment:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  -f platform/docker/compose.observability.yaml \
  down
```

Keeping observability separate allows the same Platform Lab environment to be extended in later lessons and courses without adding unnecessary components to the initial troubleshooting environment.
