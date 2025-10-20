# Use a minimal base image
FROM golang:1.24-alpine AS builder

# Set build arguments
ARG VERSION=unknown
ARG TARGETOS
ARG TARGETARCH

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-X main.version=${VERSION} -s -w" \
    -o sandboxed .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S sandboxed && \
    adduser -u 1000 -S sandboxed -G sandboxed

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/sandboxed /usr/local/bin/sandboxed

# Change ownership
RUN chown sandboxed:sandboxed /usr/local/bin/sandboxed

# Switch to non-root user
USER sandboxed

# Expose default ports
EXPOSE 8080

# Set default command
ENTRYPOINT ["/usr/local/bin/sandboxed"]
CMD ["server"]