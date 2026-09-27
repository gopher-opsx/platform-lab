# Platform Lab Storefront

Angular 21 standalone storefront for browsing products, managing a cart, placing orders, and following the final Saga state.

## Runtime behavior

The storefront calls only relative `/api` routes.

During Angular development, `proxy.conf.json` forwards those requests to the Web BFF at `http://localhost:8080`.

Default Docker/Compose value:

```text
BFF_UPSTREAM=http://web-bff:8080
```

If the BFF is unavailable, API-backed features report an error; the production Storefront does not fall back to a fake local catalog.

## Run locally

```bash
npm ci
npm start
```

Open `http://localhost:4200`. The course normally runs the Storefront as part of the complete Docker Compose stack.

## Verify

```bash
npm test -- --watch=false
npm run build
```

The production image is built from `Dockerfile` and served by Nginx.

To override the production-container BFF target manually:

```bash
docker run --rm \
  -p 4200:80 \
  -e BFF_UPSTREAM=http://host.docker.internal:8080 \
  platform-lab/storefront:local
```

