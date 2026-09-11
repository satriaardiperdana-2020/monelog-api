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

## Scope exclusions
Not a full backup or import mechanism.

## Verification
Over-page-size export; report parity; spreadsheet formula-injection safety; PDF layout/long title check; job failure/retry; expiry and foreign access.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

