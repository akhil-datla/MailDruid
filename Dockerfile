# Frontend build stage
FROM node:22-alpine AS frontend

WORKDIR /app/web

COPY web/package.json web/package-lock.json ./
RUN npm ci --legacy-peer-deps

COPY web/ ./
RUN npm run build

# Go build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy built frontend into the embed directory
COPY --from=frontend /app/web/dist ./internal/server/frontend

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /maildruid ./cmd/maildruid

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 1000 maildruid
USER maildruid

WORKDIR /app

COPY --from=builder /maildruid /app/maildruid
COPY --from=builder /app/fonts /app/fonts

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/maildruid"]
CMD ["serve"]
