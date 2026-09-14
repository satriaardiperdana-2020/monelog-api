# ISSUE-009: Owner/admin Excel and PDF exports

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api + monelog-app
Dependencies: 008
Remote issue: Not created
Requirements: FR-01, FR-08, FR-15, FR-16, FR-17
Plan: [PLAN-009](../plans/PLAN-009.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Export all matching active transactions for an owner or an admin's selected owner with accurate job authorization.

## Acceptance criteria
- [ ] Both XLSX/PDF match selected-owner active report totals/filters and include all matching pages.
- [ ] Admin can request/manage/download any selected owner's exports; regular users only their own.
- [ ] Jobs persist immutable owner_user_id, requested_by, request_mode and filters; queued jobs revalidate authority.
- [ ] Deleted transactions are excluded, while category deletion preserves historical labels.
- [ ] Expired/unready/out-of-scope downloads and canceled/demoted-requester jobs are handled correctly.

## Scope and affected areas
Export jobs/migrations/worker; report service; file generators; download handlers; web export UI.
Export is a report artifact, not a data backup/import format. Report export never includes Trash.

## Verification
More-than-one-page export; exact report parity; true rows excluded; formula injection; long-title PDF layout; A denied/B own/C admin download; stale role; target switching; job retry/expiry.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
