# PLAN-011: Transaction templates

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-011-templates.md
Repository: monelog-api + monelog-app; prerequisites: 010.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Confirm single preset versus whole-day template assumption before implementation.
2. Add template migration and full API contract.
3. Implement owner-scoped template CRUD.
4. Add save-template and apply-template UI with selected date behavior.
5. Handle archived category by requiring replacement before save.

## Access-control implementation (14 September 2026)
1. Bind every template operation to actor ID, including the category FK and the new transaction created after explicit Save.
2. Keep template controls inside My Data; selecting a user for viewing never supplies owner identity to template requests.
3. Test both ordinary and admin actors are denied on another owner's template IDs.

Policy: [access-control.md](../access-control.md). FR-01/FR-15: templates are personal convenience data; another user's visible transactions do not grant access to their templates.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.
Frontend: src views/components/services/stores/router and focused tests; native platform files only for mobile issue.
Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Applying twice creates no records; edited preset save; category mismatch/archive; cross-user access.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.
Run configured frontend unit/E2E commands and npm run build; establish exact script names from package.json.
Authorization validation: Add admin C requesting/applying B template, target tampering on save/apply and hidden template controls in Admin View.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
No recurring auto-created transactions or day bundles unless scope revised. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.
