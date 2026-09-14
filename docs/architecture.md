# Architecture
Draft v0.2, 14 September 2026. Proposed design, not a deployed system. Versions will be selected and pinned during ISSUE-001 after checking official compatibility/security documentation.

## Components
- monelog-app: Vue 3, JavaScript, router, small shared state layer and HTTP client. Same UI source for browser and Capacitor Android/iOS.
- monelog-api: Go + Echo HTTP service, PostgreSQL, sqlc queries, OpenAPI contract with oapi-codegen.
- Separate worker command later for exports/backups; no Redis or microservices needed for MVP.
- Browser deployment should proxy /api on the same origin. Native clients use the HTTPS API directly.
- Server is authoritative in MVP; mobile packaging does not imply offline synchronization.

## Request boundaries
Middleware authenticates and adds user identity → handler validates transport input → service enforces business rules → repository executes sqlc queries.
Personal handlers never accept owner identity from request bodies or query parameters; their reads and writes use the authenticated actor's user ID.
Explicit /admin/users/{user_id}/... GET handlers authenticate the actor, check the current database role is admin, resolve the target user, and create a read-only scope for that one target.
Financial repository queries always require an explicit owner ID. Personal scope uses actor_user_id; authorized admin read scope uses target_user_id. Write methods only accept the actor's own scope. Never remove the owner filter because the caller is an admin.
Database constraints are the final safety net, including composite category ownership/type relationships.
Use transactions for multi-write operations. Propagate request context/timeouts. Return sanitized errors with request IDs.

## Proposed backend directories
| Path | Responsibility |
| --- | --- |
| cmd/api/main.go | Composition and HTTP startup |
| cmd/worker/main.go | Later scheduled jobs |
| cmd/admin/main.go | Planned operator-only role provisioning command; not an HTTP endpoint |
| internal/config | Typed configuration and validation |
| internal/handlers | HTTP adapters |
| internal/middleware | Authentication, current-role guard for admin routes, logging, recovery, rate limits |
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
Reports and exports share filtering and aggregation logic but receive separately authorized owner scope. Admin report viewing never grants access to another user's export jobs. Cursor pagination uses stable ordering and binds actor, target, route family and filters.
No cached balance column. Start with indexed aggregation; optimize only after measuring.

## Deployment and operations
Separate development/staging/production configuration and databases; migration job before new app rollout, tested rollback/forward recovery.
Use database backups with retention and regular restore tests independently of personal Drive exports.
Structured logs redact titles, tokens, passwords, provider credentials and financial payloads. Health/live and health/ready endpoints.
CI runs formatting, vet/lint, tests, regeneration drift check, frontend tests/build and dependency scanning.
Commit generated sqlc/OpenAPI code; regenerate deterministically and review the source plus generated diff.
Do not commit .env, signing keys, database dumps or mobile provisioning assets.
iOS build/sign/test needs a suitable macOS/Xcode environment and signing credentials; confirm availability in mobile milestone.

## User/admin authorization
Implement [access-control.md](access-control.md) through centralized middleware and service policy. JWT sub always identifies the acting account. Read the current role from PostgreSQL on every admin request; neither a role claim nor frontend state is authoritative.
GET /me returns the current role for UI rendering. Registration sets role=user server-side; unknown role fields are rejected. Normal profile endpoints only update the actor's timezone.
An operator command provisions/revokes an admin role for one explicit existing account through a separate privileged deployment connection; it is unavailable to the HTTP database role. Record operator identity, account, old/new role and time in an operational audit log. Revoke that account's refresh sessions on role changes.
Admin viewing does not impersonate users and cannot reach other users' mutations, credentials, exports, templates or Drive resources. Database errors or unknown roles fail closed.
Each successful admin directory/profile/financial read writes a minimal admin_access_events record before responding; audit persistence failure returns 503 without financial data. Failed authenticated admin-route attempts are recorded when the audit store is available.
All authenticated API responses, including errors, use Cache-Control: no-store. Browser/native state keys include actor, mode, target and filters; clear data and cancel/invalidate requests on selection change, mode change, logout or role denial. Do not persist selected-user financial responses in localStorage, service-worker caches or offline storage.
Admin data loading uses selected-user timezone/currency; preserve date-only transactions as stored.
An authorization decision already made for an in-flight request may complete; every subsequent request after a committed demotion must see the current role.
