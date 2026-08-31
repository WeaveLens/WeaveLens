# Phase 06 — Infrastructure Graph Engine

## Objective

Build the graph engine that transforms discovered resources and relationships into an in-memory infrastructure topology.

## Responsibilities

The graph engine must support:

* node insertion;
* edge insertion;
* lookup;
* duplicate detection;
* relationship traversal;
* filtering;
* graph snapshot retrieval.

## Example

Input:

```text
VPC
Subnet
EC2
RDS
ALB
```

Output:

```text
VPC
├── contains → Subnet
│              ├── contains → EC2
│              └── contains → RDS
│
└── ...
```

## Storage

Use in-memory storage.

Do NOT add PostgreSQL or another persistent database.

The graph represents a temporary scan snapshot.

## Idempotency

Repeated insertion of the same resource must not create duplicates.

Repeated insertion of the same relationship must not create duplicate edges.

## Separation

Graph logic must remain independent from:

* AWS SDK;
* HTTP;
* gRPC;
* NATS;
* Vue.

## Testing

Test:

* node insertion;
* duplicate nodes;
* edge insertion;
* duplicate edges;
* lookup;
* traversal;
* filtering;
* empty graph;
* large graph behavior where practical.

## Acceptance Criteria

A complete graph can be constructed from canonical resources and relationships without AWS-specific knowledge.

## Git

Commit:

```text
feat(graph): add infrastructure graph engine
```

Do not proceed automatically.
