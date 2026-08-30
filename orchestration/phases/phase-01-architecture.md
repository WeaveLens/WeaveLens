# Phase 01 — Architecture & Domain Model

## Objective

Define the core WeaveLens domain model independently from AWS SDK, transport protocols, and infrastructure.

## Architectural Principle

The domain layer MUST NOT import:

* AWS SDK;
* NATS;
* gRPC;
* HTTP frameworks;
* database drivers.

The domain represents cloud infrastructure generically.

## Domain Concepts

Implement the core concepts:

* Resource
* ResourceType
* ResourceCategory
* Relationship
* RelationshipType
* Graph
* Scan

## Resource

A resource should support concepts such as:

* ID
* ARN
* Account ID
* Region
* Type
* Category
* Name
* Metadata
* Tag

Do not over-design fields that are not currently required.

## Resource Categories

Initially support:

* compute
* network
* database
* storage
* security
* integration
* other

## Relationship Types

Initially support concepts such as:

* contains
* belongs_to
* connects_to
* routes_to
* targets
* depends_on
* associated_with

## Graph

The graph must support:

* adding nodes;
* adding relationships;
* retrieving nodes;
* retrieving relationships;
* duplicate prevention;
* lookup by ID.

The graph is initially in-memory.

## Scan

A scan represents one temporary infrastructure snapshot.

A scan should have a stable ID.

Do not persist scans to a database.

## Package Structure

Prefer a structure similar to:

```text
internal/
└── domain/
    ├── resource/
    ├── relationship/
    ├── graph/
    └── scan/
```

Adapt the structure if the existing repository conventions justify another organization.

## Testing

Add unit tests for:

* resource creation;
* relationship creation;
* graph insertion;
* duplicate node handling;
* duplicate relationship handling;
* lookup;
* scan identity.

## Constraints

* No AWS SDK.
* No NATS.
* No gRPC.
* No HTTP.
* No database.
* No frontend work.
* Do not couple domain types to infrastructure.

## Acceptance Criteria

The domain model can represent:

```text
VPC
 └── Subnet
      └── EC2
```

and:

```text
ALB
 └── Target Group
      └── EC2
```

without importing AWS-specific packages.

## Git

Run:

```bash
go build ./...
go test ./...
```

Review the diff.

Create exactly one commit:

```text
feat(domain): define infrastructure graph model
```

Do not proceed to Phase 02 automatically.
