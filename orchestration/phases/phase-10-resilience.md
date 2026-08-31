# Phase 10 — Concurrency & Resilience

## Objective

Improve AWS scanning performance and reliability using Go concurrency and distributed-system resilience patterns.

## Requirements

Implement where justified:

* bounded worker pool;
* context cancellation;
* errgroup;
* rate limiting;
* retry;
* exponential backoff;
* jitter;
* AWS throttling handling;
* timeouts.

## Concurrency

Never create unbounded goroutines.

Concurrency limits must be configurable.

Example:

```text
Scan
 ├── Worker 1
 ├── Worker 2
 ├── Worker 3
 └── Worker N
```

## Cancellation

A cancelled scan must propagate cancellation through:

```text
API
 ↓
Application
 ↓
Scanner
 ↓
AWS API calls
 ↓
Workers
```

Workers must terminate promptly.

## Retry

Retry only errors that are appropriate for retry.

Do not blindly retry every error.

Use exponential backoff with jitter.

## Partial Failure

Define behavior when one scanner fails while others succeed.

The scan result should clearly represent partial failure rather than silently hiding it.

## Testing

Test:

* cancellation;
* worker limits;
* retry;
* throttling;
* timeout;
* partial scanner failure;
* context propagation.

## Acceptance Criteria

Multiple AWS resource scanners can operate concurrently with bounded resource usage and safe cancellation.

## Git

Commit:

```text
feat(discovery): add concurrent resilient scanning
```

Do not proceed automatically.
