# ISSUE-013: Full admin user and data management

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-api
Dependencies: 007
Remote issue: Not created
Requirements: FR-01, FR-15, FR-16, FR-17
Plan: [PLAN-013](../plans/PLAN-013.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Allow current admins to perform all supported application operations on any selected user's data, including CRUD and soft-delete restoration.

## Acceptance criteria
- [ ] Non-admin admin-route attempts return 403 before target lookup; current admins can perform supported reads and mutations.
- [ ] User directory/create/profile/role/delete/restore and audit listing work under admin authorization.
- [ ] Selected-owner category/transaction CRUD, Trash/restore, daily totals and reports work, preserving owner and actor attribution.
- [ ] Wrong-target related IDs/cursors are rejected; record lifecycle uses isDelete booleans and version guards.
- [ ] Admin writes and audit events commit atomically; failed audit rolls back mutations.
- [ ] Role/account changes serialize with data mutations and invalidate relevant sessions/authority.
- [ ] Contracts and shared policy authorize cross-user exports/templates/backups, implemented in their later feature issues.

## Scope and affected areas
Admin handlers; middleware/current role; authorization service; user/category/transaction/report/audit queries; OpenAPI; real PostgreSQL/HTTP suite.
No automatic production role change during planning, no physical business-row deletion, and no raw credential responses. Existing filename retains its historical slug for link compatibility.

## Verification
A/B denial versus C positive full CRUD; C-created record owner B; role promotion/demotion; user boolean deletion/restoration; active/Trash; version races; category mismatch; audit rollback; selected-user totals; expired/deleted actor.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
This task blocks Issue 008 and runs after 007 despite its stable ID/filename.
