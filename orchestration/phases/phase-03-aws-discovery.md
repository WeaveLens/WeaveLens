# Phase 03 — AWS Resource Discovery

## Objective

Implement AWS infrastructure discovery using AWS SDK for Go v2.

## Initial Resources

Support:

* VPC
* Subnet
* Route Table
* Internet Gateway
* NAT Gateway
* Security Group
* EC2
* RDS
* ALB

Do not attempt to support every AWS service.

## Architecture

Create a scanner abstraction.

Conceptually:

```go
type Scanner interface {
    Scan(ctx context.Context) ([]domain.Resource, error)
}
```

Each AWS service scanner should have a clear responsibility.

Example:

```text
internal/infrastructure/aws/
├── client/
├── vpc/
├── subnet/
├── ec2/
├── rds/
└── alb/
```

Adapt as necessary.

## AWS SDK

Use AWS SDK for Go v2.

AWS SDK models must not leak into the domain layer.

Map:

```text
AWS SDK model
      ↓
AWS adapter
      ↓
domain.Resource
```

## Credentials

Support standard AWS credential resolution.

Prefer IAM Role / STS AssumeRole for production-style usage.

Never:

* log credentials;
* store credentials in source code;
* return credentials from APIs;
* commit credentials.

## Region

Scanning must support explicit AWS regions.

Do not assume a single hard-coded region.

## Error Handling

Return meaningful errors.

Do not silently ignore AWS API failures.

Prepare the design for future retry/throttling handling, but do not implement complex resilience yet.

## Testing

Use mocks or fakes.

Unit tests MUST NOT require a real AWS account.

Test:

* AWS response mapping;
* resource normalization;
* scanner behavior;
* error propagation.

## Constraints

Do not:

* create a database;
* create microservices;
* add NATS;
* add frontend;
* expose AWS SDK types through public application contracts.

## Acceptance Criteria

The system can discover the initial AWS resource types and convert them into canonical WeaveLens resources.

## Git

Run:

```bash
go build ./...
go test ./...
```

Commit:

```text
feat(discovery): add aws resource discovery
```

Do not proceed to Phase 04 automatically.
