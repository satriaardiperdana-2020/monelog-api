# PLAN-009: Owner/admin Excel and PDF exports

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-009](../issues/ISSUE-009-exports.md)
Repository: monelog-api + monelog-app
Prerequisites: 008
Requirements: FR-01, FR-08, FR-15, FR-16, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Add export_jobs schema and personal/admin route parity with separate actor and owner.
2. Atomically create job+admin audit as needed; recheck requester/owner/role before worker execution.
3. Use a consistent source snapshot and shared active-owner filters to render every matching row.
4. Generate formula-safe XLSX and readable PDF with totals, dates/currency and selected owner label.
5. Authorize status/download by actual owner or current admin target; implement retention cleanup and UI progress/retry.
6. Test target switches cannot change a pending export and audit/provider-independent failure behavior.

## Affected areas
Export jobs/migrations/worker; report service; file generators; download handlers; web export UI.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
More-than-one-page export; exact report parity; true rows excluded; formula injection; long-title PDF layout; A denied/B own/C admin download; stale role; target switching; job retry/expiry.
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
Export is a report artifact, not a data backup/import format. Report export never includes Trash.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
