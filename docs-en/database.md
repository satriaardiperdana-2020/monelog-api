# Database design
Draft v0.3 • 14 September 2026
Logical design; executable migrations/queries arrive in Issue 002. Full admin management is Issue 013.

## Common conventions
UUID identifiers. created_at/updated_at TIMESTAMPTZ; update timestamps on mutations.
Soft-deletable entities have is_delete BOOLEAN NOT NULL DEFAULT FALSE and version INTEGER NOT NULL DEFAULT 1 CHECK(version>0).
JSON exposes isDelete (exact spelling) and Go uses IsDelete. Other JSON fields remain snake_case.
NULL is not a deletion state. Rows are active only when is_delete=false.

## Table structure (DDL)

The following DDL is a readable logical representation. Executable migrations are created and reviewed in Issue 002 and the relevant feature issues.

~~~sql
CREATE TABLE users (
    id            UUID PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'user'
                  CHECK (role IN ('user', 'admin')),
    timezone      TEXT NOT NULL DEFAULT 'Asia/Jakarta',
    currency      TEXT NOT NULL DEFAULT 'IDR'
                  CHECK (currency = 'IDR'),
    is_delete     BOOLEAN NOT NULL DEFAULT FALSE,
    version       INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE categories (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    type        TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    name        VARCHAR(80) NOT NULL CHECK (length(btrim(name)) > 0),
    is_delete   BOOLEAN NOT NULL DEFAULT FALSE,
    version     INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id, user_id, type)
);

CREATE TABLE transactions (
    id                UUID PRIMARY KEY,
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id       UUID NOT NULL,
    type              TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    amount            NUMERIC(14, 2) NOT NULL CHECK (amount > 0),
    transaction_date  DATE NOT NULL,
    title             VARCHAR(200) NOT NULL CHECK (length(btrim(title)) > 0),
    client_request_id UUID NOT NULL,
    request_hash      TEXT NOT NULL,
    created_by        UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_by        UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    is_delete         BOOLEAN NOT NULL DEFAULT FALSE,
    version           INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, client_request_id),
    FOREIGN KEY (category_id, user_id, type)
        REFERENCES categories (id, user_id, type) ON DELETE RESTRICT
);

CREATE TABLE refresh_sessions (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    family_id   UUID NOT NULL,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    replaced_by UUID REFERENCES refresh_sessions(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE admin_access_events (
    id             UUID PRIMARY KEY,
    actor_user_id  UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    target_user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    resource_type  TEXT NOT NULL,
    resource_id    UUID,
    action         TEXT NOT NULL,
    outcome        TEXT NOT NULL,
    request_id     TEXT NOT NULL,
    safe_metadata  JSONB NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Added by the relevant later milestone.
CREATE TABLE transaction_templates (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id UUID NOT NULL,
    type        TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    name        VARCHAR(80) NOT NULL CHECK (length(btrim(name)) > 0),
    amount      NUMERIC(14, 2) NOT NULL CHECK (amount > 0),
    title       VARCHAR(200) NOT NULL CHECK (length(btrim(title)) > 0),
    is_delete   BOOLEAN NOT NULL DEFAULT FALSE,
    version     INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id, user_id, type),
    FOREIGN KEY (category_id, user_id, type)
        REFERENCES categories (id, user_id, type) ON DELETE RESTRICT
);

CREATE TABLE export_jobs (
    id                UUID PRIMARY KEY,
    owner_user_id     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    requested_by      UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    request_mode      TEXT NOT NULL CHECK (request_mode IN ('personal', 'admin')),
    immutable_filters JSONB NOT NULL,
    format            TEXT NOT NULL CHECK (format IN ('xlsx', 'pdf')),
    status            TEXT NOT NULL,
    artifact_locator  TEXT,
    expires_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE drive_connections (
    id                      UUID PRIMARY KEY,
    user_id                 UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    encrypted_refresh_token TEXT NOT NULL,
    provider_account_label  TEXT NOT NULL,
    revoked_at              TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE backup_schedules (
    id            UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    authorized_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    request_mode  TEXT NOT NULL CHECK (request_mode IN ('personal', 'admin')),
    schedule      TEXT NOT NULL,
    timezone      TEXT NOT NULL,
    enabled       BOOLEAN NOT NULL DEFAULT TRUE,
    paused        BOOLEAN NOT NULL DEFAULT FALSE,
    version       INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE backup_jobs (
    id               UUID PRIMARY KEY,
    owner_user_id    UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    requested_by     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    request_mode     TEXT NOT NULL CHECK (request_mode IN ('personal', 'admin')),
    schedule_id      UUID REFERENCES backup_schedules(id) ON DELETE RESTRICT,
    scheduled_for    TIMESTAMPTZ NOT NULL,
    status           TEXT NOT NULL,
    attempt          INTEGER NOT NULL DEFAULT 0 CHECK (attempt >= 0),
    provider_file_id TEXT,
    checksum         TEXT,
    error_code       TEXT,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (schedule_id, scheduled_for)
);

CREATE UNIQUE INDEX categories_owner_type_name_uq
    ON categories (user_id, type, lower(name));

CREATE INDEX transactions_active_list_idx
    ON transactions (user_id, transaction_date DESC, created_at DESC, id DESC)
    WHERE is_delete = FALSE;

CREATE INDEX transactions_active_category_idx
    ON transactions (user_id, category_id, transaction_date)
    WHERE is_delete = FALSE;

CREATE INDEX transactions_trash_idx
    ON transactions (user_id, updated_at DESC, id DESC)
    WHERE is_delete = TRUE;

CREATE INDEX refresh_sessions_cleanup_idx
    ON refresh_sessions (user_id, family_id, expires_at);

CREATE INDEX admin_access_events_actor_idx
    ON admin_access_events (actor_user_id, created_at DESC);

CREATE INDEX admin_access_events_target_idx
    ON admin_access_events (target_user_id, created_at DESC);

CREATE INDEX export_jobs_owner_status_idx
    ON export_jobs (owner_user_id, status);

CREATE INDEX backup_jobs_owner_status_idx
    ON backup_jobs (owner_user_id, status, scheduled_for);
~~~

Job status values, retention rules, and provider states are refined in Issue 009/012. Roles describe application authority, not database superuser privileges. Foreign keys intentionally do not use cascade delete because business rows and audit history must not be physically erased.

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
Before mutation, lock involved account rows in UUID order with FOR UPDATE, recheck actor active/current admin role as needed and target account state, then mutate resource under required version.
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
