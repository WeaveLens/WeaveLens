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
├── scans/                    # host-mounted scan-history directory (scans.json lives here)
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

## Push images to Docker Hub

Create two Docker Hub repositories, for example `weavelens-backend` and
`weavelens-frontend`, then log in using your Docker Hub username and access
token:

```bash
export DOCKERHUB_USERNAME=your-dockerhub-username
export IMAGE_TAG=v1.0.0

docker login --username "$DOCKERHUB_USERNAME"
```

Tag the locally built images with both a version and `latest`, then push them:

```bash
docker tag weavelens/backend:latest "$DOCKERHUB_USERNAME/weavelens-backend:$IMAGE_TAG"
docker tag weavelens/backend:latest "$DOCKERHUB_USERNAME/weavelens-backend:latest"
docker tag weavelens/frontend:latest "$DOCKERHUB_USERNAME/weavelens-frontend:$IMAGE_TAG"
docker tag weavelens/frontend:latest "$DOCKERHUB_USERNAME/weavelens-frontend:latest"

docker push "$DOCKERHUB_USERNAME/weavelens-backend:$IMAGE_TAG"
docker push "$DOCKERHUB_USERNAME/weavelens-backend:latest"
docker push "$DOCKERHUB_USERNAME/weavelens-frontend:$IMAGE_TAG"
docker push "$DOCKERHUB_USERNAME/weavelens-frontend:latest"
```

For deployment, replace the local `image:` values in `docker-compose.yml` with
the Docker Hub references. Prefer a fixed version tag for reproducible releases:

```yaml
services:
  backend:
    image: your-dockerhub-username/weavelens-backend:v1.0.0
  frontend:
    image: your-dockerhub-username/weavelens-frontend:v1.0.0
```

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

## 3. Scan history (`scans.json`)

WeaveLens persists scan history to a JSON file. By default it resolves the path
relative to `go.mod`; in Docker neither binary ships a `go.mod`, so the path is
pinned explicitly:

```yaml
environment:
  WEAVELENS_HISTORY_FILE: /app/data/scans.json
volumes:
  - ./scans:/app/data
```

The host directory `./scans/` is bind-mounted into `/app/data`, so the
generated `scans.json` file is persisted on the host and survives container
recreates/restarts:

```bash
# The persisted scan history
cat scans/scans.json

# Reset history (delete the file, the app recreates it on next scan)
rm -f scans/scans.json
docker compose restart backend
```

All runtime files under `./scans/` are ignored by Git. The `.gitkeep`
placeholder remains tracked so the mount directory exists after cloning.

## AWS credentials

The backend uses the AWS SDK, so any mechanism the SDK supports works. Pass the
relevant variables to the `backend` service via `environment:` (or an `.env`
file). Examples:

By default, Compose mounts the host's `${HOME}/.aws` directory read-only and
runs the backend with `${HOST_UID:-1000}:${HOST_GID:-1000}` so credential files
with mode `0600` remain readable without weakening their host permissions. Set
`HOST_UID` and `HOST_GID` when the host account does not use UID/GID `1000`:

```bash
HOST_UID=$(id -u) HOST_GID=$(id -g) docker compose up -d
```

```bash
# Option A: static keys
AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... AWS_REGION=us-east-1 docker compose up -d

# Option B: AssumeRole
AWS_REGION=us-east-1 AWS_ROLE_ARN=arn:aws:iam::123456789012:role/WeaveLensRole docker compose up -d

# Option C: LocalStack (local AWS emulator)
AWS_ENDPOINT_URL=http://host.docker.internal:4566 AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1 docker compose up -d
```

The backend service maps `host.docker.internal` to Docker's host gateway, so
the LocalStack endpoint works on Linux Docker Engine as well as Docker Desktop.

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

- Both application containers run as **non-root**: Compose runs the backend with
  `${HOST_UID:-1000}:${HOST_GID:-1000}` so it can read the host's mode-`0600`
  AWS credentials, while the frontend runs as the `nginx` user (uid `101`) on
  the unprivileged port `8080` (exposed as host `5173`).
- The `nats` container exposes only its JetStream port `4222`; the NATS
  monitoring port is not published.
- No host socket (`/var/run/docker.sock`) is mounted — the backend discovers
  AWS resources via the AWS SDK, not the Docker API.
- `scans.json` is treated as runtime data: the `scans/` directory is bind-mounted
  at runtime and excluded from images via `.dockerignore`.
