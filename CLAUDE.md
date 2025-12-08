# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**Note**: This project uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking. Use `bd` commands instead of markdown TODOs. See AGENTS.md for workflow details.

## Project Overview

TinyList is a lightweight, API-first email list manager with SQLite backend designed for resource-constrained environments. It runs in <100MB RAM and provides essential email list functionality with both a Go backend and a Preact-based admin frontend.

## Tech Stack

**Backend:**
- Go 1.21+ with `modernc.org/sqlite` (pure Go, no CGO)
- `github.com/go-chi/chi` for routing
- `gopkg.in/gomail.v2` for SMTP
- SQLite3 for data storage

**Frontend:**
- Preact 10.19+ (React-compatible, ~3KB)
- Vite 7.2.6 for build tooling
- Tailwind CSS for styling
- Preact Router (~2KB) for routing
- Native fetch API for HTTP requests

## Project Architecture

### API Namespace Design

The API uses two distinct namespaces with important semantic differences:

**Public API (no namespace prefix):**
- `POST /api/subscribe` - User self-subscription from website forms
- `GET /api/verify/:token` - Email verification links
- `GET /api/unsubscribe/:token` - Unsubscribe links
- Rate-limited to prevent abuse
- Token-based authentication for verify/unsubscribe
- These routes are intentionally short for brevity in user-facing links

**Private API (explicit `/api/private/*` namespace):**
- `/api/private/subscribers/*` - Admin subscriber CRUD
- `/api/private/campaigns/*` - Campaign management
- `/api/private/settings/*` - SMTP configuration
- `/api/private/stats/*` - Analytics and dashboard statistics
- Network-level isolation in K8s (v1.0), API keys in v1.1+
- Explicit namespace for security clarity

### Directory Structure

```
tinylist/
├── cmd/server/main.go           # Application entry point
├── internal/
│   ├── db/                      # Database layer (SQLite queries, migrations)
│   ├── models/                  # Data models (subscriber, campaign)
│   ├── handlers/
│   │   ├── private/             # Admin API handlers
│   │   └── public/              # Public API handlers (no /public/ in routes)
│   ├── mailer/                  # SMTP client and email templates
│   └── config/                  # Configuration loading
├── frontend/                    # Separate Preact UI (deployed as separate container)
│   ├── src/
│   │   ├── components/          # Dashboard, Subscribers, Campaigns, Settings
│   │   └── api.js               # API client calling /api/private/* endpoints
│   ├── vite.config.js
│   └── Dockerfile               # Multi-stage: Node build → nginx serve
├── schema.sql                   # SQLite schema with 4 tables
├── Dockerfile                   # Backend container (Go multi-stage build)
└── k8s/                        # Kubernetes manifests (backend, frontend, ingress)
```

### Database Schema

Uses 4 simple tables:
- `subscribers` - Email list with status (pending/verified/unsubscribed), tokens
- `campaigns` - Email campaigns with subject, body (text/html), send stats
- `campaign_logs` - Simple send log (sent/failed per subscriber)
- `settings` - Key-value config storage for SMTP settings

Total: ~50-100KB empty, scales linearly (vs listmonk's 15+ tables)

## Common Development Commands

### Backend (Go)

```bash
# Run development server
go run cmd/server/main.go

# Build binary
go build -o tinylist cmd/server/main.go

# Run with CGO disabled (for pure Go SQLite)
CGO_ENABLED=0 go build -o tinylist cmd/server/main.go

# Run tests
go test ./...

# Run tests for specific package
go test ./internal/db
```

### Frontend (Preact)

```bash
cd frontend

# Install dependencies
npm ci

# Development server (port 5173)
npm run dev

# Production build (outputs to dist/)
npm run build

# Preview production build
npm run preview
```

### Docker

```bash
# Build backend image
docker build -t tinylist-backend:latest .

# Build frontend image
docker build -t tinylist-frontend:latest ./frontend

# Run backend locally
docker run -p 8080:8080 -v $(pwd)/data:/data tinylist-backend:latest
```

### Kubernetes

```bash
# Apply all manifests
kubectl apply -f k8s/

# Check pod status
kubectl get pods -l app=tinylist-backend
kubectl get pods -l app=tinylist-frontend

# View logs
kubectl logs -f deployment/tinylist-backend
kubectl logs -f deployment/tinylist-frontend

# Port forward for local testing
kubectl port-forward svc/tinylist-backend 8080:8080
kubectl port-forward svc/tinylist-frontend 8081:80
```

## Configuration

Configuration is loaded via `config.yaml` with environment variable overrides:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  public_url: "https://newsletter.example.com"  # For verification links

database:
  path: "./data/tinylist.db"

smtp:
  host: "smtp.gmail.com"
  port: 587
  username: "your-email@gmail.com"
  password: "your-app-password"
  from_email: "newsletter@example.com"
  from_name: "Newsletter"
  tls: true

sending:
  rate_limit: 10        # Emails per second
  max_retries: 3
  batch_size: 100
```

Environment variables use prefix `TINYLIST_` with nested paths using underscores:
- `TINYLIST_SERVER_PORT=8080`
- `TINYLIST_SMTP_HOST=smtp.gmail.com`
- `TINYLIST_DATABASE_PATH=/data/tinylist.db`

## Key Implementation Details

### Double Opt-in Flow

1. User POSTs to `/api/subscribe` (public, from website form)
2. Backend creates subscriber with status="pending", generates verify_token
3. Sends verification email with link: `/api/verify/{token}`
4. User clicks link → status changes to "verified"
5. Every campaign email includes unsubscribe link: `/api/unsubscribe/{token}`

### Campaign Sending

- Worker processes campaigns with rate limiting (configurable emails/second)
- Uses batch processing (default 100 subscribers at a time)
- Template variables: `{{name}}`, `{{email}}` in body_text and body_html
- Status tracking: draft → sending → sent/failed
- Logs each send attempt in `campaign_logs` table

### Frontend-Backend Communication

- Frontend is a separate Docker container (nginx serving static files)
- Makes API calls to `/api/private/*` endpoints
- In K8s: Ingress routes `/admin/*` to frontend, `/api/*` to backend
- Frontend build is ~45-50KB gzipped total

## Resource Targets

### Memory Usage
- Backend: 45-80 MB (Go binary + SQLite + HTTP server + campaign worker)
- Frontend: 10-15 MB (nginx + static files)
- **Total: 55-95 MB** (well under 100MB goal)

### Kubernetes Resources
Backend:
```yaml
requests: { memory: 64Mi, cpu: 50m }
limits: { memory: 128Mi, cpu: 200m }
```

Frontend:
```yaml
requests: { memory: 16Mi, cpu: 10m }
limits: { memory: 32Mi, cpu: 50m }
```

## Out of Scope (v1.0)

- Multiple lists (single list only)
- Campaign templates (plain text + basic HTML only)
- Advanced analytics (basic stats only)
- Bounce handling
- User authentication (single tenant, network-level security)
- Media uploads

## Security Notes

### v1.0 Approach
- **No authentication** - Single tenant, relies on K8s network isolation
- **Rate limiting** - Aggressive on public endpoints (e.g., 5 req/IP/hour for subscribe)
- **Token security** - Cryptographically random tokens (crypto/rand)
- **SQL injection** - Use parameterized queries (Go handles automatically)
- **Input validation** - Email format, length limits

### Future (v1.1+)
- API key authentication for private endpoints
- Session-based login for frontend
- HTTPS enforcement via ingress
- CORS configuration

## Implementation Notes

When working with this codebase:

1. **Follow the "tiny" philosophy** - Avoid adding dependencies or complexity unnecessarily
2. **Public routes have no namespace** - `/api/subscribe` not `/api/public/subscribe` (brevity for user-facing URLs)
3. **Private routes are explicit** - Always use `/api/private/*` prefix for admin endpoints
4. **Pure Go SQLite** - Use `modernc.org/sqlite` to avoid CGO dependencies
5. **Simple JSON responses** - No fancy serialization, standard library is enough
6. **Template variables** - Only `{{name}}` and `{{email}}` supported in campaign bodies
7. **Self-hosted focus** - No API versioning needed (users control deployment)
8. **React-compatible Preact** - Use standard React patterns for contributor-friendliness

## Deployment Model

This is a **self-hosted, open-source** application:
- Users run their own instances (no SaaS)
- Frontend and backend are separate containers
- Ingress handles routing between them
- SQLite data persists via PVC
- Users control when to upgrade (semantic versioning)

## Testing Strategy

- Unit tests for database queries, email templates, token generation
- Integration tests for API endpoints, SMTP mock testing
- Manual testing for real SMTP sending and K8s deployment
- Load testing target: 1000+ subscribers, 100+ emails/minute

## Notes for AI assistant
- Always prioritize simplicity and minimalism in code changes
- Ensure security best practices are followed, especially for public endpoints
- Use conventional commits principles for commit messages
- When adding features, consider resource constraints (keep memory usage low)
- Refer to AGENTS.md for workflow and issue tracking using `bd` commands
- Maintain clear separation between public and private API routes
