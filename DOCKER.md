# Docker

WeaveLens ships as two container images orchestrated with Docker Compose. The stack is
**image-driven**: every service is started from an `image:` reference (no `build:` keys in the
compose file), so the images must be built once and tagged first.

## Services

| Service   | Image                  | Container name        | Port (host) | Purpose                                  |
|-----------|------------------------|-----------------------|-------------|------------------------------------------|
| `nats`    | `nats:2.10-alpine`     | `weavelens-nats`      | `4222`      | NATS + JetStream message bus (required)  |
| `backend` | `weavelens/backend:latest` | `weavelens-backend` | `8080`      | Go API server (AWS discovery engine)     |
| `frontend`| `weavelens/frontend:latest` | `weavelens-frontend`| `5173`      | Vue.js UI served by nginx (non-root)     |

## Prerequisites

- Docker Engine 24+ and Docker Compose v2
- An AWS credential provider that the AWS SDK can resolve inside the `backend`
  container (see [AWS credentials](#aws-credentials)).

## Project layout

```
.
├── docker-compose.yml        # image-driven compose (image: refs, container_name set)
├── docker/
│   ├── backend/Dockerfile    # builds weavelens/backend:<tag>
│   └── frontend/
│       ├── Dockerfile        # builds weavelens/frontend:<tag>
│       └── nginx.conf        # nginx reverse proxy: /api/* -> backend:8080
├── scans/                    # host-mounted scan-history directory (.scans.json lives here)
└── DOCKER.md                 # this file
```

## 1. Build the images

Images are built from the dedicated Dockerfiles and tagged with explicit names
so the compose file can reference them via `image:`.

```bash
# From the repository root
docker build -t weavelens/backend:latest    -f docker/backend/Dockerfile    .
docker build -t weavelens/frontend:latest   -f docker/frontend/Dockerfile   .
```

> The compose file does **not** contain `build:` blocks, so it never rebuilds on
> `docker compose up`. Re-run the `docker build` commands above after changing
> `cmd/weavelens` or `web/src` and then `docker compose up -d --force-recreate`.

## 2. Run the stack

```bash
docker compose up -d
```

- Backend API: http://localhost:8080
  - Health: `curl http://localhost:8080/health`
- Web UI: http://localhost:5173
- The frontend reverse-proxies `/api/*` (including the `/api/scans/stream`
  Server-Sent Events feed) to the backend container over the Compose network.

Startup order is controlled in compose:

- `nats` starts first (`condition: service_started`).
- `backend` depends on `nats` started **and** exposes a `curl /health` readiness
  check, so the app is only considered ready once the HTTP server is up.
- `frontend` depends on `backend` being `service_healthy`.

> The backend exits if it cannot connect to NATS. NATS starts in well under a
> second; if there is ever a startup race, `restart: unless-stopped` retries the
> container automatically and the second attempt connects successfully.

## 3. Scan history (`.scans.json`)

WeaveLens persists scan history to a JSON file. By default it resolves the path
relative to `go.mod`; in Docker neither binary ships a `go.mod`, so the path is
pinned explicitly:

```yaml
environment:
  WEAVELENS_HISTORY_FILE: /app/data/.scans.json
volumes:
  - ./scans:/app/data
```

The host directory `./scans/` is bind-mounted into `/app/data`, so the
generated `.scans.json` file is persisted on the host and survives container
recreates/restarts:

```bash
# The persisted scan history
cat scans/.scans.json

# Reset history (delete the file, the app recreates it on next scan)
rm -f scans/.scans.json
docker compose restart backend
```

The file is git-ignored repo-wide (`/.scans.json` in `.gitignore`), and
`./scans/` contains only a `.gitkeep` placeholder.

## AWS credentials

The backend uses the AWS SDK, so any mechanism the SDK supports works. Pass the
relevant variables to the `backend` service via `environment:` (or an `.env`
file). Examples:

```bash
# Option A: static keys
AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... AWS_REGION=us-east-1 docker compose up -d

# Option B: AssumeRole
AWS_REGION=us-east-1 AWS_ROLE_ARN=arn:aws:iam::123456789012:role/WeaveLensRole docker compose up -d

# Option C: LocalStack (local AWS emulator)
AWS_ENDPOINT_URL=http://host.docker.internal:4566 AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1 docker compose up -d
```

When no AWS credentials/region are provided, the backend still starts, logs a
warning (`continuing without AWS connection`), and serves the UI and an empty
scan history — useful for a quick smoke test.

## Useful commands

```bash
docker compose ps                 # show containers + health
docker compose logs -f backend    # follow backend logs
docker compose logs -f frontend
docker compose stop               # stop stack (config is preserved)
docker compose down               # stop + remove containers/network
docker compose down -v            # ...and remove anonymous volumes
docker compose up -d --force-recreate  # rebuild containers from your images
```

## Security

- Both application containers run as **non-root**: the backend as uid `65534`
  and the frontend (nginx) as the `nginx` user (uid `101`) listening on the
  unprivileged port `8080` inside the container (exposed as host `5173`).
- The `nats` container exposes only its JetStream port `4222`; the NATS
  monitoring port is not published.
- No host socket (`/var/run/docker.sock`) is mounted — the backend discovers
  AWS resources via the AWS SDK, not the Docker API.
- `.scans.json` is treated as runtime data: the `scans/` directory is bind-mounted
  at runtime and excluded from images via `.dockerignore`.
