# Build stage - Build both VSentry and redAgent collectors
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files for main app
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy VSentry source
COPY backend/ backend/

# Build main VSentry application
RUN CGO_ENABLED=1 GOOS=linux go build -o vsentry ./backend

# Build redAgent collector for Windows
WORKDIR /app/backend/cmd/redAgent
RUN go mod download && \
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o redAgent.exe .

# Build redAgent collector for Linux
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o redAgent .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

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