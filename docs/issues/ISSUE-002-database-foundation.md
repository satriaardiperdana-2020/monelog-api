# ISSUE-002: Database migrations and sqlc foundation

Status: Backlog
Repository: monelog-api
Dependencies: 001
Remote issue: Not created
Plan: ../plans/PLAN-002.md

## Goal
Implement the core schema and repeatable SQL generation.

## Acceptance criteria
- [ ] Migrations work on an empty disposable database
- [ ] ownership/type FK and amount constraints enforced
- [ ] generated output is reproducible.

- [ ] users.role defaults/backfills to user and accepts only user/admin; existing financial owner IDs remain unchanged.
- [ ] admin_access_events schema and append-only runtime grants exist; runtime users-column grants prevent role changes.
- [ ] Generated finance queries require one explicit owner ID, including when later called for an admin view.

## Scope exclusions
Do not run destructive migrations on shared databases.

## Access-control scope (14 September 2026)
FR-01/FR-15/FR-16: role schema, role-write privilege boundary and audit records support upcoming admin reads; no OR is_admin owner-filter bypass.
See [access-control.md](../access-control.md).

## Verification
Real PostgreSQL migration and invalid FK/amount/duplicate tests; regenerate and check clean diff.

Additional authorization verification: Test role default/backfill and invalid role, attempted role update with runtime DB role, audit insert versus update/delete grants, and isolated A/B/admin fixtures using the same scoped queries.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
