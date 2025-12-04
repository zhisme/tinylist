# GitHub Copilot Instructions for TinyList

## Project Overview

**TinyList** is a lightweight, API-first email list manager with SQLite backend designed for resource-constrained environments. It runs in <100MB RAM and provides essential email list functionality with both a Go backend and a Preact-based admin frontend.

**Important**: This project uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking. Use `bd` commands instead of markdown TODOs. See AGENTS.md for workflow details.

## Tech Stack

- **Backend**: Go 1.21+ with pure Go SQLite (modernc.org/sqlite)
- **Frontend**: Preact 10.19+ with Vite and Tailwind CSS
- **Deployment**: Kubernetes with separate frontend/backend containers

## Coding Guidelines

### Go Backend
- Use pure Go SQLite (no CGO)
- Keep API responses simple (standard library JSON)
- Public routes: no namespace (e.g., `/api/subscribe`)
- Private routes: explicit `/api/private/*` prefix
- Follow the "tiny" philosophy - avoid unnecessary dependencies

### Frontend
- Use React patterns (Preact is React-compatible)
- Keep bundle size minimal (~45-50KB gzipped)
- Use native fetch API for HTTP requests
- Tailwind CSS for styling

### Testing
- Write tests for database queries and API endpoints
- Use table-driven tests in Go
- Mock SMTP for email testing
- Target: 1000+ subscribers, 100+ emails/minute

## Issue Tracking with bd

**CRITICAL**: This project uses **bd** for ALL task tracking. Do NOT create markdown TODO lists.

### Essential Commands

```bash
# Find work
bd ready --json                    # Unblocked issues
bd list --status open --json       # All open issues

# Create and manage
bd create "Title" -t bug|feature|task -p 0-4 --json
bd update <id> --status in_progress --json
bd close <id> --reason "Done" --json

# Sync (CRITICAL at end of session!)
bd sync  # Force immediate export/commit/push
```

### Workflow

1. **Check ready work**: `bd ready --json`
2. **Claim task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** `bd create "Found bug" -p 1 --deps discovered-from:<parent-id> --json`
5. **Complete**: `bd close <id> --reason "Done" --json`
6. **Sync**: `bd sync` (flushes changes to git immediately)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

## Project Structure

```
tinylist/
├── cmd/server/main.go           # Application entry point
├── internal/
│   ├── db/                      # Database layer
│   ├── models/                  # Data models
│   ├── handlers/
│   │   ├── private/             # Admin API handlers
│   │   └── public/              # Public API handlers
│   ├── mailer/                  # SMTP client
│   └── config/                  # Configuration
├── frontend/                    # Preact UI (separate container)
│   └── src/components/          # UI components
└── k8s/                        # Kubernetes manifests
```

## Key Implementation Details

- **Double opt-in flow**: subscribe → verify email → verified status
- **Rate limiting**: Aggressive on public endpoints (5 req/IP/hour)
- **Campaign sending**: Batch processing with configurable rate limits
- **Template variables**: Only `{{name}}` and `{{email}}` supported
- **Security**: Network-level isolation (v1.0), API keys planned (v1.1+)

## Resource Targets

- Backend: 45-80 MB (Go + SQLite + HTTP server)
- Frontend: 10-15 MB (nginx + static files)
- **Total: <100MB** (well under target)

## CLI Help

Run `bd <command> --help` to see all available flags for any command.
For example: `bd create --help` shows `--parent`, `--deps`, `--assignee`, etc.

## Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic bd commands
- ✅ Follow the "tiny" philosophy
- ✅ Run `bd sync` at end of sessions
- ✅ Run `bd <cmd> --help` to discover available flags
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT add unnecessary dependencies
- ❌ Do NOT use CGO for SQLite

---

**For detailed workflows and project architecture, see [CLAUDE.md](../CLAUDE.md) and [AGENTS.md](../AGENTS.md)**
