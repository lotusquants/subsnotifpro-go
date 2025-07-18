# Multi-stage Dockerfile for SubsNotifPro Go Application

# Stage 1: Build stage
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the applications
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/worker ./cmd/worker

# Stage 2: Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates postgresql-client curl

# Create non-root user
RUN addgroup -g 1001 appuser && adduser -u 1001 -G appuser -s /bin/sh -D appuser

# Set working directory
WORKDIR /app

# Copy certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binaries
COPY --from=builder /app/bin/api /app/bin/worker /app/

# Copy configuration files
COPY --from=builder /app/config ./config/
COPY --from=builder /app/database ./database/
COPY --from=builder /app/scripts ./scripts/

# Copy any migration files if they exist
RUN mkdir -p /app/migrations

# Create necessary directories
RUN mkdir -p /app/logs /app/reports /app/backups

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Default command (can be overridden)
CMD ["./api"]
