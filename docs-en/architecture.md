# Architecture
Draft v0.3 • 14 September 2026
Planning only; versions will be checked against official compatibility documentation and pinned during setup.

## Components and boundaries
monelog-api uses Go, Echo, PostgreSQL, sqlc and OpenAPI/oapi-codegen.
monelog-app uses Vue 3 + JavaScript for the browser and Capacitor Android/iOS. Backend completes before frontend work.
Use one API and later a worker process for export/backup jobs. Redis and microservices are unnecessary for the MVP design.
Browser deployment should proxy /api on the same origin; native clients call the HTTPS API directly. Server remains authoritative; mobile packaging does not imply offline synchronization.

Middleware authenticates the actor and current account state → handler validates input → service authorizes action and owner → repository executes explicitly scoped queries.
Personal scope: actor=owner. Admin scope: actor remains the signed-in admin; owner comes from the validated target path. Admin scope permits create/read/update/delete/restore and later export/template/backup operations.
The same domain rules apply to both scopes. Keep owner predicates, category/type constraints and optimistic versions on every mutation. Never authorize solely from frontend controls or JWT role claims.

## Project structure

The planned canonical structure for monelog-api is:

```
monelog-api/
├── cmd/
│   ├── api/
│   │   └── main.go              # HTTP composition and server
│   ├── worker/
│   │   └── main.go              # Later export/backup jobs
│   └── admin/
│       └── main.go              # Admin bootstrap/recovery
├── internal/
│   ├── api/                     # Generated OpenAPI code
│   ├── config/                  # Typed, validated configuration
│   ├── handlers/                # Personal and admin HTTP adapters
│   ├── middleware/              # Auth, role checks, safe logging, recovery
│   ├── service/                 # Business rules, scope, authorization, jobs
│   └── repository/
│       ├── postgresql/          # Handwritten PostgreSQL connection/pool
│       └── sqlc/                # Generated SQL queries
├── db/
│   ├── migrations/              # Versioned schema migrations
│   └── queries/                 # sqlc query sources
├── api/
│   └── openapi.yaml             # Canonical API contract
├── tests/
│   └── integration/             # PostgreSQL/HTTP tests
├── docs/                        # Canonical documentation
├── go.mod
└── go.sum
```

Business logic belongs in internal/service; internal/handlers should only translate HTTP requests to service calls. If implementation uses script/sqlc or different directory names, update the plan and documentation after the structure decision is made.

## Proposed directories
| Path | Responsibility |
| --- | --- |
| cmd/api/main.go | HTTP composition/startup/shutdown |
| cmd/worker/main.go | Later background jobs |
| cmd/admin/main.go | Initial admin bootstrap/recovery for a trusted server operator |
| internal/config | Typed validated configuration |
| internal/handlers | Personal and admin HTTP adapters |
| internal/middleware | Authentication, current account/role checks, safe logging, recovery, throttling |
| internal/service | Action authorization, scope, financial rules and job lifecycle |
| internal/repository | Handwritten repository boundary |
| internal/repository/postgresql | PostgreSQL connection and pool |
| internal/repository/sqlc | Generated SQL queries |
| internal/api | Generated OpenAPI code |
| db/migrations, db/queries | Versioned schema and sqlc source |
| api/openapi.yaml | Canonical machine-readable contract |
| tests/integration | Real PostgreSQL/HTTP tests |
| docs | Canonical planning pack |

Frontend directories: src/views, components, services, stores, router and utils, with focused tests. Admin Management reuses validated transaction forms with an explicit owner context.

## Authentication and roles
Short-lived access JWTs and random rotating refresh secrets hashed in revocable sessions.
Browser: in-memory access token; HttpOnly/Secure refresh cookie with appropriate SameSite, CSRF and origin checks.
Native: vetted OS-backed secure refresh storage; access token in memory. Never localStorage for refresh secrets.
Validate JWT algorithm, issuer, audience and expiry; sub is always the actor.
Public registration defaults to user and rejects role/owner overrides. Admin account creation/role updates use protected endpoints and current-role checks. GET /me returns current role for UI.
Bootstrap first admin through an explicit operator command; no hardcoded account or first-user auto-promotion.
Current account is_delete=true denies all protected requests, including requests with old access JWTs. Role changes/account deletion revoke refresh sessions; committed demotion blocks new admin requests.

## Transactions, authorization and audit
Read requests check current actor state/role before target lookup. All business mutations open a DB transaction, lock involved actor/target account rows in stable UUID order, recheck authorization and target activity, then validate/update the resource and persist audit before commit.
Account/role changes participate in the same locking rules; concurrent committed deletion/demotion cannot be bypassed by a stale mutation check.
Use named methods such as CreateTransaction(scope), SoftDeleteTransaction(scope,version), RestoreTransaction(scope,version); no generic client-controlled owner or role assignment.
Admin actions log actor versus owner, operation, safe resource/version metadata and outcome. Admin mutation plus audit insert is atomic; audit failure rolls back. Admin reads require audit persistence before responding.
Audit records are append-only. Admins can inspect logs through a protected endpoint; raw credentials/financial payloads are not returned or logged.

## Soft deletion
API field isDelete maps to Go IsDelete and SQL is_delete BOOLEAN NOT NULL DEFAULT FALSE.
Users/categories/transactions/templates use the flag; it replaces the earlier transaction timestamp deletion signal and category archive proposal.
Create defaults false. Delete sets true and updates version/audit time. Restore sets false after ownership, version and related-category checks.
Active queries explicitly filter transactions.is_delete=false. Trash queries explicitly request true; owner/admin authorization remains identical.
A deleted category is hidden from selectors but historical category labels remain available through scoped joins. Do not filter historical transactions out because a joined category is deleted.
Deleted user disables login/jobs without cascading child-row flags; admin may inspect retained data and restore the account. No application hard delete for these entities.

## Jobs and provider operations
Persist owner_user_id, requested_by, request_mode and immutable filters in export/backup jobs.
A regular user operates only their own owner scope; a current admin can manage any selected owner's jobs and downloads.
Before execution, recheck owner/requester account state and admin role when request_mode=admin. Canceled or unauthorized jobs make no provider call.
Recurring schedules retain owner, authorizing actor and mode; paused schedules need explicit authorized resume.
External calls follow committed job+audit creation; keep them outside DB locks and track remote success/failure/retries explicitly.
Drive operations require the selected owner's valid provider connection/consent. Admin permissions allow management but do not manufacture OAuth authorization.
Backup/restore preserves isDelete flags, validates selected-owner remapping and excludes auth roles/credentials/audit histories. Data restore cannot promote a user; dedicated admin role operations can.

## Consistency and clients
Money: NUMERIC(14,2), exact Go decimals/minor units, JSON decimal strings. No floating point or stored balance column.
Record version guards edits/deletes/restores. Create idempotency is scoped by authorized owner and client_request_id; record actual actor separately.
Reports/exports share aggregation and owner/active-row filters.
Bind cursors and cache/request keys to actor, mode, owner, endpoint, filters and deletion state.
All authenticated responses use Cache-Control: no-store. Do not persist other users' data in offline caches.
Admin UI labels the selected owner and enables management. Save/discard before switching an unsaved form; submitted operations stay bound to the original target. Discard late responses after target/session changes.

## Operations and validation
Separate development/staging/production, migration rollout and reviewed recovery plans.
Database disaster-recovery backups and restore drills are separate from personal Drive backups.
Structured logs redact finance payloads, passwords, tokens and provider secrets; health/live and health/ready disclose no configuration.
CI: formatting, vet/lint, meaningful tests, generated-code drift, frontend build and dependency checks.
Commit sqlc/OpenAPI generated code; exclude .env, signing keys, dumps and provisioning secrets.
Validate current toolchain requirements and macOS/Xcode/signing availability in the mobile milestone.
