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

## Admin tool

Interactive TUI (requires a TTY):

```bash
# local compose
docker compose exec -it jump-server admin

# Render / SSH — allocate a pseudo-TTY
ssh -t <jump-host> admin
```

Without a TTY, `admin` exits immediately (stdin EOF). Use `ssh -t` or an equivalent that requests a PTY.

Menu:

- **Backup** — writes `backups/backup-YYYYMMDD-HHMMSS.tar.gz`
- **Restore** — pick a file from `backups/`, optionally enter source `BACKUP_DB_ENCRYPTION_KEY`, confirm wipe, restore
- **Init database** — wipe and reload schema + categories (or with seed users/ads)
- **Promote / Demote admin** — by username

`./backups` on the host is mounted at `/workspace/backups`.

### Encryption keys

| Env var | Use |
|---------|-----|
| `DB_ENCRYPTION_KEY` | Live DB crypto (name/phone; journals sealed on write/restore) |
| `DB_HASH_PEPPER` | HMAC pepper for `name_hash` / `phone_hash` (must differ from encryption key) |
| `BACKUP_DB_ENCRYPTION_KEY` | Decrypt archive user fields when restoring from another env (prompted in TUI; empty uses `DB_ENCRYPTION_KEY`) |

Backup verify-decrypts users with `DB_ENCRYPTION_KEY` and stores ciphertext in the archive. Restore re-keys users to the target `DB_ENCRYPTION_KEY`, recomputes peppered lookup hashes with `DB_HASH_PEPPER`, and seals conversation journals. Restore resets the DB like init (schema + categories) before import. Ad embeddings and `vector_metadata` are included in the archive (base64 pgvector binary on each ad) so search works without recomputing after restore. Older archives without embeddings still restore; those ads get vectors from the server backfill on startup.

Ad `expires_at` / `expire_grant` are included in new archives. Older archives without those fields are still restorable: restore backfills `expires_at` from `sale_end_date` (+1 week) when present, otherwise `created_at` + 3 months, and sets `expire_grant` to 3 months.

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
| `admin` | Interactive TUI: backup/restore, init DB, promote/demote |
| `migrate_images` | One-time upload of local ad image files to MinIO |
| `quote_server` | Quote-of-the-day page on port 10000 |
| `mc` | MinIO command-line client |
