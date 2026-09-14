# Step-by-step issue workflow
Planning describes HOW an issue will be implemented. An issue describes WHAT must be achieved and acceptance criteria. A prompt tells an agent which artifacts and boundaries to follow. Planning is not coding.

1. Read requirements and access-control.md. User-only ownership and admin viewing are confirmed; review remaining flagged product choices.
2. Place docs in monelog-api; commit only after reviewing the files. Never include secrets or screenshot financial details unnecessarily.
3. Keep docs/issues.md as the index; one docs/issues/ISSUE-NNN-name.md per task. Do not add a duplicate issue.md.
4. Optionally create one GitHub Issue from each local issue. Add its URL/number to the index; retain ISSUE-NNN as the portable ID. These portable task files have no linked GitHub Issues; do not reuse historical issue numbers.
5. Work in dependency order: 001–007, then 013, then 008–012. Open ISSUE-001 and PLAN-001 first; finish the admin backend before monelog-app.
6. Ask the agent to inspect the actual repository and refine the plan only. This pack's plans are pre-repository drafts.
7. Review affected files, dependencies, migration/data risks, acceptance tests and exclusions. Approve the plan before code.
8. Create a task branch such as feature/ISSUE-001-project-setup. Do not reset an existing repository.
9. Implement only that issue. Update docs if approved design changes.
10. Run issue-specific tests and relevant regression tests; record actual commands/results, not assumed passes.
11. Review diff for security, finance correctness, generated-code drift and scope. Open PR when explicitly requested.
12. Merge after checks/review; mark local issue and index Done together; close the linked GitHub issue if used.
13. Continue to the next unblocked issue. Never mark a task Done merely because its plan exists.

## Planning prompt
Read docs/requirements.md, docs/access-control.md, docs/architecture.md, docs/database.md, docs/api.md,
docs/issues/ISSUE-001-project-setup.md and docs/plans/PLAN-001.md.
Inspect the repository without reading secret files. Refine PLAN-001 with actual paths,
steps, test commands, risks and exclusions. Do not implement code or modify remote systems.
Stop after presenting the plan for review.

## Implementation prompt
Implement only ISSUE-001 using the reviewed PLAN-001.
Preserve existing user changes and do not read .env or secret files.
Run relevant tests, report real results and summarize the diff.
Do not commit, push or open a PR unless I explicitly request it.

## Review prompt
Review the diff against ISSUE-001 acceptance criteria and PLAN-001.
Prioritize correctness, the user/admin access matrix, selected-user read scope, owner-only writes, decimal precision, migrations and missing tests.
Report findings with file references and severity. Do not change code.

Replace 001 and the filenames with the selected issue. For frontend issues work in monelog-app and reference the pinned backend API contract.
If using Jira later, map the same local stable ID to a Jira key. Choose one status owner (initially local docs) and mirror status deliberately; do not claim automatic synchronization.

## Authorization work for every issue
Identify actor_user_id and financial owner before touching a query, cache key or job payload.
On personal routes owner equals actor. On admin GET routes owner equals the explicit target only after the current-role guard. Never extend admin viewing into other users' mutations or personal exports/templates/backups.
Add acceptance tests for regular users A/B and admin C where the issue touches data. Test valid admin views as well as denial cases; an unconditional cross-user denial would now fail FR-15.
Review routes, SQL owner predicates, DTOs, pagination, async responses, workers and mobile caches together. Attach real test results to the task; keep Backlog until implementation.
Run backend 001–007 → 013 before frontend 008. Implement the selector against that tested API and reject responses whose scope no longer matches the current selection.
The docs are the canonical specification in monelog-api; frontend work uses a pinned backend documentation/contract commit.
