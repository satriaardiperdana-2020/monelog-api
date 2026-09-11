# ISSUE-011: Transaction templates

Status: Backlog
Repository: monelog-api + monelog-app
Dependencies: 010
Remote issue: Not created
Plan: ../plans/PLAN-011.md

## Goal
Add single-transaction presets after confirming screenshot behavior.

## Acceptance criteria
- [ ] Applying preset opens unsaved editable form
- [ ] only Save creates transaction
- [ ] ownership and archived categories checked.

## Scope exclusions
No recurring auto-created transactions or day bundles unless scope revised.

## Verification
Applying twice creates no records; edited preset save; category mismatch/archive; cross-user access.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

