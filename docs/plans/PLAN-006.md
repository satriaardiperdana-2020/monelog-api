# PLAN-006: Transaction CRUD and daily summaries

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-006-transactions.md
Repository: monelog-api; prerequisites: 005.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Add strict decimal parsing/date/title validation and category locking.
2. Implement create idempotency keyed by user and canonical payload.
3. Implement optimistic version update/soft delete.
4. Implement filtered transaction lists, stable cursors and daily summaries.
5. Wire generated handlers and transaction boundaries.

## Access-control implementation (14 September 2026)
1. Keep owner=actor for personal CRUD, including idempotency keys and versioned mutation predicates.
2. Pass explicit owner scope to list/detail/day query methods; include route/actor/target/filter identity in cursor validation.
3. Prepare two-target fixture assertions for 013, including mixed owner IDs and soft-deleted rows; use no all-users financial query.

Policy: [access-control.md](../access-control.md). FR-01/FR-15: add the selected-user read adapter in 013; never relax owner predicates in existing transaction methods.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.

Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Create/replay/conflicting replay; concurrent save/edit/archive; soft-delete exclusions; empty day; totals fixture; date/cursor boundaries; foreign-user IDs.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.

Authorization validation: Add distinct A/B/C totals, admin's personal CRUD attempts against B, scoped category/record mismatch, cursor swapping and unchanged ledger state after read operations.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
No wallets, transfers, recurring jobs or offline sync. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.
