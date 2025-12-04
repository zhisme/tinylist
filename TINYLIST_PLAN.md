# TinyList - Implementation Plan

A lightweight, API-first email list manager with SQLite backend for resource-constrained environments.

## Project Goals

Build a minimal viable email list manager that:
- Runs in <100MB RAM
- Uses SQLite (no PostgreSQL required)
- Provides essential email list functionality
- API-first (web UI optional for later)
- Production-ready, open-sourceable

---

## Core Features (v1.0)

### Must-Have Features
1. ✅ Subscriber management (add, list, delete)
2. ✅ Double opt-in verification flow
3. ✅ Email campaign sending to verified subscribers
4. ✅ Unsubscribe endpoint
5. ✅ SMTP configuration
6. ✅ SQLite data storage
7. ✅ REST API

### Explicitly Out of Scope (v1.0)
- ❌ Web UI (API only)
- ❌ Multiple lists
- ❌ Campaign templates
- ❌ Analytics/tracking
- ❌ Bounce handling
- ❌ Rich HTML emails (plain text + basic HTML only)
- ❌ User authentication (single tenant)
- ❌ Media uploads

---

## Technical Stack

**Language:** Go 1.21+
**Database:** SQLite3
**Libraries:**
- `modernc.org/sqlite` - Pure Go SQLite driver (no CGO)
- `github.com/go-chi/chi` - Lightweight router
- `github.com/google/uuid` - UUID generation
- `gopkg.in/gomail.v2` - SMTP email sending
- Standard library for everything else

**Why Go:**
- Compiled binary (easy deployment)
- Low memory footprint
- Excellent standard library
- Static typing
- Great concurrency support

---

## Architecture

### Project Structure
```
tinylist/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── db/
│   │   ├── db.go               # Database initialization
│   │   ├── migrations.go       # Schema migrations
│   │   └── queries.go          # SQL queries
│   ├── models/
│   │   ├── subscriber.go       # Subscriber model
│   │   └── campaign.go         # Campaign model
│   ├── handlers/
│   │   ├── subscribers.go      # Subscriber endpoints
│   │   ├── campaigns.go        # Campaign endpoints
│   │   └── public.go           # Public endpoints (verify, unsubscribe)
│   ├── mailer/
│   │   ├── mailer.go           # SMTP client wrapper
│   │   └── templates.go        # Email templates
│   └── config/
│       └── config.go           # Configuration loading
├── schema.sql                   # SQLite schema
├── config.example.yaml          # Example configuration
├── Dockerfile                   # Container image
├── k8s/                        # Kubernetes manifests
│   ├── deployment.yaml
│   └── service.yaml
├── README.md
├── LICENSE
└── go.mod
```

---

## Database Schema

### SQLite Schema (Simple)

```sql
-- subscribers table
CREATE TABLE subscribers (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid            TEXT NOT NULL UNIQUE,
    email           TEXT NOT NULL UNIQUE COLLATE NOCASE,
    name            TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL CHECK(status IN ('pending', 'verified', 'unsubscribed')) DEFAULT 'pending',
    verify_token    TEXT UNIQUE,
    unsubscribe_token TEXT NOT NULL UNIQUE,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    verified_at     TEXT,
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_subscribers_email ON subscribers(email);
CREATE INDEX idx_subscribers_status ON subscribers(status);
CREATE INDEX idx_subscribers_verify_token ON subscribers(verify_token);
CREATE INDEX idx_subscribers_unsubscribe_token ON subscribers(unsubscribe_token);

-- campaigns table
CREATE TABLE campaigns (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid            TEXT NOT NULL UNIQUE,
    subject         TEXT NOT NULL,
    body_text       TEXT NOT NULL,
    body_html       TEXT,
    status          TEXT NOT NULL CHECK(status IN ('draft', 'sending', 'sent', 'failed')) DEFAULT 'draft',
    total_count     INTEGER NOT NULL DEFAULT 0,
    sent_count      INTEGER NOT NULL DEFAULT 0,
    failed_count    INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    started_at      TEXT,
    completed_at    TEXT
);

CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_created_at ON campaigns(created_at);

-- campaign_logs (simple sending log)
CREATE TABLE campaign_logs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id     INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    subscriber_id   INTEGER NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    status          TEXT NOT NULL CHECK(status IN ('sent', 'failed')),
    error           TEXT,
    sent_at         TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(campaign_id, subscriber_id)
);

CREATE INDEX idx_campaign_logs_campaign_id ON campaign_logs(campaign_id);
CREATE INDEX idx_campaign_logs_subscriber_id ON campaign_logs(subscriber_id);

-- settings table (key-value config storage)
CREATE TABLE settings (
    key             TEXT PRIMARY KEY,
    value           TEXT NOT NULL,
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**Total Tables:** 4 (vs listmonk's 15+)
**Total Size:** ~50-100KB empty, scales linearly

---

## API Endpoints

### Subscriber Management

#### `POST /api/subscribers`
Subscribe a new email (sends verification email)

**Request:**
```json
{
  "email": "user@example.com",
  "name": "John Doe"
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "status": "pending",
  "created_at": "2025-11-29T10:00:00Z"
}
```

---

#### `GET /api/subscribers`
List all subscribers (with pagination & filtering)

**Query Params:**
- `status` - Filter by status (pending, verified, unsubscribed)
- `page` - Page number (default: 1)
- `per_page` - Items per page (default: 50, max: 100)

**Response:**
```json
{
  "subscribers": [
    {
      "id": "550e8400-...",
      "email": "user@example.com",
      "name": "John Doe",
      "status": "verified",
      "created_at": "2025-11-29T10:00:00Z",
      "verified_at": "2025-11-29T10:05:00Z"
    }
  ],
  "total": 150,
  "page": 1,
  "per_page": 50
}
```

---

#### `GET /api/subscribers/:id`
Get single subscriber

**Response:**
```json
{
  "id": "550e8400-...",
  "email": "user@example.com",
  "name": "John Doe",
  "status": "verified",
  "created_at": "2025-11-29T10:00:00Z",
  "verified_at": "2025-11-29T10:05:00Z",
  "updated_at": "2025-11-29T10:05:00Z"
}
```

---

#### `DELETE /api/subscribers/:id`
Delete subscriber permanently

**Response:**
```json
{
  "message": "Subscriber deleted"
}
```

---

### Public Endpoints (No Auth Required)

#### `GET /verify/:token`
Verify email subscription

**Response:**
```html
<html>
  <body>
    <h1>Email Verified!</h1>
    <p>Thank you for confirming your subscription.</p>
  </body>
</html>
```

---

#### `GET /unsubscribe/:token`
Unsubscribe from list

**Response:**
```html
<html>
  <body>
    <h1>Unsubscribed</h1>
    <p>You have been removed from the mailing list.</p>
  </body>
</html>
```

---

### Campaign Management

#### `POST /api/campaigns`
Create a new campaign (as draft)

**Request:**
```json
{
  "subject": "Weekly Newsletter",
  "body_text": "Hello {{name}},\n\nThis is our weekly update...",
  "body_html": "<html><body><p>Hello {{name}},</p><p>This is our weekly update...</p></body></html>"
}
```

**Response:**
```json
{
  "id": "660e8400-...",
  "subject": "Weekly Newsletter",
  "status": "draft",
  "total_count": 0,
  "created_at": "2025-11-29T11:00:00Z"
}
```

---

#### `GET /api/campaigns`
List all campaigns

**Response:**
```json
{
  "campaigns": [
    {
      "id": "660e8400-...",
      "subject": "Weekly Newsletter",
      "status": "sent",
      "total_count": 1500,
      "sent_count": 1498,
      "failed_count": 2,
      "created_at": "2025-11-29T11:00:00Z",
      "completed_at": "2025-11-29T11:30:00Z"
    }
  ]
}
```

---

#### `GET /api/campaigns/:id`
Get campaign details including stats

**Response:**
```json
{
  "id": "660e8400-...",
  "subject": "Weekly Newsletter",
  "body_text": "Hello {{name}}...",
  "body_html": "<html>...",
  "status": "sent",
  "total_count": 1500,
  "sent_count": 1498,
  "failed_count": 2,
  "created_at": "2025-11-29T11:00:00Z",
  "started_at": "2025-11-29T11:10:00Z",
  "completed_at": "2025-11-29T11:30:00Z"
}
```

---

#### `POST /api/campaigns/:id/send`
Send campaign to all verified subscribers

**Response:**
```json
{
  "id": "660e8400-...",
  "status": "sending",
  "total_count": 1500,
  "message": "Campaign sending started"
}
```

---

#### `DELETE /api/campaigns/:id`
Delete campaign (only if draft or completed)

**Response:**
```json
{
  "message": "Campaign deleted"
}
```

---

### Health & Config

#### `GET /health`
Health check endpoint

**Response:**
```json
{
  "status": "healthy",
  "database": "ok",
  "smtp": "ok"
}
```

---

## Configuration

### config.yaml

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
  rate_limit: 10              # Emails per second
  max_retries: 3
  retry_delay: "5s"
  batch_size: 100             # Process N subscribers at a time

templates:
  verification_subject: "Please verify your email"
  verification_body: |
    Hi {{name}},

    Please click the link below to verify your subscription:
    {{verify_url}}

    If you didn't sign up, please ignore this email.

logging:
  level: "info"               # debug, info, warn, error
  format: "json"              # json or text
```

### Environment Variables (Override config.yaml)

```bash
TINYLIST_SERVER_PORT=8080
TINYLIST_DATABASE_PATH=/data/tinylist.db
TINYLIST_SMTP_HOST=smtp.gmail.com
TINYLIST_SMTP_USERNAME=user@gmail.com
TINYLIST_SMTP_PASSWORD=secret
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1, Days 1-2)
**Goal:** Basic project setup and database layer

**Tasks:**
1. Initialize Go module and project structure
2. Create SQLite schema and migration system
3. Database connection and query functions
4. Configuration loading (YAML + env vars)
5. Basic logging setup

**Deliverables:**
- Project skeleton
- Database schema in `schema.sql`
- Config loading works
- Can insert/query subscribers from DB

---

### Phase 2: Core API (Week 1, Days 3-4)
**Goal:** REST API for subscriber management

**Tasks:**
1. Set up Chi router
2. Implement subscriber endpoints:
   - POST /api/subscribers
   - GET /api/subscribers
   - GET /api/subscribers/:id
   - DELETE /api/subscribers/:id
3. Input validation
4. Error handling middleware
5. JSON response helpers

**Deliverables:**
- Working subscriber CRUD API
- Can test with curl/Postman

---

### Phase 3: Email Verification (Week 1, Day 5)
**Goal:** Double opt-in flow

**Tasks:**
1. SMTP client wrapper (gomail)
2. Token generation for verification
3. Verification email template
4. Public verification endpoint (GET /verify/:token)
5. Update subscriber status on verification

**Deliverables:**
- New subscribers receive verification email
- Clicking link verifies subscription

---

### Phase 4: Campaign System (Week 2, Days 1-2)
**Goal:** Campaign creation and sending

**Tasks:**
1. Campaign CRUD endpoints
2. Campaign model and database queries
3. Template variable replacement ({{name}}, {{email}})
4. Campaign sending worker with rate limiting
5. Send status tracking (sent_count, failed_count)

**Deliverables:**
- Can create campaigns via API
- Can send campaigns to verified subscribers
- Rate limiting works

---

### Phase 5: Unsubscribe & Polish (Week 2, Day 3)
**Goal:** Complete core features

**Tasks:**
1. Unsubscribe token generation
2. Public unsubscribe endpoint (GET /unsubscribe/:token)
3. Add unsubscribe link to campaign emails
4. Health check endpoint
5. Graceful shutdown

**Deliverables:**
- Users can unsubscribe via link
- All core features complete

---

### Phase 6: Production Ready (Week 2, Days 4-5)
**Goal:** Deployment and documentation

**Tasks:**
1. Dockerfile (multi-stage build)
2. Kubernetes manifests (deployment, service, configmap)
3. Comprehensive README.md
4. API documentation (OpenAPI/Swagger optional)
5. Example config files
6. GitHub Actions CI (optional)

**Deliverables:**
- Docker image builds
- Can deploy to Kubernetes
- Documentation complete
- Ready to open-source

---

## Resource Requirements

### Memory Usage Estimate
```
Go Binary:                   ~15-25 MB
SQLite (10k subscribers):    ~5-10 MB
HTTP Server:                 ~5-10 MB
Campaign Worker:             ~10-20 MB
Buffers/Overhead:            ~10-15 MB

TOTAL:                       ~45-80 MB ✅ Fits in 100MB easily!
```

### Kubernetes Deployment
```yaml
resources:
  requests:
    memory: 64Mi
    cpu: 50m
  limits:
    memory: 128Mi
    cpu: 200m
```

---

## Example Usage Flow

### 1. User Subscribes
```bash
curl -X POST http://localhost:8080/api/subscribers \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "name": "John Doe"
  }'
```

**Response:** 201 Created, verification email sent

---

### 2. User Clicks Verification Link
```
User receives email with link:
https://newsletter.example.com/verify/abc123def456...

Clicks link → Status changes to "verified"
```

---

### 3. Admin Creates Campaign
```bash
curl -X POST http://localhost:8080/api/campaigns \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Welcome!",
    "body_text": "Hi {{name}},\n\nWelcome to our newsletter!"
  }'
```

**Response:** Campaign created with ID

---

### 4. Admin Sends Campaign
```bash
curl -X POST http://localhost:8080/api/campaigns/660e8400.../send
```

**Response:** Campaign starts sending to all verified subscribers

---

### 5. User Unsubscribes
```
User clicks unsubscribe link in email:
https://newsletter.example.com/unsubscribe/xyz789...

Status changes to "unsubscribed"
```

---

## Security Considerations

### v1.0 Security
1. **No Authentication** - Single tenant, trust network security
2. **Rate Limiting** - Prevent abuse on public endpoints
3. **SQL Injection** - Use parameterized queries (Go handles this)
4. **Token Security** - Cryptographically random tokens (crypto/rand)
5. **Input Validation** - Email format, length limits

### Future (v1.1+)
- API key authentication
- HTTPS enforcement
- CORS configuration
- Request signing
- Webhook signatures

---

## Testing Strategy

### Unit Tests
- Database queries
- Email template rendering
- Token generation
- Input validation

### Integration Tests
- Full API endpoint tests
- SMTP mock testing
- Campaign sending flow

### Manual Testing
- Real SMTP sending
- Kubernetes deployment
- Load testing (1000+ subscribers)

---

## Deployment

### Docker Build
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o tinylist ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/tinylist .
COPY schema.sql .
EXPOSE 8080
CMD ["./tinylist"]
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tinylist
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tinylist
  template:
    metadata:
      labels:
        app: tinylist
    spec:
      containers:
      - name: tinylist
        image: tinylist:latest
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: 64Mi
            cpu: 50m
          limits:
            memory: 128Mi
            cpu: 200m
        env:
        - name: TINYLIST_DATABASE_PATH
          value: /data/tinylist.db
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: tinylist-pvc
```

---

## Future Enhancements (Post v1.0)

### v1.1 - Basic UI
- Simple web interface for management
- Campaign editor
- Subscriber list view

### v1.2 - Multiple Lists
- Support multiple mailing lists
- Subscribers can join multiple lists
- Send to specific lists

### v1.3 - Analytics
- Open tracking (simple pixel)
- Click tracking
- Campaign statistics

### v1.4 - Templates
- Reusable email templates
- Template variables
- HTML editor

---

## Success Metrics

### Technical Goals
- ✅ Runs in <128MB RAM
- ✅ Handles 10k+ subscribers
- ✅ Sends 100+ emails/minute
- ✅ Single binary deployment
- ✅ Zero dependencies (except SQLite file)

### Functional Goals
- ✅ Double opt-in works reliably
- ✅ Unsubscribe works 100% of time
- ✅ No lost emails
- ✅ Clear error messages

---

## Open Source Preparation

### License
**Recommendation:** MIT or Apache 2.0 (permissive)

### Repository Setup
- Clear README with quick start
- CONTRIBUTING.md guidelines
- CODE_OF_CONDUCT.md
- Issue templates
- PR templates
- CI/CD with GitHub Actions

### Documentation
- API documentation
- Deployment guides
- Configuration reference
- Troubleshooting guide

---

## Timeline Summary

| Phase | Duration | Key Deliverable |
|-------|----------|-----------------|
| Phase 1 | 2 days | Database layer complete |
| Phase 2 | 2 days | Subscriber API working |
| Phase 3 | 1 day | Email verification working |
| Phase 4 | 2 days | Campaign sending working |
| Phase 5 | 1 day | Unsubscribe working |
| Phase 6 | 2 days | Production ready |

**Total: 10 days (2 weeks at steady pace)**

---

## Comparison with Listmonk Migration

| Aspect | Listmonk Migration | TinyList |
|--------|-------------------|----------|
| **Time** | 6-10 weeks | 2 weeks |
| **Complexity** | Very High | Low |
| **Memory** | 300-600 MB | 64-128 MB |
| **Features** | All listmonk features | Core features only |
| **Risk** | High (breaking changes) | Low (greenfield) |
| **Maintenance** | High (keep parity) | Low (simple codebase) |
| **Learning Curve** | Steep | Gentle |

---

## Conclusion

Building TinyList from scratch is:
- **Faster** than migrating listmonk (2 weeks vs 6-10 weeks)
- **Lighter** in resources (64-128MB vs 300-600MB)
- **Simpler** to maintain (500 LOC vs 50k LOC)
- **Customizable** to your exact needs
- **Production-ready** with proper testing
- **Open-sourceable** as a minimal alternative to listmonk

This approach gives you exactly what you need without the complexity you don't need.

---

**Ready to build this?** We can start with Phase 1 and have a working prototype in a few days.
