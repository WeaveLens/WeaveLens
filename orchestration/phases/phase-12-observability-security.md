# Phase 12 — Observability & Security

## Objective

Prepare WeaveLens for production-style operation.

This phase focuses on observability, credentials, API security, and operational safety.

## Observability

Use OpenTelemetry.

Trace:

```text
HTTP
 ↓
gRPC
 ↓
Application
 ↓
AWS API
 ↓
NATS
 ↓
Graph
```

## Logging

Use structured logs.

Important fields include:

* scan_id;
* trace_id;
* account_id;
* region;
* service/component;
* resource_type;
* duration;
* error.

Never log:

* AWS Secret Access Keys;
* session tokens;
* passwords;
* authorization headers;
* sensitive credential material.

## Metrics

Track:

* scan duration;
* resources discovered;
* AWS API calls;
* throttling;
* scan failures;
* NATS processing;
* graph build duration.

## AWS Credentials

Prefer:

```text
IAM Role
   ↓
STS AssumeRole
   ↓
Temporary Credentials
```

Static credentials may be supported for development through standard AWS credential resolution.

Never persist long-lived credentials in plaintext.

## API Security

Implement appropriate:

* authentication;
* authorization;
* input validation;
* secret redaction;
* audit logging.

Do not invent a complex IAM/RBAC system unless required by the current application.

## Testing

Test:

* secret redaction;
* authorization;
* invalid requests;
* telemetry initialization;
* context propagation;
* credential handling.

## Acceptance Criteria

A complete scan can be traced through the major application components.

Credential material never appears in logs.

Unauthorized operations are rejected.

## Git

Commit:

```text
feat(platform): add observability and security foundations
```

Do not proceed automatically.
