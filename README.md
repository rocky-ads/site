=== Rocky Ads Web Site ===

Development
-----------

**Prerequisites:** Node.js (includes npm). Install from https://nodejs.org or e.g. `brew install node` on macOS.

**CSS (Tailwind)**

The app uses Tailwind CSS v4 (listed in `package.json`; no need to install it separately). Install deps once, then build the stylesheet before or while running the server:

1. `npm install`
2. One-off build: `npm run build-css-dev`
   - Or watch mode (rebuild on changes): `npm run build-css`
   - Or production (minified): `npm run build-css-prod`

Input: `input.css` → output: `static/css/output.css`.

**Object storage (MinIO)**

Ad images are stored in MinIO. The server requires `MINIO_API_URL` (and related credentials) to start.

- `MINIO_API_URL` — S3 API endpoint used by the server (private-network host in production).
- `MINIO_PUBLIC_URL` — browser-facing host used when minting presigned PUT/GET URLs (e.g. `https://minio.rockyads.com` or `http://127.0.0.1:9000` locally). Defaults to `MINIO_API_URL` if unset.

Browsers upload WebP derivatives via short-lived **presigned PUT** URLs. Listing/detail pages embed **reused long-lived presigned GET** URLs so image bytes come from MinIO and remain browser-cacheable.

Local dev with docker-compose:

```bash
docker compose up -d   # starts postgres, minio, jump-server
```

Set in `.env`:

```
MINIO_API_URL=http://127.0.0.1:9000
MINIO_PUBLIC_URL=http://127.0.0.1:9000
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
MINIO_BUCKET_NAME=rockyads
```

After seeding the database (Admin TUI → Init database with seed), populate images with `go run ./cmd/gen_images` or `go run ./cmd/migrate_images`. See [doc/README.jump-server.md](doc/README.jump-server.md).
