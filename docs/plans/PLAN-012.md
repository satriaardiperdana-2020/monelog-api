# PLAN-012: Scheduled Google Drive backups and restore validation

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-012-drive-backups.md
Repository: monelog-api + monelog-app; prerequisites: 011.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Design full backup format/schema version, schedule/timezone, retention and restore policy; obtain confirmation.
2. Check current official OAuth scopes/policies and select minimal required access.
3. Implement encrypted credentials, reconnect/revocation and safe consent flow.
4. Add job tables, unique scheduled runs, retry/backoff and status UI.
5. Build versioned data backup with checksums, validate archive and restore into disposable isolated DB.
6. Document operational recovery separately from personal backups.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.
Frontend: src views/components/services/stores/router and focused tests; native platform files only for mobile issue.
Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Expired/revoked credentials; duplicate scheduler runs; cross-user isolation; corrupt backup; restore count/total parity; timezone schedule edge.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.
Run configured frontend unit/E2E commands and npm run build; establish exact script names from package.json.
Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
No unattended overwrite of production data; destructive restore needs explicit confirmation. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.

