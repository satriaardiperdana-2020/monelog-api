# Requirements
Draft v0.3 • 14 September 2026
Product: Monelog. Documentation: English. Proposed UI: Indonesian.
Repositories: monelog-api (Go backend), monelog-app (Vue 3 + JavaScript; browser and Capacitor Android/iOS).
Confirmed correction: admins have full application permissions over any user's data; regular users can CRUD only their own data. Deletion uses the boolean isDelete.

## Evidence and scope
Screenshots show Home, day detail, Add Transaction and Reports. The launcher screenshot adds no functional requirements.
Home shows today's totals, Add, older daily summaries and Home/Reports/Settings tabs.
Day detail shows dated transactions, totals and template/share-looking controls; exact secondary-control behavior is unconfirmed.
Entry shows date, income/expense type, category, amount, title, Cancel/Save and template/calculator shortcuts.
Reports show last 7/30 days or a custom range, totals/difference, category rankings and weekly/monthly/category/transaction drilldowns.
Settings, authentication and admin screens are not shown. Their flows below are design decisions supporting the user's requirements.

## Functional requirements
| ID | Requirement | Delivery / source |
| --- | --- | --- |
| FR-01 | Authenticated regular users can create, read, update, soft-delete and restore only their own data | MVP, confirmed |
| FR-02 | Income/expense entry with date, category, positive amount and title | MVP, screenshot |
| FR-03 | Edit/delete with confirmation, and restore through Trash | MVP; deletion/restore use FR-17 |
| FR-04 | Today's totals, paginated daily history and day details | MVP, screenshot |
| FR-05 | Category create/read/update/delete/restore per transaction type | MVP; deletion replaces the earlier archive proposal |
| FR-06 | Last 7 days, last 30 days and custom inclusive report dates | MVP, screenshot annotation |
| FR-07 | Income, expense, difference and weekly/monthly/category/transaction reports | MVP, screenshot |
| FR-08 | Export selected owner's filtered records to Excel/PDF | Release 1 |
| FR-09 | Responsive browser UI and Android/iOS packaging | Confirmed |
| FR-10 | Reusable transaction templates, with an editable unsaved form when applied | Later; exact template behavior remains proposed |
| FR-11 | Profile timezone, IDR display and category settings | MVP proposal |
| FR-12 | Scheduled Google Drive backups and validated restore for the selected authorized owner | Later |
| FR-13 | Offline entry and synchronization | Deferred |
| FR-14 | Graphs, calculator and native sharing | Deferred; exact control behavior unconfirmed |
| FR-15 | Admin can manage any user's accounts, roles, categories, income/expenses, reports, exports, templates and backups | Confirmed; each capability ships with its feature milestone |
| FR-16 | Backend authorization, current account/role checks and audit of admin actions | MVP technical controls |
| FR-17 | Soft deletion uses isDelete=false for active and isDelete=true for deleted records; retained rows can be restored | Confirmed |

## Permissions
See [access-control.md](access-control.md). My Data is scoped to the authenticated account. Admin Management selects an explicit owner and supports reads and writes for that owner.
Admin CRUD is permitted. No role-based restriction blocks an admin from editing/deleting another user's records, generating/downloading their exports or managing their templates/backups.
Owner identity is derived from authorization, never trusted from a submitted body. Normal users cannot select other owners or change roles.
Admins can create users, change their profile/role, soft-delete and restore accounts through protected admin operations. Public registration always creates a regular user.
Permissions do not bypass required fields, category ownership/type, version checks or provider OAuth authorization.

## Business rules
- IDR only in MVP. Store two decimal places exactly; never binary floating point. API money is a decimal string.
- Income/expense amounts are positive; difference = income minus expense, not an account balance.
- Amount range 0.01–999999999999.99. Title is trimmed, required, 1–200 characters.
- Category belongs to the transaction owner and matches its type. Admin-created records belong to the selected user, while actor attribution records the admin.
- transaction_date is a calendar date; audit timestamps are UTC. Default timezone Asia/Jakarta, configurable.
- Date presets use the financial owner's timezone, including in Admin Management. Seven days includes today plus six earlier dates; 30 days includes today plus 29.
- Week starts Monday. Custom endpoints include both dates. Filter before grouping so boundary week/month buckets contain only selected dates.
- Future transaction dates are rejected in MVP (proposal). Empty/future report ranges may return zero.
- New users, categories, transactions and templates start with isDelete=false. Only delete/restore operations change it; ordinary create/edit payloads cannot bypass lifecycle rules.
- Normal lists, daily totals, reports and exports include active transactions only. Trash explicitly requests isDelete=true; it never changes report totals.
- Soft-deleting a category removes it from selection but preserves historical transaction amounts and category labels. Existing records can be read/deleted; edits/restores require an active compatible category.
- Category names remain unique per owner/type across active/deleted rows; restore rather than duplicate a deleted name.
- Soft-deleting a user disables authentication and jobs and revokes sessions; related rows remain retained. Restore does not recover old sessions or automatically resume paused backup schedules.
- Templates use the same boolean deletion and category ownership rules.
- No physical deletion of business rows through ordinary app operations, including admin Delete.
- No wallets/transfers/account balances in MVP. An ATM withdrawal is not automatically a transfer; recording both withdrawal and later spending as expenses double-counts spending.

## Screen flows
Regular user: login → My Data → personal CRUD, reports, Trash/Restore and later exports/templates/backups.
Admin: login → My Data or Admin Management → searchable user selector → selected user's CRUD, reports, Trash/Restore and later features.
The selected user's email/identity remains visible beside management actions; forms, dialogs, pending requests and jobs are bound to that selection.
Switching users clears stale data/filters and discards delayed responses. If an unsaved form is open, require save/discard before switching. A request already sent keeps its original target; its late result never updates a new target's screen.
Admins also have user create/update/role/delete/restore controls. Login session identity stays the actor throughout.

## Acceptance examples
- Income 1000000.00 and expense 43500.00 give difference 956500.00.
- Expenses 500000.00 + 43500.00 + 226000.00 give daily expense 769500.00.
- Empty day/report returns zero totals and empty lists.
- Failed Save preserves input; duplicate submission does not create duplicate transactions.
- Regular A cannot access/change B's data or call admin routes, even knowing their IDs.
- Admin C can create/edit/delete/restore records for A and B through their respective scoped routes; selected-user reports reflect each change.
- Delete sets isDelete=true and increments version without removing the row. Active detail returns 404; Trash shows it. Restore sets false and reintroduces the amount exactly once.
- Admin target A with record/category belonging to B returns 404 rather than silently switching owner.
- Account deletion blocks even unexpired access tokens. Demotion blocks the next admin operation after the role change commits.
- Cross-user export/template/backup operations succeed for a current admin and fail for a regular user.

## Quality and release gates
HTTPS, secret-safe logs, password hashing, rotating refresh sessions, rate-limited login and backend permission checks.
Accessible forms, readable mobile layouts, labels rather than color alone, loading/empty/error/Trash states.
Proposed p95 list/summary target: below 500 ms at 20 concurrent users and 100k transactions per owner on agreed staging infrastructure.
Test the access matrix, actual PostgreSQL constraints/migrations, money/date rules, boolean deletion/restore, job authorization, browser flows and device builds.
MVP requires internet; preserve unsaved inputs and explain connection failures.

## Remaining decisions
Public self-registration versus admin-created accounts; mandatory titles/future dates; single-template versus day bundle; exact settings, sharing, graphs/calculator, online-first and IDR-only proposals.
The full admin permissions and boolean soft-deletion rules above are confirmed and supersede earlier interpretations.
