# ISSUE-007: Active transaction reports

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api
Dependencies: 006
Remote issue: Not created
Requirements: FR-01, FR-06, FR-07, FR-15, FR-17
Plan: [PLAN-007](../plans/PLAN-007.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Generate exact owner-scoped reports that consistently exclude deleted transactions.

## Acceptance criteria
- [ ] 7/30/custom dates and partial week/month buckets are consistent.
- [ ] Income/expense/difference and categories match only is_delete=false transaction sums.
- [ ] Delete/restore changes totals exactly once; category deletion does not remove historical amounts.
- [ ] Admin target scope uses selected owner's timezone and no mixed-owner aggregate.
- [ ] Empty reports return zero totals and empty groups.

## Scope and affected areas
Report handlers/service; aggregation SQL; date/decimal helpers; integration fixtures.
No graph UI, stored balance column or report totals that include Trash.

## Verification
Weekly/monthly/year/leap-date edges; inclusivity; partial buckets; deleted transaction exclusion; restored inclusion once; deleted category history; A/B/admin totals; target timezone; empty/benchmark checks.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
