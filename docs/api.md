# API draft
Prefix /api/v1. Protected routes use Bearer access token. JSON field names snake_case.
This is a design contract, not a complete OpenAPI file; ISSUE-004 must produce validated api/openapi.yaml before domain handlers.
Money JSON strings match ^[0-9]+\\.[0-9]{2}$; signed difference may be negative. Dates YYYY-MM-DD; timestamps UTC RFC3339.
UUIDs are opaque. Client never submits user_id. Unknown write fields rejected.

## Routes
| Method | Path | Input / result |
| --- | --- | --- |
| POST | /auth/register | email,password,timezone → 201 user; then log in |
| POST | /auth/login | email,password,client_type:web/native → access token + refresh transport |
| POST | /auth/refresh | browser cookie+CSRF or native refresh token → rotated session |
| POST | /auth/logout | refresh transport → revoke family, 204 |
| GET | /me | profile |
| PATCH | /me | timezone only → profile |
| GET | /categories | type optional; include_archived=false; paginated items |
| POST | /categories | name,type → 201 category |
| PATCH | /categories/{id} | name and/or archived boolean → category; type immutable |
| GET | /transactions | start_date,end_date,type?,category_id?,limit,cursor → items,next_cursor |
| POST | /transactions | create object below → 201; identical replay 200 |
| GET | /transactions/{id} | owned active transaction |
| PATCH | /transactions/{id} | mutable fields plus required version → updated transaction |
| DELETE | /transactions/{id} | required If-Match: "version" → 204 |
| GET | /daily-summaries | start_date,end_date,limit,cursor → dated income/expense/difference |
| GET | /reports/summary | start_date,end_date → totals and top categories |
| GET | /reports/breakdown | start_date,end_date,group_by:week/month/category → groups |
| POST | /exports | start_date,end_date,type?,category_id?,format:xlsx/pdf → 202 job |
| GET | /exports/{id} | owned job status queued/running/succeeded/failed |
| GET | /exports/{id}/download | owned completed job → file; 409 not ready; 410 expired |

Export routes are Release 1, not core MVP. Later templates/Drive/offline APIs are intentionally not specified until their design issues; no placeholder public endpoint promises.
Health endpoints /health/live and /health/ready outside /api/v1; readiness checks DB without exposing configuration.

## Create example
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
Success: {"data":{...create fields...,"id":"UUID","version":1,"created_at":"RFC3339","updated_at":"RFC3339"}}.
Response fields never include request_hash, password_hash or refresh hashes.
PATCH allows date,type,category_id,amount,title; require type/category consistency in final merged record.
Refresh transport: web refresh secret in HttpOnly cookie, native refresh secret in response/body; server enforces origin/CSRF policy for web and does not treat client_type as authorization.
Access response includes access_token, token_type:"Bearer", expires_in; native additionally refresh_token.
Password policy and session TTLs must be finalized in auth plan; no secrets returned in logs.

## Lists and reports
List envelope: {"data":[],"page":{"next_cursor":null}}; limit default 30, max 100.
General transactions order transaction_date DESC,created_at DESC,id DESC; bind cursor to original filters.
Daily summaries order date DESC; category list order name,id.
Report totals: {"data":{"start_date":"2026-09-01","end_date":"2026-09-08","income":"0.00","expense":"769500.00","difference":"-769500.00","top_income_categories":[],"top_expense_categories":[]}}.
Top categories: up to 5 per type, amount DESC then category_id; each {category_id,name,amount}.
Breakdown groups contain period_start for week/month or category_id/name/type for category, plus income/expense/difference. Empty results [].
Server validates both dates, start<=end, maximum 366 days per reporting/export request (proposed guardrail); lists also require an explicit bounded date range.
The UI converts last-7/30 presets to inclusive dates using profile timezone. Exports reuse transaction-list filters and contain all matching records, not only current page.
Async export metadata requires a new export_jobs migration in ISSUE-009; artifacts expire after proposed 24h, downloads reauthorize every time.

## Errors
{"error":{"code":"VALIDATION_ERROR","message":"Check the highlighted fields.","fields":{"amount":"Must be greater than zero"},"request_id":"opaque-id"}}
400 malformed input/cursor/date; 401 invalid/expired authentication; 403 CSRF/origin denial; 404 missing or foreign-owned entity; 409 stale version/idempotency conflict/category unavailable; 422 semantically invalid fields; 429 throttled; 500 sanitized internal failure.
No stack traces, SQL text or cross-user existence leaks. OpenAPI must define every request, response, nullability, header and auth scheme; examples must validate against it.

