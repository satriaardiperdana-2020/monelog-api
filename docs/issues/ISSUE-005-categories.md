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

- [ ] Personal category mutations are owner-only for ordinary users and admins.
- [ ] Category read queries can accept a service-authorized target for Issue 013 without returning other owners' categories.

## Scope exclusions
No hard category deletion or shared global category editor.

## Access-control scope (14 September 2026)
FR-01/FR-15: admin category display/filtering is read-only through Issue 013; category management remains actor-owned.
See [access-control.md](../access-control.md).

## Verification
CRUD authorization matrix; duplicate case-insensitive name; archived category visibility; invalid type.

Additional authorization verification: A cannot read/change B categories on personal paths; admin C cannot change B categories; same-target category query returns only target labels, including archived-history rules.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
