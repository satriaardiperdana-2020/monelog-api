# PLAN-010: Android/iOS owner and admin management

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-010](../issues/ISSUE-010-mobile-packaging.md)
Repository: monelog-app
Prerequisites: 009
Requirements: FR-01, FR-09, FR-15, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Check official current Capacitor/toolchain requirements and available macOS/Xcode signing setup.
2. Add platform configuration, HTTPS API connectivity and vetted secure refresh storage.
3. Port existing role-aware CRUD/Trash/admin selectors with fixed form owner context.
4. Implement scoped export file opening/sharing for own or authorized selected-user exports.
5. Verify keyboard/date inputs, safe areas, back navigation, restart/resume/session denial and offline explanation.
6. Record actual Android/iOS device or simulator results and signing/release instructions.

## Affected areas
Capacitor and native platform projects; auth/API/download adapters; shared Vue views; device smoke/E2E checks.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
A/B/C login and positive admin CRUD; delete/restore on device; stale form target; late response; app resume/restart; account deletion/demotion; scoped native export/share; accessibility.
Run the configured frontend unit/component/E2E commands and npm run build; inspect package.json for exact names. Device builds/checks are required by the mobile issue.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No guaranteed app-store approval or offline synchronization. Missing platform/signing infrastructure is a stated implementation blocker.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
