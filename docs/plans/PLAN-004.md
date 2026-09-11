# PLAN-004: OpenAPI contract and generation

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-004-api-contract.md
Repository: monelog-api; prerequisites: 003.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Reconcile draft routes with implemented auth and document choices.
2. Write complete OpenAPI schema for MVP endpoints and planned export endpoints marked release scope.
3. Add all examples, decimal patterns, pagination, status responses and If-Match.
4. Configure oapi-codegen and compilation/regeneration checks.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.

Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Validate every example; generated code compilation; contract lint; no unexpected diff after regeneration.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
Do not implement transaction handlers in this issue. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.

