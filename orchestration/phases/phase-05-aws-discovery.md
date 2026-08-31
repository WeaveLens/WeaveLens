# Phase 05 — AWS Resource Discovery

## Role

You are a senior Go engineer specializing in cloud infrastructure discovery and clean architecture.

You are working on the WeaveLens project as part of an orchestrated multi-agent development workflow.

## Context

WeaveLens is a cloud infrastructure discovery and visualization platform.

The current implementation focuses exclusively on AWS.

Azure and GCP may be supported in the future.

The architecture must therefore establish reasonable provider boundaries without implementing multi-cloud support prematurely.

Previous phases established:

```text
Phase 03
AWS Authentication & Credential Strategy
        ↓
Phase 04
AWS Client & Infrastructure Layer
        ↓
Phase 05
AWS Resource Discovery
```

The target flow is:

```text
Cloud Provider
      ↓
Resource Discovery
      ↓
Canonical WeaveLens Resources
      ↓
Relationships
      ↓
Graph Engine
```

## Objective

Implement real AWS resource discovery using the AWS clients established in Phase 04.

The implementation must:

1. call real AWS APIs;
2. discover supported AWS resources;
3. normalize AWS resources into WeaveLens domain resources;
4. discover reliable relationships between resources;
5. remain independent from HTTP, gRPC, NATS, and frontend code;
6. establish provider-neutral boundaries that can support Azure and GCP later.

## Multi-Cloud Boundary

WeaveLens is currently AWS-only.

Do NOT implement Azure or GCP.

However, avoid making the application/domain layer fundamentally dependent on AWS terminology.

Prefer provider-neutral concepts where they represent genuine domain concepts.

Example:

```text
Resource
Provider
ResourceType
ResourceID
Region
Account
Metadata
Relationship
```

AWS-specific implementation belongs under the AWS infrastructure/provider boundary.

Conceptually:

```text
internal/
├── domain/
│   └── resource/
│
├── application/
│   └── discovery/
│
└── infrastructure/
    └── cloud/
        └── aws/
            ├── auth/
            ├── client/
            └── discovery/
```

Adapt this structure to the existing repository instead of reorganizing unrelated code.

## Discovery Interface

Define an application-facing discovery abstraction.

Conceptually:

```go
type ResourceDiscovery interface {
    Discover(
        ctx context.Context,
        request DiscoveryRequest,
    ) (DiscoveryResult, error)
}
```

The exact interface must follow the existing project's architecture.

The application-facing interface MUST NOT expose:

* AWS SDK types;
* AWS SDK clients;
* AWS-specific request/response models.

AWS-specific details remain behind the infrastructure/provider implementation.

## AWS Implementation

Implement an AWS-specific discovery adapter.

Conceptually:

```text
ResourceDiscovery
       ▲
       │
AWS Resource Discovery
       │
       ▼
AWS Client Factory
       │
       ▼
AWS SDK
       │
       ▼
AWS
```

## Initial Resource Scope

Implement discovery for:

### Networking

* VPC
* Subnet
* Route Table
* Internet Gateway
* NAT Gateway
* Security Group

### Compute

* EC2 Instance

### Database

* RDS Instance

### Load Balancing

* Application Load Balancer

Do not attempt to discover every AWS service.

The implementation should make adding another AWS resource type straightforward.

## Resource Normalization

Convert AWS SDK objects into canonical WeaveLens resources.

Conceptually:

```text
AWS SDK Object
      ↓
AWS Adapter
      ↓
Canonical Resource
```

The domain model must NOT import AWS SDK packages.

Example conceptual resource:

```text
Resource
├── ID
├── Provider
├── Type
├── Name
├── Region
├── Account
├── Metadata
└── Attributes
```

Do not force every cloud provider to have identical resource semantics.

Provider-specific information may be preserved in metadata/attributes where appropriate.

## Resource Identity

Use stable resource identifiers.

Prefer:

* ARN;
* AWS resource ID;

depending on the resource.

Do not use display names as primary identity.

A resource name is metadata, not identity.

## AWS Provider

Every discovered resource must be identifiable as belonging to AWS.

Conceptually:

```text
Provider = AWS
```

Do not hard-code `"aws"` throughout unrelated application code.

Use an appropriate provider representation from the domain model.

## Relationships

Discover relationships only when they can be established reliably from AWS data.

Examples:

```text
VPC
 └── contains → Subnet

VPC
 └── contains → Route Table

VPC
 └── contains → Internet Gateway

VPC
 └── contains → NAT Gateway

Subnet
 └── contains → EC2 Instance

Route Table
 └── associated_with → Subnet

Route Table
 └── routes_to → Internet Gateway

RDS
 └── associated_with → VPC / Subnet Group

ALB
 └── associated_with → VPC / Subnets
```

Do not invent relationships merely because two resources exist in the same VPC.

## Pagination

Correctly handle AWS API pagination.

The scanner must not assume that a single AWS API response contains all resources.

Test pagination behavior.

## Region

Discovery must use the region supplied by the AWS client/configuration layer.

Do not hard-code a region.

The design should allow future scanning of multiple regions.

Multi-region orchestration itself is not required in this phase unless already established by the project.

## Account Identity

Use the effective AWS identity established by Phase 03.

Resources should be associated with the actual AWS account being scanned.

Do not trust an account ID supplied by the frontend as authoritative identity.

## API Efficiency

Avoid unnecessary AWS API calls.

Use AWS API filters when appropriate.

Avoid obvious N×M API call patterns.

Where relationships require additional API calls, document the reason.

Do not prematurely optimize at the cost of correctness.

## Concurrency

Use concurrency only where it provides a clear benefit.

Respect:

```text
context.Context
```

and AWS API throttling.

Do not create an uncontrolled goroutine for every AWS resource.

If bounded concurrency is needed, use the project's established concurrency utilities.

## Error Handling

Handle:

* AccessDenied;
* ResourceNotFound;
* throttling;
* transient AWS errors;
* malformed responses;
* context cancellation;
* partial scanner failures.

Do not silently ignore errors.

The result must distinguish between:

```text
Complete discovery
```

and:

```text
Partial discovery
```

where appropriate.

## Partial Failure

One failed AWS resource scanner must not automatically destroy successfully discovered resources from other scanners.

Example:

```text
VPC       ✓
Subnet    ✓
EC2       ✓
RDS       ✗ AccessDenied
ALB       ✓
```

The discovery result should preserve successful resources and expose the failure appropriately.

Follow existing project error/result conventions.

## Context Cancellation

Every AWS API operation must receive and respect:

```go
context.Context
```

A cancelled scan must stop unnecessary work.

## Testing

Unit tests MUST NOT require a real AWS account.

Test:

* AWS resource mapping;
* resource identity;
* provider assignment;
* pagination;
* empty results;
* API failures;
* AccessDenied;
* throttling;
* partial failures;
* context cancellation;
* relationship construction;
* duplicate resources.

Use interfaces/fakes/mocks at appropriate boundaries.

Do not over-mock internal implementation details.

## Optional AWS Integration Test

If integration testing already exists, an optional real-AWS integration test may be added.

It MUST:

* be explicitly opt-in;
* never run as part of normal unit tests;
* never require committed credentials;
* use credentials supplied externally through the AWS SDK credential chain.

Do not make CI depend on a personal AWS account.

## Security

Never log:

* Access Keys;
* Secret Access Keys;
* Session Tokens;
* Authorization headers;
* credential configuration values.

Do not include credential material in discovery results.

## Transport Independence

Discovery MUST NOT depend on:

* HTTP;
* REST;
* gRPC;
* NATS;
* Vue;
* JSON response models.

Discovery produces application/domain data.

Transport layers are responsible for serialization.

## Graph Boundary

This phase may produce relationship data suitable for the Graph Engine.

Do NOT redesign or implement the Graph Engine in this phase.

The intended boundary is:

```text
AWS Discovery
      ↓
Resources + Relationships
      ↓
Graph Engine
```

## No Database

Do not introduce:

* PostgreSQL;
* Redis;
* DynamoDB;
* any persistent database.

Resource discovery remains runtime/in-memory.

Persistent storage is outside the scope of this phase.

## Acceptance Criteria

1. WeaveLens can authenticate using the credential strategy from Phase 03.
2. AWS clients are obtained through Phase 04.
3. Real AWS APIs are called.
4. Initial AWS resource types are discovered.
5. AWS API pagination is handled.
6. Resources are normalized into canonical WeaveLens resources.
7. Resources identify AWS as their provider.
8. Reliable relationships are discovered.
9. Partial failures are represented appropriately.
10. Context cancellation is supported.
11. Unit tests do not require AWS credentials.
12. AWS SDK types do not leak into the domain layer.
13. Discovery does not depend on HTTP, gRPC, NATS, or Vue.
14. The design leaves a clean boundary for future Azure/GCP providers.
15. Azure/GCP implementation is NOT added in this phase.

## Verification

Run:

```bash
go build ./...
go test ./...
```

Run all project-configured linters and static analysis tools.

Review the complete diff.

Verify that no credentials or secrets were added to source control.

## Git

Create exactly one focused commit:

```text
feat(discovery): implement aws resource discovery
```

Do NOT automatically proceed to the next phase.
