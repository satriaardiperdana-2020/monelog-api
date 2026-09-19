# Monelog API

Monelog's REST backend for authentication, user profiles, categories, transactions, Trash/restore, daily summaries, reports, and administrator access. It is built with Go, Echo, PostgreSQL, sqlc, and an OpenAPI contract.

This documentation covers the backend only. The Indonesian version is available in [README.md](README.md).

## Available features

- Registration, login, refresh-token rotation, logout, and JWT access tokens.
- User profiles protected by optimistic locking through `version`.
- Income and expense categories with soft deletion and restoration.
- Exact-money transactions, idempotency, cursor pagination, soft deletion, and restoration.
- Daily summaries and reports grouped by week, month, or category.
- Administrator endpoints for a selected user's categories, transactions, and reports.
- PostgreSQL migrations, generated sqlc queries, OpenAPI, Swagger UI, health checks, and graceful shutdown.

All entity primary keys use `BIGSERIAL`. Identifiers and foreign keys use `BIGINT`. Money is stored as `NUMERIC(14,2)` and represented by two-decimal strings in the API.

## 1. Prerequisites

Install:

- Go `1.27.1`, as declared in [go.mod](go.mod).
- PostgreSQL.
- GNU Make.
- `golang-migrate` with the PostgreSQL driver.
- sqlc `1.31.1` if you need to regenerate database code.
- `curl`; `jq` is optional and useful for extracting response values.

Verify the tools:

```bash
go version
psql --version
migrate -version
sqlc version
make --version
```

## 2. Get the source and Go dependencies

```bash
git clone <repository-url> monelog-api
cd monelog-api
go mod download
go mod verify
```

Generated sqlc and OpenAPI files are committed. You only need sqlc when migrations, files under `db/queries`, or sqlc configuration change.

## 3. Create the PostgreSQL database and roles

The following setup separates the schema owner used for migrations from the least-privilege application role. Open PostgreSQL as an administrator:

```bash
sudo -u postgres psql
```

Run inside `psql`:

```sql
CREATE ROLE monelog_owner LOGIN PASSWORD 'change-owner-password';
CREATE ROLE monelog_runtime LOGIN PASSWORD 'change-runtime-password';
CREATE DATABASE monelog OWNER monelog_owner;
CREATE DATABASE monelog_test OWNER monelog_owner;
GRANT CONNECT ON DATABASE monelog TO monelog_runtime;
\q
```

The `monelog_test` database is used by the integration tests in step 16. The runtime role does not need access to it.

Use separate, strong passwords and never commit them.

Schema-owner connection URL:

```text
postgres://monelog_owner:change-owner-password@127.0.0.1:5432/monelog?sslmode=disable
```

Runtime connection URL:

```text
postgres://monelog_runtime:change-runtime-password@127.0.0.1:5432/monelog?sslmode=disable
```

`sslmode=disable` is suitable only for a local PostgreSQL instance. Use the TLS settings required by your staging or production provider.

## 4. Apply the DDL and migrations

The canonical DDL is stored in [db/migrations](db/migrations). Do not create the tables individually. Run every migration in order so constraints, indexes, sequences, and `schema_migrations` remain consistent.

```bash
export MIGRATION_DATABASE_URL='postgres://monelog_owner:change-owner-password@127.0.0.1:5432/monelog?sslmode=disable'

DATABASE_URL="$MIGRATION_DATABASE_URL" make migrate-up
```

The migrations create:

| Table | Purpose |
|---|---|
| `users` | Accounts, roles, timezones, currency, soft deletion, and versions. |
| `categories` | User-owned income and expense categories. |
| `transactions` | Transactions, idempotency keys, actors, soft deletion, and versions. |
| `refresh_sessions` | Refresh-token rotation and family revocation. |
| `admin_access_events` | Audit records for administrator access and mutations. |
| `schema_migrations` | Migration state maintained by `golang-migrate`. |

Verify the DDL:

```bash
psql "$MIGRATION_DATABASE_URL" -c 'SELECT version, dirty FROM schema_migrations;'
psql "$MIGRATION_DATABASE_URL" -c '\dt public.*'
psql "$MIGRATION_DATABASE_URL" -c '\d+ public.users'
psql "$MIGRATION_DATABASE_URL" -c '\d+ public.transactions'
```

The latest version must report `dirty = false`. Migration `000007` converts older UUID installations to BIGINT. Original UUID values cannot be reconstructed, so that migration intentionally rejects rollback.

Grant the application role its minimum runtime privileges after creating the DDL:

```bash
DATABASE_URL="$MIGRATION_DATABASE_URL" \
RUNTIME_DB_ROLE=monelog_runtime \
make runtime-grants
```

The runtime role receives the required table and sequence privileges, but cannot physically delete business rows or modify/delete audit records.

## 5. Configure the environment

Copy the template:

```bash
cp .env.example .env.development
```

Generate a random JWT key containing at least 32 bytes, encoded as Base64:

```bash
openssl rand -base64 32
```

Edit `.env.development`:

```dotenv
APP_ENV=development
HTTP_ADDR=127.0.0.1:8080
HTTP_READ_HEADER_TIMEOUT=5s
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=15s
HTTP_IDLE_TIMEOUT=60s
SHUTDOWN_TIMEOUT=10s
HEALTH_READY_TIMEOUT=2s
DATABASE_URL=postgres://monelog_runtime:change-runtime-password@127.0.0.1:5432/monelog?sslmode=disable
AUTH_JWT_HMAC_KEY=<output-of-openssl-rand-base64-32>
AUTH_JWT_ISSUER=monelog-api
AUTH_JWT_AUDIENCE=monelog-app
AUTH_ACCESS_TOKEN_TTL=24h
AUTH_ALLOWED_ORIGINS=http://localhost:5173
```

Configuration rules:

- With no `APP_ENV`, the application reads `.env.development`.
- `APP_ENV=staging` makes it read `.env.staging`.
- Shell environment variables always override values from the file.
- `AUTH_JWT_HMAC_KEY` must be Base64 and decode to at least 32 bytes.
- `AUTH_ACCESS_TOKEN_TTL` must be positive and no longer than 24 hours.
- `AUTH_ALLOWED_ORIGINS` accepts one or more comma-separated origins.
- Never commit `.env.*`, database passwords, access tokens, refresh tokens, or signing keys.

## 6. Bootstrap the first administrator

HTTP registration always creates a regular `user`. Create the initial administrator with the interactive CLI. It succeeds only when no active administrator exists:

```bash
go run ./cmd/admin -email admin@example.com -timezone Asia/Jakarta
```

The CLI reads the password without displaying it. The password must be 12–128 characters and pass the service validation rules.

## 7. Run the backend

Run directly:

```bash
go run ./cmd/api
```

Or build and run a binary:

```bash
make build
./bin/monelog-api
```

The default address is `http://127.0.0.1:8080`.

Check both health endpoints:

```bash
curl -i http://127.0.0.1:8080/health/live
curl -i http://127.0.0.1:8080/health/ready
```

Successful response:

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ok"}
```

`/health/live` checks the HTTP process. `/health/ready` also checks PostgreSQL and returns `503` with `{"status":"unavailable"}` when the database is unavailable.

## 8. OpenAPI and Swagger

- Source contract: [api/openapi.yaml](api/openapi.yaml)
- OpenAPI JSON while the server is running: <http://127.0.0.1:8080/api/openapi.json>
- Swagger UI: <http://127.0.0.1:8080/swagger/>

In Swagger UI, log in, copy the `access_token`, select **Authorize**, and paste the token without the `Bearer` prefix.

## 9. API request conventions

The examples below use:

```bash
export BASE_URL='http://127.0.0.1:8080'
export ACCESS_TOKEN='<access-token>'
export REFRESH_TOKEN='<refresh-token>'
export CATEGORY_ID=1
export TRANSACTION_ID=1
export USER_ID=1
```

General rules:

- Protected endpoints require `Authorization: Bearer <access_token>`.
- Every ID is a positive 64-bit integer.
- `amount` is a string such as `"43500.00"`.
- Dates use `YYYY-MM-DD`; response timestamps use RFC3339 UTC.
- `type` must be `income` or `expense`.
- Updates, deletion, and restoration use `version` for optimistic locking.
- DELETE requests use `If-Match`, for example `If-Match: "2"`.
- `client_request_id` is a positive integer unique per user. Retrying the same ID and payload returns the existing transaction with status `200`; a different payload returns `409`.
- JSON bodies reject unknown fields and multiple JSON values.

Errors use a consistent envelope:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request."
  }
}
```

Common statuses are `400` for malformed syntax/cursors, `401` for failed authentication, `403` for forbidden access, `404` for resources outside the authorized scope, `409` for state/version/idempotency conflicts, `422` for validation failures, `429` for login rate limiting, and `500`/`503` for server/dependency failures.

## 10. Authentication and profile

### Register a user

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@example.com",
    "password": "correct-horse-battery-staple",
    "timezone": "Asia/Jakarta"
  }'
```

`201 Created` response:

```json
{
  "data": {
    "id": 1,
    "email": "user@example.com",
    "role": "user",
    "timezone": "Asia/Jakarta",
    "currency": "IDR",
    "isDelete": false,
    "version": 1
  }
}
```

Registration also seeds the user's default categories.

### Native login

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@example.com",
    "password": "correct-horse-battery-staple",
    "client_type": "native"
  }'
```

`200 OK` response:

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<opaque-refresh-token>",
  "token_type": "Bearer",
  "expires_in": 86399
}
```

Store values from the response, or extract them with `jq`:

```bash
LOGIN_RESPONSE=$(curl -sS -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"correct-horse-battery-staple","client_type":"native"}')

export ACCESS_TOKEN=$(printf '%s' "$LOGIN_RESPONSE" | jq -r '.access_token')
export REFRESH_TOKEN=$(printf '%s' "$LOGIN_RESPONSE" | jq -r '.refresh_token')
```

### Native refresh

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/refresh" \
  -H 'Content-Type: application/json' \
  -d "{\"client_type\":\"native\",\"refresh_token\":\"$REFRESH_TOKEN\"}"
```

The response has the same shape as native login and contains a new access token and refresh token. Use the newly returned refresh token after every rotation.

### Native logout

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/logout" \
  -H 'Content-Type: application/json' \
  -d "{\"client_type\":\"native\",\"refresh_token\":\"$REFRESH_TOKEN\"}"
```

The response is `204 No Content` with no body.

### Web login and refresh

A web client must use HTTPS outside localhost, an allowed origin, a cookie jar, and a CSRF token. Login does not include a refresh token in the JSON response:

```bash
curl -i -c cookies.txt -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Origin: http://localhost:5173' \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"correct-horse-battery-staple","client_type":"web"}'
```

Response body:

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_in": 86399
}
```

The server stores the refresh token in the HttpOnly `__Host-monelog-refresh` cookie and the CSRF value in `__Host-monelog-csrf`. Send the CSRF cookie value through `X-CSRF-Token` for refresh/logout:

```bash
export CSRF_TOKEN='<value-of-__Host-monelog-csrf-cookie>'

curl -i -b cookies.txt -c cookies.txt \
  -X POST "$BASE_URL/api/v1/auth/refresh" \
  -H 'Origin: http://localhost:5173' \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"client_type":"web"}'
```

### Read the current profile

```bash
curl -i "$BASE_URL/api/v1/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

The `200` response uses the same user shape as registration.

### Update the profile timezone

```bash
curl -i -X PATCH "$BASE_URL/api/v1/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"timezone":"UTC","version":1}'
```

The `200` response returns `timezone: "UTC"` and increments `version` to `2`.

### Soft-delete the account

```bash
curl -i -X DELETE "$BASE_URL/api/v1/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'If-Match: "2"'
```

The response is `204 No Content`. The service also revokes all refresh sessions for the user.

## 11. Categories

### List active categories

```bash
curl -i "$BASE_URL/api/v1/categories?type=expense&isDelete=false" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

`200` response:

```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "type": "expense",
      "name": "Food",
      "isDelete": false,
      "version": 1
    }
  ]
}
```

### Create a category

```bash
curl -i -X POST "$BASE_URL/api/v1/categories" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Subscriptions","type":"expense"}'
```

The `201` response contains one category object under `data`. Names are case-insensitively unique for each user/type pair.

### Read and update a category

```bash
curl -i "$BASE_URL/api/v1/categories/$CATEGORY_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i -X PATCH "$BASE_URL/api/v1/categories/$CATEGORY_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Digital subscriptions","version":1}'
```

The `200` update response increments `version` to `2`.

### Archive and restore a category

```bash
curl -i -X DELETE "$BASE_URL/api/v1/categories/$CATEGORY_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'If-Match: "2"'

curl -i "$BASE_URL/api/v1/categories/$CATEGORY_ID?isDelete=true" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i -X POST "$BASE_URL/api/v1/categories/$CATEGORY_ID/restore" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"version":3}'
```

DELETE returns `204`. Restore returns `200`, marks the category active, and increments its version again.

## 12. Transactions

### Create a transaction

Use an active category owned by the user with a matching type:

```bash
curl -i -X POST "$BASE_URL/api/v1/transactions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{
    \"transaction_date\": \"2026-09-17\",
    \"type\": \"expense\",
    \"category_id\": $CATEGORY_ID,
    \"amount\": \"43500.00\",
    \"title\": \"Lunch\",
    \"client_request_id\": 1001
  }"
```

`201 Created` response:

```json
{
  "data": {
    "id": 1,
    "user_id": 1,
    "category_id": 1,
    "category_name": "Food",
    "transaction_date": "2026-09-17",
    "type": "expense",
    "amount": "43500.00",
    "title": "Lunch",
    "client_request_id": 1001,
    "created_by": 1,
    "updated_by": 1,
    "isDelete": false,
    "version": 1,
    "created_at": "2026-09-17T05:00:00Z",
    "updated_at": "2026-09-17T05:00:00Z"
  }
}
```

### List and paginate transactions

```bash
curl -i "$BASE_URL/api/v1/transactions?start_date=2026-09-01&end_date=2026-09-30&type=expense&category_id=$CATEGORY_ID&limit=30" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

`200` response:

```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "category_id": 1,
      "category_name": "Food",
      "transaction_date": "2026-09-17",
      "type": "expense",
      "amount": "43500.00",
      "title": "Lunch",
      "client_request_id": 1001,
      "created_by": 1,
      "updated_by": 1,
      "isDelete": false,
      "version": 1,
      "created_at": "2026-09-17T05:00:00Z",
      "updated_at": "2026-09-17T05:00:00Z"
    }
  ],
  "page": {"next_cursor": null}
}
```

When `next_cursor` is non-null, send it back unchanged:

```bash
curl -i --get "$BASE_URL/api/v1/transactions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  --data-urlencode 'start_date=2026-09-01' \
  --data-urlencode 'end_date=2026-09-30' \
  --data-urlencode 'limit=30' \
  --data-urlencode 'cursor=<next_cursor>'
```

### Read and update a transaction

```bash
curl -i "$BASE_URL/api/v1/transactions/$TRANSACTION_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i -X PATCH "$BASE_URL/api/v1/transactions/$TRANSACTION_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{
    \"transaction_date\": \"2026-09-17\",
    \"type\": \"expense\",
    \"category_id\": $CATEGORY_ID,
    \"amount\": \"50000.00\",
    \"title\": \"Lunch and coffee\",
    \"version\": 1
  }"
```

The `200` update response returns `version: 2` and records the actor in `updated_by`.

### Soft-delete, inspect Trash, and restore a transaction

```bash
curl -i -X DELETE "$BASE_URL/api/v1/transactions/$TRANSACTION_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'If-Match: "2"'

curl -i "$BASE_URL/api/v1/transactions?start_date=2026-09-01&end_date=2026-09-30&isDelete=true" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i "$BASE_URL/api/v1/transactions/$TRANSACTION_ID?isDelete=true" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i -X POST "$BASE_URL/api/v1/transactions/$TRANSACTION_ID/restore" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"version":3}'
```

DELETE returns `204`. Restore returns `200`, and the transaction appears in active history and reports again.

## 13. Summaries and reports

### Daily summaries

```bash
curl -i "$BASE_URL/api/v1/daily-summaries?start_date=2026-09-01&end_date=2026-09-30&limit=30" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

`200` response:

```json
{
  "data": [
    {
      "date": "2026-09-17",
      "income": "0.00",
      "expense": "50000.00",
      "difference": "-50000.00"
    }
  ],
  "page": {"next_cursor": null}
}
```

### Report summary

Available presets are `last_7_days` and `last_30_days`:

```bash
curl -i "$BASE_URL/api/v1/reports/summary?range=last_30_days" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

A custom range requires both dates:

```bash
curl -i "$BASE_URL/api/v1/reports/summary?range=custom&start_date=2026-09-01&end_date=2026-09-30" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Example response:

```json
{
  "data": {
    "period": {"start_date":"2026-09-01","end_date":"2026-09-30"},
    "income": "15000000.00",
    "expense": "2500000.00",
    "difference": "12500000.00",
    "top_income_categories": [
      {"category_id":2,"name":"Salary","type":"income","amount":"15000000.00"}
    ],
    "top_expense_categories": [
      {"category_id":1,"name":"Food","type":"expense","amount":"2500000.00"}
    ]
  }
}
```

### Report breakdown

`group_by` accepts `week`, `month`, or `category`:

```bash
curl -i "$BASE_URL/api/v1/reports/breakdown?range=custom&start_date=2026-09-01&end_date=2026-09-30&group_by=week" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

The response contains `period`, `group_by`, a `periods` array, and a `categories` array. Time grouping populates `periods`; category grouping populates `categories`.

## 14. Administrator endpoints

Log in with the account created by the administrator CLI, then set the administrator token and target user ID:

```bash
export ADMIN_ACCESS_TOKEN='<admin-access-token>'
export USER_ID=1
```

Administrator endpoints use the same request bodies and response schemas as personal endpoints. The `{user_id}` path segment explicitly selects the owner. Administrator actions are recorded in `admin_access_events`.

List a target user's categories and transactions:

```bash
curl -i "$BASE_URL/api/v1/admin/users/$USER_ID/categories?isDelete=false" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"

curl -i "$BASE_URL/api/v1/admin/users/$USER_ID/transactions?start_date=2026-09-01&end_date=2026-09-30" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

Create a category and transaction for the target:

```bash
curl -i -X POST "$BASE_URL/api/v1/admin/users/$USER_ID/categories" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Annual bonus","type":"income"}'

curl -i -X POST "$BASE_URL/api/v1/admin/users/$USER_ID/transactions" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"transaction_date\":\"2026-09-17\",\"type\":\"income\",\"category_id\":$CATEGORY_ID,\"amount\":\"1000000.00\",\"title\":\"Administrator correction\",\"client_request_id\":2001}"
```

Read target reports:

```bash
curl -i "$BASE_URL/api/v1/admin/users/$USER_ID/reports/summary?range=last_30_days" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"

curl -i "$BASE_URL/api/v1/admin/users/$USER_ID/reports/breakdown?range=last_30_days&group_by=category" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

Complete administrator route mapping:

| Operation | Endpoint |
|---|---|
| List/create categories | `GET/POST /api/v1/admin/users/{user_id}/categories` |
| Get/update/delete a category | `GET/PATCH/DELETE /api/v1/admin/users/{user_id}/categories/{id}` |
| Restore a category | `POST /api/v1/admin/users/{user_id}/categories/{id}/restore` |
| List/create transactions | `GET/POST /api/v1/admin/users/{user_id}/transactions` |
| Get/update/delete a transaction | `GET/PATCH/DELETE /api/v1/admin/users/{user_id}/transactions/{id}` |
| Restore a transaction | `POST /api/v1/admin/users/{user_id}/transactions/{id}/restore` |
| Daily summaries | `GET /api/v1/admin/users/{user_id}/daily-summaries` |
| Report summary | `GET /api/v1/admin/users/{user_id}/reports/summary` |
| Report breakdown | `GET /api/v1/admin/users/{user_id}/reports/breakdown` |

A regular user's token receives `403 Forbidden`:

```json
{"error":{"code":"FORBIDDEN","message":"Forbidden."}}
```

## 15. Endpoint index

| Method | Path | Authentication |
|---|---|---|
| `GET` | `/health/live` | None |
| `GET` | `/health/ready` | None |
| `POST` | `/api/v1/auth/register` | None |
| `POST` | `/api/v1/auth/login` | None |
| `POST` | `/api/v1/auth/refresh` | Refresh token/cookie |
| `POST` | `/api/v1/auth/logout` | Refresh token/cookie |
| `GET/PATCH/DELETE` | `/api/v1/me` | Bearer |
| `GET/POST` | `/api/v1/categories` | Bearer |
| `GET/PATCH/DELETE` | `/api/v1/categories/{id}` | Bearer |
| `POST` | `/api/v1/categories/{id}/restore` | Bearer |
| `GET/POST` | `/api/v1/transactions` | Bearer |
| `GET/PATCH/DELETE` | `/api/v1/transactions/{id}` | Bearer |
| `POST` | `/api/v1/transactions/{id}/restore` | Bearer |
| `GET` | `/api/v1/daily-summaries` | Bearer |
| `GET` | `/api/v1/reports/summary` | Bearer |
| `GET` | `/api/v1/reports/breakdown` | Bearer |
| Various | `/api/v1/admin/users/{user_id}/...` | Administrator bearer token |

## 16. Code generation and quality checks

After changing migrations or SQL queries:

```bash
make sqlc-generate
make sqlc-vet
```

After changing [api/openapi.yaml](api/openapi.yaml):

```bash
make oapi-generate
```

Run backend checks:

```bash
make fmt
make test
make test-race
make vet
make build
make check
```

Integration tests create temporary schemas in a test database:

```bash
export TEST_DATABASE_URL='postgres://monelog_owner:change-owner-password@127.0.0.1:5432/monelog_test?sslmode=disable'
make test-integration
```

Never point `TEST_DATABASE_URL` at production.

## 17. Important backend directories

```text
api/                    OpenAPI contract and generator configuration
cmd/api/                HTTP API entrypoint
cmd/admin/              initial-administrator bootstrap CLI
db/migrations/          up/down DDL migrations
db/queries/             sqlc query sources
db/roles/               runtime role grants
internal/api/           generated OpenAPI code
internal/auth/          JWT, password hashing, refresh secrets
internal/handlers/      HTTP transport
internal/middleware/    authentication, CORS, logging, rate limiting
internal/repository/    PostgreSQL access and generated sqlc code
internal/service/       business rules
```

Additional backend documentation is available under [docs-en](docs-en), especially [database](docs-en/database.md), [API](docs-en/api.md), [architecture](docs-en/architecture.md), and [access control](docs-en/access-control.md).
