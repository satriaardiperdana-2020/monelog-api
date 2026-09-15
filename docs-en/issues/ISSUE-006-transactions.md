# ISSUE-006: Transaction CRUD, soft delete and daily summaries

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api
Dependencies: 005
Remote issue: Not created
Requirements: FR-01, FR-02, FR-03, FR-04, FR-15, FR-17
Plan: [PLAN-006](../plans/PLAN-006.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Implement exact-money transaction operations and daily history with active/Trash separation and restoration.

## Acceptance criteria
- [ ] Create defaults false; ordinary edits cannot set isDelete or change ownership.
- [ ] Owner CRUD and service-authorized admin CRUD retain correct owner with actor attribution.
- [ ] Delete sets true; restore sets false; each successful transition increments version and preserves row/idempotency identity.
- [ ] Active lists/day totals exclude true rows; authorized Trash lists/details expose them.
- [ ] Stale edit/delete/restore races return 409; foreign resource/category returns 404.
- [ ] Example day totals 769500.00 and recalculates correctly after delete/restore.

## Scope and affected areas
Transaction handlers/services; scoped queries/sqlc; daily aggregation; integration tests.
No wallets, transfers, recurring automatic transactions or offline synchronization.

## Verification
Duplicate/conflicting create replay; positive scoped admin CRUD; owner/actor distinction; true/false flag; wrong scope; stale versions; concurrent edit/delete/restore; category state; empty day; money/date/cursor boundaries.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
