# Monelog API

The Go backend foundation for Monelog. ISSUE-001 provides validated configuration, PostgreSQL connectivity, health checks, safe HTTP logging, and graceful shutdown.

## Language / Bahasa

- [English](README.en.md)
- [Bahasa Indonesia](README.md)

## Confirmed behavior
- Regular users can CRUD only their own data.
- Admins can perform all application operations on any selected user's data, including CRUD, roles, reports, exports, templates and backups.
- Soft deletion uses isDelete=false for active records and true for deleted records. Delete retains the row; restore changes the flag back to false.

## Documents
1. [Requirements](docs-en/requirements.md)
2. [Access control and soft deletion](docs-en/access-control.md)
3. [Architecture](docs-en/architecture.md)
4. [Database](docs-en/database.md)
5. [API](docs-en/api.md)
6. [Roadmap](docs-en/roadmap.md)
7. [Issue index](docs-en/issues.md)
8. [Workflow](docs-en/workflow.md)

See the [project structure](docs-en/architecture.md#project-structure) for the planned backend directory layout.

Start [ISSUE-001](docs-en/issues/ISSUE-001-project-setup.md) with [PLAN-001](docs-en/plans/PLAN-001.md).
Finish backend 001–007 and [ISSUE-013](docs-en/issues/ISSUE-013-admin-viewing.md) before frontend Issue 008.
Issue 013 now covers full admin management; its original filename is retained to preserve links.
Then deliver exports, Android/iOS, templates and Drive backups in their listed milestones.
Task Markdown files are the portable source of truth and may link to GitHub Issues without changing their IDs. Keep docs canonical in monelog-api; monelog-app references a pinned documentation/API commit.

## Running the API

Prerequisites:

- Go 1.27.1
- PostgreSQL accessible through a local connection URL
- sqlc 1.31.1 for query regeneration
- golang-migrate 4.18 or compatible for migrations

Copy `.env.example` to `.env.development` and replace the placeholders with local PostgreSQL credentials. When `APP_ENV` is unset, the application loads `.env.development`; when it is set, the application loads `.env.<APP_ENV>`. Shell and deployment environment variables always override file values.

```bash
go run ./cmd/api
```

Optional configuration and default values are documented in [.env.example](.env.example). Do not commit `.env`, passwords, tokens, or keys.

Health endpoints:

- `GET /health/live` returns `200` while the API process is alive.
- `GET /health/ready` returns `200` when PostgreSQL is available and `503` while it is unavailable.

```bash
# Confirm that the HTTP API process is alive.
curl --fail http://127.0.0.1:8080/health/live
# Confirm that PostgreSQL and runtime dependencies are ready.
curl --fail http://127.0.0.1:8080/health/ready
```

## OpenAPI and Swagger

The source contract is [api/openapi.yaml](api/openapi.yaml). While the API is running, open [Swagger UI](http://127.0.0.1:8080/swagger/) or the [OpenAPI JSON document](http://127.0.0.1:8080/api/openapi.json). Select `Authorize` in Swagger UI and enter the access token returned by login.

Register an account for a new standard user:

```bash
curl --location 'http://127.0.0.1:8080/api/v1/auth/register' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "email": "you@example.com",
    "password": "correct-horse-battery-staple",
    "timezone": "Asia/Jakarta"
  }'
```

Login for Android/iOS. The API returns the native refresh token in JSON:

```bash
curl --location 'http://127.0.0.1:8080/api/v1/auth/login' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "email": "you@example.com",
    "password": "correct-horse-battery-staple",
    "client_type": "native"
  }'
```

Browser login (`web`) requires an allowed `Origin` and sends the refresh token in an HttpOnly cookie:

```bash
curl --location 'http://127.0.0.1:8080/api/v1/auth/login' \
  --header 'Origin: https://app.example.com' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "email": "you@example.com",
    "password": "correct-horse-battery-staple",
    "client_type": "web"
  }'
```

`client_type` must use the underscore spelling and must be either `native` or `web`. Run `make oapi-generate` after changing the contract; commit the generated code in `internal/api`.

Transaction examples after obtaining an `access_token` from login. The `Authorization` header is required for the application endpoints below:

```bash
ACCESS_TOKEN='<access-token-from-login>'

# Create an income category that can be selected for a new income transaction.
curl --location 'http://127.0.0.1:8080/api/v1/categories' \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data '{"name":"Salary","type":"income"}'

# Record one transaction. amount is an exact money string and client_request_id makes retries safe.
curl --location 'http://127.0.0.1:8080/api/v1/transactions' \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data '{"transaction_date":"2026-09-17","type":"income","category_id":"<category-uuid>","amount":"15000000.00","title":"September salary","client_request_id":"<request-uuid>"}'

# List active transaction history for a date range; repeat with page.next_cursor for the next page.
curl --location 'http://127.0.0.1:8080/api/v1/transactions?start_date=2026-09-01&end_date=2026-09-30&limit=30' \
  --header "Authorization: Bearer $ACCESS_TOKEN"

# List daily income, expense, and difference for the main page summary or date navigation.
curl --location 'http://127.0.0.1:8080/api/v1/daily-summaries?start_date=2026-09-01&end_date=2026-09-30' \
  --header "Authorization: Bearer $ACCESS_TOKEN"

# List soft-deleted transactions from Trash.
curl --location 'http://127.0.0.1:8080/api/v1/transactions?start_date=2026-09-01&end_date=2026-09-30&isDelete=true' \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

Swagger UI documents the purpose of every endpoint, parameter, access scope, amount format, cursor pagination, optimistic locking, and idempotency rules. Admin endpoints use `/api/v1/admin/users/{user_id}/...`; the target `user_id` must be selected explicitly in the URL.

## Development

```bash
make fmt
make test
make vet
make build
make check
```

## Database

Run migrations as the schema owner. Use `migrate-down` only on a disposable database after reviewing its data impact.

```bash
DATABASE_URL="$DATABASE_URL" make migrate-up
make sqlc-generate
make sqlc-vet
TEST_DATABASE_URL="$TEST_DATABASE_URL" make test-integration
```

The application runtime must use a role separate from the migration owner. After the tables exist, grant minimum runtime privileges with:

```bash
DATABASE_URL="$DATABASE_URL" RUNTIME_DB_ROLE=monelog_runtime make runtime-grants
```

The runtime role can read and write required application data, but it cannot physically delete business rows or update/delete audit records. User and admin authorization remains enforced by the service in later issues.
