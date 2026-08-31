# ADR-001: AWS Credential Provider Strategy

## Status

Accepted

## Context

WeaveLens is an infrastructure discovery and visualization tool that does not own an AWS account. It must obtain AWS credentials at runtime without hard-coding or committing secrets.

The application layer must not know how AWS credentials are obtained. Credential acquisition must be isolated under the AWS infrastructure layer.

## Decision

We will use a **provider-based credential strategy** with the following design:

### Provider Interface

```go
type Provider interface {
    Load(ctx context.Context, region string) (aws.Config, error)
}
```

This simple interface isolates credential acquisition from the rest of the application.

### Implemented Providers

1. **DefaultProvider** — Uses the AWS SDK for Go v2 default credential chain. This supports:
   - `AWS_PROFILE`
   - `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_SESSION_TOKEN`
   - `~/.aws/credentials` and `~/.aws/config`
   - EC2 instance metadata / ECS task role / workload identity

2. **AssumeRoleProvider** — Wraps a base provider and optionally assumes an IAM role via STS AssumeRole. Supports:
   - `AWS_ROLE_ARN`
   - `AWS_ROLE_SESSION_NAME` (defaults to `weavelens-session`)
   - `AWS_EXTERNAL_ID`

### Identity Verification

After loading credentials, the system calls `sts:GetCallerIdentity` to verify the effective AWS identity. This provides:
- Account ID
- ARN
- User ID

This identity is informational only and must not be treated as authorization.

### Configuration

AWS configuration is loaded from the application config:

```go
type Config struct {
    AWSRegion        string
    AWSRoleARN       string
    AWSRoleSessionName string
    AWSExternalID    string
}
```

Configuration is loaded from environment variables:
- `AWS_REGION` / `AWS_DEFAULT_REGION`
- `AWS_ROLE_ARN`
- `AWS_ROLE_SESSION_NAME`
- `AWS_EXTERNAL_ID`

### Error Handling

Typed errors distinguish between:
- Missing credentials
- Invalid credentials
- Expired credentials
- AssumeRole failure
- AccessDenied
- Invalid Role ARN
- STS failure

Error messages never expose secret material.

## Consequences

### Positive

- Credential acquisition is isolated in the infrastructure layer
- The application layer only knows about configuration, not credential mechanics
- Future credential strategies (IAM Identity Center, Workload Identity) can be added by implementing the `Provider` interface
- Local development is supported through standard AWS mechanisms
- Cross-account access is supported via AssumeRole

### Negative

- Slightly more complex than passing `aws.Config` directly
- Requires careful testing with mocked STS endpoints

## Security Considerations

- Never hard-code credentials
- Never commit credentials
- Never log secret access keys or session tokens
- Never return credentials through an API
- Never expose credentials to the frontend
- Never store long-lived credentials in application state unnecessarily

## Testing

Tests use mocked HTTP endpoints for STS and do not require real AWS credentials:
- Default provider configuration
- Invalid configuration
- Role configuration
- Role ARN validation
- STS identity mapping
- AssumeRole behavior with mocks
- Credential error handling
- Secret redaction

## References

- [AWS SDK for Go v2 Configuration](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/)
- [AWS STS AssumeRole](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole.html)
- [AWS STS GetCallerIdentity](https://docs.aws.amazon.com/STS/latest/APIReference/API_GetCallerIdentity.html)
