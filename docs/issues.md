# Issue index
Updated 14 September 2026 (v0.3) for full admin permissions and boolean soft deletion.
All tasks remain Backlog. These portable task files have no linked GitHub Issues; historical closed repository issues are separate.
Backend 001–007 and 013 complete before monelog-app Issue 008.

| ID | Task | Repository | Depends on | Status | Plan |
| --- | --- | --- | --- | --- | --- |
| ISSUE-001 | [Backend project setup](issues/ISSUE-001-project-setup.md) | monelog-api | None | Backlog | [PLAN-001](plans/PLAN-001.md) |
| ISSUE-002 | [Database migrations and sqlc foundation](issues/ISSUE-002-database-foundation.md) | monelog-api | 001 | Backlog | [PLAN-002](plans/PLAN-002.md) |
| ISSUE-003 | [Authentication, account state and role authorization](issues/ISSUE-003-authentication.md) | monelog-api | 002 | Backlog | [PLAN-003](plans/PLAN-003.md) |
| ISSUE-004 | [Owner/admin API and soft-delete contract](issues/ISSUE-004-api-contract.md) | monelog-api | 003 | Backlog | [PLAN-004](plans/PLAN-004.md) |
| ISSUE-005 | [Category CRUD, Trash and restore](issues/ISSUE-005-categories.md) | monelog-api | 004 | Backlog | [PLAN-005](plans/PLAN-005.md) |
| ISSUE-006 | [Transaction CRUD, soft delete and daily summaries](issues/ISSUE-006-transactions.md) | monelog-api | 005 | Backlog | [PLAN-006](plans/PLAN-006.md) |
| ISSUE-007 | [Active transaction reports](issues/ISSUE-007-reports.md) | monelog-api | 006 | Backlog | [PLAN-007](plans/PLAN-007.md) |
| ISSUE-013 | [Full admin user and data management](issues/ISSUE-013-admin-viewing.md) | monelog-api | 007 | Backlog | [PLAN-013](plans/PLAN-013.md) |
| ISSUE-008 | [Vue owner and admin management UI](issues/ISSUE-008-web-mvp.md) | monelog-app | 013 | Backlog | [PLAN-008](plans/PLAN-008.md) |
| ISSUE-009 | [Owner/admin Excel and PDF exports](issues/ISSUE-009-exports.md) | monelog-api + monelog-app | 008 | Backlog | [PLAN-009](plans/PLAN-009.md) |
| ISSUE-010 | [Android/iOS owner and admin management](issues/ISSUE-010-mobile-packaging.md) | monelog-app | 009 | Backlog | [PLAN-010](plans/PLAN-010.md) |
| ISSUE-011 | [Owner/admin templates with soft delete](issues/ISSUE-011-templates.md) | monelog-api + monelog-app | 010 | Backlog | [PLAN-011](plans/PLAN-011.md) |
| ISSUE-012 | [Owner/admin Drive backup and validated restore](issues/ISSUE-012-drive-backups.md) | monelog-api + monelog-app | 011 | Backlog | [PLAN-012](plans/PLAN-012.md) |

Issue 013 retains its original filename to preserve links, but its scope is full admin management. IDs remain stable; execute in the table's dependency order.

## Changes across the process
| Issues | Permission/lifecycle work |
| --- | --- |
| 001 | Service boundaries and current confirmed policy |
| 002 | is_delete bool/default/backfill, roles, versioned lifecycle, actor attribution and audit schema |
| 003 | Active-account authentication, current-role guards for reads/writes, explicit bootstrap |
| 004 | Concrete owner/admin CRUD contracts, exact isDelete field, Trash/delete/restore and versions |
| 005–007 | Scoped categories/transactions, boolean deletion/restoration, active-only reports |
| 013 | Full admin account/role and selected-user CRUD, audit and authorization suite |
| 008 | Enabled admin forms/actions and own/admin Trash/restore, immutable form target |
| 009 | Selected-owner admin exports and job/download reauthorization |
| 010 | Android/iOS owner/admin CRUD and flag lifecycle parity |
| 011 | Owner/admin template management, soft delete and restore |
| 012 | Owner/admin Drive/backup management, flag-preserving validated restore |

Use [access-control.md](access-control.md) as the permission matrix and [workflow.md](workflow.md) for each task.
A singular duplicate issue.md is unnecessary. If remote issues are created later, add their real URLs to the matching files/index and preserve these portable IDs.
