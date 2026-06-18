# syntax=docker/dockerfile:1

# ---- Build stage ------------------------------------------------------------
FROM golang:1.21-alpine AS builder

# Build dependencies (git for VCS info, ca-certs/tzdata copied into final image)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Download modules first for better layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a fully static binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -trimpath \
    -o /poker-server ./cmd/server

# ---- Final stage ------------------------------------------------------------
# scratch keeps the image minimal and reduces attack surface.
FROM scratch

# CA certificates (for outbound TLS) and timezone data.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Create an unprivileged user by copying a minimal passwd entry. scratch has no
# adduser, so we provide the entry from the builder's /etc/passwd via a static
# definition. Running as non-root (uid 65532) limits blast radius.
COPY --from=builder /poker-server /poker-server

# Static admin panel assets (served at /admin/). The server resolves these
# relative to its working directory, so keep them at /web/admin.
COPY --from=builder /app/web /web

USER 65532:65532

EXPOSE 8080

# The orchestrator should hit /healthz (liveness) and /readyz (readiness).
ENTRYPOINT ["/poker-server"]
