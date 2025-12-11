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

# Runtime stage
FROM nginx:alpine

# Copy our nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Remove templates directory (prevents envsubst script from failing)
# and ensure conf.d is writable by nginx user for entrypoint scripts
RUN rm -rf /etc/nginx/templates && \
    chown -R nginx:nginx /etc/nginx/conf.d

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Set ownership for nginx directories (running as non-root)
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/run/nginx.pid

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
