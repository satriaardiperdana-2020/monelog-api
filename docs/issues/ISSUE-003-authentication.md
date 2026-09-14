# ISSUE-003: Authentication, account state and role authorization

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api
Dependencies: 002
Remote issue: Not created
Requirements: FR-01, FR-15, FR-16, FR-17
Plan: [PLAN-003](../plans/PLAN-003.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Authenticate browser/native accounts and enforce current actor activity and role before personal/admin operations.

## Acceptance criteria
- [ ] Register/login/refresh/logout and GET/PATCH /me work; public registration defaults to regular user.
- [ ] Personal role/owner/isDelete overrides are rejected.
- [ ] Deleted actor cannot authenticate/refresh/use protected routes with old access tokens.
- [ ] Current-role admin guard denies a non-admin before target lookup and supports both admin reads and writes.
- [ ] Own account soft deletion revokes sessions and pauses owned schedules when that feature exists.
- [ ] Initial admin bootstrap is explicit; protected admin user/role CRUD is implemented in Issue 013.

## Scope and affected areas
cmd/admin bootstrap; internal/middleware; auth/profile handlers/services; user/session queries.
No raw password/session/provider secret exposure. Full target account/role management endpoints are completed in 013.

## Verification
Invalid credentials/JWT; issuer/audience/expiry; refresh replay/race; CSRF/origin; registration role injection; deleted account; self-delete version conflict; bootstrap; demotion with old JWT.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
