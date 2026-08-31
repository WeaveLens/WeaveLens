# Phase 05 — Real AWS Resource Discovery

## Role

You are a senior Go engineer specializing in AWS infrastructure discovery.

You are working on WeaveLens as part of an orchestrated multi-agent development workflow.

## Context

Previous phases established:

```text
Phase 03
AWS Credential Strategy
        ↓
Phase 04
AWS Client Factory
        ↓
Phase 05
AWS Resource Discovery
```

WeaveLens must now make real AWS API calls.

## Objective

Implement the first production-oriented AWS resource discovery capability.

The discovery layer must:

1. use AWS clients created by Phase 04;
2. call real AWS APIs;
3. normalize AWS resources into WeaveLens domain resources;
4. discover relevant relationships;
5. return canonical WeaveLens resources;
6. remain independent from transport and presentation layers.

## Initial Resource Scope

Start with:

### Networking

* VPC
* Subnet
* Route Table
* Internet Gateway
* NAT Gateway
* Security Group

### Compute

* EC2

### Database

* RDS

### Load Balancing

* Application Load Balancer

Do NOT attempt to discover every AWS service.

## Architecture

Target flow:

```text
Application
    ↓
Discovery Interface
    ↓
AWS Discovery Adapter
    ↓
AWS Client Factory
    ↓
AWS SDK
    ↓
AWS
```

The frontend, gRPC layer, and NATS must not directly call AWS SDK APIs.

## Discovery Interface

Define an application-facing abstraction.

Conceptually:

```go
type ResourceDiscovery interface {
    Discover(ctx context.Context, request DiscoveryRequest) (DiscoveryResult, error)
}
```

Adapt the exact API to existing domain conventions.

The interface must not expose AWS SDK-specific types.

## Service Scanners

Create focused scanners.

Conceptually:

```text
AWS Discovery
├── VPC Scanner
├── Subnet Scanner
├── Route Table Scanner
├── Internet Gateway Scanner
├── NAT Gateway Scanner
├── Security Group Scanner
├── EC2 Scanner
├── RDS Scanner
└── ALB Scanner
```

Each scanner should have one clear responsibility.

Avoid creating one huge:

```text
aws_scanner.go
```

containing all AWS services.

## Resource Normalization

Convert AWS SDK objects into canonical WeaveLens resources.

Conceptually:

```text
AWS SDK Object
      ↓
AWS Adapter
      ↓
domain.Resource
```

The domain layer must not import AWS SDK packages.

## Relationships

Discover relationships where AWS data provides sufficient evidence.

Examples:

```text
VPC
 └── contains → Subnet

Subnet
 └── contains → EC2

VPC
 └── contains → Internet Gateway

VPC
 └── contains → NAT Gateway

Route Table
 └── routes_to → Internet Gateway

ALB
 └── targets → Target Group / EC2 where applicable

RDS
 └── belongs_to → VPC / Subnet Group where applicable
```

Do not invent relationships that cannot be reliably established.

## Resource Identity

Use stable identifiers.

Prefer AWS identifiers such as:

* ARN;
* resource ID;

where appropriate.

Do not use display names as primary identity.

## Region

Discovery must operate against the configured AWS region.

Do not hard-code a region.

## Account

Associate discovered resources with the effective AWS account identity obtained through the authentication layer.

Do not trust an account ID supplied by the user as authoritative identity.

## Pagination

AWS APIs may return paginated results.

Discovery MUST correctly handle pagination for APIs that support it.

Do not assume the first response contains all resources.

## AWS API Efficiency

Avoid unnecessary API calls.

Use appropriate AWS API filters when available.

Do not implement naive:

```text
N × M
```

API call patterns without justification.

## Error Handling

Define behavior for:

* AccessDenied;
* ResourceNotFound;
* throttling;
* transient AWS errors;
* malformed responses;
* partial scanner failures;
* context cancellation.

Do not silently swallow errors.

A partial scan must be distinguishable from a successful complete scan.

## Context

All AWS API calls must accept and propagate:

```go
context.Context
```

A cancelled scan must stop unnecessary AWS work.

## Testing

Tests MUST NOT depend on a real AWS account.

Create unit tests using mocks/fakes for AWS API interactions.

Test:

* resource mapping;
* pagination;
* empty results;
* API failures;
* AccessDenied;
* partial failures;
* context cancellation;
* relationship construction;
* duplicate resource handling.

## Optional Integration Test

If the repository already supports integration testing, provide an optional real-AWS integration test.

It MUST:

* be explicitly opt-in;
* never run during normal unit tests;
* never require committed credentials.

For example, use an environment flag.

Do not make CI depend on a personal AWS account.

## Security

Never log:

* access keys;
* secret keys;
* session tokens;
* authorization headers.

Do not return credentials in discovery results.

## Output

Discovery should produce canonical WeaveLens data suitable for:

```text
Discovery
    ↓
Graph Engine
```

Do not couple the output to:

* Vue;
* HTTP JSON;
* gRPC;
* NATS.

Transport-specific serialization belongs elsewhere.

## Constraints

Do NOT:

* add a database;
* persist AWS resources permanently;
* add frontend logic;
* directly expose AWS SDK types;
* introduce microservices;
* redesign the graph engine unnecessarily;
* introduce NATS event publishing unless it already belongs to an existing integration point.

NATS event integration will be handled by its dedicated phase.

## Acceptance Criteria

1. WeaveLens can authenticate using the credential strategy from Phase 03.
2. WeaveLens can construct AWS clients using Phase 04.
3. WeaveLens can make real AWS API calls.
4. Initial AWS resource types are discovered.
5. Pagination is handled correctly.
6. AWS resources are normalized into canonical WeaveLens resources.
7. Reliable relationships are discovered.
8. Partial failures are represented explicitly.
9. Context cancellation works.
10. Unit tests do not require AWS credentials.
11. No credential material is logged or exposed.
12. Discovery remains independent of HTTP, gRPC, NATS, and Vue.

## Manual Verification

With valid user-provided AWS credentials available through the standard AWS credential chain, verify that WeaveLens can:

```text
1. authenticate
2. identify the AWS account
3. connect to the configured region
4. discover resources
5. produce canonical resources
6. produce relationships
7. construct a graph
```

Credentials must be supplied externally.

Never commit them.

## Verification

Run:

```bash
go build ./...
go test ./...
```

Run linting.

If an opt-in AWS integration test exists, document exactly how to run it.

Review the complete diff.

## Git

Create exactly one focused commit:

```text
feat(discovery): implement real aws resource discovery
```

Do NOT automatically proceed to Phase 06.
