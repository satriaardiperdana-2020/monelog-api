# PLAN-008: Vue owner and admin management UI

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-008](../issues/ISSUE-008-web-mvp.md)
Repository: monelog-app
Prerequisites: 013
Requirements: FR-01, FR-02, FR-03, FR-04, FR-05, FR-06, FR-07, FR-11, FR-15, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Initialize pinned Vue JavaScript/router/state/API setup and in-memory access authentication.
2. Build personal main tabs, transaction/category CRUD, reports and version-aware error flows.
3. Add Trash/detail/restore UI using strict isDelete queries and lifecycle endpoints.
4. Render Admin Management using current /me role; build user directory/create/role/account lifecycle controls.
5. Reuse forms with immutable owner context and explicit admin endpoints; keep management actions enabled.
6. Handle save/discard, cancel stale reads, scope/generation-check responses and clear data on role/session changes.
7. Verify accessibility, narrow/wide layouts and A/B/C end-to-end flows.

## Affected areas
src/views/components/services/stores/router/utils; focused unit/component/E2E tests.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
Login/refresh; owner CRUD; positive admin CRUD/roles; A direct admin denial; active versus Trash; delete/restore totals; delayed responses; unsaved form target switch; logout/relogin; 360px/desktop accessibility.
Run the configured frontend unit/component/E2E commands and npm run build; inspect package.json for exact names. Device builds/checks are required by the mobile issue.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No Capacitor packaging, offline write queue, calculator or graph in browser MVP.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
