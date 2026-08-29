# Phase 02 — Go Application Skeleton

## Objective

Create the runnable WeaveLens Go application around the domain layer.

## Architecture

Use a modular-monolith structure:

```text
cmd/
└── weavelens/

internal/
├── domain/
├── application/
├── infrastructure/
└── transport/
```

The exact structure may be adapted to existing conventions.

## Requirements

Implement:

* application entrypoint;
* configuration;
* dependency wiring;
* structured logging;
* HTTP server;
* health endpoint;
* readiness endpoint;
* graceful shutdown.

## Configuration

Configuration must come from environment variables or configuration files.

Do not hard-code:

* AWS credentials;
* ports;
* secrets;
* environment-specific values.

## Graceful Shutdown

Handle:

* SIGINT;
* SIGTERM.

Use `context.Context`.

Shutdown should respect a timeout.

## Dependency Injection

Avoid global mutable state.

Dependencies should be explicitly constructed and injected.

## HTTP

Implement minimal endpoints:

```text
GET /health
GET /ready
```

Do not implement business APIs yet.

## Testing

Add tests for:

* configuration;
* health endpoint;
* application startup where practical;
* graceful shutdown where practical.

## Constraints

Do not:

* introduce microservices;
* add AWS SDK;
* add NATS;
* add database;
* add frontend functionality.

## Acceptance Criteria

The following must work:

```bash
go run ./cmd/weavelens
```

and:

```text
GET /health
GET /ready
```

must return successful responses.

## Git

Run:

```bash
go build ./...
go test ./...
```

Create:

```text
feat(core): add go application foundation
```

Do not proceed to Phase 03 automatically.
