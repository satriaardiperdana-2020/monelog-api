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

## Scope exclusions
No guaranteed app store approval or offline sync; missing signing authority is a blocker.

## Verification
Android/iOS login, save, report and download; restart refresh; offline message; device accessibility smoke test.

## Definition of done
Acceptance tests pass; relevant regressions pass; diff reviewed; documentation/contract updated; actual verification results recorded. No status change before implementation.

