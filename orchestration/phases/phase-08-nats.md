# Phase 08 — NATS JetStream Event Architecture

## Objective

Introduce NATS JetStream as the asynchronous event backbone of WeaveLens.

## Architectural Rule

Use:

### gRPC

For synchronous request/response.

### NATS JetStream

For asynchronous events and decoupled processing.

## Events

Create versioned event subjects:

```text
weavelens.scan.started.v1
weavelens.scan.completed.v1
weavelens.scan.failed.v1
weavelens.resource.discovered.v1
weavelens.relationship.discovered.v1
weavelens.graph.completed.v1
```

## Event Envelope

Events should contain:

* event ID;
* event type;
* version;
* occurred timestamp;
* scan ID;
* account ID;
* region.

Resource events should use canonical WeaveLens resource representations.

## JetStream

Implement:

* connection management;
* stream initialization;
* publisher abstraction;
* subscriber abstraction;
* durable consumers;
* explicit ACK;
* retry behavior;
* graceful shutdown.

## Idempotency

Consumers must be idempotent.

Processing the same event more than once must not produce duplicate graph nodes or edges.

## Architecture

At this stage, NATS may operate inside the modular-monolith architecture.

Do NOT immediately split services.

Conceptually:

```text
Application
    │
    ▼
NATS JetStream
    │
    ▼
Graph Component
```

## Testing

Test:

* publish;
* consume;
* ACK;
* retry;
* duplicate delivery;
* consumer restart;
* graceful shutdown.

Use a local NATS test environment.

Tests must not depend on production infrastructure.

## Observability

Log:

* event ID;
* subject;
* scan ID;
* consumer;
* processing duration;
* error.

Never log secrets.

## Acceptance Criteria

A discovery operation can publish resource events and a graph component can consume them.

Duplicate event processing is safe.

## Git

Commit:

```text
feat(events): add nats jetstream messaging
```

Do not proceed automatically.
