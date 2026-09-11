# PLAN-001: Backend project setup

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-001-project-setup.md
Repository: monelog-api; prerequisites: None.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Inspect existing tree and Go constraints; pin compatible versions using official docs.
2. Create cmd/api, internal/config, handlers, middleware, service and repository boundaries.
3. Add graceful shutdown, recovery, safe request logging and health endpoints.
4. Add example configuration with placeholders and local PostgreSQL setup instructions.
5. Add CI formatting/test/lint checks and README startup steps.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.

Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Config validation tests; HTTP live test; readiness with unavailable DB; startup/shutdown smoke test.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
No auth, migrations, finance logic, deployment or remote repo changes. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.

