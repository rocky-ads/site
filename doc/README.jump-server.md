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
docker compose exec -it jump-server admin
```

Menu:

- **Backup** — writes `backups/backup-YYYYMMDD-HHMMSS.tar.gz`
- **Restore** — pick a file from `backups/`, optionally enter source `BACKUP_DB_ENCRYPTION_KEY`, confirm wipe, restore
- **Init database** — wipe and reload schema + categories (or with seed users/ads)
- **Promote / Demote admin** — by username

Non-interactive init (scripts / CI):

```bash
admin init              # schema + categories
admin init -load-seed   # also seed users and ads
```

`./backups` on the host is mounted at `/workspace/backups`.

### Encryption keys

| Env var | Use |
|---------|-----|
| `DB_ENCRYPTION_KEY` | Live DB crypto (name/phone; journals sealed on write/restore) |
| `BACKUP_DB_ENCRYPTION_KEY` | Decrypt archive user fields when restoring from another env (prompted in TUI; empty uses `DB_ENCRYPTION_KEY`) |

<<<<<<< Updated upstream
### Backup and restore ads

To rebuild the database from a backup:

```bash
backup_db backup -out /workspace/backups/prod
backup_db restore -from /workspace/backups/prod
```

`backup` / `restore` use a `.tar.gz` archive (`.tar.gz` is appended when omitted). The staging directory is not kept.

`backup_db` exports all ads, along with dependent users, locations, facets, bookmarks, clicks, conversations, messages, and MinIO images. `restore` first resets the database like `init_db` (schema + categories, no seed users/ads), then imports the archive. Archive format v2 uses creation-order refs instead of database IDs; restored ads get new IDs. Embeddings are excluded and recomputed by the server after restore.

`restore` requires `USER_ENCRYPTION_KEY` to re-encrypt user data under new IDs. If the backup came from an environment with a different key, set `BACKUP_USER_ENCRYPTION_KEY` to the source key for decrypt; encrypt always uses `USER_ENCRYPTION_KEY`.

Backups are persisted under `./backups` on the host via the docker-compose volume mount.
=======
Backup verify-decrypts users with `DB_ENCRYPTION_KEY` and stores ciphertext in the archive. Restore re-keys users to the target `DB_ENCRYPTION_KEY` and seals conversation journals. Restore resets the DB like init (schema + categories) before import. Embeddings are omitted and recomputed by the server after restore.
>>>>>>> Stashed changes

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
| `admin` | TUI + `admin init [-load-seed]`: backup/restore, init DB, promote/demote |
| `migrate_images` | One-time upload of local ad image files to MinIO |
| `quote_server` | Quote-of-the-day page on port 10000 |
| `mc` | MinIO command-line client |
