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
8. ✅ Web UI (Preact-based admin interface)

### Explicitly Out of Scope (v1.0)
- ❌ Multiple lists
- ❌ Campaign templates
- ❌ Advanced analytics/tracking (basic stats only)
- ❌ Bounce handling
- ❌ Rich HTML emails (plain text + basic HTML only)
- ❌ User authentication (single tenant)
- ❌ Media uploads

---

## Technical Stack

### Backend (Go)
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

### Frontend (Preact)
**Framework:** Preact 10.19+ (React-compatible, ~3KB)
**Build Tool:** Vite 7.2.6 (fast dev server & builds)
**Styling:** Tailwind CSS
**Routing:** Preact Router (~2KB)
**HTTP Client:** Native fetch API

**Why Preact:**
- React-compatible API (easy for contributors)
- Tiny bundle size (~3KB vs React's 42KB)
- Large ecosystem & community support
- Production bundle <50KB gzipped
- Excellent performance

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
│   │   ├── private/            # Private API handlers (admin UI)
│   │   │   ├── subscribers.go  # Subscriber management
│   │   │   ├── campaigns.go    # Campaign management
│   │   │   ├── settings.go     # SMTP & settings
│   │   │   └── stats.go        # Statistics & analytics
│   │   └── public/             # Public API handlers (subscribers)
│   │       ├── verify.go       # Email verification
│   │       └── unsubscribe.go  # Unsubscribe handler
│   ├── mailer/
│   │   ├── mailer.go           # SMTP client wrapper
│   │   └── templates.go        # Email templates
│   └── config/
│       └── config.go           # Configuration loading
├── frontend/                    # Preact UI (separate container)
│   ├── src/
│   │   ├── main.jsx            # Entry point
│   │   ├── App.jsx             # Root component
│   │   ├── components/
│   │   │   ├── Dashboard.jsx   # Stats overview
│   │   │   ├── Subscribers.jsx # Subscriber management
│   │   │   ├── Campaigns.jsx   # Campaign list & creation
│   │   │   ├── CampaignStats.jsx  # Campaign analytics
│   │   │   └── Settings.jsx    # SMTP & app settings
│   │   ├── api.js              # API client (calls /api/private/*)
│   │   └── styles.css
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   ├── Dockerfile              # Frontend container
│   └── .gitignore
├── schema.sql                   # SQLite schema
├── config.example.yaml          # Example configuration
├── Dockerfile                   # Backend container
├── k8s/                        # Kubernetes manifests
│   ├── backend-deployment.yaml
│   ├── frontend-deployment.yaml
│   ├── backend-service.yaml
│   ├── frontend-service.yaml
│   └── ingress.yaml            # Route public/private traffic
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

## API Architecture

### Namespace Design

The API is organized into two distinct namespaces:

#### **Public API** (default routes)
- **Purpose:** Subscriber-facing endpoints (no authentication required)
- **Audience:** End users on your website/blog and email links
- **Endpoints:**
  - `POST /api/subscribe` - Subscribe to newsletter (from website form)
  - `GET /api/verify/:token` - Email verification
  - `GET /api/unsubscribe/:token` - Unsubscribe from list
- **Security:** Token-based authentication for verify/unsubscribe; rate limiting on subscribe
- **Rate Limiting:** Aggressive rate limiting to prevent abuse
- **Note:** Public routes don't need explicit `/public/` namespace - brevity matters

#### **Private API** (`/api/private/*` - explicit namespace)
- **Purpose:** Admin UI management endpoints
- **Audience:** Frontend application only
- **Endpoints:**
  - Subscriber CRUD (`/api/private/subscribers/*`)
  - Campaign management (`/api/private/campaigns/*`)
  - SMTP settings (`/api/private/settings/*`)
  - Statistics & analytics (`/api/private/stats/*`)
- **Security:** Network-level isolation (K8s internal traffic only in v1.0)
- **Future:** API key authentication in v1.1+
- **Note:** Private routes use explicit `/private/` namespace for security clarity

### Traffic Routing

```
                          ┌─────────────┐
                          │   Ingress   │
                          └──────┬──────┘
                                 │
                ┌────────────────┼────────────────┐
                │                │                │
      /api/subscribe          /api/private/*      /admin/*
      /api/verify/:token         │                │
      /api/unsubscribe/:token    │                │
                │                │                │
        ┌───────▼────────┐  ┌───▼────────┐  ┌───▼────────┐
        │  Go Backend    │  │ Go Backend │  │  Frontend  │
        │  (Public API)  │  │(Private API)│  │  (Preact)  │
        └────────────────┘  └─────┬──────┘  └─────┬──────┘
                                  │                │
                                  └────────────────┘
                                   API calls to /api/private/*
```

---

## API Endpoints

### Private API (Admin UI)

#### Subscriber Management

##### `POST /api/private/subscribers`
Create a new subscriber (admin-initiated, for manual additions via UI)

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

##### `GET /api/private/subscribers`
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

##### `GET /api/private/subscribers/:id`
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

##### `DELETE /api/private/subscribers/:id`
Delete subscriber permanently

**Response:**
```json
{
  "message": "Subscriber deleted"
}
```

---

### Public API (Subscriber-Facing)

#### `POST /api/subscribe`
Public subscription endpoint (for website/blog forms)

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
  "message": "Verification email sent. Please check your inbox.",
  "email": "user@example.com"
}
```

**Notes:**
- Sends verification email with token
- Rate limited to prevent abuse (e.g., 5 requests per IP per hour)
- Primary method for users to subscribe themselves
- Used in website forms, embedded widgets, etc.

**Example HTML Form:**
```html
<form id="subscribe-form">
  <input type="email" name="email" placeholder="Enter your email" required>
  <input type="text" name="name" placeholder="Your name (optional)">
  <button type="submit">Subscribe</button>
</form>

<script>
document.getElementById('subscribe-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const formData = new FormData(e.target);

  const response = await fetch('/api/subscribe', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email: formData.get('email'),
      name: formData.get('name') || ''
    })
  });

  const result = await response.json();
  alert(result.message); // "Verification email sent. Please check your inbox."
});
</script>
```

---

#### `GET /api/verify/:token`
Verify email subscription (token-based authentication)

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

#### `GET /api/unsubscribe/:token`
Unsubscribe from list (token-based authentication)

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

#### Campaign Management

##### `POST /api/private/campaigns`
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

##### `GET /api/private/campaigns`
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

##### `GET /api/private/campaigns/:id`
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

##### `POST /api/private/campaigns/:id/send`
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

##### `DELETE /api/private/campaigns/:id`
Delete campaign (only if draft or completed)

**Response:**
```json
{
  "message": "Campaign deleted"
}
```

---

#### Settings Management

##### `GET /api/private/settings`
Get all settings (SMTP config, etc.)

**Response:**
```json
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "user@gmail.com",
    "from_email": "newsletter@example.com",
    "from_name": "Newsletter",
    "tls": true
  },
  "sending": {
    "rate_limit": 10,
    "max_retries": 3,
    "batch_size": 100
  }
}
```

---

##### `PUT /api/private/settings/smtp`
Update SMTP configuration

**Request:**
```json
{
  "host": "smtp.gmail.com",
  "port": 587,
  "username": "user@gmail.com",
  "password": "app-password",
  "from_email": "newsletter@example.com",
  "from_name": "Newsletter",
  "tls": true
}
```

**Response:**
```json
{
  "message": "SMTP settings updated successfully"
}
```

---

#### Statistics & Analytics

##### `GET /api/private/stats/overview`
Get dashboard overview statistics

**Response:**
```json
{
  "subscribers": {
    "total": 1500,
    "verified": 1450,
    "pending": 30,
    "unsubscribed": 20
  },
  "campaigns": {
    "total": 25,
    "sent": 20,
    "draft": 3,
    "sending": 2
  },
  "recent_activity": {
    "last_campaign_sent": "2025-11-29T11:30:00Z",
    "emails_sent_today": 1498,
    "new_subscribers_today": 15
  }
}
```

---

##### `GET /api/private/stats/campaigns`
Get campaign statistics with time-series data

**Query Params:**
- `period` - Time period (7d, 30d, 90d, all)

**Response:**
```json
{
  "campaigns": [
    {
      "id": "660e8400-...",
      "subject": "Weekly Newsletter",
      "sent_count": 1498,
      "failed_count": 2,
      "sent_at": "2025-11-29T11:30:00Z"
    }
  ],
  "timeline": [
    {
      "date": "2025-11-29",
      "emails_sent": 1498,
      "campaigns": 1
    }
  ]
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

## Frontend Architecture

### Technology Stack
- **Preact 10.19+** - React-compatible, 3KB core
- **Vite 7.2.6** - Lightning-fast dev server & build tool
- **Tailwind CSS** - Utility-first CSS framework
- **Preact Router** - Client-side routing (~2KB)
- **Native Fetch API** - HTTP requests (no axios bloat)

### Component Structure

```
src/
├── main.jsx              # App initialization
├── App.jsx               # Root component with routing
├── components/
│   ├── Dashboard.jsx     # Overview stats & charts
│   ├── Subscribers.jsx   # Subscriber table with filters
│   ├── Campaigns.jsx     # Campaign list & create form
│   ├── CampaignStats.jsx # Campaign analytics & graphs
│   └── Settings.jsx      # SMTP configuration form
├── api.js                # API client wrapper
└── styles.css            # Global styles + Tailwind
```

### Routing

```javascript
// Routes in App.jsx
<Router>
  <Dashboard path="/admin/" />
  <Subscribers path="/admin/subscribers" />
  <Campaigns path="/admin/campaigns" />
  <CampaignStats path="/admin/campaigns/:id" />
  <Settings path="/admin/settings" />
</Router>
```

### API Client

```javascript
// api.js - Centralized API calls
const API_BASE = '/api/private';

export const api = {
  // Subscribers
  getSubscribers: (params) =>
    fetch(`${API_BASE}/subscribers?${new URLSearchParams(params)}`).then(r => r.json()),

  // Campaigns
  getCampaigns: () =>
    fetch(`${API_BASE}/campaigns`).then(r => r.json()),

  createCampaign: (data) =>
    fetch(`${API_BASE}/campaigns`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    }).then(r => r.json()),

  // Settings
  getSettings: () =>
    fetch(`${API_BASE}/settings`).then(r => r.json()),

  updateSMTP: (data) =>
    fetch(`${API_BASE}/settings/smtp`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    }).then(r => r.json()),

  // Stats
  getOverviewStats: () =>
    fetch(`${API_BASE}/stats/overview`).then(r => r.json())
};
```

### Build Output
- **Development:** `npm run dev` (Vite dev server on port 5173)
- **Production:** `npm run build` → `dist/` folder
  - Minified bundle: ~45-50KB gzipped (HTML + JS + CSS)
  - Served via nginx in Docker container

### Container Deployment
- **Image:** nginx:alpine (~5MB base)
- **Build:** Multi-stage (Node for build, nginx for serve)
- **Size:** ~15-20MB total image size
- **Memory:** ~10-15MB runtime

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

### Phase 6: Frontend Development (Week 2, Days 4-6)
**Goal:** Build Preact admin interface

**Tasks:**
1. Initialize Vite + Preact project in `frontend/`
2. Set up Tailwind CSS + PostCSS
3. Create Dashboard component with stats overview
4. Build Subscribers component (list, pagination, filters)
5. Build Campaigns component (list, create form, send)
6. Build CampaignStats component (charts & analytics)
7. Build Settings component (SMTP configuration form)
8. Implement API client (`api.js`)
9. Add Preact Router for navigation
10. Frontend Dockerfile (multi-stage: Node build → nginx serve)

**Deliverables:**
- Fully functional admin UI
- Production-optimized build (<50KB gzipped)
- Frontend Docker image

---

### Phase 7: Production Ready (Week 3, Days 1-2)
**Goal:** Deployment and documentation

**Tasks:**
1. Backend Dockerfile (multi-stage build)
2. Frontend Dockerfile (Node build → nginx)
3. Kubernetes manifests:
   - Backend deployment & service
   - Frontend deployment & service
   - Ingress (route /api/* to backend, /admin/* to frontend)
   - ConfigMaps & Secrets
   - PVC for SQLite database
4. Comprehensive README.md
5. API documentation
6. Example config files
7. GitHub Actions CI (optional)

**Deliverables:**
- Both Docker images build successfully
- Can deploy full stack to Kubernetes
- Documentation complete
- Ready to open-source

---

## Resource Requirements

### Memory Usage Estimate

#### Backend (Go)
```
Go Binary:                   ~15-25 MB
SQLite (10k subscribers):    ~5-10 MB
HTTP Server:                 ~5-10 MB
Campaign Worker:             ~10-20 MB
Buffers/Overhead:            ~10-15 MB

TOTAL BACKEND:               ~45-80 MB
```

#### Frontend (Nginx)
```
Nginx:                       ~5-8 MB
Static files (in memory):    ~5-10 MB

TOTAL FRONTEND:              ~10-15 MB
```

#### Combined System
```
Backend:                     ~45-80 MB
Frontend:                    ~10-15 MB

TOTAL SYSTEM:                ~55-95 MB ✅ Well under 100MB!
```

### Kubernetes Deployment

#### Backend Resources
```yaml
resources:
  requests:
    memory: 64Mi
    cpu: 50m
  limits:
    memory: 128Mi
    cpu: 200m
```

#### Frontend Resources
```yaml
resources:
  requests:
    memory: 16Mi
    cpu: 10m
  limits:
    memory: 32Mi
    cpu: 50m
```

---

## Example Usage Flow

### 1. User Subscribes from Website
```bash
# User fills out subscription form on your blog
curl -X POST http://localhost:8080/api/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "name": "John Doe"
  }'
```

**Response:** Verification email sent to user

---

### 2. User Clicks Verification Link
```
User receives email with link:
https://newsletter.example.com/api/verify/abc123def456...

Clicks link → Status changes to "verified"
```

---

### 3. Admin Creates Campaign (via UI or API)
```bash
curl -X POST http://localhost:8080/api/private/campaigns \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Welcome!",
    "body_text": "Hi {{name}},\n\nWelcome to our newsletter!"
  }'
```

**Response:** Campaign created with ID

---

### 4. Admin Sends Campaign (via UI or API)
```bash
curl -X POST http://localhost:8080/api/private/campaigns/660e8400.../send
```

**Response:** Campaign starts sending to all verified subscribers

---

### 5. User Unsubscribes
```
User clicks unsubscribe link in email:
https://newsletter.example.com/api/unsubscribe/xyz789...

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

### Backend Dockerfile
```dockerfile
# Backend Dockerfile (root directory)
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

### Frontend Dockerfile
```dockerfile
# Frontend Dockerfile (frontend/Dockerfile)
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Kubernetes Deployment

#### Backend Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tinylist-backend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tinylist-backend
  template:
    metadata:
      labels:
        app: tinylist-backend
    spec:
      containers:
      - name: backend
        image: tinylist-backend:latest
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
---
apiVersion: v1
kind: Service
metadata:
  name: tinylist-backend
spec:
  selector:
    app: tinylist-backend
  ports:
  - port: 8080
    targetPort: 8080
```

#### Frontend Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tinylist-frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tinylist-frontend
  template:
    metadata:
      labels:
        app: tinylist-frontend
    spec:
      containers:
      - name: frontend
        image: tinylist-frontend:latest
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: 16Mi
            cpu: 10m
          limits:
            memory: 32Mi
            cpu: 50m
---
apiVersion: v1
kind: Service
metadata:
  name: tinylist-frontend
spec:
  selector:
    app: tinylist-frontend
  ports:
  - port: 80
    targetPort: 80
```

#### Ingress Configuration
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: tinylist-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: newsletter.example.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: tinylist-backend
            port:
              number: 8080
      - path: /admin
        pathType: Prefix
        backend:
          service:
            name: tinylist-frontend
            port:
              number: 80
      - path: /
        pathType: Prefix
        backend:
          service:
            name: tinylist-frontend
            port:
              number: 80
```

---

## Future Enhancements (Post v1.0)

### v1.1 - Authentication & Security
- API key authentication for private endpoints
- Session-based login for frontend
- HTTPS enforcement
- CORS configuration

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
| Phase 6 | 3 days | Frontend UI complete |
| Phase 7 | 2 days | Production deployment ready |

**Total: 13 days (~2.5 weeks at steady pace)**

---

## Comparison with Listmonk Migration

| Aspect | Listmonk Migration | TinyList |
|--------|-------------------|----------|
| **Time** | 6-10 weeks | 2.5 weeks |
| **Complexity** | Very High | Low-Medium |
| **Memory** | 300-600 MB | 55-95 MB (backend + UI) |
| **Features** | All listmonk features | Core features + Admin UI |
| **UI** | Complex React dashboard | Lightweight Preact UI |
| **Risk** | High (breaking changes) | Low (greenfield) |
| **Maintenance** | High (keep parity) | Low (simple codebase) |
| **Learning Curve** | Steep | Gentle (React-compatible) |
| **Contributors** | Need to learn listmonk | Easy (standard React/Preact) |

---

## Conclusion

Building TinyList from scratch is:
- **Faster** than migrating listmonk (2.5 weeks vs 6-10 weeks)
- **Lighter** in resources (55-95MB total vs 300-600MB)
- **Simpler** to maintain (~1000 LOC vs 50k LOC)
- **Modern UI** with Preact (React-compatible, contributor-friendly)
- **Fully featured** for v1.0 (API + Admin UI + Statistics)
- **Production-ready** with proper testing & K8s deployment
- **Open-sourceable** as a minimal alternative to listmonk

### What You Get in v1.0:
✅ Go backend with REST API (clean public routes + explicit `/private/*` namespace)
✅ **Public subscription endpoint** for website/blog forms (`/api/subscribe`)
✅ SQLite database (lightweight, zero dependencies)
✅ Double opt-in email verification
✅ Campaign management & sending
✅ **Preact admin UI** with:
   - Dashboard with stats & charts
   - Subscriber management
   - Campaign creation & analytics
   - SMTP settings configuration
✅ Separate Docker containers (backend + frontend)
✅ Full Kubernetes deployment manifests
✅ Production-optimized builds (<50KB frontend bundle)

This approach gives you exactly what you need without the complexity you don't need, and it's contributor-friendly with React-compatible Preact.

---

**Ready to build this?** We can start with Phase 1 and have a working full-stack application in ~2.5 weeks.
