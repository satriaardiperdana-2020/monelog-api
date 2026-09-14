# PLAN-012: Owner/admin Drive backup and validated restore

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-012](../issues/ISSUE-012-drive-backups.md)
Repository: monelog-api + monelog-app
Prerequisites: 011
Requirements: FR-01, FR-12, FR-15, FR-16, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Design schema-versioned backup format, schedule/timezone, retention and restore conflict policy before implementation.
2. Check current official provider OAuth requirements and bind authorization state to actor/target/mode.
3. Implement encrypted connections and own/admin management; do not infer OAuth consent from application role.
4. Implement versioned schedules and jobs with authority rechecks, unique runs, retry/backoff, cancel/pause/resume and audit.
5. Include active/deleted data with isDelete; validate schemas/checksums and map records only to authorized target.
6. Run a restore drill in an isolated DB, comparing total row counts, each flag group and active report totals.
7. Build owner/admin progress/settings UI and document operational recovery separately.

## Affected areas
Drive OAuth/service/worker; connection/schedule/job migrations; backup/restore schema; scoped API; UI settings; isolated restore tests.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
Own/admin positive operations and A/B denial; expired/revoked OAuth; requester demotion; deleted owner; duplicate runs; target payload tampering; corrupted backup; malicious role/owner fields; flag/count/active-total parity.
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
No unattended overwrite of production data, plaintext credentials in backups, or claim that app admin role authorizes a Google account without provider consent.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
