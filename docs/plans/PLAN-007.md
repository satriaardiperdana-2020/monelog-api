# PLAN-007: Active transaction reports

Status: Draft — refine against actual repository before implementation.
Updated: 14 September 2026 (v0.3)
Issue: [ISSUE-007](../issues/ISSUE-007-reports.md)
Repository: monelog-api
Prerequisites: 006
Requirements: FR-01, FR-06, FR-07, FR-15, FR-17

## Before implementation
Read requirements.md, access-control.md, architecture.md, database.md, api.md and the linked issue.
Check existing code/instructions and user changes; confirm prerequisite tasks are complete. Use actual repository filenames and commands during refinement.
Confirmed policy: regular users CRUD only their own data; admin can manage any selected user's data. isDelete=false means active; true means soft-deleted.

## Implementation sequence
1. Implement explicit-owner and active-row report filters shared with later exports.
2. Define Monday/month buckets and deterministic top-five category sorting.
3. Keep historical category joins without applying deleted-category exclusion to transactions.
4. Wire personal report handlers and reusable scoped service for 013.
5. Check performance on representative data and validate true/false transitions against source records.

## Affected areas
Report handlers/service; aggregation SQL; date/decimal helpers; integration fixtures.
These are planned areas. Narrow them to exact files during repository inspection; do not edit unrelated modules.

## Validation
Weekly/monthly/year/leap-date edges; inclusivity; partial buckets; deleted transaction exclusion; restored inclusion once; deleted category history; A/B/admin totals; target timezone; empty/benchmark checks.
Run go test ./..., go vet ./... and the configured PostgreSQL/HTTP integration suite where relevant. Verify generated-code drift for changed SQL/OpenAPI sources.
Record real commands/results. Do not claim runtime authorization or lifecycle correctness from documentation review alone.

## Authorization and lifecycle review
Trace actor, selected owner, action, version and isDelete state through every affected boundary.
Personal operations use actor ownership; admin operations use authorized target ownership. Keep category/resource/cursor/job scope consistent.
For admin writes, validation and audit must succeed with the data transaction. For external jobs, record authorization/job/audit before provider work and revalidate at execution.
Active financial totals exclude true transactions. Trash/restore retain owner constraints. Never infer physical deletion or role changes from imported financial data.
Apply only the checks relevant to this issue's actual scope.

## Scope and recovery
No graph UI, stored balance column or report totals that include Trash.
Preserve user changes. Test schema changes against disposable fixtures and use reviewed forward recovery before rollout; do not erase shared rows.
After implementation, compare every acceptance criterion with evidence, then update issue/index status under the user's current workflow authorization.

## Review checkpoint
Present the refined file scope, steps, unresolved decisions and tests before coding, unless the user has already authorized implementation.
