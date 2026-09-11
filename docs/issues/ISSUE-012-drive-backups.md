# ISSUE-012: Scheduled Google Drive backups and restore validation

Status: Backlog
Repository: monelog-api + monelog-app
Dependencies: 011
Remote issue: Not created
Plan: ../plans/PLAN-012.md

## Goal
Provide opt-in per-user scheduled backups with a verified recovery path.

## Acceptance criteria
- [ ] User can connect/disconnect
- [ ] scheduler retries without duplicate jobs
- [ ] only owner's data exported
- [ ] recovery drill validates counts and totals.

## Scope exclusions
No unattended overwrite of production data; destructive restore needs explicit confirmation.

## Verification
Expired/revoked credentials; duplicate scheduler runs; cross-user isolation; corrupt backup; restore count/total parity; timezone schedule edge.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

