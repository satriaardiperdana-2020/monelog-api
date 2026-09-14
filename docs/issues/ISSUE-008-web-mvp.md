# ISSUE-008: Responsive Vue browser MVP

Status: Backlog
Repository: monelog-app
Dependencies: 013 (which depends on backend 007)
Remote issue: Not created
Plan: ../plans/PLAN-008.md

## Goal
Build Indonesian-facing core flows from screenshots using Vue JavaScript.

## Acceptance criteria
- [ ] Login, Home, day detail, add/edit/delete, categories, settings and reports work on narrow/wide screens
- [ ] failures preserve form input.

- [ ] Ordinary users see only personal screens; admin routes/selector are unavailable to them.
- [ ] Admins can search/select any one account and see its read-only transactions, daily totals and reports with target identity always visible.
- [ ] Switching targets clears old records/cursors/filters and discards delayed responses from the prior target.
- [ ] Admin View hides mutation, export/share, template and backup/settings actions; logout/role denial clears selected-user state.

## Scope exclusions
No Capacitor packaging, offline queue, calculator or graph.

## Access-control scope (14 September 2026)
FR-01/FR-15: implement only after backend Issue 013. The UI uses the same session and distinct admin GET routes; selected-user state never overrides personal owner identity.
See [access-control.md](../access-control.md).

## Verification
Unit money/date-format tests; component validation; E2E login/create/edit/delete/report; session expiry; 360px and desktop layouts.

Additional authorization verification: Add E2E ordinary-user direct admin navigation, admin select A/B, delayed A response after switching to B, target timezone presets, logout/login as A, role demotion, read-only controls and 360px selector layout.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
