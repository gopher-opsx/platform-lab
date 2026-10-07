# Platform Lab

Platform Lab is a local-first microservices environment for learning Docker, containers, service dependencies, troubleshooting, and platform engineering through hands-on practice.

The project provides a realistic e-commerce application that runs entirely on your local machine using Docker Compose.

Rather than troubleshooting isolated toy containers, you operate a connected system with frontend, backend services, databases, caching, messaging, health checks, and controlled failures.

## What is Platform Lab?

Platform Lab represents a small e-commerce platform where customers can:

- browse products
- add products to a cart
- create orders
- process inventory
- process payments
- receive order notifications

Behind that simple customer experience is a distributed system composed of multiple services and infrastructure dependencies.

The goal is to learn how to understand, operate, investigate, troubleshoot, and recover that system.

## Architecture

```text
Customer
   │
   ▼
Storefront :4200
   │
   ▼
Web BFF :8080
   │
   ├── Catalog Service :8081 ───── PostgreSQL
   │
   ├── Cart Service :8082 ──────── Redis
   │
   └── Order Service :8083
             │
             ▼
           Kafka
             │
       ┌─────┼───────────┐
       ▼     ▼           ▼
 Inventory  Payment   Notification
  :8084     :8085       :8086
       │       │           │
       └────── PostgreSQL ─┘
```

The environment intentionally contains multiple dependency types so students can investigate realistic failure conditions.

## Core Components

| Component | Port | Purpose |
|---|---:|---|
| Storefront | `4200` | Customer-facing web application |
| Web BFF | `8080` | Backend-for-frontend and request routing |
| Catalog Service | `8081` | Product catalog |
| Cart Service | `8082` | Shopping cart |
| Order Service | `8083` | Order workflow |
| Inventory Service | `8084` | Inventory processing |
| Payment Service | `8085` | Payment processing |
| Notification Service | `8086` | Order notifications |
| PostgreSQL | `5432` | Persistent service data |
| Redis | `6379` | Cart storage |
| Kafka | `9092` | Asynchronous events |

Internally, Kafka services communicate using the Docker network listener on port `29092`.

## Requirements

Before starting Platform Lab, install:

- Git
- Docker
- Docker Compose

Verify Docker:

```bash
docker --version
docker compose version
```

## Quick Start

Clone the repository and enter the project directory:

```bash
git clone <repository-url>
cd platform-lab
```

Start the complete Platform Lab environment:

```bash
docker compose -f platform/docker/compose.yaml up -d --build --wait
```

Or use the Makefile:

```bash
make compose-local-up
```

Docker will start the infrastructure dependencies, initialize the databases and Kafka topics, build the application services, and start the storefront.

## Verify the Environment

Check container status:

```bash
docker compose -f platform/docker/compose.yaml ps
```

Or:

```bash
make compose-local-ps
```

The main long-running services should report a healthy state.

Some initialization containers such as PostgreSQL bootstrap and Kafka initialization are expected to complete and exit successfully.

## Open the Application

Open:

```text
http://localhost:4200
```

You should be able to browse products, use the cart, and place orders.

This healthy application becomes the baseline for later troubleshooting exercises.

## Health Checks

The Go services expose health endpoints:

```text
http://localhost:8080/healthz
http://localhost:8081/healthz
http://localhost:8082/healthz
http://localhost:8083/healthz
http://localhost:8084/healthz
http://localhost:8085/healthz
http://localhost:8086/healthz
```

You can check them with:

```bash
make health
```

## View Logs

Follow logs for the complete environment:

```bash
docker compose -f platform/docker/compose.yaml logs -f --tail=100
```

View one service:

```bash
docker compose -f platform/docker/compose.yaml logs -f --tail=100 catalog-service
```

Replace `catalog-service` with the service you want to inspect.

## Stop Platform Lab

Stop the environment while preserving PostgreSQL and Redis data:

```bash
docker compose -f platform/docker/compose.yaml down
```

Or:

```bash
make compose-local-down
```

Named volumes preserve local data across ordinary shutdowns.

## Clean Rebuild

To remove the running environment and any orphan containers:

```bash
docker compose -f platform/docker/compose.yaml down --remove-orphans
```

Rebuild the application images without Docker build cache:

```bash
docker compose -f platform/docker/compose.yaml build --no-cache --pull
```

Start the environment again:

```bash
docker compose -f platform/docker/compose.yaml up -d --wait
```

## Full Environment Reset

To completely reset Platform Lab, including PostgreSQL and Redis data:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  down --remove-orphans -v
```

Then rebuild and start:

```bash
docker compose -f platform/docker/compose.yaml build --no-cache --pull

docker compose -f platform/docker/compose.yaml up -d --wait
```

> **Warning:** Using `-v` removes the local PostgreSQL and Redis volumes. Use it only when you intentionally want a fresh training environment.

## Smoke Test

Platform Lab includes a local smoke test:

```bash
bash scripts/smoke-local.sh
```

Or:

```bash
make compose-local-smoke
```

## Observability

Platform Lab also contains an optional observability environment with components such as:

- Prometheus
- Grafana
- OpenTelemetry Collector
- Tempo

Observability is intentionally separated from the core environment and is **not required for the initial Platform Lab troubleshooting course environment**.

When observability is needed, it can be added using:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  -f platform/docker/compose.observability.yaml \
  up -d --build --wait
```

This keeps the initial learning environment focused while allowing later courses and labs to extend the same platform.

## Companion Tools

Platform Lab is designed to work with two companion command-line tools.

### Platform Lab CLI

The Platform Lab CLI introduces controlled and repeatable failures into the training environment.

It is used to create scenarios such as:

- service failures
- dependency failures
- configuration problems
- restart loops
- DNS problems
- resource pressure
- Redis failures
- Kafka failures

The purpose of the Lab CLI is to create the incident.

The student still investigates the system and determines the root cause.

### Platform Doctor CLI

Platform Doctor is an investigation assistant.

It collects troubleshooting evidence from the Platform Lab environment such as:

- container state
- health information
- logs
- configuration
- networking information
- resource information
- dependency state

Doctor is designed to assist investigation rather than automatically repair the environment.

This distinction is intentional:

```text
Lab CLI
   │
   ▼
introduces failure
   │
   ▼
student investigates
   │
   ▼
Doctor collects evidence
   │
   ▼
root cause
   │
   ▼
fix and verify
```

## Repository Structure

```text
platform-lab/
├── apps/
│   └── storefront/
│
├── services/
│   ├── web-bff/
│   ├── catalog-service/
│   ├── cart-service/
│   ├── order-service/
│   ├── inventory-service/
│   ├── payment-service/
│   └── notification-service/
│
├── contracts/
│   ├── openapi/
│   └── asyncapi/
│
├── platform/
│   ├── docker/
│   ├── postgres/
│   └── observability/
│
├── scripts/
├── .github/
├── Makefile
└── README.md
```

## Development and Validation

Run the local CI checks with:

```bash
make ci-local
```

The project includes validation for the Go services, frontend, Docker Compose configuration, and container images.

## Training Environment

Platform Lab is intentionally designed as a local training environment.

Credentials, networking decisions, exposed ports, and infrastructure configuration are optimized for learning and local development.

Do not deploy the environment directly to production without applying appropriate production security, networking, secrets management, resilience, and operational controls.

## Contributing

Contributions, bug reports, improvements, and documentation fixes are welcome.

See `CONTRIBUTING.md` for contribution guidelines.

## Security

For security-related information and responsible reporting guidance, see `SECURITY.md`.

## License

Platform Lab source code is licensed under the Apache License 2.0.

See the `LICENSE` file for details.

Course videos, narration, slides, PDFs, and other training materials are separate from this repository's source-code license unless explicitly stated otherwise.
