# Platform Lab local observability

This stack receives OpenTelemetry traces and metrics and makes them available through Grafana.

## Components

- OpenTelemetry Collector: OTLP gRPC `4317`, OTLP HTTP `4318`, health `13133`
- Tempo: trace storage and query API on `3200`
- Prometheus: metrics on `9090`
- Grafana: dashboards and trace exploration on `3000`

## Start

From the repository root:

```bash
docker compose -f platform/docker/compose.yaml -f platform/docker/compose.observability.yaml up -d --wait
```

Open Grafana at <http://localhost:3000>. Sign in with `admin` / `platform`, or use anonymous viewer access. The Prometheus and Tempo data sources and the Platform Lab Overview dashboard are provisioned automatically.

## Verify

```bash
curl http://localhost:13133/
curl http://localhost:9090/-/ready
curl http://localhost:3200/ready
curl http://localhost:3000/api/health
```

The dashboard panels remain empty until the Go services send OTLP data. That application instrumentation is the next development batch.

## Stop

```bash
docker compose -f platform/docker/compose.yaml -f platform/docker/compose.observability.yaml down
```
