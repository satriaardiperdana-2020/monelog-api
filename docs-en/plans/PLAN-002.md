# PLAN-002: Database migrations and sqlc foundation

Status: Implemented — ready for review.
Updated: 16 September 2026 (v0.4)
Issue: [ISSUE-002](../issues/ISSUE-002-database-foundation.md)
Repository: monelog-api
Prerequisite: 001
Requirements: FR-01, FR-15, FR-16, FR-17

## Implementation decisions

- The repository has no deployed schema, so ISSUE-002 uses a clean initial migration without legacy backfill.
- Migrations use golang-migrate up/down pairs. Down is limited to disposable fixtures.
- sqlc 1.31.1 generates a pgx/v5 package in internal/repository/sqlc, and generated code is committed to the repository.
- The migration owner and runtime role are separate. Runtime grants live in db/roles/runtime.sql; runtime can only read and append audit records.
- Role authorization, HTTP responses, and atomic admin writes with audit remain the responsibility of ISSUE-003, ISSUE-004, and ISSUE-013.

## Implementation

1. db/migrations contains one up/down migration pair per table for users, categories, transactions, refresh_sessions, and admin_access_events with the required constraints and indexes.
2. db/queries provides core owner-scoped queries for accounts, categories, transactions, sessions, audit, active/Trash data, and versioned lifecycle operations.
3. sqlc.yaml deterministically generates internal/repository/sqlc with pgx/v5.
4. db/roles/runtime.sql grants minimum runtime access without physical business-data deletion or audit mutation privileges.
5. integration_test.go applies the migration in a random schema on disposable PostgreSQL and checks defaults, constraints, ownership, actors, idempotency keys, lifecycle behavior, audit, and sessions.

## Affected areas

db/migrations; db/queries; db/roles; sqlc.yaml; internal/repository/sqlc; Makefile; README; ISSUE-002 documentation and this plan.

## Validation

- Migration up, one-version down, and up again on disposable PostgreSQL 16: passed.
- PostgreSQL integration suite for the schema and generated queries: passed.
- Runtime role: audit SELECT/INSERT allowed; audit UPDATE/DELETE and transaction DELETE denied.
- sqlc generate and sqlc vet: passed.
- Go checks, formatting, vet, build, race, and generated-code drift checks: passed.

## Boundaries and recovery

There is no legacy migration because no deployed schema was found. If an older schema is discovered later, add a separate forward migration that preserves ownership and deletion state.
Use down only on disposable fixtures. Recover shared databases through a reviewed forward migration.
