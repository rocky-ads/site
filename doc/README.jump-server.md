# Jump Server

The jump server image includes database utilities, image tools, and the MinIO client (`mc`).

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

Images are stored in MinIO (required). After seeding the database, populate images with one of:

### Generate new images (AI)

```bash
export MINIO_API_URL="http://minio:9000"
export MINIO_USERNAME="minioadmin"
export MINIO_PASSWORD="minioadmin"
export MINIO_BUCKET_NAME="rockyads"
export FAL_API_KEY="your-key"

gen_images
```

Generates WebP images (160w, 480w, 1200w) and uploads them to MinIO for each ad.

### Migrate existing local files (one-time)

If you have files under `static/images/ad/` from before MinIO:

```bash
export MINIO_API_URL="http://minio:9000"
export MINIO_USERNAME="minioadmin"
export MINIO_PASSWORD="minioadmin"
export MINIO_BUCKET_NAME="rockyads"

migrate_images -dir static/images/ad
```

Use `-dry-run` to preview without uploading.

### Manual copy with mc (optional)

The MinIO client is pre-installed:

```bash
mc alias set local http://minio:9000 minioadmin minioadmin
mc mb local/rockyads --ignore-existing
mc mirror static/images/ad/ local/rockyads/
```

## Using mc

```bash
mc --version
mc ls local/rockyads
mc cp local/rockyads/23/1-480w.webp ./
```

See the [MinIO Client documentation](https://docs.min.io/docs/minio-client-quickstart-guide.html).

## Binaries

| Binary | Purpose |
|--------|---------|
| `rebuild_db` | Drop/recreate schema and import seed data |
| `gen_images` | Generate AI ad images and upload to MinIO |
| `migrate_images` | One-time upload of local `static/images/ad/` files to MinIO |
| `quote_server` | Quote-of-the-day page on port 10000 |
| `mc` | MinIO command-line client |
