# ISSUE-007: Summary and breakdown reports

Status: Backlog
Repository: monelog-api
Dependencies: 006
Remote issue: Not created
Plan: ../plans/PLAN-007.md

## Goal
Deliver simple date-filtered numeric reports.

## Acceptance criteria
- [ ] 7/30/custom semantics consistent
- [ ] weekly/monthly/category totals match transaction sums
- [ ] empty totals zero.

- [ ] Personal reports for user/admin accounts use only the actor's records.
- [ ] The report service can produce identical totals for a service-authorized target scope without combining users.
- [ ] Presets use the financial owner's timezone; missing scope never means all users.

## Scope exclusions
No graph UI or stored balances.

## Access-control scope (14 September 2026)
FR-01/FR-15: authorize selected-target report requests at Issue 013's admin boundary; keep report aggregation owner-filtered.
See [access-control.md](../access-control.md).

## Verification
Week/month/year/leap-day edges; date inclusivity; partial buckets; income minus expense; ownership and empty sets; benchmark target.

Additional authorization verification: Run distinct A/B/C aggregates, personal admin-own totals, target-isolated category/weekly/monthly totals, different-timezone presets and empty target cases.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
