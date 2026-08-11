# Build Stage - Use latest official Go Alpine image to support Go 1.24+ / 1.25+
FROM golang:alpine AS builder

ENV GOTOOLCHAIN=auto

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

# Production Stage - Minimal, secure Alpine runtime
FROM alpine:3.19 AS runner

WORKDIR /app

# Security: Install CA certificates and create non-root user
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S flashcart && adduser -S flashcart -G flashcart

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/web /app/web

USER flashcart

EXPOSE 8080

ENTRYPOINT ["/app/server"]
