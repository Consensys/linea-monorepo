# Database Migration Strategy

This project uses Prisma Migrate for database schema management with tracked migration history.

## Migration Workflow

### Development: Creating Schema Changes

1. Edit `prisma/schema.prisma`
2. Generate a migration:
   ```bash
   make test-db-migrate
   # Or manually:
   DATABASE_URL=postgresql://... pnpm db:migrate --name describe_change
   ```
3. Review the generated SQL in `prisma/migrations/<timestamp>_describe_change/`
4. Commit the migration with your schema change:
   ```bash
   git add prisma/
   git commit -m "feat: add X to schema"
   ```

### Testing: Applying Migrations

```bash
# Start test DB and apply all migrations
make test-db

# Check migration status
DATABASE_URL=postgresql://... pnpm db:status

# Clean up
make clean
```

### Production: Deploying Migrations

Migrations are applied automatically via init container before the application starts.

## Commands Reference

| Command | Environment | Description |
|---------|-------------|-------------|
| `pnpm db:migrate --name X` | Development | Generate new migration from schema changes |
| `pnpm db:deploy` | Production/CI | Apply pending migrations (no prompts) |
| `pnpm db:status` | Any | Check which migrations have been applied |
| `pnpm db:push` | Development | Push schema without migration history (legacy) |
| `pnpm db:generate` | Any | Regenerate Prisma Client |

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make test-db` | Start test DB and apply migrations |
| `make test-db-migrate` | Start test DB and generate new migration |
| `make clean` | Stop test DB and remove volumes |

## Kubernetes Deployment

### Helm Chart Configuration

Add an init container to run migrations before the application starts:

```yaml
# values.yaml or deployment template
initContainers:
  - name: db-migrate
    image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
    command: ["npx", "prisma", "migrate", "deploy"]
    env:
      - name: DATABASE_URL
        valueFrom:
          secretKeyRef:
            name: {{ .Values.database.secretName }}
            key: {{ .Values.database.secretKey }}
```

### Key Points

- `migrate deploy` is production-safe (no interactive prompts)
- Init container runs before the main application container starts
- If migration fails, the pod won't start (deployment is blocked)
- Only one replica runs the init container at a time (Kubernetes handles this)

### Verifying Migrations in Production

```bash
# Check init container logs
kubectl logs <pod-name> -c db-migrate

# Or from local machine with DB access
DATABASE_URL=postgresql://... pnpm db:status
```

## Migration Files

Migrations are stored in `prisma/migrations/`:

```
prisma/migrations/
├── 20260130121027_init/
│   └── migration.sql
└── migration_lock.toml
```

Each migration directory contains:
- `migration.sql` - The SQL statements to apply
- Timestamp prefix for ordering

The `migration_lock.toml` file locks the database provider (PostgreSQL).

## Troubleshooting

### Migration Failed in Production

1. Check init container logs: `kubectl logs <pod> -c db-migrate`
2. Fix the issue in the migration SQL if needed
3. Redeploy

### Schema Drift

If the database schema doesn't match migrations:

```bash
# Check current state
pnpm db:status

# In development, reset and reapply all migrations
DATABASE_URL=... pnpm exec prisma migrate reset
```

### Reverting a Migration

Prisma doesn't support automatic rollbacks. To revert:

1. Create a new migration that undoes the changes
2. Or restore from database backup
