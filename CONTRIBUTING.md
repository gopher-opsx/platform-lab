# Contributing to Platform Lab

Thank you for your interest in contributing to Platform Lab.

Platform Lab is a local-first training environment for learning Docker, distributed systems, troubleshooting, and platform engineering.

Contributions should preserve that purpose.

## Contribution Philosophy

Platform Lab should remain:

- practical
- understandable
- local-first
- reproducible
- troubleshooting-friendly
- suitable for students
- easy to run with Docker Compose

Avoid adding unnecessary infrastructure complexity unless it directly supports a learning objective.

The project intentionally favors clarity over production-scale abstraction.

## Ways to Contribute

Useful contributions include:

- bug fixes
- documentation improvements
- service fixes
- Docker Compose improvements
- test improvements
- troubleshooting scenarios
- local developer experience improvements
- accessibility improvements
- security improvements
- dependency updates

Large architectural changes should be discussed before implementation.

## Development Requirements

Before contributing, install:

- Git
- Docker
- Docker Compose
- Go
- Node.js and npm

Exact language and runtime requirements may vary by component.

## Fork and Clone

Fork the repository and clone your fork:

```bash
git clone <your-fork-url>
cd platform-lab
```

Add the upstream repository if needed:

```bash
git remote add upstream <upstream-repository-url>
```

Verify:

```bash
git remote -v
```

## Create a Branch

Create a focused branch for your change:

```bash
git checkout -b fix/catalog-health-check
```

Examples:

```text
fix/catalog-health-check
docs/update-platform-readme
feat/add-training-scenario
test/order-service
chore/update-dependencies
```

Keep each branch focused on one logical change.

## Start the Local Environment

Start the core Platform Lab environment:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  up -d --build --wait
```

Check status:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  ps
```

The core environment should be healthy before testing your change.

## Run Local Validation

Run the project's local CI workflow:

```bash
make ci-local
```

This should be run before opening a pull request.

Depending on the component you modify, also run the relevant service or frontend tests.

## Validate Docker Compose

Always validate the Compose configuration after modifying platform files:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  config
```

If your change affects the observability overlay, also validate:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  -f platform/docker/compose.observability.yaml \
  config
```

## Run the Smoke Test

With Platform Lab running:

```bash
bash scripts/smoke-local.sh
```

Or:

```bash
make compose-local-smoke
```

The normal application workflow should continue to work unless your contribution intentionally changes that behavior.

## Verify Container Health

Check the runtime environment:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  ps
```

Long-running application services should be running and healthy.

Initialization containers such as PostgreSQL bootstrap or Kafka initialization may complete and exit successfully.

That behavior is expected.

## View Logs

When testing a change, inspect the affected service logs:

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

Follow logs if needed:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  logs -f --tail=100 catalog-service
```

## Clean Testing

If your change affects:

- Dockerfiles
- startup behavior
- initialization scripts
- service dependencies
- database bootstrap
- environment configuration

perform a clean rebuild before submitting the contribution.

Remove the environment:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  down --remove-orphans
```

Rebuild without cache:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  build --no-cache --pull
```

Start again:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  up -d --wait
```

If database initialization must also be tested from scratch:

```bash
docker compose \
  -f platform/docker/compose.yaml \
  down --remove-orphans -v
```

> Warning: `-v` removes the local PostgreSQL and Redis volumes.

## Code Guidelines

Keep changes:

- focused
- readable
- easy to troubleshoot
- consistent with the existing structure
- appropriate for a training environment

Avoid introducing unnecessary abstraction.

Students should be able to inspect the code and understand how the system works.

When introducing new behavior, consider whether students can observe it through:

- logs
- container state
- health endpoints
- configuration
- network behavior
- dependency behavior

Observability of system behavior is important to the Platform Lab learning model.

## Docker Guidelines

When modifying containers:

- keep images reasonably small
- avoid unnecessary packages
- preserve useful health checks
- avoid embedding real secrets
- prefer deterministic startup behavior
- preserve meaningful logs
- keep container responsibilities clear

Changes should work with the local Docker Compose environment.

## Configuration Guidelines

Never commit real credentials.

Use:

- local development values
- example values
- environment variables
- `.env.example`

Do not commit:

```text
.env
private keys
cloud credentials
API tokens
production passwords
personal access tokens
real customer data
```

## Training Scenarios

Platform Lab may intentionally support failure scenarios.

A training failure should be:

- controlled
- reproducible
- understandable
- reversible
- limited to the local lab
- safe for the student's machine

Do not create scenarios that intentionally damage the host system, access unrelated resources, or expose sensitive information.

## Documentation Changes

Documentation contributions are welcome.

When changing behavior, update the relevant documentation where necessary.

Keep documentation:

- concise
- technically accurate
- command-oriented
- easy for students to follow

Commands should be tested whenever practical.

## Commit Messages

Use short, descriptive commit messages.

Recommended format:

```text
type(scope): description
```

Examples:

```text
fix(catalog): improve database health handling

docs(platform): clarify clean rebuild workflow

feat(order): add event validation

test(cart): cover redis connection failure

chore(deps): update Go dependencies
```

Common types:

```text
feat
fix
docs
test
refactor
chore
ci
build
```

## Pull Requests

Before opening a pull request, verify that:

- the project builds
- relevant tests pass
- `make ci-local` passes
- Compose configuration is valid
- the core Platform Lab environment starts
- affected services are healthy
- the smoke test passes where applicable
- documentation has been updated if behavior changed
- no secrets or sensitive data were added

Keep pull requests focused.

Describe:

1. what changed
2. why it changed
3. how it was tested
4. any impact on students or training scenarios

## Breaking Changes

Avoid breaking the student workflow unnecessarily.

Changes affecting any of the following should be clearly documented:

- service names
- container names
- ports
- health endpoints
- Compose paths
- environment variables
- API contracts
- Kafka topics
- database schema
- CLI integration

These identifiers may be referenced directly by course lessons and companion tools.

## Course Compatibility

Platform Lab is used as a training environment.

Some repository versions may correspond to specific course snapshots or releases.

When changing behavior that could invalidate an existing lesson, scenario, command, or recording, clearly call it out in the pull request.

Compatibility with published course material should be considered before merging breaking changes.

## Security

Please review `SECURITY.md` before reporting security issues.

Do not publish sensitive vulnerability details in a public issue when private reporting is more appropriate.

## Code of Conduct

Contributors are expected to follow the project's `CODE_OF_CONDUCT.md`.

## License

By contributing to Platform Lab, you agree that your contributions will be licensed under the Apache License 2.0 used by this repository.
