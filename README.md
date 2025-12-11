# TinyList

A lightweight, API-first email list manager with SQLite backend. Designed for resource-constrained environments, running in under 100MB RAM.

## Features

- **Double opt-in** subscription flow with email verification
- **Campaign management** with text and HTML email support
- **Admin UI** for managing subscribers, campaigns, and SMTP settings
- **SQLite storage** - no external database required
- **Pure Go backend** - single binary, no CGO dependencies
- **Minimal footprint** - backend ~64MB RAM, frontend ~16MB RAM

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Ingress                             │
│                    (nginx/traefik)                          │
└─────────────────┬───────────────────────┬───────────────────┘
                  │ /                     │ /api/*
                  ▼                       ▼
┌─────────────────────────┐   ┌─────────────────────────────┐
│       Frontend          │   │          Backend            │
│    (Preact + nginx)     │   │       (Go + SQLite)         │
│                         │   │                             │
│  - Admin dashboard      │   │  - REST API                 │
│  - Subscriber mgmt      │   │  - Email sending (SMTP)     │
│  - Campaign editor      │   │  - Double opt-in flow       │
│  - SMTP settings        │   │  - Rate limiting            │
│                         │   │                             │
│  Port: 80               │   │  Port: 8080                 │
└─────────────────────────┘   └──────────────┬──────────────┘
                                             │
                                             ▼
                              ┌─────────────────────────────┐
                              │      SQLite Database        │
                              │   (PersistentVolume in K8s) │
                              └─────────────────────────────┘
```

### API Structure

| Endpoint | Description |
|----------|-------------|
| `POST /api/subscribe` | Public - User subscription from website forms |
| `GET /api/verify/:token` | Public - Email verification links |
| `GET /api/unsubscribe/:token` | Public - Unsubscribe links |
| `/api/private/*` | Private - Admin API (subscribers, campaigns, settings) |

## Deployment Options

### Option 1: Docker Compose (Simplest)

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
      - "8081:80"
    environment:
      - BACKEND_URL=http://backend:8080
    depends_on:
      - backend
    restart: unless-stopped
```

Create a `config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  public_url: "https://newsletter.example.com"  # Your public URL

database:
  path: "./data/tinylist.db"

sending:
  rate_limit: 10
  max_retries: 3
  batch_size: 100
```

Then run:

```bash
docker compose up -d
```

Access the admin UI at `http://localhost:8081` and configure SMTP settings.

### Option 2: Kubernetes with Helm (Recommended for Production)

```bash
# Install from OCI registry
helm install tinylist oci://ghcr.io/zhisme/tinylist/charts/tinylist \
  --namespace tinylist \
  --create-namespace \
  --set config.publicUrl=https://newsletter.example.com \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set "ingress.hosts[0].host=newsletter.example.com" \
  --set "ingress.hosts[0].paths[0].path=/" \
  --set "ingress.hosts[0].paths[0].pathType=Prefix" \
  --set "ingress.hosts[0].paths[0].service=frontend" \
  --set "ingress.hosts[0].paths[1].path=/api" \
  --set "ingress.hosts[0].paths[1].pathType=Prefix" \
  --set "ingress.hosts[0].paths[1].service=backend"
```

Or with a values file:

```yaml
# values.yaml
config:
  publicUrl: "https://newsletter.example.com"

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: newsletter.example.com
      paths:
        - path: /
          pathType: Prefix
          service: frontend
        - path: /api
          pathType: Prefix
          service: backend
  tls:
    - secretName: tinylist-tls
      hosts:
        - newsletter.example.com

persistence:
  enabled: true
  size: 1Gi
```

```bash
helm install tinylist oci://ghcr.io/zhisme/tinylist/charts/tinylist \
  --namespace tinylist \
  --create-namespace \
  -f values.yaml
```

### Option 3: Docker (Manual)

```bash
# Create network
docker network create tinylist

# Run backend
docker run -d \
  --name tinylist-backend \
  --network tinylist \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  ghcr.io/zhisme/tinylist/backend:latest

# Run frontend
docker run -d \
  --name tinylist-frontend \
  --network tinylist \
  -p 8081:80 \
  -e BACKEND_URL=http://tinylist-backend:8080 \
  ghcr.io/zhisme/tinylist/frontend:latest
```

### Option 4: Build from Source

```bash
# Backend
CGO_ENABLED=0 go build -o tinylist cmd/server/main.go
./tinylist

# Frontend
cd frontend
npm ci
npm run build
# Serve dist/ with nginx or any static file server
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Backend server port | `8080` |
| `BACKEND_URL` | Frontend → Backend URL (nginx proxy) | `http://localhost:8080` |

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
```

SMTP settings are configured via the admin UI (Settings page) and stored in the database.

## Resource Requirements

| Component | Memory Request | Memory Limit | CPU Request | CPU Limit |
|-----------|---------------|--------------|-------------|-----------|
| Backend   | 64Mi          | 128Mi        | 50m         | 200m      |
| Frontend  | 16Mi          | 32Mi         | 10m         | 50m       |

## License

MIT
