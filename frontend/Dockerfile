# Build stage
FROM node:22-alpine AS builder

WORKDIR /app

# Copy package files for caching
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build production bundle
RUN npm run build

# Runtime stage - use specific version to avoid surprises
FROM nginx:1.27-alpine

# Replace the ENTIRE nginx.conf (not conf.d) to avoid entrypoint script issues
COPY nginx.conf /etc/nginx/nginx.conf

# Remove all entrypoint scripts that cause problems
# These try to modify configs at runtime and fail with read-only filesystems
RUN rm -rf /docker-entrypoint.d/*

# Copy built assets to /tinylist/ subdirectory (matches Vite base path)
COPY --from=builder /app/dist /usr/share/nginx/html/tinylist

# Create temp directories for non-root nginx operation
RUN mkdir -p /tmp/client_temp /tmp/proxy_temp /tmp/fastcgi_temp /tmp/uwsgi_temp /tmp/scgi_temp && \
    chown -R nginx:nginx /tmp && \
    chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx

# Run as non-root
USER nginx

# Non-privileged port
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
