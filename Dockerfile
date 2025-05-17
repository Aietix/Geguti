# Stage 1: Build the Go application
FROM golang:1.20-buster AS builder

WORKDIR /app

# Initialize a Go module and download dependencies
COPY app/main.go .
RUN ls -la && \
    go mod init screenshot-service && \
    go get github.com/chromedp/chromedp && \
    go mod tidy

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o screenshot-service main.go

# Stage 2: Create a lightweight runtime image
FROM debian:bullseye-slim

# Install Chromium
RUN apt-get update && apt-get install -y chromium \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/screenshot-service .

# Expose port 8080 for the HTTP server
EXPOSE 8080

# Set the default environment variable for the output path
ENV OUTPUT_PATH=/mnt/storage/

# Create the output directory
RUN mkdir -p /mnt/storage/

# Command to run the application
CMD ["./screenshot-service"]
