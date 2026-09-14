# PLAN-002: Database migrations and sqlc foundation

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-002-database-foundation.md
Repository: monelog-api; prerequisites: 001.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Translate database.md into ordered migrations for users (with role)/categories/transactions/sessions/admin_access_events.
2. Configure sqlc and basic scoped queries with exact decimal mapping.
3. Add migration/generation commands and local fixture strategy.
4. Commit-ready generated files and deterministic regeneration drift check.

## Access-control implementation (14 September 2026)
1. Add users.role with default user, NOT NULL and user/admin constraint; use a forward migration for populated fixtures.
2. Add admin_access_events with validated actions/outcomes, nullable resolved target and append-only runtime privileges.
3. Restrict runtime users INSERT/UPDATE columns so only the operator connection can change roles.
4. Generate scoped read methods requiring owner ID and verify two ordinary users plus an admin have independent records.

Policy: [access-control.md](../access-control.md). FR-01/FR-15/FR-16: role schema, role-write privilege boundary and audit records support upcoming admin reads; no OR is_admin owner-filter bypass.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.

Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Real PostgreSQL migration and invalid FK/amount/duplicate tests; regenerate and check clean diff.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.

Authorization validation: Test role default/backfill and invalid role, attempted role update with runtime DB role, audit insert versus update/delete grants, and isolated A/B/admin fixtures using the same scoped queries.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
Do not run destructive migrations on shared databases. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.
