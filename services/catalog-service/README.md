# Catalog Service

Read-only product catalog API backed by PostgreSQL `catalog_db`.

## Endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /products`
- `GET /products/{id}`

## Configuration

| Variable | Default |
|---|---|
| `HTTP_ADDR` | `:8081` |
| `DATABASE_URL` | `postgres://platform:platform@localhost:5432/catalog_db?sslmode=disable` |

Run as part of the complete Docker Compose stack from the repository root.
