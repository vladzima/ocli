# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy SSH server files
COPY cmd/ocli-ssh/go.mod cmd/ocli-ssh/go.sum ./
RUN go mod download

# Copy all SSH server source files
COPY cmd/ocli-ssh/*.go ./

# Build the SSH server
RUN CGO_ENABLED=0 GOOS=linux go build -o ocli-ssh-server .

# Runtime stage  
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Create user and directories
RUN adduser -D -s /bin/sh ocli && \
    mkdir -p /data && \
    chown -R ocli:ocli /data

WORKDIR /app

# Copy binary
COPY --from=builder /app/ocli-ssh-server .
RUN chmod +x ocli-ssh-server

# Set environment variables
ENV OCLI_SSH_DATA_DIR=/data
ENV OCLI_SSH_AUTO_REGISTER=true

# Switch to non-root user  
USER ocli

# Expose port
EXPOSE 8080

# Start server
CMD ["./ocli-ssh-server"]