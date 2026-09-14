# ISSUE-001: Backend project setup

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api
Dependencies: None
Remote issue: Not created
Requirements: FR-01, FR-15, FR-16, FR-17
Plan: [PLAN-001](../plans/PLAN-001.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Create a reproducible Go/Echo service skeleton with clear authentication/service/repository boundaries.

## Acceptance criteria
- [ ] Service starts using documented commands and shuts down gracefully.
- [ ] Live and readiness checks differ correctly, including unavailable PostgreSQL.
- [ ] Missing required config fails clearly; examples contain no production credentials.
- [ ] README links the full admin and boolean deletion specification; no automatic first-user admin.

## Scope and affected areas
cmd/api; internal/config/handlers/middleware/service/repository; README and CI.
No finance/auth implementation, migrations or deployment in this setup task.

## Verification
Configuration errors; liveness; readiness with unavailable DB; graceful shutdown. Confirm role/owner authority is not configured from public client input.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
