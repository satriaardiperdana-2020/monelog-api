# ISSUE-005: Category management

Status: Backlog
Repository: monelog-api
Dependencies: 004
Remote issue: Not created
Plan: ../plans/PLAN-005.md

## Goal
Implement per-user typed categories and archival.

## Acceptance criteria
- [ ] Users can create/rename/archive their categories
- [ ] type immutable
- [ ] archived history readable
- [ ] ownership leaks prevented.

## Scope exclusions
No hard category deletion or shared global category editor.

## Verification
CRUD authorization matrix; duplicate case-insensitive name; archived category visibility; invalid type.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

