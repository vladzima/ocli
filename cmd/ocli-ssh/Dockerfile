# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod and source files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source files explicitly
COPY main.go server.go auth.go ssh_model.go model.go bullet.go persistence.go ./

# Build the SSH server with explicit output name and verify
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ocli-ssh-server . && \
    ls -la ocli-ssh-server && \
    file ocli-ssh-server

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
COPY --from=builder /app/ocli-ssh-server ./ocli-ssh-server

# Debug: verify binary
RUN ls -la ./ocli-ssh-server && \
    file ./ocli-ssh-server && \
    chmod +x ./ocli-ssh-server

# Set permissions for data directory
RUN chown -R ocli:ocli /data /app

# Expose SSH port (Railway will override with PORT env var)  
EXPOSE 2222

# Set default environment variables
ENV OCLI_SSH_DATA_DIR=/data
ENV OCLI_SSH_AUTO_REGISTER=true

# Add debug startup script
RUN echo '#!/bin/sh' > start.sh && \
    echo 'echo "=== Debug Info ==="' >> start.sh && \
    echo 'whoami' >> start.sh && \
    echo 'pwd' >> start.sh && \
    echo 'ls -la' >> start.sh && \
    echo 'echo "=== Environment ==="' >> start.sh && \
    echo 'env | grep OCLI' >> start.sh && \
    echo 'echo "=== Starting server ==="' >> start.sh && \
    echo './ocli-ssh-server' >> start.sh && \
    chmod +x start.sh

# Switch to non-root user
USER ocli

# Start with debug script
CMD ["./start.sh"]