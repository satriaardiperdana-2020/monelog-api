# Monelog planning pack
Draft v0.3 • 14 September 2026 • Specification and task plans; application implementation is pending.

## Language / Bahasa

- [English](README.en.md)
- [Bahasa Indonesia](README.md)

## Confirmed behavior
- Regular users can CRUD only their own data.
- Admins can perform all application operations on any selected user's data, including CRUD, roles, reports, exports, templates and backups.
- Soft deletion uses isDelete=false for active records and true for deleted records. Delete retains the row; restore changes the flag back to false.

## Documents
1. [Requirements](docs-en/requirements.md)
2. [Access control and soft deletion](docs-en/access-control.md)
3. [Architecture](docs-en/architecture.md)
4. [Database](docs-en/database.md)
5. [API](docs-en/api.md)
6. [Roadmap](docs-en/roadmap.md)
7. [Issue index](docs-en/issues.md)
8. [Workflow](docs-en/workflow.md)

Start [ISSUE-001](docs-en/issues/ISSUE-001-project-setup.md) with [PLAN-001](docs-en/plans/PLAN-001.md).
Finish backend 001–007 and [ISSUE-013](docs-en/issues/ISSUE-013-admin-viewing.md) before frontend Issue 008.
Issue 013 now covers full admin management; its original filename is retained to preserve links.
Then deliver exports, Android/iOS, templates and Drive backups in their listed milestones.
Task Markdown files are the portable source of truth and have no linked GitHub Issues yet. Keep docs canonical in monelog-api; monelog-app references a pinned documentation/API commit.
