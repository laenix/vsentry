# Build stage - Build VSentry and collectors
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Copy go mod files for main app
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy VSentry source
COPY backend/ backend/

# Build main VSentry application
WORKDIR /app/backend
RUN go mod tidy && CGO_ENABLED=0 go build -o vsentry .

# Build collectors for Linux (no CGO needed)
WORKDIR /app/backend/cmd/collectors
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o vsentry-agent .

# Build collectors for Windows (uses syscall, no CGO)
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o vsentry-agent.exe .

# Final stage - use distroless for minimal image
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy main app
COPY --from=builder /app/backend/vsentry .

# Copy collectors
COPY --from=builder /app/backend/cmd/collectors/vsentry-agent /app/collector/vsentry-agent
COPY --from=builder /app/backend/cmd/collectors/vsentry-agent.exe /app/collector/vsentry-agent.exe

# Copy config
COPY --from=builder /app/backend/config/ ./config/

EXPOSE 8088

USER root
CMD ["./vsentry"]