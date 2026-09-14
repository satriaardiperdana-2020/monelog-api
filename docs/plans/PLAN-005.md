# PLAN-005: Category management

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-005-categories.md
Repository: monelog-api; prerequisites: 004.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Add owner-scoped category queries and uniqueness handling.
2. Implement create/list/patch services and HTTP adapters.
3. Enforce type immutability, trimmed names and archived-name uniqueness.
4. Test safe read semantics for categories used by historical transactions.

## Access-control implementation (14 September 2026)
1. Keep all create/rename/archive paths tied to actor ID for both roles.
2. Make category read projection and pagination reusable for the future authorized admin target adapter; include no credentials.
3. Test foreign-category and type/owner mismatches without adding an admin write exception.

Policy: [access-control.md](../access-control.md). FR-01/FR-15: admin category display/filtering is read-only through Issue 013; category management remains actor-owned.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.

Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
CRUD authorization matrix; duplicate case-insensitive name; archived category visibility; invalid type.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.

Authorization validation: A cannot read/change B categories on personal paths; admin C cannot change B categories; same-target category query returns only target labels, including archived-history rules.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
No hard category deletion or shared global category editor. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.
