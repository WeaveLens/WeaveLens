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

Option A - Using AWS Profile:

```bash
# ~/.aws/credentials
[weavelens]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

```bash
# From project root (WeaveLens/)
AWS_PROFILE=weavelens AWS_REGION=us-east-1 go run ./cmd/weavelens
```

Or export variables first:

```bash
export AWS_PROFILE=weavelens
export AWS_REGION=us-east-1
go run ./cmd/weavelens
```

Option B - Using environment variables:

```bash
# From project root (WeaveLens/)
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_REGION=us-east-1
go run ./cmd/weavelens
```

Or inline:

```bash
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE AWS_SECRET_ACCESS_KEY=... AWS_REGION=us-east-1 go run ./cmd/weavelens
```

Option C - Using LocalStack (local AWS emulator):

```bash
# Start LocalStack
docker run -d --name localstack -p 4566:4566 localstack/localstack

# Configure and run
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_REGION=us-east-1
go run ./cmd/weavelens
```

Option C - Using IAM Role (EC2/ECS):

```bash
# From project root (WeaveLens/)
# No credentials needed - uses instance role
export AWS_REGION=us-east-1
go run ./cmd/weavelens
```

Option D - Using STS Assume Role:

```bash
# From project root (WeaveLens/)
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
