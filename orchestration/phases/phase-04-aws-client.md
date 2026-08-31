# Phase 04 — AWS Client & Infrastructure Layer

## Role

You are a senior Go engineer specializing in AWS SDK architecture.

You are working on WeaveLens as part of an orchestrated multi-agent development workflow.

## Context

Phase 03 established how WeaveLens obtains AWS credentials.

This phase establishes how those credentials are transformed into reusable AWS SDK clients.

The responsibility boundary is:

```text
Phase 03
AWS Identity / Credentials
        ↓
Phase 04
AWS Configuration & Client Factory
        ↓
Phase 05
AWS Resource Discovery
```

## Objective

Implement the AWS client and infrastructure layer using AWS SDK for Go v2.

The client layer must:

* load AWS configuration;
* use the credential strategy from Phase 03;
* configure region;
* construct AWS service clients;
* provide clients to discovery components.

This phase MUST NOT implement resource discovery.

## Architecture

Target structure:

```text
internal/
└── infrastructure/
    └── aws/
        ├── auth/
        ├── config/
        ├── client/
        └── ...
```

Adapt to existing project conventions.

Do not reorganize unrelated code.

## AWS SDK

Use AWS SDK for Go v2.

Do not use deprecated AWS SDK versions.

AWS SDK types must remain inside the infrastructure boundary whenever practical.

Do not expose AWS SDK-specific models through the domain layer.

## AWS Configuration

Create a clear mechanism for constructing AWS configuration.

Conceptually:

```text
Credential Strategy
       ↓
AWS Config
       ↓
AWS Client Factory
```

Configuration should support:

* region;
* credential provider;
* optional AssumeRole configuration.

Do not hard-code AWS regions.

## Client Factory

Provide a mechanism for constructing service clients.

Initial clients should be limited to services required by Phase 05.

Expected candidates include:

```text
EC2
STS
RDS
Elastic Load Balancing v2
```

Only add clients that are actually required.

Do not create clients for every AWS service.

## Client Lifetime

AWS clients should be reusable.

Do not create a new AWS client for every resource API call.

Prefer:

```text
Application
    ↓
AWS Client Factory
    ↓
Reusable Service Clients
    ↓
Discovery
```

rather than:

```text
Scan Resource
    ↓
Create AWS Client
    ↓
API Call
    ↓
Destroy Client
```

## Region

AWS region must be explicitly configurable or resolved through the AWS SDK configuration mechanism.

The implementation must support scanning different regions.

Do not assume:

```text
ap-southeast-1
```

or any other region as a permanent default.

## Account / Identity

The AWS client layer should make it possible for the application to obtain the effective AWS account identity established in Phase 03.

Do not trust user-provided account IDs as proof of identity.

## Error Handling

Return meaningful errors when:

* AWS configuration cannot be loaded;
* credentials cannot be resolved;
* region configuration is invalid;
* STS identity verification fails;
* client initialization fails.

Do not expose secrets.

## Testing

Tests MUST NOT require a real AWS account.

Use interfaces, mocks, or fakes where appropriate.

Test:

* AWS configuration creation;
* region configuration;
* credential provider integration;
* client factory;
* invalid configuration;
* error propagation.

Avoid mocking every AWS SDK internal type unnecessarily.

Keep tests maintainable.

## Dependency Injection

AWS clients must be injected into consumers.

Avoid global AWS clients.

Avoid global mutable state.

Conceptually:

```text
main
 ↓
construct AWS dependencies
 ↓
construct application services
 ↓
inject dependencies
```

## Constraints

Do NOT implement:

* VPC scanning;
* EC2 scanning;
* RDS scanning;
* ALB scanning;
* graph construction;
* NATS;
* gRPC changes;
* frontend changes;
* database.

Phase 04 is infrastructure preparation for Phase 05.

## Acceptance Criteria

1. WeaveLens can construct AWS SDK clients using Phase 03 credentials.
2. AWS clients can be configured for different regions.
3. Clients are reusable.
4. Clients are injected rather than globally created.
5. AWS SDK implementation remains inside the infrastructure boundary.
6. No resource discovery logic is implemented.
7. Tests do not require real AWS credentials.

## Verification

Run:

```bash
go build ./...
go test ./...
```

Run the project's configured linting tools.

Review the complete diff.

## Commit

Create exactly one focused commit:

```text
feat(aws): add aws client factory
```

Do NOT automatically proceed to Phase 05.
