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

Local dev with docker-compose:

```bash
docker compose up -d   # starts postgres, minio, jump-server
```

Set in `.env`:

```
MINIO_API_URL=http://127.0.0.1:9000
MINIO_USERNAME=minioadmin
MINIO_PASSWORD=minioadmin
MINIO_BUCKET_NAME=rockyads
```

After `rebuild_db`, populate images with `go run ./cmd/gen_images` or `go run ./cmd/migrate_images`. See [doc/README.jump-server.md](doc/README.jump-server.md).
