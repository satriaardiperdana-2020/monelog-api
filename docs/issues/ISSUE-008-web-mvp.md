# ISSUE-008: Vue owner and admin management UI

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-app
Dependencies: 013
Remote issue: Not created
Requirements: FR-01, FR-02, FR-03, FR-04, FR-05, FR-06, FR-07, FR-11, FR-15, FR-17
Plan: [PLAN-008](../plans/PLAN-008.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Build responsive Indonesian-facing personal finance and full admin management flows using Vue 3 + JavaScript.

## Acceptance criteria
- [ ] Own login/Home/day/detail/category/transaction/settings/reports work on mobile-width and desktop.
- [ ] Admin can search/select a user and use enabled create/edit/delete/restore actions plus account/role management.
- [ ] Active and Trash screens use false/true correctly; reports remain active-only.
- [ ] Forms and confirmations show fixed owner identity; unsaved edits require save/discard before selection change.
- [ ] Late responses or submitted operations never apply to a newly selected owner; logout/denial clears state.
- [ ] Failed requests preserve appropriate inputs and display clear validation/conflict feedback.

## Scope and affected areas
src/views/components/services/stores/router/utils; focused unit/component/E2E tests.
No Capacitor packaging, offline write queue, calculator or graph in browser MVP.

## Verification
Login/refresh; owner CRUD; positive admin CRUD/roles; A direct admin denial; active versus Trash; delete/restore totals; delayed responses; unsaved form target switch; logout/relogin; 360px/desktop accessibility.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
