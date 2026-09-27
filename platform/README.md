# Local Platform

`platform/docker/compose.yaml` defines the full Platform-Lab environment: PostgreSQL, Redis, Kafka, topic initialization, seven Go containers, Web BFF, and the Nginx storefront.

## Commands

```bash
docker compose \
  -f platform/docker/compose.yaml \
  -f platform/docker/compose.observability.yaml \
  up -d --build --wait

docker compose \
  -f platform/docker/compose.yaml \
  -f platform/docker/compose.observability.yaml \
  ps

bash scripts/smoke-local.sh

docker compose \
  -f platform/docker/compose.yaml \
  -f platform/docker/compose.observability.yaml \
  logs --tail=100 SERVICE_NAME

docker compose \
  -f platform/docker/compose.yaml \
  -f platform/docker/compose.observability.yaml \
  down
```

Named volumes preserve PostgreSQL and Redis data across ordinary shutdowns. The normal `docker compose ... down` command does not delete volumes. Never add `--volumes` unless intentionally resetting local data.

Database init scripts run only when PostgreSQL creates a fresh volume. Migrations for existing databases are currently applied explicitly; migration automation is a remaining platform improvement.
