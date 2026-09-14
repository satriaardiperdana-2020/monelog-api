# Monelog planning pack
Draft v0.3 • 14 September 2026 • Specification and task plans; application implementation is pending.

## Confirmed behavior
- Regular users can CRUD only their own data.
- Admins can perform all application operations on any selected user's data, including CRUD, roles, reports, exports, templates and backups.
- Soft deletion uses isDelete=false for active records and true for deleted records. Delete retains the row; restore changes the flag back to false.

## Documents
1. [Requirements](docs/requirements.md)
2. [Access control and soft deletion](docs/access-control.md)
3. [Architecture](docs/architecture.md)
4. [Database](docs/database.md)
5. [API](docs/api.md)
6. [Roadmap](docs/roadmap.md)
7. [Issue index](docs/issues.md)
8. [Workflow](docs/workflow.md)

Start [ISSUE-001](docs/issues/ISSUE-001-project-setup.md) with [PLAN-001](docs/plans/PLAN-001.md).
Finish backend 001–007 and [ISSUE-013](docs/issues/ISSUE-013-admin-viewing.md) before frontend Issue 008.
Issue 013 now covers full admin management; its original filename is retained to preserve links.
Then deliver exports, Android/iOS, templates and Drive backups in their listed milestones.
Task Markdown files are the portable source of truth and have no linked GitHub Issues yet. Keep docs canonical in monelog-api; monelog-app references a pinned documentation/API commit.
