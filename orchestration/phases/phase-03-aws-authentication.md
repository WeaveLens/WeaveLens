# Phase 03 — AWS Authentication & Credential Strategy

## Role

You are a senior Go engineer specializing in AWS identity and security.

You are working on the WeaveLens project as part of an orchestrated multi-agent development workflow.

## Context

WeaveLens is an infrastructure discovery and visualization tool.

The system will eventually:

```text
AWS
 ↓
Discovery
 ↓
Canonical Resources
 ↓
Graph
 ↓
gRPC / NATS
 ↓
Web Visualization
```

WeaveLens does NOT own an AWS account and must NOT depend on a hard-coded AWS credential.

The AWS identity must be supplied at runtime.

## Objective

Design and implement the AWS credential acquisition layer.

This phase is ONLY responsible for:

* obtaining AWS credentials;
* selecting a credential strategy;
* validating AWS identity;
* supporting future credential strategies.

This phase MUST NOT implement AWS resource discovery.

## Architectural Principle

The application layer must not know how AWS credentials are obtained.

Use an infrastructure abstraction.

Conceptually:

```text
Application
    ↓
AWS Credential / Client abstraction
    ↓
AWS SDK
```

Credential acquisition must be isolated under the AWS infrastructure layer.

## Initial Credential Strategy

Use the AWS SDK for Go v2 default credential chain as the primary mechanism.

Support local development through standard AWS mechanisms such as:

```text
AWS_PROFILE
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
AWS_SESSION_TOKEN
~/.aws/credentials
~/.aws/config
```

Do not manually parse credential files unless there is a compelling reason.

Let the AWS SDK resolve credentials.

## AssumeRole

Support STS AssumeRole as the preferred cross-account mechanism.

Example:

```text
User AWS Identity
       ↓
STS AssumeRole
       ↓
WeaveLensScanner Role
       ↓
Temporary Credentials
       ↓
AWS APIs
```

Support configuration such as:

```text
AWS_ROLE_ARN
AWS_ROLE_SESSION_NAME
AWS_EXTERNAL_ID
```

Only implement configuration that is actually required.

## Identity Verification

Use STS GetCallerIdentity to verify the effective AWS identity.

The system should be able to determine:

* Account ID
* ARN
* User ID

This identity is informational and must not be treated as authorization by itself.

Authorization is determined by AWS IAM.

## Security Requirements

NEVER:

* hard-code credentials;
* commit credentials;
* log secret access keys;
* log session tokens;
* return credentials through an API;
* expose credentials to the frontend;
* store credentials in the database;
* store long-lived credentials in application state unnecessarily.

The frontend must never directly call AWS using WeaveLens-managed credentials.

## Access Key Policy

Static access keys may be supported for development through the AWS SDK credential chain.

They must NOT become the only supported credential strategy.

Do not create a frontend form that requires users to submit long-lived AWS credentials in this phase.

## Credential Provider Design

Create a design that can support future strategies:

```text
Default Credential Chain
Static Credentials
AssumeRole
IAM Identity Center
Workload Identity
```

Do not implement unused providers merely for abstraction completeness.

Prefer simple interfaces.

Do not create a large generic authentication framework.

## Configuration

Use the project's existing configuration mechanism.

Do not introduce another configuration framework without justification.

Configuration must distinguish between:

```text
Credential source
Role ARN
Role session configuration
AWS region
```

Do not mix AWS authentication configuration with discovery-specific configuration.

## Error Handling

Errors must clearly distinguish:

* missing credentials;
* invalid credentials;
* expired credentials;
* AssumeRole failure;
* AccessDenied;
* invalid Role ARN;
* STS failure.

Do not expose secret material in errors.

## Testing

Tests MUST NOT require a real AWS account.

Test:

* default credential configuration;
* invalid configuration;
* role configuration;
* Role ARN validation;
* STS identity mapping;
* AssumeRole behavior using mocks/fakes;
* credential error handling;
* secret redaction.

## Documentation

Add an ADR documenting:

* supported credential strategies;
* default credential chain;
* AssumeRole;
* security considerations;
* local development;
* future enterprise authentication options.

Suggested location:

```text
docs/adr/
```

## Constraints

Do NOT implement:

* VPC discovery;
* EC2 discovery;
* RDS discovery;
* ALB discovery;
* graph construction;
* NATS;
* frontend changes;
* database.

## Acceptance Criteria

The implementation must satisfy all of the following:

1. WeaveLens can obtain credentials through the AWS SDK default credential chain.
2. WeaveLens can optionally assume a configured IAM Role.
3. The effective AWS identity can be verified through STS.
4. Credentials never reach the frontend.
5. Credentials never appear in logs.
6. AWS credential logic is isolated from domain logic.
7. Tests do not require real AWS credentials.
8. The design allows future credential strategies without modifying discovery logic.

## Verification

Run:

```bash
go build ./...
go test ./...
```

Run linting using the project's configured tooling.

Review the complete diff.

## Commit

Create one focused commit:

```text
feat(auth): add aws credential provider strategy
```

Do NOT automatically proceed to Phase 04.
