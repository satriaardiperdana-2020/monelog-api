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

## Scope exclusions
Do not implement transaction handlers in this issue.

## Verification
Validate every example; generated code compilation; contract lint; no unexpected diff after regeneration.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

