# ===== Build Stage =====
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install required system packages (git is needed for module fetching)
RUN apk add --no-cache git

# Copy go.mod and go.sum to leverage Docker caching for dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source code
COPY . .

# Build the application from the correct main package directory
RUN go build -o explore ./cmd/server

# ===== Final Stage =====
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/explore .

# Expose the gRPC port (default is 50051)
EXPOSE 50051

# Run the service
CMD ["./explore"]
