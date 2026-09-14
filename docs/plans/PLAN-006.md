# PLAN-006: Transaction CRUD, soft delete and daily summaries

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-006](../issues/ISSUE-006-transactions.md)
Repository: monelog-api
Prerequisites: 005
Requirements: FR-01, FR-02, FR-03, FR-04, FR-15, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Implement strict amount/title/date/category validation and account/category locking.
2. Add create idempotency per authorized owner and request hash, recording actor in attribution.
3. Implement versioned edit, SoftDeleteTransaction and RestoreTransaction with expected flag state.
4. Implement active and Trash list/detail queries and scope/filter/deletion-bound cursors.
5. Implement active-only daily summaries and wire core handlers for later admin reuse.
6. Test owner isolation, authorized admin service scope, retained deleted rows and exact totals.

## Affected areas
Transaction handlers/services; scoped queries/sqlc; daily aggregation; integration tests.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
Duplicate/conflicting create replay; positive scoped admin CRUD; owner/actor distinction; true/false flag; wrong scope; stale versions; concurrent edit/delete/restore; category state; empty day; money/date/cursor boundaries.
Run go test ./..., go vet ./... and the configured PostgreSQL/HTTP integration suite where relevant. Verify generated-code drift for changed SQL/OpenAPI sources.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No wallets, transfers, recurring automatic transactions or offline synchronization.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
