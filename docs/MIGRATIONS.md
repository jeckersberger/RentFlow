# Database migration policy

CrateDesk database migrations are append-only production history.

## Rules

- New migration files use `<numeric-version>_<description>.sql` inside the owning service's `migrations/` directory.
- Applied migration files are immutable. The runner records a SHA-256 checksum and refuses to continue when an already-recorded file changes.
- SQL failures are fatal. Migrations are executed with PostgreSQL `ON_ERROR_STOP` inside a transaction.
- Re-running the migration command is expected and must be idempotent through the migration ledger, not through ignored SQL errors.
- Generic automatic down migrations are intentionally disabled. Rollback of data/schema changes must be designed explicitly for the affected release.
- `scripts/audit-migrations.sh` is a required CI check and rejects invalid filenames or new duplicate numeric versions.

## Historical duplicate versions

The repository already contained three duplicate version numbers before the versioned runner was introduced. They are recorded in `scripts/migration-legacy-duplicates.txt` and are the only allowed exceptions:

- `auth-service` version `005`
- `expense-service` version `003`
- `project-service` version `002`

Do not add new entries to that allow-list to bypass CI. Resolve any newly introduced collision before merge.

## Commands

Run all service migrations through the same runner used by Docker Compose:

```sh
make migrate-all
```

Run one service explicitly:

```sh
make migrate-up SERVICE=auth-service
```

The PostgreSQL integration test exercises first application, a no-op second application, ledger persistence, and checksum-drift rejection:

```sh
bash scripts/test-migrate-internal.sh
```
