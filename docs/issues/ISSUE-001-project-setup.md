# ISSUE-001: Backend project setup

Status: Backlog
Repository: monelog-api
Dependencies: None
Remote issue: Not created
Plan: ../plans/PLAN-001.md

## Goal
Create a reproducible Go/Echo service skeleton without finance features.

## Acceptance criteria
- [ ] Service starts from documented commands
- [ ] live/readiness checks differ correctly
- [ ] missing required configuration fails clearly
- [ ] no secrets tracked.

- [ ] Project documentation references the user/admin access model and planned admin route boundary without exposing role-setting configuration to public clients.

## Scope exclusions
No auth, migrations, finance logic, deployment or remote repo changes.

## Access-control scope (14 September 2026)
FR-01/FR-15 groundwork only: keep an explicit place for authentication and authorization in middleware/service. Implement roles in 003 and admin viewing in 013.
See [access-control.md](../access-control.md).

## Verification
Config validation tests; HTTP live test; readiness with unavailable DB; startup/shutdown smoke test.

Additional authorization verification: Review config examples and startup docs: no public admin toggle, automatic first-user promotion or seeded production admin credentials.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
