# ISSUE-010: Android and iOS packaging

Status: Backlog
Repository: monelog-app
Dependencies: 009
Remote issue: Not created
Plan: ../plans/PLAN-010.md

## Goal
Package shared Vue app with Capacitor and verify platform behavior.

## Acceptance criteria
- [ ] Both platform builds tested on device/simulator
- [ ] auth secure storage and export open/share work
- [ ] offline behavior honest.

- [ ] Android/iOS use the same user/admin permission matrix and initially empty admin selector as browser.
- [ ] Target switching, navigation/back, logout and demotion never display a prior target's cached financial data.
- [ ] Selected-user data is not persisted offline; native sharing/export remains personal.

## Scope exclusions
No guaranteed app store approval or offline sync; missing signing authority is a blocker.

## Access-control scope (14 September 2026)
FR-01/FR-15: reuse the tested Vue admin viewer with server enforcement; packaging does not create more admin permissions.
See [access-control.md](../access-control.md).

## Verification
Android/iOS login, save, report and download; restart refresh; offline message; device accessibility smoke test.

Additional authorization verification: Add real-device/simulator A/B/C selector, fast switch with delayed response, app restart, logout/relogin, deep-link role denial and native export/share ownership checks.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.
