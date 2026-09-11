# Database design
Draft logical schema; ISSUE-002 produces executable migrations and tests.

## Tables
All IDs UUID. Audit timestamps TIMESTAMPTZ, created_at/updated_at default current timestamp; application updates updated_at.
| Table | Columns / constraints |
| --- | --- |
| users | id PK; email normalized lowercase UNIQUE NOT NULL; password_hash NOT NULL; timezone NOT NULL default Asia/Jakarta; currency NOT NULL default IDR CHECK currency='IDR'; created_at; updated_at |
| categories | id PK; user_id FK users NOT NULL; type CHECK IN ('income','expense'); name VARCHAR(80) NOT NULL; archived_at nullable; created_at; updated_at; UNIQUE(id,user_id,type) |
| transactions | id PK; user_id FK users NOT NULL; category_id NOT NULL; type NOT NULL CHECK IN ('income','expense'); amount NUMERIC(14,2) NOT NULL CHECK(amount>0); transaction_date DATE NOT NULL; title VARCHAR(200) NOT NULL; version INTEGER NOT NULL default 1 CHECK(version>0); client_request_id UUID NOT NULL; request_hash TEXT NOT NULL; deleted_at nullable; created_at; updated_at; UNIQUE(user_id,client_request_id) |
| refresh_sessions | id PK; user_id FK users; family_id UUID; token_hash UNIQUE NOT NULL; expires_at NOT NULL; revoked_at nullable; replaced_by nullable self FK; created_at |
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
Evaluate additional report indexes with EXPLAIN on realistic data. No index on title/amount until a demonstrated query need.

## Mutation rules
Create uses client_request_id and hash of canonical create fields. Repeated identical requests return original record; changed payload with same key returns 409. Keep key after deletion; replay of a deleted result returns 409, never silently creates another record.
Update uses WHERE id=$id AND user_id=$user AND version=$expected AND deleted_at IS NULL; increment version. Missing/foreign IDs return 404; owned stale versions return 409.
Delete uses the same version guard, sets deleted_at and increments version.
Every list/report excludes deleted_at IS NOT NULL. Soft-deleted records remain private and subject to a future documented retention/purge policy.

## Reporting queries
Use transaction_date >= start AND transaction_date <= end, always with user_id and deleted_at filters.
Totals: COALESCE(SUM(amount) FILTER (WHERE type='income'),0) and corresponding expense sum.
Group by transaction_date for daily data, date_trunc('week',transaction_date::timestamp)::date for Monday buckets, and date_trunc('month',...) for months.
Filter dates before grouping; boundary buckets are partial. Difference is computed, never stored.
Daily summary returns only dates with records; today's zero summary is returned separately.
Order same-date transactions by created_at DESC,id DESC; cursor contains both values when listing within a day.

## Migration and test plan
Create users → categories → transactions → sessions; seed default categories separately per user during registration.
Integration tests: invalid amounts, wrong category type/owner, duplicate request IDs, stale versions, archived category race, soft-delete exclusion, leap day, week/month boundaries, empty aggregates.
Test migration up on empty database and forward upgrade with fixtures. Down migrations only against disposable databases until data-loss consequences are approved.
No production SQL is included in this draft.

