# PLAN-002: Database migrations and sqlc foundation

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-002](../issues/ISSUE-002-database-foundation.md)
Repository: monelog-api
Prerequisites: 001
Requirements: FR-01, FR-15, FR-16, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Inspect for existing schema; use the fresh design if none, otherwise add a forward migration preserving prior deleted/archived status.
2. Create users, categories, transactions, sessions and admin_access_events in FK order.
3. Add false defaults, version constraints, actor attribution, composite category ownership and active/Trash indexes.
4. Generate scoped create/read/update/delete/restore queries, including authorization account-row locking and audit insert.
5. Set runtime privileges separately from migration ownership; audit insert/read allowed, audit update/delete denied.
6. Test A/B/C fixtures, bool defaults, constraints and generation drift on real PostgreSQL.

## Affected areas
db/migrations; db/queries; sqlc configuration; internal/repository/sqlc; PostgreSQL integration fixtures.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
Fresh migration; applicable legacy backfill; boolean NOT NULL/default; wrong owner/type/amount; duplicate request key; active versus Trash queries; runtime audit grants; deterministic regeneration.
Run go test ./..., go vet ./... and the configured PostgreSQL/HTTP integration suite where relevant. Verify generated-code drift for changed SQL/OpenAPI sources.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No destructive migration on shared data or automatic promotion of existing accounts.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
