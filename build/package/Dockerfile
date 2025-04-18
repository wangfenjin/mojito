# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/mojito ./cmd/mojito

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/mojito /app/mojito

# Copy config files
COPY --from=builder /app/config /app/config

# Create a non-root user to run the application
RUN adduser -D -g '' mojito
USER mojito

# Expose the application port
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release

# Run the application
CMD ["/app/mojito"]