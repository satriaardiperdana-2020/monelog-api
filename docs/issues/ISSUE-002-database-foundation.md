# ISSUE-002: Database migrations and sqlc foundation

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api
Dependencies: 001
Remote issue: Not created
Requirements: FR-01, FR-15, FR-16, FR-17
Plan: [PLAN-002](../plans/PLAN-002.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Implement core role, ownership, boolean lifecycle, actor attribution and audit schema with repeatable SQL generation.

## Acceptance criteria
- [ ] users/categories/transactions have is_delete BOOLEAN NOT NULL DEFAULT FALSE and version constraints.
- [ ] Roles default to user; amount/type/owner composite constraints reject invalid data.
- [ ] Admin changes can retain selected-owner user_id and record actual actor.
- [ ] Required owner-filtered active/Trash queries and versioned delete/restore operations generate reproducibly.
- [ ] Fresh migrations and any applicable legacy true/false backfill preserve rows and existing deletion state.

## Scope and affected areas
db/migrations; db/queries; sqlc configuration; internal/repository/sqlc; PostgreSQL integration fixtures.
No destructive migration on shared data or automatic promotion of existing accounts.

## Verification
Fresh migration; applicable legacy backfill; boolean NOT NULL/default; wrong owner/type/amount; duplicate request key; active versus Trash queries; runtime audit grants; deterministic regeneration.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
