# Development Conventions

## Go Style

- Use standard `gofmt` formatting.
- Keep packages focused and small.
- Avoid unnecessary abstractions.
- Prefer interfaces in the consumer package.

## Commit Messages

- Use conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `test:`, `refactor:`
- Keep commits focused and atomic.

## Build & Test

- Run `make build` before committing.
- Run `make test` and ensure all tests pass.
- Run `make lint` to check for issues.

## Module Structure

- `cmd/` — application entrypoints
- `internal/` — private application packages
- `proto/` — protocol buffer definitions
- `web/` — web assets and handlers
- `tests/` — integration and end-to-end tests
- `orchestration/` — phase documentation and workflows
