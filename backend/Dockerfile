# Build stage - Build both VSentry and redAgent collectors
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Install build dependencies for cross-compilation
RUN apt-get update && apt-get install -y --no-install-recommends git gcc libc6-dev gcc-mingw-w64 && rm -rf /var/lib/apt/lists/*

# Copy go mod files for main app
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy VSentry source
COPY backend/ backend/

# Build main VSentry application
WORKDIR /app/backend
RUN CGO_ENABLED=1 GOOS=linux go build -o vsentry .

# Build redAgent collector for Windows
WORKDIR /app/backend/cmd/redAgent
RUN go mod download && \
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o redAgent.exe .

# Build redAgent collector for Linux
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o redAgent .

# Final stage - use debian slim for compatibility
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy main app and config
COPY --from=builder /app/vsentry .
COPY --from=builder /app/backend/config/ ./config/

# Copy redAgent collectors
RUN mkdir -p /app/collector
COPY --from=builder /app/backend/cmd/redAgent/redAgent.exe /app/collector/
COPY --from=builder /app/backend/cmd/redAgent/redAgent /app/collector/

# Copy redAgent source for reference
COPY --from=builder /app/backend/cmd/redAgent ./redAgent-source/

ENV TZ=UTC

EXPOSE 8088

CMD ["./vsentry"]