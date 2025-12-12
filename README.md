# TinyList

A lightweight, API-first email list manager with SQLite backend. Designed for resource-constrained environments, running in under 100MB RAM.

## Features

- **Double opt-in** subscription flow with email verification
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

### Docker Compose

```yaml
# docker-compose.yml
services:
  backend:
    image: ghcr.io/zhisme/tinylist/backend:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./config.yaml:/app/config.yaml:ro
    restart: unless-stopped

  frontend:
    image: ghcr.io/zhisme/tinylist/frontend:latest
    ports:
      - "8081:8080"
    depends_on:
      - backend
    restart: unless-stopped
```

Note: For Docker Compose, you'll need a reverse proxy (nginx/traefik) to route `/tinylist/api/*` to backend and `/tinylist/*` to frontend.

Create `config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  public_url: "https://newsletter.example.com"

database:
  path: "./data/tinylist.db"

sending:
  rate_limit: 10
  max_retries: 3
  batch_size: 100

# REQUIRED - server will not start without auth password
auth:
  username: admin
  password: your-secure-password
```

```bash
docker compose up -d
```

Access admin UI at `http://localhost:8081/tinylist`, login with your configured credentials, and configure SMTP in Settings.

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
│  - Double opt-in flow     │   │                             │
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

### config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  public_url: "https://newsletter.example.com"  # Required for email links

database:
  path: "./data/tinylist.db"

sending:
  rate_limit: 10        # Emails per second
  max_retries: 3        # Retry failed sends
  batch_size: 100       # Subscribers per batch

# REQUIRED - server will not start without this
auth:
  username: admin
  password: your-secure-password
```

### Helm Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.publicUrl` | Public URL for email links | `""` |
| `config.apiBasePath` | Base path for API routes | `"/tinylist"` |
| `config.auth.username` | Admin username | `admin` |
| `config.auth.password` | Admin password (required) | `""` |
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `""` |
| `persistence.enabled` | Enable SQLite persistence | `true` |
| `persistence.size` | PVC size | `1Gi` |

See [values.yaml](helm/tinylist/values.yaml) for all options.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Backend server port | `8080` |

SMTP settings are configured via the admin UI (Settings page) and stored in the database.

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
