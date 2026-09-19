# Database design
Draft v0.3 • 14 September 2026
Logical design; executable migrations/queries arrive in Issue 002. Full admin management is Issue 013.

## Common conventions
Primary keys use BIGSERIAL and relational identifiers use BIGINT. created_at/updated_at use TIMESTAMPTZ; update timestamps on mutations.
Soft-deletable entities have is_delete BOOLEAN NOT NULL DEFAULT FALSE and version INTEGER NOT NULL DEFAULT 1 CHECK(version>0).
JSON exposes isDelete (exact spelling) and Go uses IsDelete. Other JSON fields remain snake_case.
NULL is not a deletion state. Rows are active only when is_delete=false.

## Tables
| Table | Main fields and constraints |
| --- | --- |
| users | id PK; email normalized lowercase UNIQUE NOT NULL; password_hash NOT NULL; role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('user','admin')); timezone NOT NULL default Asia/Jakarta; currency NOT NULL default IDR CHECK(currency='IDR'); is_delete; version; created_at; updated_at |
| categories | id PK; user_id FK users NOT NULL; type NOT NULL CHECK IN ('income','expense'); name VARCHAR(80) NOT NULL; is_delete; version; created_at; updated_at; UNIQUE(id,user_id,type) |
| transactions | id BIGSERIAL PK; user_id BIGINT FK users NOT NULL; category_id BIGINT NOT NULL; type NOT NULL CHECK IN ('income','expense'); amount NUMERIC(14,2) NOT NULL CHECK(amount>0); transaction_date DATE NOT NULL; title VARCHAR(200) NOT NULL; client_request_id BIGINT NOT NULL; request_hash TEXT NOT NULL; created_by/updated_by BIGINT NOT NULL FK users; is_delete; version; created_at; updated_at; UNIQUE(user_id,client_request_id) |
| refresh_sessions | id BIGSERIAL PK; user_id BIGINT FK users; family_id BIGINT; token_hash UNIQUE NOT NULL; expires_at NOT NULL; revoked_at nullable; replaced_by BIGINT nullable self FK; created_at |
| admin_access_events | id BIGSERIAL PK; actor_user_id BIGINT NOT NULL FK users; target_user_id BIGINT nullable FK users; resource_type TEXT NOT NULL; resource_id BIGINT nullable; action TEXT NOT NULL; outcome TEXT NOT NULL; request_id TEXT NOT NULL; safe_metadata JSONB NOT NULL default '{}'; created_at TIMESTAMPTZ NOT NULL default current_timestamp |
| transaction_templates (later) | id PK; user_id FK users; category_id; type; name VARCHAR(80); amount NUMERIC(14,2) CHECK(amount>0); title; is_delete; version; created_at; updated_at |
| export_jobs (later, 009) | id PK; owner_user_id FK users; requested_by FK users; request_mode personal/admin; immutable filters/format; status; artifact locator; expires_at; created_at/updated_at |
| drive_connections (later, 012) | id PK; user_id UNIQUE FK users; encrypted_refresh_token; provider_account_label; revoked_at nullable; created_at/updated_at |
| backup_schedules (later, 012) | id PK; owner_user_id FK users; authorized_by FK users; request_mode personal/admin; schedule/timezone; enabled/paused state; version; created_at/updated_at |
| backup_jobs (later, 012) | id PK; owner_user_id FK users; requested_by FK users; request_mode personal/admin; schedule_id nullable; scheduled_for; status; attempt; provider_file_id/checksum/error_code nullable; started_at/completed_at; UNIQUE(schedule_id,scheduled_for) for scheduled runs |

Future job tables need full nullability/status definitions in their feature issues. Common flag/version definitions above apply to users, categories, transactions and templates.
CHECK(length(btrim(name/title))>0) where applicable. Amount/title limits remain in requirements.md.
Transactions enforce FOREIGN KEY(category_id,user_id,type) REFERENCES categories(id,user_id,type); templates use the same constraint.
Roles describe application authority, not database superuser privileges. Admin edits retain selected owner's user_id; created_by/updated_by attribute actual actor.
Email and per-owner/type category-name uniqueness includes deleted rows to avoid ambiguous recovery; restore an existing row instead of recreating its identity.
Foreign keys restrict physical deletion; no cascading erase of financial/audit history.

## Indexes
- categories: UNIQUE(user_id,type,lower(name)), including deleted rows.
- transactions: (user_id,transaction_date DESC,id DESC) WHERE is_delete=false for normal lists/totals.
- transactions: (user_id,category_id,transaction_date) WHERE is_delete=false for category reports.
- transactions: (user_id,updated_at DESC,id DESC) WHERE is_delete=true for Trash.
- refresh_sessions: (user_id,family_id) and expires_at for cleanup.
- admin_access_events: (actor_user_id,created_at DESC) and (target_user_id,created_at DESC).
- jobs: owner/status and runnable status/time indexes, defined in 009/012.
Do not add a standalone low-selectivity boolean index. Measure EXPLAIN on representative owner-scoped queries; no title/amount index without a demonstrated query.

## Authorization and concurrency
Service creates explicit scope {actor_user_id,owner_user_id,mode}. Personal owner=actor; admin owner=validated path target.
All scoped reads/writes include owner ID, even for admins. A related record from another target yields 404.
Before mutation, lock involved account rows in stable ID order with FOR UPDATE, recheck actor active/current admin role as needed and target account state, then mutate resource under required version.
Account delete/restore/role change follows the same locking protocol. Do not hold these locks while calling Google or rendering files.
Domain rules allow a regular owner or current admin; repository scope cannot come straight from a request body.

## Create, update, delete and restore
Create inserts is_delete=false, version=1. Transactions record actual actor in created_by and updated_by.
Hash canonical create fields and enforce UNIQUE(owner,client_request_id). Same request replays original result; changed fields or replay against a deleted result returns 409. Retain idempotency key after deletion.
Editing an active transaction uses WHERE id=$id AND user_id=$owner AND version=$expected AND is_delete=false; increments version and updated_by/updated_at.
Soft delete uses the same predicate, SET is_delete=true, version=version+1, updated_by=$actor, updated_at=now(). No DELETE FROM business tables.
Restore requires is_delete=true and the expected version; sets false and increments version. Validate target account and category active/type/ownership before restoring.
Owned but stale/already-deleted/already-restored lifecycle attempts return 409. Truly missing or wrong-owner rows return 404.
Users/categories/templates also use flag+version; attribution for their admin mutations is in admin_access_events.
Ordinary PATCH/POST DTOs reject isDelete; delete/restore routes are the only lifecycle writers.

## Category/account behavior
Category soft deletion hides it from selectors without deleting transactions or changing their sums. Historical joins keep same-owner category labels even if category is_delete=true.
New/edit/restored transactions or templates must use an active compatible category; restore the category first or choose another active category during an edit. Existing historical records remain readable and deletable.
User soft deletion sets the account flag and revokes sessions in the same transaction, pauses schedules and blocks queued jobs at execution. Child flags do not change.
Only an admin can restore a deleted account, because deleted users cannot authenticate. Restore does not resurrect revoked sessions or resume schedules automatically.
A trusted operator bootstrap/recovery command remains available; there is no automatic role promotion.

## Reads, reports and exports
Active transaction detail/list: WHERE user_id=$owner AND is_delete=false. Trash uses the same owner predicate with is_delete=true and updated_at DESC,id DESC.
Reports/exports always filter active transactions regardless of Trash UI state. Use inclusive transaction_date >= start AND <= end.
SUM income/expense separately with COALESCE(...,0); difference is computed. Filter dates before Monday-week/month grouping; boundary groups are partial.
Daily history returns only populated dates; today's summary is computed separately and may be zero.
Do not add a deleted-category predicate to the transaction join that would erase historical amounts.
Cursors bind actor, owner, mode, endpoint, date/category/type filters and deletion state.

## Audit and roles
Admin access events now cover read/create/update/delete/restore, exports/jobs, account and role operations. resource_type/action/outcome use a service whitelist defined with the contract; safe_metadata may contain versions/changed-field names and old/new role, never finance payloads/secrets.
Admin writes and success audit inserts commit atomically. Reads persist success auditing before returning. Failed authorized attempts log denial/conflict where possible without masking the primary error.
Runtime can insert and read audit records for the protected admin log service, but cannot update/delete them. It is not the migration owner.
Registration SQL excludes role and uses the default. Personal profile SQL only updates allowed fields. Admin account/role SQL is invoked exclusively after current-role service authorization and audit setup; update is now permitted through protected admin operations.
Role updates revoke target refresh sessions. Locking/rechecks prevent a transaction authorized before concurrent demotion from committing after a contradictory role change without serialization.

## Jobs, backups and restoration
Persist owner and requester separately. Admin role does not make requester the financial owner.
Owner can access own jobs; any current admin can manage jobs under their verified target route, including jobs created by another admin.
Workers revalidate requester/owner and request_mode. Admin jobs require requester still admin. Invalid jobs are canceled/paused before execution; IDs/filters never change on retries.
Backup data includes isDelete and row relationships; reports exclude deleted rows but backups preserve them for recovery.
Exclude roles/passwords/session/provider secrets and audit records from personal-data backups. Imports validate strict schemas, reject privilege fields and remap only into the authorized owner, preserving deletion flags.
Snapshot restore may preserve an active historical transaction referencing a deleted same-owner category; validate its owner/type foreign key without silently dropping the transaction or reactivating the category. Individual transaction restore and new edits still require an active category.
Version/restore conflict policy and provider retention are finalized in Issue 012.

## Migration and verification
Fresh install: users → categories → transactions → sessions → admin_access_events; later templates/jobs as their issues land.
If upgrading an existing timestamp-deletion schema, add is_delete=false then backfill true where deleted_at IS NOT NULL; keep the old column only during a staged rollout and retire it after verification. Never reset deleted rows to active.
For the earlier category archive design, map archived_at IS NOT NULL to is_delete=true when replacing archive semantics. This repository has no deployed schema yet; use the fresh schema unless inspection proves otherwise.
Backfill actor attribution on legacy transactions with known owner only where historical actor information is absent, and document that limitation.
Migrations preserve all rows/owner IDs. Tests verify defaults, true/false backfill, constraints, role escalation denial, A/B/C admin CRUD, stale versions, delete/restore races, report/Trash behavior, category history and atomic admin audits.
Run migration down only on disposable fixtures until data implications are reviewed. No production SQL has been executed by this documentation update.
