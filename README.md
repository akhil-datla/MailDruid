# MailDruid

[![CI](https://github.com/akhil-datla/maildruid/actions/workflows/ci.yml/badge.svg)](https://github.com/akhil-datla/maildruid/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/akhil-datla/maildruid)](https://goreportcard.com/report/github.com/akhil-datla/maildruid)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Version](https://img.shields.io/github/go-mod/go-version/akhil-datla/maildruid)](go.mod)

**Automated email summarization service** that connects to your inbox via IMAP, intelligently summarizes messages by topic, generates word cloud visualizations, and delivers periodic digest emails.

## Features

- **Smart Email Filtering** — Filter emails by tags (subject keywords), sender blacklists, and date ranges
- **AI-Powered Summaries** — Automatic text summarization using the TLDR algorithm
- **Word Cloud Generation** — Visual keyword extraction with RAKE algorithm and PNG word clouds
- **Scheduled Digests** — Configurable periodic summaries delivered straight to your inbox
- **RESTful API** — Clean JSON API with JWT authentication, input validation, and rate limiting
- **Modern Web UI** — React + TypeScript + Tailwind CSS dashboard, embedded in a single binary
- **Production Ready** — Structured logging, graceful shutdown, health checks, Docker support

## Architecture

```
┌─────────────────────────────────────────────────┐
│              REST API (Echo v4)                  │
│         /api/v1/* with JWT Auth                  │
├──────────┬──────────┬──────────┬────────────────┤
│  Users   │ Schedules│ Summaries│    Health       │
├──────────┴──────────┴──────────┴────────────────┤
│              Domain Services                     │
│         User Service │ Summary Service           │
├──────────────────────────────────────────────────┤
│            Infrastructure Layer                  │
│  PostgreSQL │ IMAP │ SMTP │ Encryption │ WordCloud│
└──────────────────────────────────────────────────┘
```

## Quick Start

### Using Docker Compose (recommended)

```bash
# Clone the repository
git clone https://github.com/akhil-datla/maildruid.git
cd maildruid

# Configure environment
export MAILDRUID_AUTH_SIGNING_KEY="your-secret-signing-key"
export MAILDRUID_AUTH_ENCRYPTION_KEY="your-32-byte-encryption-key!!"  # exactly 16, 24, or 32 bytes
export MAILDRUID_SMTP_HOST="smtp.gmail.com"
export MAILDRUID_SMTP_PORT=587
export MAILDRUID_SMTP_EMAIL="you@gmail.com"
export MAILDRUID_SMTP_PASSWORD="your-app-password"

# Start services
docker compose up -d
```

### From Source

```bash
# Prerequisites: Go 1.23+, PostgreSQL

# Build
make build

# Configure
cp config.example.yaml config.yaml
# Edit config.yaml with your settings

# Run database migrations
./bin/maildruid migrate

# Start the server
./bin/maildruid serve
```

## Configuration

MailDruid can be configured via:

1. **Config file** — `config.yaml` (see [config.example.yaml](config.example.yaml))
2. **Environment variables** — Prefixed with `MAILDRUID_` (e.g., `MAILDRUID_DATABASE_HOST`)
3. **CLI flags** — `--config /path/to/config.yaml`

| Variable | Description | Default |
|---|---|---|
| `MAILDRUID_SERVER_PORT` | HTTP server port | `8080` |
| `MAILDRUID_DATABASE_HOST` | PostgreSQL host | `localhost` |
| `MAILDRUID_DATABASE_PORT` | PostgreSQL port | `5432` |
| `MAILDRUID_DATABASE_NAME` | Database name | `maildruid` |
| `MAILDRUID_AUTH_SIGNING_KEY` | JWT signing key | **required** |
| `MAILDRUID_AUTH_ENCRYPTION_KEY` | AES encryption key (16/24/32 bytes) | **required** |
| `MAILDRUID_SMTP_HOST` | SMTP server host | **required** |
| `MAILDRUID_SMTP_EMAIL` | Sender email address | **required** |
| `MAILDRUID_SMTP_PASSWORD` | Sender email password | **required** |
| `MAILDRUID_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |
| `MAILDRUID_LOG_FORMAT` | Log format (text/json) | `text` |

## API Reference

### Authentication

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/users` | Register a new user |
| `POST` | `/api/v1/auth/login` | Login and receive JWT token |

### User Management (requires JWT)

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/users/me` | Get user profile |
| `PATCH` | `/api/v1/users/me` | Update user profile |
| `DELETE` | `/api/v1/users/me` | Delete user account |

### Email Configuration (requires JWT)

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/users/me/folders` | List IMAP folders |
| `PATCH` | `/api/v1/users/me/folder` | Set target folder |
| `PUT` | `/api/v1/users/me/tags` | Set email filter tags |
| `PUT` | `/api/v1/users/me/blacklist` | Set sender blacklist |
| `PATCH` | `/api/v1/users/me/start-time` | Set start time filter |
| `PATCH` | `/api/v1/users/me/summary-count` | Set summary sentence count |

### Scheduling (requires JWT)

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/schedules` | List all scheduled tasks |
| `POST` | `/api/v1/schedules` | Create a scheduled task |
| `PATCH` | `/api/v1/schedules` | Update task interval |
| `DELETE` | `/api/v1/schedules` | Remove a scheduled task |

### Summaries (requires JWT)

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/summaries/generate` | Generate summary on demand |

### Health Checks

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/healthz` | Liveness probe |
| `GET` | `/readyz` | Readiness probe (checks DB) |

### Example: Create User

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@company.com",
    "receivingEmail": "jane@gmail.com",
    "password": "imap-password",
    "domain": "imap.company.com",
    "port": 993
  }'
```

### Example: Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "jane@company.com", "password": "imap-password"}'

# Response: {"token": "eyJhbGciOiJIUzI1NiIs..."}
```

### Example: Generate Summary

```bash
curl -X POST http://localhost:8080/api/v1/summaries/generate \
  -H "Authorization: Bearer <your-token>"

# Response: {"summary": "...", "image": "<base64-png>"}
```

## CLI Commands

```bash
maildruid serve      # Start the HTTP server
maildruid migrate    # Run database migrations
maildruid version    # Print version information
```

## Development

```bash
# Install all dependencies (Go + frontend)
make deps

# Build everything (frontend + Go binary)
make build

# Run locally
make run

# Run frontend dev server with hot reload (proxies API to :8080)
make dev

# Run tests
make test

# Run linter
make lint

# Format code
make fmt

# Build Docker image
make docker-build
```

## Project Structure

```
cmd/maildruid/          # CLI entry point (Cobra)
internal/
  config/               # Configuration (Viper)
  domain/
    user/               # User model, repository interface, service
    summary/            # Email summarization pipeline
  infrastructure/
    postgres/           # PostgreSQL repository implementation
    imap/               # IMAP email client
    smtp/               # SMTP email sender
    encryption/         # AES-256-CFB encryption
    wordcloud/          # Text summarization & word cloud generation
  scheduler/            # Periodic task scheduler
  server/
    handlers/           # HTTP request handlers
    middleware/         # JWT auth, rate limiting
    frontend/           # Embedded frontend (built from web/)
web/                    # React + TypeScript + Tailwind CSS frontend
fonts/                  # Font files for word cloud rendering
```

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Authors

- **Akhil Datla** — [GitHub](https://github.com/akhil-datla)
- **Alexander Ott**
