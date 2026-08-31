# Phase 07 — gRPC Contracts

## Objective

Introduce Protocol Buffers and gRPC as the synchronous application/service contract.

## Principle

Use gRPC for synchronous request/response operations.

Do not expose internal Go domain structures directly.

## Proto

Create versioned contracts under:

```text
proto/
```

Suggested structure:

```text
proto/
├── common/
├── discovery/
└── graph/
```

## Initial Operations

Support concepts such as:

```text
StartScan
GetScanStatus
CancelScan
GetGraph
GetResource
```

Only implement operations that are actually needed.

## API Design

Do not expose AWS SDK structures.

Use stable API DTOs.

## Requirements

Implement:

* protobuf definitions;
* generated Go code;
* gRPC server;
* gRPC client;
* validation;
* context propagation;
* error mapping.

## Testing

Add:

* gRPC server tests;
* client/server integration tests;
* validation tests;
* error mapping tests.

## Constraints

Do not convert the application into microservices yet.

The gRPC server may run inside the existing modular monolith.

Do not introduce Kubernetes.

Do not introduce service discovery.

## Acceptance Criteria

A client can start a scan and retrieve graph-related data through the gRPC contract.

## Git

Commit:

```text
feat(grpc): add weavelens service contracts
```

Do not proceed automatically.
