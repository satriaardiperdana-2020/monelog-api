# PLAN-007: Summary and breakdown reports

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-007-reports.md
Repository: monelog-api; prerequisites: 006.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Implement shared filter validation and aggregation queries.
2. Define partial-period buckets and deterministic top-five rankings.
3. Implement summary and breakdown endpoints.
4. Benchmark representative data and inspect query plans.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.

Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Week/month/year/leap-day edges; date inclusivity; partial buckets; income minus expense; ownership and empty sets; benchmark target.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
No graph UI or stored balances. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.

