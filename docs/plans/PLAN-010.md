# PLAN-010: Android and iOS packaging

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-010-mobile-packaging.md
Repository: monelog-app; prerequisites: 009.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Check official Capacitor/toolchain requirements and available macOS signing environment.
2. Add platform projects/config with HTTPS API and allowed origins.
3. Implement vetted secure token storage and platform-specific export handoff.
4. Verify keyboard/date input, navigation/back, safe areas and session expiry.
5. Document signing/release steps and secrets exclusion.

## Access-control implementation (14 September 2026)
1. Carry role-aware navigation and explicit admin routes from the tested web implementation without storing target identity in token storage.
2. Verify back navigation, restart and resume clear or reauthorize selected-user view; show no financial data until that check succeeds.
3. Exclude admin responses from persistent caches and native share/export paths; test late responses on each platform.
4. Run A/B/C switching and role-demotion flows on Android and iOS as release gates.

Policy: [access-control.md](../access-control.md). FR-01/FR-15: reuse the tested Vue admin viewer with server enforcement; packaging does not create more admin permissions.

## Affected areas

Frontend: src views/components/services/stores/router and focused tests; native platform files only for mobile issue.
Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Android/iOS login, save, report and download; restart refresh; offline message; device accessibility smoke test.

Run configured frontend unit/E2E commands and npm run build; establish exact script names from package.json.
Authorization validation: Add real-device/simulator A/B/C selector, fast switch with delayed response, app restart, logout/relogin, deep-link role denial and native export/share ownership checks.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
No guaranteed app store approval or offline sync; missing signing authority is a blocker. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.
