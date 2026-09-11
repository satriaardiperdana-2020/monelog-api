# Requirements
Status: draft for confirmation. Product: Monelog. Language of documentation: English; proposed UI: Indonesian.
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
| FR-01 | Register/login/logout; isolate every user's data | MVP proposal for hosted private records |
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

## Business rules
- MVP currency is IDR only. Amounts have two decimal places internally; never use binary floating point for money.
- Income and expense are separate types; amounts are always positive. Difference = income − expense; this is NOT an account balance.
- Category belongs to the logged-in user and has the same type as its transaction.
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
- User A cannot read, edit, delete or export user B's records, even with a known UUID.

## Quality and release gates
HTTPS; no secrets in Git/logs; password hashing and token rotation; backend ownership enforcement; rate-limited login.
Keyboard-accessible forms, labeled amounts (not color alone), readable narrow-screen layout, loading/empty/error states.
Proposed performance target: p95 ordinary list/summary request below 500 ms at 20 concurrent users with 100k transactions per user in an agreed staging environment.
Test SQL integration, business rules, auth isolation, end-to-end save/edit/delete/report flow and actual mobile builds.
MVP requires internet; show an offline explanation and preserve unsaved input. Do not claim offline synchronization.

## Decisions to confirm
Are login and registration needed? Are titles mandatory? Are future dates allowed?
Are templates single transactions or whole-day bundles? What should Settings and share/export icons do?
Should graphs/calculator enter MVP? Is online-first acceptable? Is IDR-only sufficient?
Confirm these before implementing affected issues; unrelated setup can proceed.

