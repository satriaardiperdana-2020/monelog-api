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

- [ ] Template read/create/update/apply paths are owner-only for ordinary users and admins.
- [ ] Admin View cannot read/apply/save a selected user's templates or create a transaction for that user.

## Scope exclusions
No recurring auto-created transactions or day bundles unless scope revised.

## Access-control scope (14 September 2026)
FR-01/FR-15: templates are personal convenience data; another user's visible transactions do not grant access to their templates.
See [access-control.md](../access-control.md).

## Verification
Applying twice creates no records; edited preset save; category mismatch/archive; cross-user access.

Additional authorization verification: Add admin C requesting/applying B template, target tampering on save/apply and hidden template controls in Admin View.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
