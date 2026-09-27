# Cart Service

Redis-backed shopping-cart API with a default 24-hour TTL.

## Endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /cart`
- `POST /cart/items`
- `PATCH /cart/items/{productId}`
- `DELETE /cart/items/{productId}`
- `DELETE /cart`

Cart routes require `X-Customer-ID`. The storefront sends it to the BFF, which forwards it to this service.

## Configuration

| Variable | Default |
|---|---|
| `HTTP_ADDR` | `:8082` |
| `REDIS_ADDR` | `localhost:6379` |
| `REDIS_PASSWORD` | empty |
| `REDIS_DB` | `0` |
| `CART_TTL` | `24h` |

Local Docker Redis does not require TLS, so the default remains `false`.

Run as part of the complete Docker Compose stack from the repository root.
