# Phase 07 — Web Visualization

## Objective

Build the first usable WeaveLens web application.

## Technology

Use:

* Vue 3
* TypeScript
* Vite
* Pinia
* TanStack Query
* Cytoscape.js

## UI

Implement:

* scan configuration;
* scan status;
* topology visualization;
* zoom;
* pan;
* search;
* filtering;
* resource detail panel;
* legend.

## Graph

The frontend consumes the canonical graph representation.

The frontend MUST NOT depend on AWS SDK models.

## Visualization Metadata

Centralize resource visualization metadata.

Example categories:

```text
Compute
Network
Database
Storage
Security
Integration
Other
```

Do not scatter resource-specific colors/styles throughout components.

## UX

The user should be able to:

1. start a scan;
2. see scan progress;
3. view topology;
4. select a resource;
5. inspect relationships;
6. filter resource categories;
7. search resources.

## Testing

Add appropriate:

* component tests;
* store tests;
* graph transformation tests;
* API integration tests where practical.

## Acceptance Criteria

A WeaveLens scan can be rendered as an interactive infrastructure graph.

## Constraints

Do not introduce unnecessary frontend frameworks.

Do not add a database.

Do not couple UI code to AWS SDK.

## Git

Commit:

```text
feat(web): add infrastructure topology visualization
```

Do not proceed automatically.
