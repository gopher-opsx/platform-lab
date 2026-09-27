# Platform Lab Web BFF

Browser-facing Backend for Frontend on port `8080`. It gives the Angular storefront one API origin while keeping internal service addresses private.

## Routing

- `/api/products...` -> Catalog Service
- `/api/cart...` -> Cart Service
- `/api/orders...` -> Order Service
- `GET /healthz` -> process liveness
- `GET /readyz` -> Catalog, Cart, and Order readiness

The BFF forwards content negotiation, `X-Customer-ID`, and W3C `traceparent` headers. It also provides development CORS policy for `http://localhost:4200`.

## Configuration

| Variable | Default |
|---|---|
| `HTTP_ADDR` | `:8080` |
| `CATALOG_SERVICE_URL` | `http://localhost:8081` |
| `CART_SERVICE_URL` | `http://localhost:8082` |
| `ORDER_SERVICE_URL` | `http://localhost:8083` |
| `ALLOWED_ORIGIN` | `http://localhost:4200` |

Run as part of the complete Docker Compose stack from the repository root. For service-only development, use the normal Go commands inside `services/web-bff`.
