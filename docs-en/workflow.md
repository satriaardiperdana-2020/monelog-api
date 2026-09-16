# Step-by-step development workflow
Requirements describe what the app does. An issue defines one deliverable and acceptance criteria. Its plan describes implementation steps and verification. A prompt selects the task and instructions.

1. Read requirements.md and access-control.md: owner-only CRUD, full admin management and isDelete deletion are confirmed.
2. Keep docs canonical in monelog-api. Review repository instructions and existing changes before editing.
3. Use docs/issues.md plus one ISSUE-NNN file and PLAN-NNN per task. Preserve stable IDs across GitHub/Bitbucket.
4. When a GitHub Issue is created, link its actual URL from the task file and preserve the portable ID. ISSUE-002 maps to GitHub #6.
5. Work backend 001–007 → 013, then frontend 008 → 009–012 in dependency order.
6. Inspect the actual repository and refine the chosen plan to exact files, commands and remaining decisions before implementation.
7. Implement the selected issue under the user's current authorization. Keep branch/commit/PR actions within that authorization.
8. Trace actor, owner and requested operation across handler, service, query, response, cache and job payload.
9. For each affected entity, implement is_delete=false defaults, true soft deletion and false restoration with version checks. No physical business-row delete.
10. Run meaningful issue tests, including positive admin operations and regular-user cross-owner denials. Record real commands/results; blocked infrastructure is not a pass.
11. Review the diff for owner predicates, role checks, deletion filters, category history, job target and audit atomicity. Validate contract/generated-code consistency.
12. Complete review/merge as authorized; update the issue and index together when acceptance criteria pass.
13. Proceed to the next unblocked task. A documentation update does not mark implementation Done.

## Planning prompt
Read docs/requirements.md, docs/access-control.md, docs/architecture.md,
docs/database.md, docs/api.md, docs/issues/ISSUE-001-project-setup.md
and docs/plans/PLAN-001.md. Inspect the repository without reading secret files.
Refine this issue's plan to actual paths, implementation steps, test commands
and unresolved decisions. This request is planning only.

## Implementation prompt
Implement the selected issue using its reviewed plan and current access-control rules.
A regular user CRUDs only their own data; an admin may manage any explicitly selected owner.
Use API isDelete / Go IsDelete / SQL is_delete; delete retains rows and restore reactivates them.
Preserve unrelated user changes, do not read secret files, run relevant tests and report actual results.
Follow my current instructions for commit, push and PR actions.

## Review prompt
Review the diff against the selected issue and plan.
Check owner isolation and successful admin CRUD, account/role changes, exact-money totals,
boolean soft deletion and restore races, correct selected-user jobs, audit atomicity and missing tests.
Report findings with file references and severity. This request is review only.

Replace ISSUE-001 and PLAN-001 with the task being worked. Frontend work uses monelog-app and a pinned backend API/permission specification.
Later Jira/Bitbucket mapping retains the same portable ID. Choose one status owner and mirror deliberately; no automatic synchronization is claimed.

## Authorization checklist for implementation
Personal owner=actor. Admin owner=verified target, with an active current admin and supported action. Admin writes are authorized; do not retain obsolete blanket cross-user write denials.
Every query and mutation keeps a single financial owner. New admin-created data belongs to the target, with actor attribution.
Lifecycle status is a boolean, not physical erasure. Active totals use transactions.is_delete=false and do not disappear when a historical category is deleted.
Forms, pending writes, exports, templates, schedules and restore jobs bind immutable target context. Never switch an in-flight operation to a newly selected owner.
Use A/B regular accounts and C admin for both allowed and denied paths; include Trash, restore, stale versions, job reauthorization, deleted accounts and audit failure.
