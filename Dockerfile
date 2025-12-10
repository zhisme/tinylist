# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies for pure Go SQLite
RUN apk add --no-cache ca-certificates

# Copy go module files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO disabled (pure Go SQLite)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o tinylist cmd/server/main.go

# Runtime stage
FROM alpine:3.21

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Create data directory
RUN mkdir -p /data && chown appuser:appuser /data

# Copy binary from builder
COPY --from=builder /app/tinylist .

# Copy config if exists (optional, can be overridden via env vars)
COPY --from=builder /app/config.yaml* ./

# Ensure appuser can read all files
RUN chown -R appuser:appuser /app

USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
CMD ["./tinylist"]
