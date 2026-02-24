FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies including gcc for CGO
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build with CGO enabled (required for go-sqlite3)
RUN CGO_ENABLED=1 GOOS=linux go build -o vsentry .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/vsentry .
COPY config/ ./config/

ENV TZ=UTC

EXPOSE 8080

CMD ["./vsentry"]