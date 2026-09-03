# WeaveLens

WeaveLens is an infrastructure observability tool that discovers AWS resources, maps their relationships, and visualizes them in a web-based graph interface.

## Features

- **AWS Resource Discovery**: Automatically discovers EC2, RDS, ALB, VPC, Subnet, Security Group, and other AWS resources
- **Relationship Mapping**: Builds a graph of resource relationships (contains, connects to, depends on, etc.)
- **Web Visualization**: Interactive graph visualization using Cytoscape.js
- **Graph Export**: Export infrastructure graphs as JSON, Draw.io, or SVG
- **Concurrent Scanning**: Bounded worker pool for parallel AWS API calls
- **Resilience**: Retry with exponential backoff, rate limiting, and graceful handling of throttling
- **Security**: Secret redaction, security headers, API key authentication, and credential source tracking

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Web UI    │────▶│  HTTP API   │────▶│ Application │
│  (Vue.js)   │     │   (Go)      │     │   Services  │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                         ┌─────────────────────┼─────────────────────┐
                         │                     │                     │
                         ▼                     ▼                     ▼
                  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
                  │  Discovery  │      │    Graph    │      │   Export    │
                  │  Service    │      │   Service   │      │   Service   │
                  └──────┬──────┘      └─────────────┘      └─────────────┘
                         │
                         ▼
                  ┌─────────────┐
                  │  AWS SDK    │
                  │  (EC2,RDS,  │
                  │   ELBv2)    │
                  └─────────────┘
```

## Prerequisites

- Go 1.24+
- Node.js 18+ (for web frontend)
- NATS Server with JetStream enabled
- AWS credentials configured

## Quick Start

### 1. Start NATS Server

Using Docker:

```bash
docker run -d --name nats -p 4222:4222 -p 8222:8222 nats:latest -js
```

Or install NATS locally and run with JetStream:

```bash
nats-server -js
```

### 2. Configure AWS Credentials

Credential resolution order (highest priority first):

1. **Environment variables** (`AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY`)
2. **AWS_PROFILE** env var → uses that named profile
3. **Default profile** → uses `weavelens` profile from `~/.aws/credentials`
4. **IAM Role** (EC2/ECS instance role, no config needed)

Option A - AWS Profile `weavelens` (default):

```bash
# ~/.aws/credentials
[weavelens]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

```bash
# ~/.aws/config
[profile weavelens]
region = us-east-1
```

```bash
# From project root (WeaveLens/)
go run ./cmd/weavelens
```

The app uses the `weavelens` profile by default. If it doesn't exist, it falls back to `default`.

Option B - Custom profile via env:

```bash
export AWS_PROFILE=my-profile
export AWS_REGION=us-east-1
go run ./cmd/weavelens
```

Option C - Environment variables (no profile):

```bash
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_REGION=us-east-1
go run ./cmd/weavelens
```

When `AWS_ACCESS_KEY_ID` is set, the app uses environment credentials and ignores the profile setting.

Option D - LocalStack (local AWS emulator):

```bash
# Start LocalStack
docker run -d --name localstack -p 4566:4566 localstack/localstack

# Configure dummy credentials + endpoint
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_REGION=us-east-1
go run ./cmd/weavelens
```

Option E - IAM Role (EC2/ECS):

```bash
# No credentials needed - uses instance role
export AWS_REGION=us-east-1
go run ./cmd/weavelens
```

Option F - STS Assume Role:

```bash
export AWS_REGION=us-east-1
export AWS_ROLE_ARN=arn:aws:iam::123456789012:role/WeaveLensRole
go run ./cmd/weavelens
```

### 3. Start Backend

From the project root directory:

```bash
go run ./cmd/weavelens
```

### 4. Start Frontend

```bash
cd web
npm ci
npm run dev
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `ENV` | `development` | Environment name |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `NATS_URL` | `nats://localhost:4222` | NATS server URL |
| `AWS_REGION` | - | AWS region |
| `AWS_PROFILE` | `weavelens` | AWS profile name from `~/.aws/credentials` |
| `AWS_ACCESS_KEY_ID` | - | Override with env credentials (bypasses profile) |
| `AWS_SECRET_ACCESS_KEY` | - | Override with env credentials (bypasses profile) |
| `AWS_ENDPOINT_URL` | - | Custom endpoint (e.g., LocalStack) |
| `AWS_ROLE_ARN` | - | Optional IAM role ARN for AssumeRole |
| `AWS_ROLE_SESSION_NAME` | `weavelens-session` | Session name for assumed role |
| `AWS_EXTERNAL_ID` | - | External ID for role assumption |
| `API_KEY` | - | Optional API key for endpoint authentication |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/ready` | Readiness check |
| `GET` | `/api/connection` | AWS connection status |
| `POST` | `/api/scans` | Start a new scan |
| `GET` | `/api/scans/{scanId}/status` | Get scan status |
| `GET` | `/api/scans/{scanId}/graph` | Get resource graph |
| `GET` | `/api/scans/{scanId}/export` | Export graph (json, drawio, svg) |
| `GET` | `/api/resources/{resourceId}` | Get resource details |
| `GET` | `/api/resources/{resourceId}/relationships` | Get resource relationships |

## Development

### Git Hooks (Lefthook)

This project uses [Lefthook](https://github.com/evilmartians/lefthook) for managing Git hooks.
Pre-commit hooks verify Go formatting, run `go vet`, and check frontend types.

Install Lefthook (one-time):

```bash
go install github.com/evilmartians/lefthook@latest
```

Install hooks:

```bash
make install-hooks
```

To skip hooks for a commit:

```bash
git commit -k "message"
```

```bash
# Run tests
make test

# Build
make build

# Lint
make lint

# Run locally
make run
```

## Security

- AWS credentials are never logged or exposed to the frontend
- Secret redaction prevents credential leakage in logs
- Security headers (X-Frame-Options, X-Content-Type-Options, etc.)
- Optional API key authentication
- Error messages are sanitized before reaching clients

## License

MIT
