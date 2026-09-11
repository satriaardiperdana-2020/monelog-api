# PLAN-009: Excel and PDF report exports

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-009-exports.md
Repository: monelog-api + monelog-app; prerequisites: 008.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Add export_jobs schema, ownership, bounded workload and job lifecycle.
2. Reuse report filters and stream/paginate all matching source records using a consistent snapshot.
3. Generate XLSX/PDF with dates, currency, totals and safe user text.
4. Add authenticated downloads, artifact retention cleanup and UI progress/retry.
5. Document job execution and operational limits.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.
Frontend: src views/components/services/stores/router and focused tests; native platform files only for mobile issue.
Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Over-page-size export; report parity; spreadsheet formula-injection safety; PDF layout/long title check; job failure/retry; expiry and foreign access.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.
Run configured frontend unit/E2E commands and npm run build; establish exact script names from package.json.
Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
Not a full backup or import mechanism. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.

