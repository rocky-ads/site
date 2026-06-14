# Jump Server

The jump server image includes database utilities, image migration, and the MinIO client (`mc`).

## Building

```bash
docker build -f dockerfiles/Dockerfile.jump-server -t rockyads-jump-server:dev .
```

With docker-compose (includes Postgres and MinIO):

```bash
docker compose up -d
```

## Database seeding

```bash
export DATABASE_URL="postgres://postgres:postgres@postgres:5432/rockyads?sslmode=disable"
rebuild_db
```

`rebuild_db` recreates the schema and imports seed data. It sets `image_count` on ads but does **not** upload image files.

## Ad images

Images are stored in MinIO (required). After seeding the database, populate images from your dev machine with `go run ./cmd/gen_images` (requires `FAL_API_KEY`) or upload existing local files with `migrate_images` on the jump server (see below).

### Migrate existing local files (one-time)

If you have files under `static/images/ad/` (or another directory):

```bash
migrate_images -dir static/images/ad
```

Use `-dry-run` to preview without uploading. MinIO connection uses `MINIO_API_URL`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`, and `MINIO_BUCKET_NAME` from the service environment.

On the jump server, you can also mirror files with `mc` (see below).

## Using mc

The MinIO client is pre-installed:

```bash
mc alias set local http://minio:9000 minioadmin minioadmin
mc mb local/rockyads --ignore-existing
mc mirror static/images/ad/ local/rockyads/
mc --version
mc ls local/rockyads
mc cp local/rockyads/23/1-480w.webp ./
```

See the [MinIO Client documentation](https://docs.min.io/docs/minio-client-quickstart-guide.html).

## Binaries

| Binary | Purpose |
|--------|---------|
| `rebuild_db` | Drop/recreate schema and import seed data |
| `migrate_images` | One-time upload of local ad image files to MinIO |
| `quote_server` | Quote-of-the-day page on port 10000 |
| `mc` | MinIO command-line client |
