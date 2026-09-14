# ISSUE-011: Owner/admin templates with soft delete

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api + monelog-app
Dependencies: 010
Remote issue: Not created
Requirements: FR-01, FR-10, FR-15, FR-17
Plan: [PLAN-011](../plans/PLAN-011.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Provide reusable templates editable by their owner or an admin for the selected owner, with boolean deletion and restore.

## Acceptance criteria
- [ ] Owner/admin template CRUD/Trash/Restore uses authorized owner, isDelete and versions.
- [ ] Applying an active template opens an unsaved form; only explicit Save creates a transaction for the same owner.
- [ ] Deleted/wrong-owner templates and incompatible/deleted categories are rejected appropriately.
- [ ] Admin can manage any selected owner's templates; a regular user cannot access another owner's templates.

## Scope and affected areas
Template migration/queries/handlers/service; API contract; Vue template and transaction-form integration.
No recurring automatic creation or whole-day bundles unless the requirement is revised.

## Verification
Own CRUD and admin C managing B; A denied; isDelete false/true/restore; version race; apply twice creates no records; explicit Save; category lifecycle; owner switching.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
