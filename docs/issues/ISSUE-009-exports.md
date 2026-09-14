# ISSUE-009: Excel and PDF report exports

Status: Backlog
Repository: monelog-api + monelog-app
Dependencies: 008
Remote issue: Not created
Plan: ../plans/PLAN-009.md

## Goal
Export all matching filtered records consistently with reports.

## Acceptance criteria
- [ ] Both formats match report totals and filters
- [ ] files private
- [ ] expired/foreign downloads rejected.

- [ ] Export owner always equals the authenticated requester for both roles; workers and download checks use that stored owner.
- [ ] An admin's selected-user view cannot submit another owner or fetch that user's export job; export/share UI is absent in Admin View.

## Scope exclusions
Not a full backup or import mechanism.

## Access-control scope (14 September 2026)
FR-01/FR-15: viewing another user's report grants no cross-user export/download permission. Admins may export their own records through My Data.
See [access-control.md](../access-control.md).

## Verification
Over-page-size export; report parity; spreadsheet formula-injection safety; PDF layout/long title check; job failure/retry; expiry and foreign access.

Additional authorization verification: Add owner override rejection, C attempting B's job status/download, UI target switching before export, worker owner tampering fixtures and owner-only totals.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
