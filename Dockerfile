# PancakesGit - Privacy-focused Git Service
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev \
    sqlite-dev \
    openssh-server \
    openssh-keygen

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o pancakesgit ./cmd/pancakesgit

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    openssh-server \
    openssh-keygen \
    sqlite \
    tzdata \
    curl \
    bash

# Create pancakesgit user
RUN addgroup -g 1000 pancakesgit && \
    adduser -D -u 1000 -G pancakesgit -s /bin/bash pancakesgit

# Create directories
RUN mkdir -p /data /config /app && \
    chown -R pancakesgit:pancakesgit /data /config /app

# Copy binary from builder
COPY --from=builder /app/pancakesgit /app/pancakesgit
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static

# Set up SSH
RUN mkdir -p /etc/ssh && \
    ssh-keygen -A

# Copy configuration files
COPY docker/app.ini /config/app.ini
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Set permissions
RUN chown -R pancakesgit:pancakesgit /app

# Expose ports
EXPOSE 3000 22

# Switch to pancakesgit user
USER pancakesgit

# Set working directory
WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/api/healthz || exit 1

# Entry point
ENTRYPOINT ["/entrypoint.sh"]
CMD ["./pancakesgit"]
