# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /workspace

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the registry API binary
# CGO_ENABLED=0 for static binary
# GOOS=linux for linux containers
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o thv-registry-api ./cmd/thv-registry-api

# Final stage - minimal runtime image
FROM gcr.io/distroless/static:nonroot

# Copy the binary from builder stage
COPY --from=builder /workspace/thv-registry-api /usr/local/bin/thv-registry-api

# Copy ca-certificates for TLS verification
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Use non-root user for security
USER 65532:65532

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/thv-registry-api"]

# Default command-line arguments
CMD ["--port=8080"]

# Health check - Note: distroless doesn't include curl/wget, so we rely on Kubernetes probes

# Labels for better container management
LABEL org.opencontainers.image.title="ToolHive Registry API" \
      org.opencontainers.image.description="REST API service for ToolHive MCPRegistry resources" \
      org.opencontainers.image.vendor="Stacklok" \
      org.opencontainers.image.source="https://github.com/stacklok/toolhive" \
      org.opencontainers.image.documentation="https://github.com/stacklok/toolhive/tree/main/cmd/thv-registry-api" \
      org.opencontainers.image.licenses="Apache-2.0"

# Expose the default port
EXPOSE 8080