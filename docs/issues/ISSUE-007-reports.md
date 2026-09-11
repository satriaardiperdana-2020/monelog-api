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

## Scope exclusions
No graph UI or stored balances.

## Verification
Week/month/year/leap-day edges; date inclusivity; partial buckets; income minus expense; ownership and empty sets; benchmark target.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

