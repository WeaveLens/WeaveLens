# Phase 00 — Project Foundation

## Role

You are a senior Go engineer working on the WeaveLens project.

You are operating as part of an orchestrated multi-agent development workflow.

## Objective

Initialize the WeaveLens project from an empty repository.

At this phase, establish the project structure, development conventions, documentation foundation, and Go module.

Do NOT implement AWS discovery, graph logic, NATS, gRPC business logic, or microservices yet.

## Product Context

WeaveLens will eventually:

* discover AWS infrastructure resources;
* discover relationships between resources;
* construct an infrastructure graph;
* visualize the graph on the web;
* optionally export the graph;
* evolve from a modular monolith into microservices.

The initial implementation MUST remain a modular monolith.

## Tasks

### 1. Initialize Go

Create:

* `go.mod` with Go 1.26
* appropriate Go version
* basic module configuration

### 2. Create project structure

Use a clean structure similar to:

```text
cmd/
internal/
proto/
web/
docs/
tests/
orchestration/
```

Do not create unnecessary packages yet.

### 3. Create documentation and `.gitignore`

Create:

```text
docs/
├── architecture/
└── development/
```

Also create `.gitignore` covering:

* Go build artifacts (`bin/`, `*.test`, `*.out`, `weavelens`)
* Go module cache (`.cache/`, `vendor/`)
* IDE and OS files (`.idea/`, `.vscode/`, `.DS_Store`, `Thumbs.db`)
* Environment files (`.env`, `.env.local`)

Document the initial project direction.

### 4. Create Makefile

Provide commands for at least:

```text
make build
make test
make lint
make run
```

Only add commands that can actually work.

### 5. Create README

README should explain:

* what WeaveLens is;
* current project status;
* high-level roadmap;
* development prerequisites;
* how to run the project.

## Constraints

* Do not add AWS SDK yet.
* Do not add NATS yet.
* Do not add gRPC implementation yet.
* Do not add database.
* Do not create microservices.
* Do not introduce unnecessary dependencies.
* Do not modify unrelated tooling.

## Testing

Verify:

```bash
go build ./...
go test ./...
```

Both must succeed.

## Acceptance Criteria

* Go project builds successfully.
* Repository structure is established.
* README exists.
* Makefile exists.
* `.gitignore` exists and covers Go artifacts, IDE files, and environment files.
* Architecture direction is documented.
* No unnecessary runtime dependency has been introduced.

## Git

Review the complete diff before committing.

Create exactly one focused commit:

```text
chore: initialize weavelens project
```

Do not proceed to Phase 01 automatically.
