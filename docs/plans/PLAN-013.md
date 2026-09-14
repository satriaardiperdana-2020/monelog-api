# PLAN-013: Admin user selection and read-only finance APIs

Status: Draft — refine against repository before implementation.
Issue: [ISSUE-013](../issues/ISSUE-013-admin-viewing.md)
Repository: monelog-api
Prerequisites: 007, transitively 002–006
Blocks: 008
Policy: [access-control.md](../access-control.md)

## Before implementation
Read requirements, architecture, database, API and the full access matrix. Verify roles/audit schema from 002, current-role guard/operator provisioning from 003 and generated contract from 004.
Inspect implemented domain services from 005–007 and current repository instructions. Preserve existing changes; resolve missing prerequisites before coding.
The user has confirmed regular-user ownership and admin visibility. This task interprets admin visibility as read-only selected-user finance access.

## Implementation sequence
1. Add explicit admin GET route group guarded by authentication and a current database role check; keep JWT sub as actor. Add guarded method fallback returning 405.
2. Implement paginated normalized-email directory search using minimal DTOs. Resolve one target by UUID; no target means no financial query.
3. Construct service-level read scope {actor_user_id,target_user_id,mode:admin_read}; never allow raw target input to bypass role checks. Personal and write methods retain actor ownership.
4. Reuse scoped category, transaction, daily and report queries. Enforce target consistency for related IDs and bind cursors to actor/target/endpoint/filters.
5. Add scope metadata, safe errors and no-store headers. Return target timezone/currency through the metadata endpoint; retain existing date/money calculations.
6. Persist minimal allowed admin-read audit events before writing responses. Record authenticated denied/not-found attempts where possible; return 503 without data if required audit or role lookup fails.
7. Implement the A/B/C matrix, including a role demotion with an old JWT, target swapping, report totals and method/write rejection. Verify finance data and owner IDs are unchanged.
8. Record contract diff and test evidence; update issue status only after actual implementation passes. Hand stable API to Issue 008.

## Files to inspect and refine
- internal/middleware: authentication/current-role guard and method fallback.
- internal/handlers: admin directory/finance HTTP adapters.
- internal/service: explicit authorized read scope.
- db/queries and internal/repository: owner-filtered reads, user directory and audit insert.
- api/openapi.yaml and generated interfaces.
- tests/integration: PostgreSQL/HTTP authorization and report fixtures.

Use the repository's actual filenames and query generator setup; these are planned areas, not a claim of existing code.

## Validation
Use A/B regular accounts and C admin, each with distinct records/categories/timezones. Run all cases in access-control.md, including scope/role/cursor tampering and outage behavior.
Run go test ./..., go vet ./... and the configured real-PostgreSQL integration commands. Use a concurrent test for role revocation/next request and audit/response ordering.
Compare target totals to the owner's personal report under identical dates; test empty/soft-deleted records and verify no ledger values/versions change from admin reads.
Compilation and documentation checks alone do not count as security verification.

## Risks and recovery
Wrong target propagation, stale role decisions, permissive fallback routes and delayed frontend responses can expose data. Keep all queries scoped, use read-only service capabilities and fail closed on role/audit errors.
Existing owner-only endpoints remain unchanged in authority. Revert an application rollout through a reviewed prior build if needed; do not delete financial rows or role history.
No production rollout, frontend implementation or user-role mutation is performed by this planning update.

## Review checkpoint
Present actual files, refined sequence and test commands before implementation. Follow the user's current authorization for subsequent actions.
