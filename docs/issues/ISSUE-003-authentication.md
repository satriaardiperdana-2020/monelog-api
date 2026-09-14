# ISSUE-003: Authentication, roles and profile

Status: Backlog
Repository: monelog-api
Dependencies: 002
Remote issue: Not created
Plan: ../plans/PLAN-003.md

## Goal
Protect user-owned records across browser/native clients.

## Acceptance criteria
- [ ] Register/login/refresh/logout and profile work
- [ ] refresh replay revokes family
- [ ] secrets never logged
- [ ] unauthorized cross-user attempts fail; current-role admin guard supports the approved read-only routes implemented in Issue 013.

- [ ] Public registration assigns role=user; registration/profile role overrides return 400.
- [ ] GET /me returns current user/admin role; JWT sub always identifies the actor.
- [ ] Admin middleware checks current database role on every request and rejects non-admins before target lookup.
- [ ] Operator-only role provisioning targets an explicit existing account, records the change and revokes refresh sessions; a committed demotion blocks the next admin request with an old JWT.

## Scope exclusions
No Google login or provider backup credentials.

## Access-control scope (14 September 2026)
FR-01/FR-15/FR-16: role assignment through trusted server operation only. No public role editor, automatic first-admin registration or impersonation; actual admin data routes arrive in 013.
See [access-control.md](../access-control.md).

## Verification
Wrong password; duplicate email; expired JWT; invalid issuer/audience; refresh replay/race; logout; CSRF failure; user isolation.

Additional authorization verification: Add role escalation attempts, forged/stale role claim, denied admin route before target lookup, operator promotion/demotion, refresh revocation and unexpired JWT after demotion.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
