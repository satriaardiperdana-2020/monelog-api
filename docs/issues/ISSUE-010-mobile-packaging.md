# ISSUE-010: Android/iOS owner and admin management

Status: Backlog
Updated: 14 September 2026 (v0.3)
Repository: monelog-app
Dependencies: 009
Remote issue: Not created
Requirements: FR-01, FR-09, FR-15, FR-17
Plan: [PLAN-010](../plans/PLAN-010.md)
Policy: [Access control and soft deletion](../access-control.md)

## Goal
Package the tested Vue app with Capacitor and preserve full admin/owner CRUD and boolean lifecycle behavior on devices.

## Acceptance criteria
- [ ] Android and iOS builds run with secure auth storage, API access and authorized export opening/sharing.
- [ ] Admin management actions and personal CRUD/Trash/Restore match browser behavior.
- [ ] Target switching, restart, resume and logout do not expose previous-owner records or redirect mutations.
- [ ] Native file sharing reuses correct authorized export owner; other-user data is not stored in offline caches.

## Scope and affected areas
Capacitor and native platform projects; auth/API/download adapters; shared Vue views; device smoke/E2E checks.
No guaranteed app-store approval or offline synchronization. Missing platform/signing infrastructure is a stated implementation blocker.

## Verification
A/B/C login and positive admin CRUD; delete/restore on device; stale form target; late response; app resume/restart; account deletion/demotion; scoped native export/share; accessibility.

## Definition of done
Acceptance criteria pass with actual test evidence; contract/schema/docs and affected generated code are consistent; diff reviewed; relevant regressions pass.
Document unavailable infrastructure explicitly. A plan or documentation update does not complete this issue.
