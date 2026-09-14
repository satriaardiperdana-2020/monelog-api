# PLAN-001: Backend project setup

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-001](../issues/ISSUE-001-project-setup.md)
Repository: monelog-api
Prerequisites: None
Requirements: FR-01, FR-15, FR-16, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Inspect repository instructions and current Go/dependency constraints; pin compatible versions.
2. Create cmd/api and internal/config, handlers, middleware, service and repository boundaries.
3. Add graceful shutdown, recovery, safe request logging, timeouts and health endpoints.
4. Document placeholder-only configuration and local PostgreSQL startup.
5. Add focused config/health checks and CI formatting, vet/lint and tests.

## Affected areas
cmd/api; internal/config/handlers/middleware/service/repository; README and CI.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
Configuration errors; liveness; readiness with unavailable DB; graceful shutdown. Confirm role/owner authority is not configured from public client input.
Run go test ./..., go vet ./... and the configured PostgreSQL/HTTP integration suite where relevant. Verify generated-code drift for changed SQL/OpenAPI sources.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No finance/auth implementation, migrations or deployment in this setup task.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
