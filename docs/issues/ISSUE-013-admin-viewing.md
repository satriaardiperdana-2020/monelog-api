# ISSUE-013: Admin user selection and read-only finance APIs

Status: Backlog
Repository: monelog-api
Dependencies: 007 (includes 002 role/audit schema, 003 role guard and 004 API contract)
Blocks: 008
Remote issue: Not created
Plan: [PLAN-013](../plans/PLAN-013.md)
Requirements: FR-01, FR-15, FR-16
Policy: [access-control.md](../access-control.md)

## Goal
Allow an authenticated admin to select any existing user and view only that user's income, expenses and reports, while regular users retain access only to their own records.

## Scope
Implement GET /admin/users and GET /admin/users/{user_id} with minimal directory/metadata DTOs.
Implement target-specific category lists, transaction list/detail, daily summaries and report summary/breakdown GETs defined in api.md.
Use existing scoped domain services, current database role checks, required target selection, scope-bound cursors, no-store responses and admin read audit records.
Use selected user's timezone for metadata/presets and preserve exact money/date behavior.

## Acceptance criteria
- [ ] Every admin endpoint checks current role; unauthenticated requests return 401 and non-admins return 403 before target lookup.
- [ ] Directory search is paginated and returns only id,email,timezone,currency; there is no cross-user financial aggregate.
- [ ] Admin C can view A or B explicitly; every list/detail/report returns only that selected owner's active data.
- [ ] Mismatched category/transaction owner returns 404; cross-target/actor/filter cursors return 400.
- [ ] Responses include target scope metadata and Cache-Control: no-store; empty targets have zero totals and empty lists.
- [ ] Successful reads persist minimal audit records first; audit/role storage failure yields 503 without financial data.
- [ ] Admin mutations return 405 after auth/role checks; personal mutation/export/template/backup ownership cannot be bypassed.
- [ ] C's unexpired JWT loses admin access on the next request following a committed demotion.
- [ ] Personal endpoints for A/B/C still return only their own data; target reads do not change finance rows or JWT identity.

## Implementation areas
internal/middleware role guard; internal/handlers admin adapters; internal/service authorization scope; internal/repository and db/queries directory/audit reads/writes; api/openapi.yaml; tests/integration.
Only add a migration if inspection shows the prerequisite schema is missing/incomplete; do not duplicate Issue 002.

## Verification
Run the full A/B/C permission matrix in access-control.md against real PostgreSQL and HTTP handlers.
Test positive viewing, invalid/missing targets, forged role fields, record/category/cursor swapping, revoked role, audit failure, owner-only mutations and aggregate equality for a selected user.
Test route-method fallback so unauthorized mutation attempts cannot reach a personal write handler.
No frontend build is part of this backend task.

## Scope exclusions
No impersonation, public role editor, cross-user mutations/exports/templates/Drive operations, all-user financial report or viewing soft-deleted records.
Frontend user selection belongs to Issue 008; native parity belongs to Issue 010.

## Definition of done
Backend matrix passes with actual recorded results; contract and queries reviewed; no finance mutation through admin reads; Issue 008 can consume the stable API.
Preserve existing task IDs: 013 runs after 007 and before 008.
