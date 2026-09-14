# Requirements
Status: draft v0.2, updated 14 September 2026. User ownership and admin viewing are confirmed by the user; remaining proposals are labeled. Product: Monelog. Language of documentation: English; proposed UI: Indonesian.
Repositories: monelog-api (Go), monelog-app (Vue 3 + JavaScript; browser and Capacitor mobile).

## Evidence and scope
Four application screenshots show: Home, day detail, Add Transaction, Reports. The phone launcher screenshot gives no functional requirements.
Home: UTAMA/LAPORAN/SETELAN tabs; today's totals, Tambah, older daily summaries.
Day detail: dated transactions and totals, Tambah, template buttons, export/share-looking icons. Icon behavior is unconfirmed.
Entry: date, expense/income type, category, amount, title, Cancel/Save, template shortcut and calculator-looking icon.
Reports: last 7/30 days or custom range (user annotation), income/expense/difference, category rankings, monthly/weekly/category/transaction drilldowns; graph button visible.
No Settings, authentication, edit/delete, template dialogs or export dialogs were supplied. Their behavior below is proposed, not observed.

## Functional requirements
| ID | Requirement | Priority / source |
| --- | --- | --- |
| FR-01 | Require authenticated accounts; regular users view only their own income, expenses and derived data | MVP access rule confirmed 14 September 2026; public self-registration remains a proposal |
| FR-02 | Create income or expense with date, category, positive amount and title | MVP screenshot |
| FR-03 | Edit/delete a transaction with delete confirmation | MVP proposal |
| FR-04 | Show today's totals and paginated historical daily summaries; open day detail | MVP screenshot |
| FR-05 | List/create/rename/archive categories per transaction type | MVP inferred category support |
| FR-06 | Reports for last 7 days, last 30 days, custom inclusive date range | MVP screenshot annotation |
| FR-07 | Show income, expense and difference; monthly, weekly, category and transaction lists | MVP screenshot |
| FR-08 | Export the same filtered records as Excel and PDF | Release 1 proposal |
| FR-09 | Responsive browser UI and Android/iOS packaging | Explicit user request |
| FR-10 | Save reusable transaction presets; applying one opens an editable unsaved form | Later, proposed behavior for visible template controls |
| FR-11 | Settings: profile timezone, IDR display, category management | MVP proposal; settings screen missing |
| FR-12 | Scheduled user-authorized Google Drive backup with restore validation | Later proposal; separate from database disaster recovery |
| FR-13 | Offline entry and conflict-safe synchronization | Deferred; not established by screenshots |
| FR-14 | Graphs, calculator helper, native sharing | Deferred; controls visible, exact behavior unknown |
| FR-15 | Admin selects any user and views that user's transactions, daily summaries, categories and reports | MVP admin viewing confirmed 14 September 2026; read-only scope is the implementation interpretation |
| FR-16 | Server-managed user/admin roles, current-role checks and admin-view audit records | MVP technical controls supporting FR-01/FR-15 |

## Business rules
- MVP currency is IDR only. Amounts have two decimal places internally; never use binary floating point for money.
- Income and expense are separate types; amounts are always positive. Difference = income − expense; this is NOT an account balance.
- Category belongs to the transaction owner and has the same type as its transaction. In My Data that owner is the logged-in user; in an admin view it is the selected user.
- A transaction date is a calendar date, independent of its UTC audit timestamps.
- Default timezone Asia/Jakarta, configurable. Last 7 days includes today and six previous dates; last 30 includes today and 29 previous dates.
- Week starts Monday. Custom endpoints include both dates. Partial week/month totals include only dates inside the requested range.
- Title required, trimmed, 1–200 characters; category required; amount from 0.01 to 999999999999.99.
- Future transaction dates rejected in MVP (proposal). Reporting arbitrary empty/future ranges remains permitted.
- Archived categories remain on historical transactions but cannot be newly selected; their type cannot change. Renaming changes historical display names.
- Delete removes records from normal totals; soft deletion is an implementation choice, not a promise of recovery UI.
- No wallets/transfers/account balances in MVP. An ATM withdrawal is not automatically classified as a transfer; warn in help that tracking both withdrawal and spending as expenses double-counts spending.

## Acceptance examples
- Income 1000000.00 and expense 43500.00 yield difference 956500.00.
- Example day expenses 500000.00 + 43500.00 + 226000.00 yield 769500.00.
- Empty day/report returns zero totals and empty lists, not fabricated transactions.
- Save succeeds only after server confirmation; failure preserves typed form values. Double-tapping Save must not create duplicates.
- Editing/deleting a record updates day totals, report totals and category totals consistently.
- A regular user A cannot read, edit, delete or export user B's records, even with a known UUID.
- Admin C can select A or B and view only that selected user's active transactions and reports. A record from B requested inside A's admin URL returns 404.
- Admin C cannot create/edit/delete or export A's records, change A's settings, use A's templates or manage A's personal backups. My Data remains scoped to C.
- Switching from A to B clears A's displayed data immediately; late responses for A cannot populate B's screen.
- A non-admin calling an admin endpoint receives 403 before any target lookup. Demotion committed before the next request blocks that next admin request, even if the old access JWT is unexpired.
- Public registration and profile updates cannot set role; new and migrated accounts default to user. No automatic first-account admin.

## Quality and release gates
HTTPS; no secrets in Git/logs; password hashing and token rotation; backend ownership and admin-role enforcement on every protected admin request; rate-limited login.
Keyboard-accessible forms, labeled amounts (not color alone), readable narrow-screen layout, loading/empty/error states.
Proposed performance target: p95 ordinary list/summary request below 500 ms at 20 concurrent users with 100k transactions per user in an agreed staging environment.
Test SQL integration, business rules, the full [access matrix](access-control.md), end-to-end personal CRUD and admin view/switch flows, and actual mobile builds.
MVP requires internet; show an offline explanation and preserve unsaved input. Do not claim offline synchronization.

## Decisions to confirm
Authentication and the user/admin visibility distinction are required. Confirm public self-registration versus operator-created regular accounts. Are titles mandatory? Are future dates allowed?
Are templates single transactions or whole-day bundles? What should Settings and share/export icons do?
Should graphs/calculator enter MVP? Is online-first acceptable? Is IDR-only sufficient?
Confirm these before implementing affected issues; unrelated setup can proceed.

## Role-aware screen flow
Regular user: login → My Data → own Home/day detail/reports and own record actions.
Admin: login → My Data or Admin View → searchable paginated user selector → one selected user's read-only Home/day detail/reports.
No user is preselected in Admin View. The selected user's email, timezone and a read-only label remain visible; date presets use that user's timezone.
Add/Edit/Delete, export/share, template actions and backup/settings changes are hidden in Admin View. Return to My Data restores the admin's own account context.
The selected account never changes the authenticated session or token subject. The complete policy and test cases are in [access-control.md](access-control.md).
