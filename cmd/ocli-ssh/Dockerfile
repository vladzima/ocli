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

# Build the SSH server with explicit output name
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ocli-ssh-server .

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
RUN chmod +x ocli-ssh-server

# Switch to non-root user
USER ocli

# Expose SSH port (Railway will override with PORT env var)
EXPOSE 2222

# Set default environment variables
ENV OCLI_SSH_DATA_DIR=/data
ENV OCLI_SSH_AUTO_REGISTER=true

# Start server (will auto-generate host key if needed)
CMD ["./ocli-ssh-server"]