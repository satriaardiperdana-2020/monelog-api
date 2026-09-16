# ISSUE-002: Database migrations and sqlc foundation

Status: Implemented — ready for review
Updated: 16 September 2026 (v0.4)
Repository: monelog-api
Dependencies: 001
Remote issue: [GitHub #6](https://github.com/satriaardiperdana-2020/monelog-api/issues/6)
Requirements: FR-01, FR-15, FR-16, FR-17
Plan: [PLAN-002](../plans/PLAN-002.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Build the PostgreSQL and sqlc foundation for accounts, categories, transactions, sessions, audit, soft-deletion lifecycle, versioning, and actor attribution. This issue provides the data boundary required by later work; authentication, HTTP contracts, and complete admin-management flows belong to their respective issues.

## Acceptance criteria
- [x] Fresh migrations create users, categories, transactions, refresh_sessions, and admin_access_events consistently with the database design.
- [x] Data boundaries for ownership, category/transaction type, amount, is_delete, and version reject invalid state; every soft-deletable record starts active.
- [x] Core sqlc queries for owner scope, active/Trash data, and versioned lifecycle operations regenerate deterministically.
- [x] The actor and audit foundation preserves the distinction between the selected owner and the actor making a change, without granting application roles through migrations or ordinary registration queries.
- [x] Migrations and queries have evidence from a disposable PostgreSQL database. Legacy backfill was unnecessary because no deployed schema was found.

## Scope and affected areas
db/migrations; db/queries; sqlc configuration; internal/repository/sqlc; PostgreSQL integration fixtures; affected database documentation.
No destructive migration on shared data or automatic promotion of existing accounts.

## Verification
Fresh migration; boolean/default and version; invalid owner/type/amount; duplicate request key; active versus Trash queries; owner/actor attribution; runtime audit access; deterministic regeneration. Test legacy backfill only when a real legacy schema exists.

## Implementation evidence

- golang-migrate up, one-version down, and up again on disposable PostgreSQL 16: passed.
- `go test -tags=integration ./internal/repository/sqlc`: passed for defaults, constraints, owner/actor attribution, duplicate request keys, active/Trash, stale versions, restore, audit, and sessions.
- Runtime-role grant check: audit SELECT/INSERT `true`; audit UPDATE/DELETE and transaction DELETE `false`.
- `make check` and `make test-race`: passed. sqlc generation/vetting and drift checking are included in the quality gate.

## Definition of done
Acceptance criteria pass with actual test evidence; schema, documentation, and affected generated sqlc code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.

## Boundaries

Role authorization, HTTP status behavior, and atomic admin-write workflows are verified in ISSUE-003, ISSUE-004, and ISSUE-013 using this foundation.
