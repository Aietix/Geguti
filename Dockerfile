# Stage 1: Build the Go application
FROM golang:1.24-bookworm AS builder

WORKDIR /app

COPY app/go.mod app/go.sum .
RUN go mod download

COPY app/main.go .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o screenshot-service main.go

# Stage 2: Lightweight runtime
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y chromium \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/screenshot-service .

EXPOSE 8080

ENV OUTPUT_PATH=/mnt/storage/
RUN mkdir -p /mnt/storage/

CMD ["./screenshot-service"]
