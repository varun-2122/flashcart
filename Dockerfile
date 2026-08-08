# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build tools
RUN apk add --no-cache git ca-certificates tzdata

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically compiled binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/server ./cmd/server

# Production Stage
FROM alpine:3.19 AS runner

WORKDIR /app

# Security: Install CA certificates and create non-root user
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S flashcart && adduser -S flashcart -G flashcart

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/.env /app/.env

USER flashcart

EXPOSE 8080

ENTRYPOINT ["/app/server"]
