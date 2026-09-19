# PLAN-011: Owner/admin templates with soft delete

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-011](../issues/ISSUE-011-templates.md)
Repository: monelog-api + monelog-app
Prerequisites: 010
Requirements: FR-01, FR-10, FR-15, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Confirm single-transaction template versus day-bundle proposal before coding affected behavior.
2. Add template schema with boolean flag/version/composite category ownership and full personal/admin contract.
3. Implement scoped CRUD/Trash/Restore and validate active category on creation/edit/restore/apply.
4. Add template-management/apply UI in both My Data and Admin Management, with immutable owner context.
5. Test soft-deleted state and target switching cannot create a transaction for an unintended user.

## Affected areas
Template migration/queries/handlers/service; API contract; Vue template and transaction-form integration.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
Own CRUD and admin C managing B; A denied; isDelete false/true/restore; version race; apply twice creates no records; explicit Save; category lifecycle; owner switching.
Run go test ./..., go vet ./... and the configured PostgreSQL/HTTP integration suite where relevant. Verify generated-code drift for changed SQL/OpenAPI sources.
Run the configured frontend unit/component/E2E commands and npm run build; inspect package.json for exact names. Device builds/checks are required by the mobile issue.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No recurring automatic creation or whole-day bundles unless the requirement is revised.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
