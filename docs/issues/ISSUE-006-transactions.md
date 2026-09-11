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
- [ ] foreign records return 404.

## Scope exclusions
No wallets, transfers, recurring jobs or offline sync.

## Verification
Create/replay/conflicting replay; concurrent save/edit/archive; soft-delete exclusions; empty day; totals fixture; date/cursor boundaries; foreign-user IDs.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

