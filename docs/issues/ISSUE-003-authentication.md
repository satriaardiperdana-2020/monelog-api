# ISSUE-003: Authentication and profile

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
- [ ] cross-user attempts fail.

## Scope exclusions
No Google login or provider backup credentials.

## Verification
Wrong password; duplicate email; expired JWT; invalid issuer/audience; refresh replay/race; logout; CSRF failure; user isolation.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

