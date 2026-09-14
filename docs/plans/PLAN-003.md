# PLAN-003: Authentication, account state and role authorization

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-003](../issues/ISSUE-003-authentication.md)
Repository: monelog-api
Prerequisites: 002
Requirements: FR-01, FR-15, FR-16, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Finalize password/session policy, JWT validation and browser/native refresh transport.
2. Register with role=user, is_delete=false and category seeding; reject disallowed fields.
3. Implement rotating hashed refresh sessions, reuse detection and logout.
4. Load current actor state for protected requests; implement current-role guard for all supported admin methods.
5. Implement versioned self profile/account deletion with session revocation and account-locking rules.
6. Document explicit cmd/admin bootstrap/recovery and test unexpired-token denial after account deletion/demotion.

## Affected areas
cmd/admin bootstrap; internal/middleware; auth/profile handlers/services; user/session queries.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
Invalid credentials/JWT; issuer/audience/expiry; refresh replay/race; CSRF/origin; registration role injection; deleted account; self-delete version conflict; bootstrap; demotion with old JWT.
Run go test ./..., go vet ./... and the configured PostgreSQL/HTTP integration suite where relevant. Verify generated-code drift for changed SQL/OpenAPI sources.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No raw password/session/provider secret exposure. Full target account/role management endpoints are completed in 013.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
