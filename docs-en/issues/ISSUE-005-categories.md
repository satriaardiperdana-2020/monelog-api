# ISSUE-005: Category CRUD, Trash and restore

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api
Dependencies: 004
Remote issue: Not created
Requirements: FR-01, FR-05, FR-15, FR-17
Plan: [PLAN-005](../plans/PLAN-005.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Implement scoped category management using boolean deletion while preserving historical transaction labels.

## Acceptance criteria
- [ ] Regular users CRUD/restore only their own categories; scoped services support admin management through 013.
- [ ] Delete sets true, restore sets false, and versioned operations retain the row.
- [ ] Deleted categories disappear from active selectors but historical transaction totals/labels remain.
- [ ] Type remains immutable and names unique per owner/type across active/deleted rows.
- [ ] New/edit/restored transactions/templates cannot select a deleted or wrong-owner category.

## Scope and affected areas
Category handlers/service; db category queries; internal/repository; active selectors and historical label integration tests.
No physical category deletion or resurrecting history by recreating a deleted name.

## Verification
CRUD ownership; wrong type; duplicate name/case; soft-delete row retained; Trash; restore; stale version; category delete with existing transaction; concurrent selection/category deletion.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
