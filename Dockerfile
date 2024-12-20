# Build stage
FROM golang:1.21-alpine AS builder

# Build arguments for version information
ARG VERSION=0.1.0
ARG GIT_COMMIT
ARG BUILD_TIME

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with version information
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s \
    -X github.com/news-reader/internal/version.Version=${VERSION} \
    -X github.com/news-reader/internal/version.BuildTime=${BUILD_TIME} \
    -X github.com/news-reader/internal/version.GitCommit=${GIT_COMMIT}" \
    -o /news-reader ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /news-reader .
COPY web/templates/ web/templates/
COPY web/static/ web/static/

# Create directory for preferences file
RUN mkdir -p /data

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8082

# Expose port
EXPOSE 8082

# Run the binary
CMD ["./news-reader", "-port", "8082", "-prefs", "/data/preferences.json", "-debug=false"]
