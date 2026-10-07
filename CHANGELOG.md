# Changelog

All notable changes to Platform Lab will be documented in this file.

The project follows semantic versioning where practical.

## [Unreleased]

### Planned

- additional troubleshooting scenarios
- expanded observability exercises
- additional platform-engineering labs
- continued integration with Platform Lab CLI
- continued integration with Platform Doctor CLI

---

## [1.1.0] - 2026

### Added

- Course 1 Platform Lab training baseline
- Docker Compose-based local microservices environment
- Angular storefront
- Web BFF
- Catalog Service
- Cart Service
- Order Service
- Inventory Service
- Payment Service
- Notification Service
- PostgreSQL
- Redis
- Kafka
- PostgreSQL bootstrap initialization
- Kafka topic initialization
- service health checks
- local smoke testing
- Docker Compose health-based startup dependencies
- optional observability Compose overlay
- Prometheus support
- Grafana support
- OpenTelemetry Collector support
- Tempo support
- local CI validation workflow

### Changed

- redesigned storefront for the Platform Lab training environment
- aligned the repository with Course 1 troubleshooting exercises
- improved order-processing behavior
- separated the core training environment from optional observability services
- documented clean startup, rebuild, and reset workflows

### Training

The `1.1.0` release represents the baseline used for the first Platform Lab troubleshooting course.

The core learning environment is started with:

```bash id="3kq8rc"
docker compose -f platform/docker/compose.yaml up -d
```

The observability stack is optional and can be enabled separately when required.

---

## [1.0.0] - 2026

### Added

- initial Platform Lab project structure
- initial microservices application
- Docker-based local environment
- initial application services and infrastructure dependencies

This release established the original Platform Lab foundation.
