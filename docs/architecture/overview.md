# Architecture Overview

WeaveLens is designed as a modular monolith that may evolve into microservices in the future.

## Core Components

- **Discovery**: AWS resource discovery layer
- **Graph**: Infrastructure relationship graph construction
- **API**: gRPC service layer
- **Transport**: NATS event streaming
- **Web**: HTTP/GraphQL visualization layer
- **Export**: Graph export utilities

## Technology Stack

- Language: Go
- Architecture: Modular monolith
- Communication: gRPC, NATS
- Visualization: Web interface

## Constraints

- No external dependencies unless necessary
- Clear separation between internal packages
- No business logic in main package
