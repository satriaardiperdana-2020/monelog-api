# Issue index

Updated 14 September 2026 for FR-01/FR-15. Canonical portable task IDs. All statuses are Backlog; plans are drafts, not completed code. These 13 task files have no linked GitHub Issues; historical closed repository issues are separate.

| ID | Task | Repository | Depends on | Status | Plan |
| --- | --- | --- | --- | --- | --- |
| ISSUE-001 | [Backend project setup](issues/ISSUE-001-project-setup.md) | monelog-api | None | Backlog | [PLAN-001](plans/PLAN-001.md) |
| ISSUE-002 | [Database migrations and sqlc foundation](issues/ISSUE-002-database-foundation.md) | monelog-api | 001 | Backlog | [PLAN-002](plans/PLAN-002.md) |
| ISSUE-003 | [Authentication, roles and profile](issues/ISSUE-003-authentication.md) | monelog-api | 002 | Backlog | [PLAN-003](plans/PLAN-003.md) |
| ISSUE-004 | [OpenAPI contract and generation](issues/ISSUE-004-api-contract.md) | monelog-api | 003 | Backlog | [PLAN-004](plans/PLAN-004.md) |
| ISSUE-005 | [Category management](issues/ISSUE-005-categories.md) | monelog-api | 004 | Backlog | [PLAN-005](plans/PLAN-005.md) |
| ISSUE-006 | [Transaction CRUD and daily summaries](issues/ISSUE-006-transactions.md) | monelog-api | 005 | Backlog | [PLAN-006](plans/PLAN-006.md) |
| ISSUE-007 | [Summary and breakdown reports](issues/ISSUE-007-reports.md) | monelog-api | 006 | Backlog | [PLAN-007](plans/PLAN-007.md) |
| ISSUE-013 | [Admin user selection and read-only finance APIs](issues/ISSUE-013-admin-viewing.md) | monelog-api | 007 | Backlog | [PLAN-013](plans/PLAN-013.md) |
| ISSUE-008 | [Responsive Vue browser MVP](issues/ISSUE-008-web-mvp.md) | monelog-app | 013 | Backlog | [PLAN-008](plans/PLAN-008.md) |
| ISSUE-009 | [Excel and PDF report exports](issues/ISSUE-009-exports.md) | monelog-api + monelog-app | 008 | Backlog | [PLAN-009](plans/PLAN-009.md) |
| ISSUE-010 | [Android and iOS packaging](issues/ISSUE-010-mobile-packaging.md) | monelog-app | 009 | Backlog | [PLAN-010](plans/PLAN-010.md) |
| ISSUE-011 | [Transaction templates](issues/ISSUE-011-templates.md) | monelog-api + monelog-app | 010 | Backlog | [PLAN-011](plans/PLAN-011.md) |
| ISSUE-012 | [Scheduled Google Drive backups and restore validation](issues/ISSUE-012-drive-backups.md) | monelog-api + monelog-app | 011 | Backlog | [PLAN-012](plans/PLAN-012.md) |

Use one issue file per task, plus this index. A second singular issue.md would duplicate the same source of truth. Each local file can become one GitHub Issue later. Append the real remote link only after creation.

## Access-control work across issues
Read [access-control.md](access-control.md) for the permission matrix and test fixtures.

| Issues | Required change |
| --- | --- |
| 001 | Preserve auth/service boundaries; no role authority from client configuration |
| 002 | role=user default/check, admin audit schema, explicit owner queries and DB grants |
| 003 | Server-assigned roles, current-role admin guard and operator-only role provisioning |
| 004 | Complete personal/admin contract, selected-user scope and denial responses |
| 005–007 | Owner-scoped domain operations; queries reusable only with authorized scope |
| 013 | Admin directory, selected-user GETs, auditing and cross-role integration suite |
| 008 | Browser user selector and read-only mode with request/cache isolation |
| 009 | Personal exports only; selection cannot redirect ownership |
| 010 | Android/iOS selector parity and session/selection cleanup |
| 011–012 | Personal templates, Drive jobs and restore stay owner-scoped |

Issue 013 is appended to preserve all existing IDs, but it runs after 007 and before 008. Backend first remains mandatory.
