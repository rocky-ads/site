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

## Database initialization

```bash
export DATABASE_URL="postgres://postgres:postgres@postgres:5432/rockyads?sslmode=disable"
init_db
```

`init_db` recreates the schema and loads categories. It does **not** load seed users or ads unless you pass `-load-seed`. It sets `image_count` on seeded ads but does **not** upload image files.

For local/dev with test users and ads:

```bash
init_db -load-seed
```

### Promote a user to admin

After registering a user via the UI (typical without `-load-seed`):

```bash
set_admin promote <name>
set_admin demote <name>
```

### Backup and restore non-test ads

To preserve production ads across a database rebuild:

```bash
backup_db backup -out /workspace/backups/prod
init_db
backup_db restore -from /workspace/backups/prod
```

`backup_db` exports ads not owned by the seeded `test` user (when that user exists; otherwise all ads), along with dependent users, locations, facets, bookmarks, clicks, conversations, messages, and MinIO images. Archive format v2 uses creation-order refs instead of database IDs; restored ads are appended with new IDs. Embeddings are excluded and recomputed by the server after restore.

`restore` requires `USER_ENCRYPTION_KEY` to re-encrypt user data under new IDs. If the backup came from an environment with a different key, set `BACKUP_USER_ENCRYPTION_KEY` to the source key for decrypt; encrypt always uses `USER_ENCRYPTION_KEY`.

`restore` also runs an idempotent one-time schema bridge (`migrate-schema`) so a target DB still on the old phone lifecycle constraints (global unique `phone_hash`; `phone_verification` without `purpose` / `user_id`) is upgraded before import. To upgrade an in-place DB without restore:

```bash
backup_db migrate-schema
```

Backups are persisted under `./backups` on the host via the docker-compose volume mount.

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
| `init_db` | Drop/recreate schema and categories; `-load-seed` also loads users/ads |
| `set_admin` | Promote or demote a user by name (`promote` / `demote`) |
| `backup_db` | Backup/restore non-test ads (DB rows + MinIO images); one-time `migrate-schema` |
| `migrate_images` | One-time upload of local ad image files to MinIO |
| `quote_server` | Quote-of-the-day page on port 10000 |
| `mc` | MinIO command-line client |
