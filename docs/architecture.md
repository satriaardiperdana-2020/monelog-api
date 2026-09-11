# Architecture
Proposed design, not a deployed system. Versions will be selected and pinned during ISSUE-001 after checking official compatibility/security documentation.

## Components
- monelog-app: Vue 3, JavaScript, router, small shared state layer and HTTP client. Same UI source for browser and Capacitor Android/iOS.
- monelog-api: Go + Echo HTTP service, PostgreSQL, sqlc queries, OpenAPI contract with oapi-codegen.
- Separate worker command later for exports/backups; no Redis or microservices needed for MVP.
- Browser deployment should proxy /api on the same origin. Native clients use the HTTPS API directly.
- Server is authoritative in MVP; mobile packaging does not imply offline synchronization.

## Request boundaries
Middleware authenticates and adds user identity → handler validates transport input → service enforces business rules → repository executes sqlc queries.
Handlers never accept owner identity from a request body. Every repository read/write is scoped by authenticated user ID.
Database constraints are the final safety net, including composite category ownership/type relationships.
Use transactions for multi-write operations. Propagate request context/timeouts. Return sanitized errors with request IDs.

## Proposed backend directories
| Path | Responsibility |
| --- | --- |
| cmd/api/main.go | Composition and HTTP startup |
| cmd/worker/main.go | Later scheduled jobs |
| internal/config | Typed configuration and validation |
| internal/handlers | HTTP adapters |
| internal/middleware | Authentication, logging, recovery, rate limits |
| internal/service | Finance and authorization rules |
| internal/repository | Handwritten query adapter |
| internal/repository/sqlc | Generated database code |
| internal/api | Generated OpenAPI types/interfaces |
| db/migrations | Versioned SQL schema |
| db/queries | sqlc source queries |
| api/openapi.yaml | Canonical machine-readable API contract |
| tests/integration | Real PostgreSQL tests |
| docs | This planning pack |

Frontend: src/views, src/components, src/services, src/stores, src/router, src/utils; tests alongside modules or under tests/.
Use service and repository interfaces only where useful for tests; avoid layers that merely rename methods.

## Authentication proposal
Short-lived JWT access tokens; random rotating refresh tokens stored hashed server-side in revocable sessions.
Browser: access token in memory; refresh cookie HttpOnly/Secure with suitable SameSite policy, scoped path, origin checks and CSRF protection.
Native: refresh token in OS-backed secure storage through a separately vetted plugin, not localStorage; bearer access token in memory.
Refresh reuse revokes the session family. JWT validation pins algorithm, issuer, audience and expiry. Logout revokes refresh sessions; existing access tokens may remain valid until their short expiry.
Never place Google OAuth credentials or refresh tokens in the client bundle. Encrypt provider credentials at rest using a managed external key.

## Consistency
Money: NUMERIC(14,2) in PostgreSQL, exact decimal or integer minor-unit representation in Go; JSON money strings.
Transactions use version for optimistic locking; duplicate create protection uses client_request_id per user.
Reports and exports share the same filtering and aggregation service. Cursor pagination uses stable ordering.
No cached balance column. Start with indexed aggregation; optimize only after measuring.

## Deployment and operations
Separate development/staging/production configuration and databases; migration job before new app rollout, tested rollback/forward recovery.
Use database backups with retention and regular restore tests independently of personal Drive exports.
Structured logs redact titles, tokens, passwords, provider credentials and financial payloads. Health/live and health/ready endpoints.
CI runs formatting, vet/lint, tests, regeneration drift check, frontend tests/build and dependency scanning.
Commit generated sqlc/OpenAPI code; regenerate deterministically and review the source plus generated diff.
Do not commit .env, signing keys, database dumps or mobile provisioning assets.
iOS build/sign/test needs a suitable macOS/Xcode environment and signing credentials; confirm availability in mobile milestone.

