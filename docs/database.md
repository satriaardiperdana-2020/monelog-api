# Database design
Draft v0.2, 14 September 2026. ISSUE-002 produces executable migrations and tests; ISSUE-003 implements role provisioning; ISSUE-013 implements admin reads.

## Tables
All IDs UUID. Audit timestamps TIMESTAMPTZ, created_at/updated_at default current timestamp; application updates updated_at.
| Table | Columns / constraints |
| --- | --- |
| users | id PK; email normalized lowercase UNIQUE NOT NULL; password_hash NOT NULL; role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('user','admin')); timezone NOT NULL default Asia/Jakarta; currency NOT NULL default IDR CHECK currency='IDR'; created_at; updated_at |
| categories | id PK; user_id FK users NOT NULL; type CHECK IN ('income','expense'); name VARCHAR(80) NOT NULL; archived_at nullable; created_at; updated_at; UNIQUE(id,user_id,type) |
| transactions | id PK; user_id FK users NOT NULL; category_id NOT NULL; type NOT NULL CHECK IN ('income','expense'); amount NUMERIC(14,2) NOT NULL CHECK(amount>0); transaction_date DATE NOT NULL; title VARCHAR(200) NOT NULL; version INTEGER NOT NULL default 1 CHECK(version>0); client_request_id UUID NOT NULL; request_hash TEXT NOT NULL; deleted_at nullable; created_at; updated_at; UNIQUE(user_id,client_request_id) |
| refresh_sessions | id PK; user_id FK users; family_id UUID; token_hash UNIQUE NOT NULL; expires_at NOT NULL; revoked_at nullable; replaced_by nullable self FK; created_at |
| admin_access_events | id PK; actor_user_id UUID NOT NULL FK users; target_user_id UUID nullable FK users; action TEXT NOT NULL CHECK IN ('users.list','user.read','categories.list','transactions.list','transaction.read','daily_summaries.read','reports.summary','reports.breakdown'); resource_id UUID nullable; outcome TEXT NOT NULL CHECK IN ('allowed','forbidden','not_found'); request_id TEXT NOT NULL; created_at TIMESTAMPTZ NOT NULL default current_timestamp |
| transaction_templates (later) | id PK; user_id FK; category_id; type; amount NUMERIC(14,2) CHECK(amount>0); title; name VARCHAR(80); created_at; updated_at |
| drive_connections (later) | id PK; user_id UNIQUE FK; encrypted_refresh_token; provider_account_label; revoked_at; created_at; updated_at |
| backup_jobs (later) | id PK; user_id FK; scheduled_for; status; attempt; provider_file_id nullable; checksum nullable; safe_error_code nullable; started_at; completed_at; UNIQUE(user_id,scheduled_for) |

NOT NULL applies to required functional fields; specify exact nullability/defaults in migrations. For future tables schema is provisional pending their issue's design gate.
Transactions enforce FOREIGN KEY(category_id,user_id,type) REFERENCES categories(id,user_id,type).
Templates need the equivalent FK. Trimmed non-empty names/titles enforced with CHECK(length(btrim(...))>0).
Archived category check is enforced by the service inside the write transaction; lock selected category during creation to prevent archive races.
Users/categories use restricted deletion; no cascade erasing financial history.

## Indexes
- categories: UNIQUE(user_id,type,lower(name)) — includes archived names; unarchive instead of creating an ambiguous duplicate.
- transactions: (user_id,transaction_date DESC,id DESC) WHERE deleted_at IS NULL.
- transactions: (user_id,category_id,transaction_date) WHERE deleted_at IS NULL.
- refresh_sessions: (user_id,family_id); expires_at for cleanup.
- admin_access_events: (actor_user_id,created_at DESC) and (target_user_id,created_at DESC).
- users: existing unique normalized-email index supports stable email,id selector ordering; evaluate search performance before adding another index.
Evaluate additional report indexes with EXPLAIN on realistic data. No index on title/amount until a demonstrated query need.

## Mutation rules
Create uses client_request_id and hash of canonical create fields. Repeated identical requests return original record; changed payload with same key returns 409. Keep key after deletion; replay of a deleted result returns 409, never silently creates another record.
Update uses WHERE id=$id AND user_id=$user AND version=$expected AND deleted_at IS NULL; increment version. On personal mutation routes, missing/foreign IDs return 404 for regular users and admins alike; owned stale versions return 409.
Delete uses the same version guard, sets deleted_at and increments version.
Every list/report excludes deleted_at IS NOT NULL. Soft-deleted records remain private and subject to a future documented retention/purge policy.

## Reporting queries
Use transaction_date >= start AND transaction_date <= end, always with an explicit authorized owner user_id and deleted_at filters. Owner = actor on personal routes; owner = validated selected target on admin GET routes. Never use an OR is_admin clause to bypass the owner predicate.
Totals: COALESCE(SUM(amount) FILTER (WHERE type='income'),0) and corresponding expense sum.
Group by transaction_date for daily data, date_trunc('week',transaction_date::timestamp)::date for Monday buckets, and date_trunc('month',...) for months.
Filter dates before grouping; boundary buckets are partial. Difference is computed, never stored.
Daily summary returns only dates with records; today's zero summary is returned separately.
Order same-date transactions by created_at DESC,id DESC; cursor contains both values when listing within a day.

## Migration and test plan
Create users (including role) → categories → transactions → sessions → admin_access_events; seed default categories separately per user during registration.
For an existing users table, add the role column with user as the default/backfill and the enum check in a forward migration. Do not auto-promote any account. Test against populated fixtures; rollouts do not change any transaction owner.
Integration tests: invalid amounts, wrong category type/owner, duplicate request IDs, stale versions, archived category race, soft-delete exclusion, leap day, week/month boundaries, empty aggregates.
Test migration up on empty database and forward upgrade with fixtures. Down migrations only against disposable databases until data-loss consequences are approved.
No production SQL is included in this draft.

## Role and scope invariants
users.role describes application permissions, not a PostgreSQL superuser role. It never changes ownership of transactions/categories.
The runtime database role may read users.role but cannot update it. Grant user-insert privileges only for required registration columns (role supplied by database default), and profile-update privileges only for allowed columns. Use a separate privileged operator connection for explicit role changes.
Select actor role before resolving any target user. The admin directory projects id,email,timezone,currency only; it never loads password_hash/session/provider columns.
Admin financial reads reuse owner-filtered sqlc queries after service authorization. Category filters and transaction IDs must match the same target; mismatch returns 404. Missing target scope is a validation error, never an all-user query.
Admin cursors bind actor ID, target ID (null only for directory), endpoint and filters; reject reuse across targets. Personal cursors similarly bind the actor's own scope.
Admin read access is not a bypass around category composite FKs, version checks, idempotency ownership or soft deletion.

## Audit and future jobs
admin_access_events is append-only to the runtime role: insert only; no application update/delete/read endpoint. Operational audit access/retention is managed separately by the server operator. Proposed retention is 90 days, to be confirmed before release.
target_user_id is null for directory reads, unauthorized callers (authorization precedes lookup), or unresolved targets; use a resolved target FK only when valid. resource_id is optional transaction UUID without a FK so not-found lookups can be recorded safely.
Store no titles, amounts, search strings, email lists, request bodies, credentials or response payloads. Success auditing must persist before data is returned; denied attempts are best-effort if persistence is unavailable.
Future export_jobs contains owner_user_id fixed to the requester; jobs and downloads always use that owner. Admin viewing introduces no second export owner.
Personal backup formats exclude roles, sessions, admin audit events and provider credentials. Restore binds all imported financial rows to the authenticated backup owner, validates category ownership and cannot promote roles; reject privileged fields. Restore ownership remapping needs the later backup design.
Additional integration fixtures: two regular users plus an admin with distinct records. Verify default/invalid roles, runtime-role update denial, per-target aggregates, current-role demotion, append-only audit grants and no financial changes after admin reads.
