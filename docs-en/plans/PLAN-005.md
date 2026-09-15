# PLAN-005: Category CRUD, Trash and restore

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-005](../issues/ISSUE-005-categories.md)
Repository: monelog-api
Prerequisites: 004
Requirements: FR-01, FR-05, FR-15, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Implement scoped category list/detail/create/update/delete/restore queries and service validation.
2. Use false active default and true Trash filter, with row-version checks on mutations.
3. Keep category uniqueness across deleted rows and return conflict instead of ambiguous recreation.
4. Preserve same-owner historical joins without filtering out transactions whose category is deleted.
5. Expose core handlers and reuse authorized scope in 013's admin adapters.

## Affected areas
Category handlers/service; db category queries; internal/repository; active selectors and historical label integration tests.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
CRUD ownership; wrong type; duplicate name/case; soft-delete row retained; Trash; restore; stale version; category delete with existing transaction; concurrent selection/category deletion.
Run go test ./..., go vet ./... and the configured PostgreSQL/HTTP integration suite where relevant. Verify generated-code drift for changed SQL/OpenAPI sources.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No physical category deletion or resurrecting history by recreating a deleted name.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
