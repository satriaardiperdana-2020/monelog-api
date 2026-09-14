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

- [ ] Drive connection, schedules, jobs and restore are bound to the signed-in owner for both roles; admin selection cannot change them.
- [ ] Personal backup payloads exclude roles, session/audit/provider secrets; restore rejects privileged fields and cannot promote an account.
- [ ] Restored records and category relationships are validated/remapped to the authorized backup owner.

## Scope exclusions
No unattended overwrite of production data; destructive restore needs explicit confirmation.

## Access-control scope (14 September 2026)
FR-01/FR-15: admin finance viewing grants no permission over another user's Drive tokens, backup files, schedules or restoration. Operational database recovery is separate.
See [access-control.md](../access-control.md).

## Verification
Expired/revoked credentials; duplicate scheduler runs; cross-user isolation; corrupt backup; restore count/total parity; timezone schedule edge.

Additional authorization verification: Add C attempting B Drive/job/restore paths, selected-target payload tampering, malicious imported role/session/owner IDs and per-owner restored totals.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
