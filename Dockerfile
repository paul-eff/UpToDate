# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download
# Copy source code
COPY . .
# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o uptodate .

# Final stage
FROM alpine:latest
# Install chromium for rod browser functionality
RUN apk add --no-cache ca-certificates chromium
WORKDIR /app
# Copy binary from builder stage
COPY --from=builder /app/uptodate .
# Default command
CMD ["./uptodate", "-config", "config.json"]