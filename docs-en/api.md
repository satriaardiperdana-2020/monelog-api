# API design
Draft v0.3 • 14 September 2026
Base /api/v1. Protected operations use Bearer access JWTs and server-side authorization.
This is a design guide; Issue 004 produces validated api/openapi.yaml and generated interfaces before domain implementation.
JSON names are snake_case except the explicitly requested isDelete boolean. Go IsDelete maps to SQL is_delete.
Money uses exact decimal strings, e.g. "43500.00"; unsigned amounts match ^[0-9]+\.[0-9]{2}$, while difference may be negative.
Dates YYYY-MM-DD, timestamps UTC RFC3339, IDs opaque UUIDs.

## Authorization
Personal paths always scope owner to actor. Admin paths require current active admin and scope data to the explicit target user.
Admin CRUD, exports, templates and backups are permitted. A regular user cannot use these admin paths.
Unknown body fields/owner overrides are rejected; admin targets come from the path, not a body override.
Public registration defaults to user. role is writable only through protected admin account operations.
Current deleted actor gets 401; current non-admin on an admin path gets 403 before target lookup.
For a current admin: malformed target gets 400; nonexistent/wrong-target resource gets 404. Reading retained data of a deleted target is allowed; financial writes require account restoration first.
All authenticated responses/errors use Cache-Control: no-store. Admin mutations require atomic data+audit persistence; storage failure returns 503 with no committed mutation.

## Account and auth routes
| Method | Path | Input / result |
| --- | --- | --- |
| POST | /auth/register | email,password,timezone → 201 regular user; role/isDelete overrides rejected |
| POST | /auth/login | email,password,client_type:web/native → access token and refresh transport |
| POST | /auth/refresh | browser refresh cookie+CSRF or native refresh secret → rotated session |
| POST | /auth/logout | refresh transport → revoke family, 204 |
| GET | /me | own id,email,role,timezone,currency,isDelete,version |
| PATCH | /me | timezone,version → updated own profile |
| DELETE | /me | If-Match version → soft-delete own account and revoke sessions, 204 |
| GET | /admin/users | q?,isDelete=false,limit,cursor → paginated user directory |
| POST | /admin/users | email,password,timezone,role:user/admin → 201 new active user; omitted role defaults user |
| GET | /admin/users/{user_id} | isDelete=false or true → selected user's metadata |
| PATCH | /admin/users/{user_id} | timezone and/or role, required version → updated active user |
| DELETE | /admin/users/{user_id} | If-Match version → soft-delete selected user and revoke sessions, 204 |
| POST | /admin/users/{user_id}/restore | version → restored user, isDelete=false |
| GET | /admin/audit-events | actor_user_id?,target_user_id?,action?,start_time,end_time,limit,cursor → safe audit metadata |

Admin directory DTO: id,email,role,timezone,currency,isDelete,version. Email is immutable in this initial profile contract; account identity/auth changes need their own validated flow.
Directory q is a trimmed literal email substring, 1–100 chars if present; parameterize and escape wildcard semantics. Ordering email ASC,id ASC; limit default 30, max 100.
Audit list uses created_at DESC,id DESC and a maximum 366-day time range; contains no passwords, provider tokens or financial payloads. Audit events are not business rows available for editing.
Session/JWT sub stays actor, never selected target. Role changes revoke refresh sessions and take effect for the next admin request.
An account cannot use self routes after deletion; an active admin restores it.
Public registration and admin creation hash supplied passwords; never return or log them.

## Personal finance routes
| Method | Path | Input / result |
| --- | --- | --- |
| GET | /categories | type?,isDelete=false,limit,cursor → own categories |
| POST | /categories | name,type → 201 active category |
| GET | /categories/{id} | isDelete=false or true → own category |
| PATCH | /categories/{id} | name,version → updated active category; type immutable |
| DELETE | /categories/{id} | If-Match version → isDelete=true, 204 |
| POST | /categories/{id}/restore | version → isDelete=false category |
| GET | /transactions | start_date,end_date,type?,category_id?,isDelete=false,limit,cursor → own transactions |
| POST | /transactions | create object → 201; identical idempotent replay 200 |
| GET | /transactions/{id} | isDelete=false or true → own transaction |
| PATCH | /transactions/{id} | date/type/category/amount/title fields plus version → updated active transaction |
| DELETE | /transactions/{id} | If-Match version → isDelete=true, 204 |
| POST | /transactions/{id}/restore | version → isDelete=false transaction |
| GET | /daily-summaries | start_date,end_date,limit,cursor → active daily totals |
| GET | /reports/summary | start_date,end_date → active income/expense/difference, top categories |
| GET | /reports/breakdown | start_date,end_date,group_by:week/month/category → active grouped totals |

Admin supports the same methods, payloads and statuses for every finance route above by replacing the leading slash with /admin/users/{user_id}/.
For example:
- POST /admin/users/{user_id}/transactions creates a record owned by that user.
- PATCH /admin/users/{user_id}/transactions/{id} edits that user's record.
- DELETE /admin/users/{user_id}/transactions/{id} sets that record's flag to true.
- POST /admin/users/{user_id}/transactions/{id}/restore sets it back to false.
- GET /admin/users/{user_id}/reports/summary reports that user's active rows.
Issue 004 must expand the mapping into distinct OpenAPI operations for every method/path; Issue 013 implements the admin adapters after core domain services.

## Soft-delete contract
isDelete is a required response field on user/category/transaction/template DTOs and accepts only JSON booleans. New rows default false server-side.
Ordinary create/PATCH bodies cannot include isDelete. DELETE and /restore are authorized lifecycle transitions.
For resource lists/detail, query isDelete accepts only false or true, default false. true selects Trash, including detail retrieval needed for restore.
Normal detail for a deleted record returns 404. Authorized Trash detail returns its current version.
DELETE requires If-Match: "3" for version 3. Successful deletion increments version to 4 and returns 204. Fetch Trash detail for the current version before restoration.
POST /restore body {"version":4} returns the active resource with version 5. No manual setting of isDelete through generic updates.
Missing version → 400; stale/already-deleted/already-restored lifecycle operation → 409; foreign/missing record → 404.
The flag does not change authorization. No physical business-row deletion endpoint is provided.
Reports/exports/daily summaries always use active transactions; reject isDelete on those routes with 400 rather than including Trash in financial totals.
A deleted category's labels still appear on historical transactions; active category required for new/edit/restored records.

## Example: admin creates an owner's transaction
POST /admin/users/6b3ab04b-22cb-4777-b3ec-15f3247b123d/transactions
```json
{
  "transaction_date": "2026-09-08",
  "type": "expense",
  "category_id": "bda088ec-2694-473c-997d-cb93161456f1",
  "amount": "43500.00",
  "title": "Groceries",
  "client_request_id": "22f60b83-db97-48f7-b591-b403b93c4c12"
}
```
Illustrative 201 response:
```json
{
  "data": {
    "id": "a1e29497-1328-496c-9c29-542815d5c2fd",
    "user_id": "6b3ab04b-22cb-4777-b3ec-15f3247b123d",
    "transaction_date": "2026-09-08",
    "type": "expense",
    "category_id": "bda088ec-2694-473c-997d-cb93161456f1",
    "amount": "43500.00",
    "title": "Groceries",
    "isDelete": false,
    "version": 1
  },
  "scope": {
    "mode": "admin",
    "owner_user_id": "6b3ab04b-22cb-4777-b3ec-15f3247b123d"
  }
}
```
Actual transaction DTO also includes client_request_id, created_by, updated_by, created_at and updated_at. Actor fields come from authentication.
Success resource payloads use data; list payloads add page.next_cursor (null at end).
Selected-target admin responses also include scope.mode="admin" and scope.owner_user_id; personal responses keep their existing envelopes. A 204 has no response body.
No request_hash, password_hash, session hash or provider secrets are included.

## Lists and reports
Default page limit 30, max 100. Active transactions order transaction_date DESC,created_at DESC,id DESC; Trash orders updated_at DESC,id DESC.
Categories order name,id; daily summaries date DESC. Cursors bind actor/mode/owner/endpoint/filters/isDelete; cross-scope reuse returns 400.
Date bounds required for transaction lists/reports/exports: start<=end, maximum 366 inclusive days (proposal). A category filter from another owner returns 404.
Presets use the owner's timezone, including admin target. Empty lists [], empty sums "0.00".
Totals data includes start_date,end_date,income,expense,difference,top_income_categories,top_expense_categories.
Top categories: up to five per type, ordered amount DESC then category_id, each {category_id,name,amount}.
Breakdown returns period_start for week/month or category_id/name/type for category plus totals. Boundary periods are partial after filtering.
Every calculation includes only transactions with is_delete=false. Deleting/restoring a transaction affects owner reports and exports consistently.

## Exports (Issue 009)
Personal /exports paths and equivalent /admin/users/{user_id}/exports paths:
| Method | Suffix | Behavior |
| --- | --- | --- |
| POST | /exports | start_date,end_date,type?,category_id?,format:xlsx/pdf → 202 job |
| GET | /exports/{id} | scoped queued/running/succeeded/failed/canceled status |
| GET | /exports/{id}/download | authorize owner/current admin and stream artifact; 409 not ready, 410 expired |

Jobs store owner_user_id, requested_by and request_mode. Regular users access only jobs owned by themselves; admins can manage any selected owner's jobs.
Workers recheck requester/owner state and admin authority if applicable; filters and owner are immutable.
Exports contain all matching active transactions, not only a page. Proposed artifact expiry 24 hours; downloads always reauthorize.

## Templates and Drive (later contracts)
Issue 011 defines personal /templates and the same /admin/users/{user_id}/templates routes for list/get/create/update/delete/restore/apply-to-form; use isDelete/version and target-compatible categories.
Issue 012 defines personal /drive/connection, /backup-schedules, /backups and /restores plus matching selected-user admin routes.
Admin may initiate/disconnect target Drive connection, change schedules, create/cancel/retry backup jobs, access authorized backups and run validated restores.
OAuth callbacks bind actor, target and request mode in protected state and recheck authorization. A provider-authorized target connection is required.
Persist owner/requester/mode for jobs. Recurring schedules carry authorizing actor/mode; revoked authority or deleted account pauses execution.
Backup/restore data preserves isDelete but excludes roles/password/session/provider secrets/audit records; strict import cannot smuggle role changes.
Snapshot restore preserves valid historical category relationships, including a deleted same-owner category, without changing its flag; ordinary transaction restore still requires an active category.
Full schemas, concurrency and provider-specific states for these later routes are defined before their implementation, not implied to exist now.

## Error contract
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Check the highlighted fields.",
    "fields": {"amount": "Must be greater than zero"},
    "request_id": "opaque-id"
  }
}
```
400 malformed IDs/cursors/booleans/missing versions or disallowed owner/role/isDelete fields; 401 absent/invalid token or deleted actor; 403 insufficient admin role/CSRF denial; 404 missing/out-of-scope row; 409 version/lifecycle/idempotency conflict or inactive target/category; 422 field validation; 429 throttled; 500 sanitized internal failure; 503 unavailable required authorization/audit storage.
405 applies only to unsupported methods, not to supported admin CRUD.
Health /health/live and /health/ready sit outside /api/v1 and expose no data/configuration.
OpenAPI must define each concrete operation, envelope, nullability, version header, lifecycle boolean, auth requirement and error; validate examples and generated interfaces.
