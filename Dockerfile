# syntax=docker/dockerfile:1.7-labs

# --- Builder stage ---
FROM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache ca-certificates git

WORKDIR /build

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /build/server ./cmd/server

# --- Runtime stage ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /build/server /app/server

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/server"]

LABEL org.opencontainers.image.source="https://github.com/Cerebrovinny/fizz-buzz-rest"
LABEL org.opencontainers.image.description="FizzBuzz REST API with statistics"
LABEL org.opencontainers.image.licenses="MIT"
