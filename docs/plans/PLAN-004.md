# PLAN-004: Owner/admin API and soft-delete contract

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-004](../issues/ISSUE-004-api-contract.md)
Repository: monelog-api
Prerequisites: 003
Requirements: FR-01, FR-15, FR-16, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Reconcile auth implementation with api.md and enumerate every mapped admin finance path.
2. Define user/category/transaction DTOs, exact isDelete spelling, current role and actor/owner scope.
3. Define own/admin delete/restore version rules, Trash detail and report active-only behavior.
4. Define protected admin account creation/profile/role/delete/restore and audit listing operations.
5. Mark exports/templates/Drive contracts by later milestone; require admin parity when expanded.
6. Validate examples, route security/method coverage and generated code.

## Affected areas
api/openapi.yaml; oapi-codegen configuration; internal/api generated interfaces; contract validation.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
JSON examples; required boolean and invalid string/null flag; owner/role/lifecycle overrides; admin mutation operations; missing/stale version schema; generation compile/drift.
Run go test ./..., go vet ./... and the configured PostgreSQL/HTTP integration suite where relevant. Verify generated-code drift for changed SQL/OpenAPI sources.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No domain handler implementation or unimplemented routes presented as running.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
