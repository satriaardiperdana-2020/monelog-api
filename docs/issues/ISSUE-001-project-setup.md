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

## Scope exclusions
No auth, migrations, finance logic, deployment or remote repo changes.

## Verification
Config validation tests; HTTP live test; readiness with unavailable DB; startup/shutdown smoke test.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

