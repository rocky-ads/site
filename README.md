# Rocky Ads Web Site

[![Go Reference](https://pkg.go.dev/badge/github.com/rocky-ads/site.svg)](https://pkg.go.dev/github.com/rocky-ads/site)
![Go Version](https://img.shields.io/github/go-mod/go-version/rocky-ads/site)

[Rocky Ads](https://rockyads.com) is classified ads for the Internet.

<p align="center">
  <img src="static/images/classifieds.jpg" alt="Newspaper classifieds" height="500">
</p>

Remember classified ads? Easy, simple. Post an ad with your phone number and folks call you. Rocky Ads works the same way, except your phone number stays hidden.

&emsp;🚫 No email  
&emsp;🚫 No Facebook friends  
&emsp;🚫 No posting fees  
&emsp;🚫 No credit cards

All you need is your phone number to [get started](https://rockyads.com/register).

## Development

```mermaid
flowchart LR
  Browser -->|HTTP / HTMX| Site["Site<br/>Go / Fiber"]
  Browser -->|presigned PUT/GET| MinIO
  Site --> Postgres[(Postgres)]
  Site --> MinIO
  Site --> Ollama
  Site --> Twilio
  Site --> Grok["Grok API<br/>x.ai"]
  Twilio -->|SMS webhooks| Site
```

### Prerequisites

- Go 1.26.5+ — https://go.dev/dl
- Node.js (includes npm) — https://nodejs.org, or e.g. `brew install node` on macOS

### CSS (Tailwind)

The app uses Tailwind CSS v4 (listed in `package.json`; no need to install it separately). Install deps once, then build the stylesheet before or while running the server:

1. `npm install`
2. One-off build: `npm run build-css-dev`
   - Or watch mode (rebuild on changes): `npm run build-css`
   - Or production (minified): `npm run build-css-prod`

Input: `input.css` → output: `static/css/output.css`.

### Object storage (MinIO)

Ad images are stored in MinIO. The server requires `MINIO_API_URL` (and related credentials) to start.

- `MINIO_API_URL` — S3 API endpoint used by the server (private-network host in production).
- `MINIO_PUBLIC_URL` — browser-facing host used when minting presigned PUT/GET URLs (e.g. `https://minio.rockyads.com` or `http://127.0.0.1:9000` locally). Defaults to `MINIO_API_URL` if unset.

Browsers upload WebP derivatives via short-lived **presigned PUT** URLs. Listing/detail pages embed **reused long-lived presigned GET** URLs so image bytes come from MinIO and remain browser-cacheable.

### SMS (Twilio)

Phone verification, account recovery, and notification SMS use Twilio. Required at server start unless `ALLOW_TEST_REGISTRATION=true` (which skips Twilio validation for local/test).

- `TWILIO_ACCOUNT_SID` — Twilio account SID (must start with `AC`).
- `TWILIO_AUTH_TOKEN` — Twilio auth token.
- `TWILIO_FROM_NUMBER` — sender number in E.164 (e.g. `+12025550123`).
- `TWILIO_WEBHOOK_URL` — public base URL for Twilio webhooks and SMS links. Must be a non-localhost `http`/`https` URL (use ngrok locally).

### DB (Postgres)

- `DATABASE_URL` — PostgreSQL DSN. Default: `postgres://localhost:5432/rockyads?sslmode=disable`. With docker-compose Postgres: `postgres://postgres:postgres@localhost:5432/rockyads?sslmode=disable`.
- `DB_ENCRYPTION_KEY` — required. Base64-encoded 32-byte (256-bit) key for encrypting user name/phone.

### LLM APIs

- `GROK_API_KEY` — xAI Grok API key (chat completions for rock opinions / AI features).
- `OLLAMA_URL` — Ollama HTTP endpoint for embeddings (`nomic-embed-text`). Default: `http://localhost:11434`.
- `FAL_API_KEY` — fal.ai key for `go run ./cmd/gen_images` only (not required by the server).

### JWT

- `JWT_SECRET` — required. Secret for signing auth cookies; at least 32 characters.

### Other settings

- `PORT` — HTTP listen port. Default: `10000`.
- `PORT_TEST` — port used by integration tests. Default: `10001`.
- `APP_NAME` — display name. Default: `Rocky Ads`.
- `CONTACT_EMAIL` — contact address shown in the UI. Default: `contact@rockyads.com`.
- `LOCAL_DEVELOPMENT` — set to `true` so auth cookies work over plain HTTP (otherwise cookies are Secure-only).
- `ALLOW_TEST_REGISTRATION` — set to `true` to skip SMS verification for `+1555010xxxx` phones and relax Twilio startup checks (dev/test only).
- `LOG_LEVEL` — `debug` / `info` / … Default: `info`.
- `LOG_FORMAT` — `json` or text. Default: `json`.
- `LOG_FILE` — optional log file path; empty logs to stderr.

### Local dev with docker-compose

```bash
docker compose up -d   # starts postgres, minio, jump-server, ollama
```

Set in `.env`:

```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/rockyads?sslmode=disable
DB_ENCRYPTION_KEY=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
JWT_SECRET=local-dev-jwt-secret-key-min-32-chars
LOCAL_DEVELOPMENT=true
ALLOW_TEST_REGISTRATION=true
MINIO_API_URL=http://127.0.0.1:9000
MINIO_PUBLIC_URL=http://127.0.0.1:9000
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
MINIO_BUCKET_NAME=rockyads
OLLAMA_URL=http://localhost:11434
```

Add Twilio (`TWILIO_*`), `GROK_API_KEY`, and `FAL_API_KEY` when you need real SMS, Grok, or image generation.

Initialize the database via the jump-server admin TUI:

```bash
docker compose exec -it jump-server admin
```

Choose **Init database**.

Run the site (loads `.env` into the environment):

```bash
set -a && source .env && set +a
go run ./cmd/server
```

Open http://localhost:10000.
