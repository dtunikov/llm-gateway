# syntax=docker/dockerfile:1

# Stage 1: Build the Go application
FROM --platform=$BUILDPLATFORM golang:1.24.3-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./ 
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /llm-gateway ./cmd/llm-gateway

# Stage 2: Create the final lightweight image
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Copy the compiled binary from the builder stage
COPY --from=builder /llm-gateway /usr/local/bin/llm-gateway

# Copy the configuration file and OpenAPI spec
COPY config.yml /app/config.yml
COPY api /app/api

# Expose the port the application listens on
EXPOSE 8080

# Set the entrypoint to run the application
ENTRYPOINT ["/usr/local/bin/llm-gateway"]
