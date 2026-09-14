# ISSUE-006: Transaction CRUD and daily summaries

Status: Backlog
Repository: monelog-api
Dependencies: 005
Remote issue: Not created
Plan: ../plans/PLAN-006.md

## Goal
Implement exact-money recording and daily navigation.

## Acceptance criteria
- [ ] CRUD recalculates totals
- [ ] duplicate create is safe
- [ ] stale edits conflict
- [ ] daily sample totals 769500.00
- [ ] records outside the request's authorized owner scope return 404.

- [ ] All personal transaction lists, detail, mutations and daily summaries remain actor-owned for both roles.
- [ ] Scoped read methods support one authorized target for Issue 013; wrong-target record/category IDs return 404 and cross-scope cursors fail.
- [ ] Admin reads preserve transaction values, owner IDs, version and soft-deletion behavior.

## Scope exclusions
No wallets, transfers, recurring jobs or offline sync.

## Access-control scope (14 September 2026)
FR-01/FR-15: add the selected-user read adapter in 013; never relax owner predicates in existing transaction methods.
See [access-control.md](../access-control.md).

## Verification
Create/replay/conflicting replay; concurrent save/edit/archive; soft-delete exclusions; empty day; totals fixture; date/cursor boundaries; foreign-user IDs.

Additional authorization verification: Add distinct A/B/C totals, admin's personal CRUD attempts against B, scoped category/record mismatch, cursor swapping and unchanged ledger state after read operations.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
