# ISSUE-002: Database migrations and sqlc foundation

Status: Backlog
Repository: monelog-api
Dependencies: 001
Remote issue: Not created
Plan: ../plans/PLAN-002.md

## Goal
Implement the core schema and repeatable SQL generation.

## Acceptance criteria
- [ ] Migrations work on an empty disposable database
- [ ] ownership/type FK and amount constraints enforced
- [ ] generated output is reproducible.

## Scope exclusions
Do not run destructive migrations on shared databases.

## Verification
Real PostgreSQL migration and invalid FK/amount/duplicate tests; regenerate and check clean diff.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

