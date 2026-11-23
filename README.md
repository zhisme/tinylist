# TinyList

A lightweight, API-first email list manager with SQLite backend. Designed for resource-constrained environments, running in under 100MB RAM.

## Features

- **Configurable opt-in flow** — double opt-in (email verification) or single opt-in
- **Campaign management** with text and HTML email support
- **Admin UI** for managing subscribers, campaigns, and SMTP settings
- **Basic Auth** protection for admin endpoints
- **SQLite storage** - no external database required
- **Pure Go backend** - single binary, no CGO dependencies
- **Minimal footprint** - backend ~64MB RAM, frontend ~16MB RAM

## Quick Start

### Kubernetes with Helm (Recommended)

```bash
helm install tinylist oci://ghcr.io/zhisme/tinylist/charts/tinylist \
  --namespace tinylist \
  --create-namespace \
  --set config.publicUrl=https://example.com \
  --set config.auth.password=your-secure-password \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set "ingress.hosts[0].host=example.com"
```

Access the admin UI at `https://example.com/tinylist`

### Docker

The config file is optional — you can configure everything with environment variables:

```bash
docker run -d \
  -p 8080:8080 \
  -e TINYLIST_AUTH_PASSWORD=changeme \
  -e TINYLIST_SERVER_PUBLIC_URL=https://newsletter.example.com \
  -v tinylist-data:/app/data \
  ghcr.io/zhisme/tinylist/backend:latest
```

See the [Environment Variables](#environment-variables) reference for the full list.

### Docker Compose

```yaml
# docker-compose.yml
services:
  backend:
    image: ghcr.io/zhisme/tinylist/backend:latest
    ports:
      - "8080:8080"
    environment:
      TINYLIST_AUTH_PASSWORD: changeme
      TINYLIST_SERVER_PUBLIC_URL: https://newsletter.example.com
    volumes:
      - ./data:/app/data
    restart: unless-stopped

  frontend:
    image: ghcr.io/zhisme/tinylist/frontend:latest
    ports:
      - "8081:8080"
    depends_on:
      - backend
    restart: unless-stopped
```

> **Note:** A reverse proxy (nginx/traefik) is needed to route `/tinylist/api/*` to the backend and `/tinylist/*` to the frontend. Alternatively, you can use a `config.yaml` file (see [Configuration Reference](#configuration-reference)) and mount it as a volume if you prefer file-based config.

```bash
docker compose up -d
```

Access the admin UI at `http://localhost:8081/tinylist`, log in with your configured credentials, and configure SMTP in Settings.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Ingress                             │
│                    (nginx/traefik)                          │
└───────────────┬─────────────────────────┬───────────────────┘
                │ /tinylist/api/*         │ /tinylist/*
                ▼                         ▼
┌───────────────────────────┐   ┌─────────────────────────────┐
│         Backend           │   │         Frontend            │
│      (Go + SQLite)        │   │     (Preact + nginx)        │
│                           │   │                             │
│  - REST API               │   │  - Admin dashboard          │
│  - Email sending (SMTP)   │   │  - Static files only        │
│  - Configurable opt-in flow │   │                             │
│  - Basic Auth             │   │  Port: 8080                 │
│                           │   └─────────────────────────────┘
│  Port: 8080               │
└─────────────┬─────────────┘
              │
              ▼
┌─────────────────────────────┐
│      SQLite Database        │
│   (PersistentVolume in K8s) │
└─────────────────────────────┘
```

### URL Structure

All routes are served under `/tinylist`:

| URL | Description |
|-----|-------------|
| `/tinylist/` | Admin UI (frontend static files) |
| `/tinylist/api/subscribe` | Public subscription endpoint |
| `/tinylist/api/verify/:token` | Email verification links |
| `/tinylist/api/unsubscribe/:token` | Unsubscribe links |
| `/tinylist/api/private/*` | Admin API (Basic Auth protected) |

### API Endpoints

The backend serves API routes directly at `/tinylist/api/*` (configurable via `api_base_path`):

| Endpoint | Auth | Description |
|----------|------|-------------|
| `POST /tinylist/api/subscribe` | Public | User subscription from website forms |
| `GET /tinylist/api/verify/:token` | Public | Email verification links |
| `GET /tinylist/api/unsubscribe/:token` | Public | Unsubscribe links |
| `/tinylist/api/private/*` | Basic Auth | Admin API (subscribers, campaigns, settings) |

## Helm Deployment

```yaml
# values.yaml
config:
  publicUrl: "https://example.com"
  auth:
    username: admin
    password: "your-secure-password"

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: example.com
  tls:
    - secretName: tinylist-tls
      hosts:
        - example.com
```

```bash
helm install tinylist oci://ghcr.io/zhisme/tinylist/charts/tinylist \
  --namespace tinylist \
  --create-namespace \
  -f values.yaml
```

The admin UI will be available at `https://example.com/tinylist`

### Build from Source

```bash
# Backend
CGO_ENABLED=0 go build -o tinylist cmd/server/main.go
./tinylist

# Frontend (for development)
cd frontend
npm ci
npm run dev

# Frontend (for production)
npm run build
# Serve dist/ with nginx or any static file server
```

## Configuration Reference

Configuration is loaded in the following priority order (highest wins):

1. **Environment variables** — always applied last, override everything
2. **`config.yaml`** — optional; if the file is absent, defaults + env vars are used
3. **Built-in defaults**

This means a Docker deployment can omit the config file entirely.

### config.yaml

The config file is optional. All fields have environment variable equivalents (see below).

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  public_url: "https://newsletter.example.com"  # Used in email links and CORS
  api_base_path: ""         # Set to "/tinylist" for subpath deployments
  require_verification: true  # Set to false to skip the double opt-in confirmation email
  # extra_origins:           # Additional CORS origins for local development only
  #   - "http://localhost:5173"

database:
  path: "./data/tinylist.db"

sending:
  rate_limit: 10        # Emails per second
  max_retries: 3        # Retry failed sends
  batch_size: 100       # Subscribers per batch

# REQUIRED — server will not start without auth.password
# In Kubernetes, credentials are injected via a Secret instead (see Helm section).
auth:
  username: admin
  password: your-secure-password
```

> SMTP settings (host, port, credentials, from address) are configured through the **admin UI → Settings** page and stored in the database. There are no env vars or config file fields for SMTP.

### Environment Variables

Every `config.yaml` field has a corresponding `TINYLIST_*` environment variable. Env vars take precedence over the config file.

| Variable | Config file equivalent | Default | Notes |
|---|---|---|---|
| `TINYLIST_AUTH_PASSWORD` | `auth.password` | *(none)* | **Required** |
| `TINYLIST_AUTH_USERNAME` | `auth.username` | `admin` | |
| `TINYLIST_SERVER_PUBLIC_URL` | `server.public_url` | `http://localhost:8080` | Used in email links & CORS |
| `TINYLIST_SERVER_HOST` | `server.host` | `0.0.0.0` | |
| `TINYLIST_SERVER_PORT` | `server.port` | `8080` | |
| `TINYLIST_SERVER_API_BASE_PATH` | `server.api_base_path` | `""` | e.g. `/tinylist` |
| `TINYLIST_SERVER_REQUIRE_VERIFICATION` | `server.require_verification` | `true` | `true` or `false` |
| `TINYLIST_SERVER_EXTRA_ORIGINS` | `server.extra_origins` | `""` | Comma-separated; dev only |
| `TINYLIST_DATABASE_PATH` | `database.path` | `./data/tinylist.db` | |
| `TINYLIST_SENDING_RATE_LIMIT` | `sending.rate_limit` | `10` | Emails per second |
| `TINYLIST_SENDING_MAX_RETRIES` | `sending.max_retries` | `3` | |
| `TINYLIST_SENDING_BATCH_SIZE` | `sending.batch_size` | `100` | |

### Helm Values

In the Helm chart, `auth.username` and `auth.password` are stored in a Kubernetes **Secret** (not the ConfigMap) and injected into the pod via the env vars above. All other settings go in the ConfigMap via `values.yaml`.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.publicUrl` | Public URL for email links | `""` |
| `config.apiBasePath` | Base path for API routes | `"/tinylist"` |
| `config.auth.username` | Admin username (stored in Secret) | `admin` |
| `config.auth.password` | Admin password — **required** (stored in Secret) | `""` |
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `""` |
| `persistence.enabled` | Enable SQLite persistence | `true` |
| `persistence.size` | PVC size | `1Gi` |

See [values.yaml](helm/tinylist/values.yaml) for all options.

## Resource Requirements

| Component | Memory Request | Memory Limit | CPU Request | CPU Limit |
|-----------|---------------|--------------|-------------|-----------|
| Backend   | 64Mi          | 128Mi        | 50m         | 200m      |
| Frontend  | 16Mi          | 32Mi         | 10m         | 50m       |

**Total: ~80Mi request, ~160Mi limit**

## TinyList vs listmonk

| Feature | TinyList | listmonk |
|---------|----------|----------|
| Memory usage | ~80MB | ~500MB+ |
| Database | SQLite | PostgreSQL |
| Setup complexity | Single binary + config | Multiple services |
| Multiple lists | No (single list) | Yes |
| Templates | Basic (text/HTML) | Advanced templating |
| Analytics | Basic stats | Detailed analytics |
| Bounce handling | No | Yes |
| Media uploads | No | Yes |

**Choose TinyList if**: You need a simple newsletter for a small site, run on limited resources, or want minimal operational overhead.

**Choose listmonk if**: You need multiple lists, advanced templating, detailed analytics, or enterprise features.

## License

MIT
