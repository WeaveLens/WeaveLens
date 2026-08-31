# Phase 13 — Microservice Extraction

## Objective

Extract WeaveLens bounded contexts into independently deployable services.

This phase MUST NOT begin until the modular monolith is stable and the previous phases are working.

## Target Architecture

```text
                    Web
                     │
                     ▼
                API Gateway
                     │
          ┌──────────┼──────────┐
          │          │          │
         gRPC       gRPC       gRPC
          │          │          │
          ▼          ▼          ▼
     Discovery     Graph      Export
      Service      Service    Service
          │          ▲
          │          │
          └── NATS ──┘
```

## Services

Extract:

### API Gateway

Responsibilities:

* HTTP;
* authentication;
* authorization;
* request validation;
* API composition.

### Discovery Service

Responsibilities:

* AWS authentication;
* AWS resource scanners;
* scan orchestration;
* publishing discovery events.

### Graph Service

Responsibilities:

* consuming discovery events;
* building graph;
* maintaining temporary graph state;
* graph queries.

### Export Service

Responsibilities:

* converting canonical graph data into export formats.

## Communication

Use gRPC for synchronous calls.

Use NATS JetStream for asynchronous events.

Do not introduce direct database coupling between services.

## Service Boundaries

Each service must have:

* its own application layer;
* its own infrastructure layer;
* clear interfaces;
* independent configuration;
* health endpoints;
* graceful shutdown;
* tests.

Shared code must remain minimal.

Do not create a huge shared package that couples every service together.

## Deployment

Prepare for:

* Docker;
* independent containers;
* Kubernetes deployment;
* independent scaling.

Do not add Kubernetes-specific complexity until local service execution works.

## Distributed-System Requirements

Introduce:

* timeout;
* retry;
* idempotency;
* graceful degradation;
* health checks;
* correlation IDs;
* trace propagation.

## Testing

Add:

* service-level tests;
* gRPC integration tests;
* NATS integration tests;
* end-to-end scan flow.

## Migration Strategy

Do not rewrite the entire system.

Extract one bounded context at a time.

Verify behavior after each extraction.

## Acceptance Criteria

The system works with:

```text
API Gateway
      ↓
Discovery Service
      ↓
NATS JetStream
      ↓
Graph Service
      ↓
API Gateway
      ↓
Web
```

and Export Service can operate independently.

## Git

Review architecture and deployment changes carefully.

Commit:

```text
refactor(architecture): extract bounded contexts into microservices
```

This phase completes the initial WeaveLens architecture roadmap.
