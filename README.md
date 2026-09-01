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
- NATS Server (with JetStream enabled)
- AWS credentials configured

## Quick Start

### Backend

```bash
# Set environment variables
export AWS_REGION=us-east-1
export NATS_URL=nats://localhost:4222

# Run the server
go run ./cmd/weavelens
```

### Frontend

```bash
cd web
npm ci
npm run dev
```

### Docker (optional)

```bash
docker-compose up
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
