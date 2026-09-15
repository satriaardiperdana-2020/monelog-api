# ISSUE-004: Owner/admin API and soft-delete contract

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api
Dependencies: 003
Remote issue: Not created
Requirements: FR-01, FR-15, FR-16, FR-17
Plan: [PLAN-004](../plans/PLAN-004.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Produce complete OpenAPI schemas and generated interfaces for owner/admin operations and boolean deletion.

## Acceptance criteria
- [ ] Every core personal and mapped admin operation in api.md is an explicit validated OpenAPI path/method.
- [ ] Supported admin POST/PATCH/DELETE/restore operations require current admin permission and selected owner.
- [ ] isDelete is a required boolean response field and strictly boolean list/detail query parameter; no generic lifecycle write field.
- [ ] Versions, 204 delete, restore responses, active/Trash filters and error envelopes validate.
- [ ] Generated interfaces compile and regeneration is deterministic.

## Scope and affected areas
api/openapi.yaml; oapi-codegen configuration; internal/api generated interfaces; contract validation.
No domain handler implementation or unimplemented routes presented as running.

## Verification
JSON examples; required boolean and invalid string/null flag; owner/role/lifecycle overrides; admin mutation operations; missing/stale version schema; generation compile/drift.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
