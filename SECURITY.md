# Security Policy

Platform Lab is a local-first training environment designed for learning Docker, distributed systems, troubleshooting, and platform engineering.

It intentionally favors simplicity and visibility over production-grade security.

## Training Environment

Platform Lab is designed to run on a developer workstation using Docker Compose.

The repository may include:

- local development credentials
- predictable database usernames and passwords
- locally exposed service ports
- development-oriented network configuration
- debugging and health endpoints
- intentionally introduced failure conditions

These choices make the environment easier to understand and troubleshoot during training.

They are **not production security recommendations**.

Do not deploy Platform Lab directly to a public or production environment without reviewing and hardening the configuration.

## Local Credentials

Some infrastructure components use simple credentials intended only for local training.

For example, local PostgreSQL credentials may be defined directly in the Compose configuration.

These credentials:

- are not secrets
- are intentionally visible
- are intended only for the local lab
- must not be reused in production systems

Production environments should use an appropriate secrets-management system and stronger access controls.

## Exposed Ports

Platform Lab exposes several application and infrastructure ports to the local host so students can inspect and interact with individual components.

Examples include:

```text id="a3z7lc"
4200  storefront
8080  web-bff
8081  catalog-service
8082  cart-service
8083  order-service
8084  inventory-service
8085  payment-service
8086  notification-service
5432  PostgreSQL
6379  Redis
9092  Kafka
```

These ports are provided for local learning and debugging.

Production systems should apply appropriate network isolation, firewalls, authentication, encryption, and service-access policies.

## Intentional Failure Scenarios

Platform Lab is used together with controlled troubleshooting scenarios.

Some exercises may deliberately introduce conditions such as:

- unavailable services
- invalid configuration
- dependency failures
- restart loops
- DNS failures
- resource restrictions
- Redis failures
- Kafka failures

These scenarios are intentional parts of the training environment.

A deliberately broken lab scenario should not automatically be treated as a security vulnerability.

## Production Use

Platform Lab is **not intended to be deployed directly to production**.

Before adapting any part of the project for production use, review areas including:

- secrets management
- TLS and encryption
- authentication
- authorization
- network segmentation
- firewall policies
- container privileges
- image provenance
- dependency vulnerabilities
- database security
- Kafka security
- Redis security
- resource limits
- logging and audit controls
- backup and recovery
- availability and redundancy

The repository should be treated as educational source code rather than a production reference architecture.

## Reporting a Security Vulnerability

If you discover a genuine security vulnerability in Platform Lab, please avoid publishing sensitive exploitation details in a public issue.

When the repository is hosted on GitHub and private vulnerability reporting is available, use the repository's **Security** section to submit a private security report.

Please include:

- a clear description of the issue
- the affected component
- steps required to reproduce it
- the potential impact
- any suggested mitigation, if available

For non-sensitive bugs that do not create a security risk, use the normal GitHub issue tracker.

## Please Do Not Report

The following are generally expected characteristics of the local training environment and should not normally be reported as vulnerabilities:

- documented local development credentials
- services exposed to `localhost`
- intentionally simplified authentication
- deliberately injected troubleshooting failures
- containers stopped as part of a lab exercise
- intentionally broken configuration created by a training scenario

If an issue allows unintended access beyond the documented local lab behavior, it may still represent a valid security problem.

## Sensitive Information

Do not commit real credentials or sensitive information to this repository.

This includes:

- cloud access keys
- API tokens
- GitHub tokens
- private SSH keys
- certificates containing private keys
- production database passwords
- personal access tokens
- real customer data
- personally identifiable information
- `.env` files containing secrets

Use example or placeholder values for training configuration.

## Dependency Security

Platform Lab uses third-party software, container images, libraries, and packages.

Contributors should avoid introducing dependencies with known critical security vulnerabilities where practical.

Security and dependency scanning may be added or expanded as the project evolves.

## Supported Versions

Security fixes are applied to the actively maintained version of Platform Lab.

| Version | Supported |
|---|---|
| Latest release | Yes |
| Older training snapshots | Best effort |
| Modified forks | No |

Users running older course snapshots should compare their version with the latest repository release before reporting an issue.

## Responsible Disclosure

Please provide maintainers reasonable time to investigate and address legitimate security vulnerabilities before publicly disclosing technical exploitation details.

Thank you for helping keep Platform Lab useful and safe for learners and contributors.
