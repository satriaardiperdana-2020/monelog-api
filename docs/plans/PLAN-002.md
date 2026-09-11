# PLAN-002: Database migrations and sqlc foundation

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-002-database-foundation.md
Repository: monelog-api; prerequisites: 001.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Translate database.md into ordered migrations for users/categories/transactions/sessions.
2. Configure sqlc and basic scoped queries with exact decimal mapping.
3. Add migration/generation commands and local fixture strategy.
4. Commit-ready generated files and deterministic regeneration drift check.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.

Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Real PostgreSQL migration and invalid FK/amount/duplicate tests; regenerate and check clean diff.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
Do not run destructive migrations on shared databases. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.

