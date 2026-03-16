# syntax=docker/dockerfile:1
# Multi-stage build for Pebblr CRM

# ── Stage 1: Go builder ──────────────────────────────────────────────────────
FROM golang:1.26-alpine AS go-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api

# ── Stage 2: Frontend builder ─────────────────────────────────────────────────
FROM node:20-alpine AS web-builder

WORKDIR /app/web
COPY web/package.json web/bun.lockb* ./
RUN npm install -g bun && bun install --frozen-lockfile

COPY web/ ./
RUN bun run build

# ── Stage 3: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19

RUN addgroup -S pebblr && adduser -S pebblr -G pebblr

WORKDIR /app

COPY --from=go-builder /bin/api .
COPY --from=web-builder /app/web/dist ./web/dist

# Secrets are read from file mounts — never from environment variables
# Mount secrets at /run/secrets/ (e.g., /run/secrets/db-password)

USER pebblr

EXPOSE 8080

ENTRYPOINT ["/app/api"]
