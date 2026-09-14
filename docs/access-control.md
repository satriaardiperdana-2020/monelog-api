# Access control
Version 0.2 • 14 September 2026
User instruction: regular users see only their own income/expenses; an admin may select any user and view that user's data.
FR-01 and FR-15 are confirmed requirements. Read-only admin scope and the technical controls below are the implementation interpretation.

## Permission matrix
| Capability | Regular user | Admin in My Data | Admin viewing selected user |
| --- | --- | --- | --- |
| View transactions, daily totals, category labels and reports | Own only | Own only | Selected user only |
| List/search registered users | Denied | Admin View directory only | Minimal directory permitted |
| Create/edit/delete transactions or categories | Own only | Own only | Denied |
| Change profile/timezone | Own only | Own only | Denied |
| Excel/PDF export creation, status or download | Own only | Own only | Denied |
| Template read/create/apply/update | Own only | Own only | Denied |
| Connect Drive, schedule backup or restore | Own only | Own only | Denied |
| Assign roles through app/API | Denied | Denied | Denied |
| Read passwords, session secrets or provider credentials | Denied | Denied | Denied |

Admins may choose any existing account, including another admin or themselves. Admin View stays read-only even when selecting self; use My Data for own actions.
No aggregate across all users is included. Exports/sharing of another user's data are outside this viewing request.

## Policy
- Actor is the authenticated account. Target is the selected financial owner. Never mutate JWT sub or impersonate the target.
- Every ordinary route uses actor ownership, even for an admin. Do not add an optional user_id override.
- Only /admin/users/{user_id}/... GET routes can use another owner, after checking the actor's current role in PostgreSQL. /admin/users is a minimal directory exception, containing no financial data.
- Require explicit target selection; every financial query retains owner filtering. Record IDs, category IDs and cursors must belong to that same scope.
- Registration defaults to user; reject client-supplied roles. A trusted deployment operator provisions an admin explicitly through cmd/admin; there is no first-user auto-promotion or public role editor.
- Role changes revoke refresh sessions. The next admin request after committed demotion is denied even with an unexpired JWT. Do not authorize from cached frontend state or a stale JWT role claim.
- Missing authentication → 401; non-admin admin-route attempt → 403 before target lookup; missing/mismatched target resource → 404; tampered/mismatched cursor → 400; unsupported admin write method → 405 after auth/role guard.
- No active data returns for soft-deleted transactions, even to admins. Owner/category constraints remain enforced.
- Personal jobs and backup/restore data never inherit a selected admin target. Imports cannot assign roles or trust source owner IDs.

## Screen behavior
Regular users see only My Data and cannot open the admin selector.
Admins open Admin View, search/select one account, then see its dated records and reports with email and read-only label always visible. Use selected account timezone for presets.
Initial selection is empty. Switch A → B: clear A immediately, reset cursors/category filters, abort old requests and ignore late responses through a request generation check.
Keep selected target and returned data in memory, keyed by actor/mode/target/filters. Send no financial request when target is empty.
Logout, role denial and switching back to My Data discard selected-user state. Reauthentication must not restore another user's cached records.
Hide transaction/category mutations, exports/sharing, templates and backup/settings actions in Admin View. UI controls supplement backend enforcement.

## Audit
Persist actor, resolved target if any, action, optional resource UUID, outcome, request ID and UTC time. Never log transaction titles/amounts, emails/search text, tokens or full payloads.
Directory reads have null target. Non-admin attempts are denied before target resolution and recorded with null target where audit storage is available.
Successful admin reads require persisted audit records before responding; otherwise return 503 without data. The operator can inspect audit records; no new app audit UI is included.
All authenticated API responses use Cache-Control: no-store, including errors.

## Required acceptance tests
Use regular users A/B and admin C with deliberately different income, expense, category and date fixtures.

| Case | Expected |
| --- | --- |
| A lists own data/reports | Only A's active records and exact totals |
| A submits B's UUID on personal GET/PATCH/DELETE | 404, B unchanged |
| A adds user_id/role override | 400, no broader access |
| A calls any admin endpoint | 403 before target existence lookup |
| Anonymous caller hits admin endpoint | 401 |
| C uses My Data | Only C's records |
| C selects A, then B | Only selected owner's records/totals; no mixed aggregates |
| C requests B's transaction/category inside A's admin URL | 404 |
| Cursor from A reused for B | 400 |
| C attempts target mutation on admin route | 405; no records altered |
| C attempts B's mutation/export/download via personal routes | 404 or 400 for forbidden owner override; B unchanged |
| Register/PATCH profile with role=admin | 400; role remains user |
| Operator demotes C while old JWT remains valid | Next admin request 403 |
| C switches target while old request is delayed | Late old response discarded; no stale data flashes |
| Logout/login as A after viewing B | No B data in browser/native state |
| Admin success with unavailable audit store | 503 without selected-user data |
| Target has no data / only soft-deleted records | Zero totals and empty lists |
| Target timezone differs from admin timezone | Presets use target dates; stored date fields do not shift |
| Admin financial reads | Records and owner IDs unchanged; minimal audit event present |
| Personal backup import contains role/session fields | Rejected; no privilege escalation |

Automate middleware/service and PostgreSQL tests during backend tasks; then browser E2E and Android/iOS checks. Documentation review itself does not prove runtime authorization.

## Delivery
Issues 002–004 define role schema, policy and contract; 005–007 retain scoped domain operations; [ISSUE-013](issues/ISSUE-013-admin-viewing.md) implements admin reads.
Only after backend Issues 001–007 and 013 pass, implement the selector/read-only UI in Issue 008, then mobile parity in Issue 010.
Issues 009/011/012 explicitly keep personal exports/templates/backups owner-scoped.
Reference: [OWASP Authorization Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html), checked 14 September 2026. Its request-by-request authorization and default-denial guidance supports the server checks; Monelog's permission matrix is the product-specific design.
