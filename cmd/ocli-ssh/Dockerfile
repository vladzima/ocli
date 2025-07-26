# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod and source files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source files
COPY *.go ./

# Build the SSH server
RUN go build -o ocli-ssh .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates openssh-keygen

# Create non-root user
RUN adduser -D -s /bin/sh ocli

# Create directories
RUN mkdir -p /data /ssh-keys && \
    chown -R ocli:ocli /data /ssh-keys

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/ocli-ssh .
RUN chmod +x ocli-ssh

# Switch to non-root user
USER ocli

# Expose SSH port (Railway will override with PORT env var)
EXPOSE 2222

# Volume for persistent data
VOLUME ["/data"]

# Set default environment variables
ENV OCLI_SSH_DATA_DIR=/data
ENV OCLI_SSH_AUTO_REGISTER=true

# Generate host key on startup if not exists and start server
CMD if [ ! -f /ssh-keys/ssh_host_ed25519_key ]; then \
        ssh-keygen -t ed25519 -f /ssh-keys/ssh_host_ed25519_key -N ''; \
    fi && \
    ./ocli-ssh --key /ssh-keys/ssh_host_ed25519_key