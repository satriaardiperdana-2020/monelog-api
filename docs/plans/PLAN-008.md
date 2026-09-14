# PLAN-008: Responsive Vue browser MVP

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-008-web-mvp.md
Repository: monelog-app; prerequisites: 013 (backend Issues 001–007 and 013 complete).

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Initialize Vue 3 JavaScript project with pinned compatible dependencies.
2. Add router, API wrapper, in-memory access auth and browser refresh flow.
3. Build main tabs and day-detail navigation.
4. Build validated entry/edit form and delete confirmation.
5. Add reports presets/custom filters and drilldown lists.
6. Add loading/empty/error/offline explanation and accessibility checks.

## Access-control implementation (14 September 2026)
1. Wait for backend 001–007 and 013, then use GET /me current role for navigation rendering and handle server 403.
2. Build an initially empty Admin View with paginated email search and one-user selector; fetch metadata and finance via explicit admin routes.
3. Display the selected email/read-only label and apply that user's timezone to date presets.
4. Use separate read-only components or capabilities so Add/Edit/Delete/export/share/template/backup controls cannot run in Admin View.
5. Key pending requests/state by actor/mode/target/filters; clear immediately on changes, abort requests and ignore stale generations.
6. Verify role denial/logout and switch back to My Data discard selected-user data without changing JWT sub.

Policy: [access-control.md](../access-control.md). FR-01/FR-15: implement only after backend Issue 013. The UI uses the same session and distinct admin GET routes; selected-user state never overrides personal owner identity.

## Affected areas

Frontend: src views/components/services/stores/router and focused tests; native platform files only for mobile issue.
Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Unit money/date-format tests; component validation; E2E login/create/edit/delete/report; session expiry; 360px and desktop layouts.

Run configured frontend unit/E2E commands and npm run build; establish exact script names from package.json.
Authorization validation: Add E2E ordinary-user direct admin navigation, admin select A/B, delayed A response after switching to B, target timezone presets, logout/login as A, role demotion, read-only controls and 360px selector layout.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
No Capacitor packaging, offline queue, calculator or graph. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.
