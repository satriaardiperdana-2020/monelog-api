# PLAN-013: Full admin user and data management

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-013](../issues/ISSUE-013-admin-viewing.md)
Repository: monelog-api
Prerequisites: 007
Requirements: FR-01, FR-15, FR-16, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Reuse current actor/role guards on all admin methods and implement explicit target scope.
2. Add directory and account create/profile/role/delete/restore handlers, using versions, locks and session revocation.
3. Implement admin category/transaction read/create/update/delete/restore and active/Trash endpoints through scoped domain services.
4. Wire selected-owner daily/report reads and scope metadata; validate related IDs and cursors.
5. Add admin audit listing; atomically audit all mutations and persist successful read events before response.
6. Provide shared owner/admin job policy for later export/template/backup features, without prematurely implementing them.
7. Run positive and negative A/B/C matrix and hand stable backend to frontend Issue 008.

## Affected areas
Admin handlers; middleware/current role; authorization service; user/category/transaction/report/audit queries; OpenAPI; real PostgreSQL/HTTP suite.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
A/B denial versus C positive full CRUD; C-created record owner B; role promotion/demotion; user boolean deletion/restoration; active/Trash; version races; category mismatch; audit rollback; selected-user totals; expired/deleted actor.
Run go test ./..., go vet ./... and the configured PostgreSQL/HTTP integration suite where relevant. Verify generated-code drift for changed SQL/OpenAPI sources.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No automatic production role change during planning, no physical business-row deletion, and no raw credential responses. Existing filename retains its historical slug for link compatibility.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
Complete backend management before frontend Issue 008; retain this task's historical filename for existing links.
