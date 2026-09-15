# ISSUE-012: Owner/admin Drive backup and validated restore

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api + monelog-app
Dependencies: 011
Remote issue: Not created
Requirements: FR-01, FR-12, FR-15, FR-16, FR-17
Plan: [PLAN-012](../plans/PLAN-012.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Allow owners and admins to manage authorized per-user Drive backups and restore retained data while preserving boolean deletion and owner boundaries.

## Acceptance criteria
- [ ] Own/admin connection/schedule/job/download/restore operations use selected owner and valid provider authorization.
- [ ] Jobs/schedules persist owner, authorizing actor and mode; revoked authority or deleted accounts cancel/pause execution.
- [ ] Backup data preserves true and false business rows, excludes roles/credentials/audit records and contains checksums/schema version.
- [ ] Restore validates target ownership/category relationships and preserves flags, including active historical transactions linked to deleted categories; imported fields cannot elevate role.
- [ ] Actor/target audit and explicit external job states support safe retries and a real recovery drill.

## Scope and affected areas
Drive OAuth/service/worker; connection/schedule/job migrations; backup/restore schema; scoped API; UI settings; isolated restore tests.
No unattended overwrite of production data, plaintext credentials in backups, or claim that app admin role authorizes a Google account without provider consent.

## Verification
Own/admin positive operations and A/B denial; expired/revoked OAuth; requester demotion; deleted owner; duplicate runs; target payload tampering; corrupted backup; malicious role/owner fields; flag/count/active-total parity.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
