# ISSUE-004: OpenAPI contract and generation

Status: Backlog
Repository: monelog-api
Dependencies: 003
Remote issue: Not created
Plan: ../plans/PLAN-004.md

## Goal
Turn api.md into a validated canonical contract.

## Acceptance criteria
- [ ] Core endpoints, all error envelopes, examples, security and concurrency headers validate
- [ ] generated interfaces compile.

- [ ] OpenAPI defines GET /admin/users and selected-user metadata/category/transaction/day/report routes with a current-role requirement.
- [ ] Admin responses contain scope metadata; contract specifies 401/403/404/400/405/503 and scope-bound cursors.
- [ ] Role and owner fields are not writable through public personal routes; no cross-user mutation/export/Drive/template contract is exposed.

## Scope exclusions
Do not implement transaction handlers in this issue.

## Access-control scope (14 September 2026)
FR-01/FR-15/FR-16: preserve personal routes as actor-owned and define separate read-only admin routes. Bearer authentication alone does not prove the required current role.
See [access-control.md](../access-control.md).

## Verification
Validate every example; generated code compilation; contract lint; no unexpected diff after regeneration.

Additional authorization verification: Validate personal/admin response examples, read-only role schema, route coverage, required target parameter, unknown override rejection and error schemas; generated interface compiles.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
