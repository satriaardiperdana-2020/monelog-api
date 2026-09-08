# Monelog product requirements

Status: Draft for review
Date: 2026-09-08
Owner: Satria Ardi Perdana
Canonical location: monelog-api/docs/requirements.md

## 1. Purpose

Monelog helps individuals quickly record income and expenses, understand their cash flow, export clear reports, and back up their financial records to Google Drive.

This document defines shared product behavior for monelog-api and monelog-app. It is a planning baseline, not a claim that features are implemented. The existing repository currently contains project initialization files only.

## 2. Confirmed constraints

- Backend repository: monelog-api.
- Backend stack: Go, Echo, PostgreSQL, sqlc, JWT, and oapi-codegen.
- Frontend repository: monelog-app.
- Frontend stack: Vue 3 with JavaScript and Capacitor. No TypeScript or Ionic.
- Targets: web, Android, and iOS, using the same backend.
- Main menus: Utama, Laporan, Setelan.
- Record income and expenses with a date, type, category, amount, and title.
- Show history grouped by date and a daily transaction detail screen.
- Keep reports simple; complex charts are unnecessary.
- Support Excel and PDF report exports.
- Support automatic Google Drive backup.
- Plan before implementing feature code; use GitHub issues to track delivery.

## 3. Proposed defaults requiring review

These are working proposals, not previously approved requirements.

| Decision | Proposal |
|---|---|
| Currency | IDR only for version 1 |
| Ledger | One personal ledger per user |
| Connectivity | Online-only; no offline transaction queue or synchronization |
| Interface language | Bahasa Indonesia initially |
| Timezone | Asia/Jakarta initially, stored per user |
| Amount precision | Up to two decimal places, represented exactly |
| Authentication | Email/password registration and login, JWT access token, revocable refresh sessions |
| Backup schedule | Daily at 02:00 in the saved user timezone |
| Backup retention | Latest 7 successful backups |
| Backup format | Versioned compressed JSON containing financial records |
| Restore mode | Replace the current user's financial records after preview and explicit confirmation |
| Deferred features | Templates, calculator, budgets, recurring entries, multiple wallets, transfers, multi-currency |

Review these decisions before implementing affected features. Password recovery and account deletion need a release-scope decision before public launch.

## 4. Users and boundaries

A registered user manages only their own records, categories, settings, reports, and backups. Administrative screens are outside the initial product scope.

Net cash flow means income minus expenses for the selected period. It must not be labeled as a bank balance because opening balances and wallets are not included.

## 5. Functional requirements and acceptance criteria

### AUTH-01: Account and sessions

- Register, log in, refresh an eligible session, and log out.
- Store securely hashed passwords; never return passwords or password hashes.
- Reject unauthenticated financial requests.
- Expired or revoked sessions cannot grant continued access.
- Logging out invalidates the applicable refresh session; access-token expiry behavior must be documented.
- Invalid login responses do not disclose whether an account exists.
- User A cannot read, edit, delete, export, back up, or restore user B's records by changing identifiers.

### CAT-01: Categories

- List and create categories for income or expense.
- Rename categories and archive categories used by historical records.
- Archived categories remain readable in history but cannot be selected for new transactions.
- Reject another user's category and a category incompatible with the selected transaction type.
- Seed categories, if provided, must have a documented ownership strategy.

### TXN-01: Add a transaction

Fields: transaction date, income/expense type, category, positive amount, title.

- Date defaults to today in the user's timezone and can be changed.
- Validate required fields on both client and server.
- Reject zero, negative, malformed, or over-precision amounts.
- Store exact amounts; do not use binary floating-point for financial calculations.
- Save a valid transaction once and refresh the relevant list and totals.
- Prevent duplicate entries from repeated submission or retry using a documented idempotency strategy.
- On save failure, preserve entered values and show an actionable error.
- Cancel returns without saving.
- Title length, maximum amount, and future-date policy must be finalized in the API design.

### TXN-02: History and daily detail

- Utama shows today's income and expense totals, an Add action, and dates with recorded activity ordered newest first.
- Each date row shows its income and expense totals.
- Selecting a date opens its transactions and totals.
- Daily rows show a readable title and formatted amount, with a clear income/expense distinction.
- Empty dates show zero totals and an invitation to add a record.
- Paginate long lists; totals cover all matching records, not just the visible page.

### TXN-03: Edit and delete

- Edit an owned transaction using the same validation as creation.
- Changing its date or type updates both old and new affected summaries.
- Ask for confirmation before deletion.
- Failed edits/deletions leave the saved record unchanged.
- Successful changes update the relevant totals and reports.

### REP-01: Simple reports

- Period options: last 7 days, last 30 days, this month, and a custom inclusive date range.
- Proposed semantics: rolling periods include today; this month runs from month start through today.
- Reject a start date after the end date.
- Display income, expenses, and difference (income minus expenses).
- Display simple category totals, separated by income and expense, and a transaction-detail action.
- No chart, graph menu, or separate weekly/monthly screen is needed initially.
- Show explicit empty states rather than misleading missing totals.
- Use the same date/timezone and amount rules throughout all screens and exports.

### EXP-01: Excel and PDF exports

- Export the selected report period to a real .xlsx file or a readable PDF.
- Include the period, currency, income, expenses, difference, and matching transaction details.
- Export all matching records rather than only the current screen page.
- Export totals match the displayed report for the same data snapshot and filters.
- Enforce user ownership for generation and download.
- Browser users can download; Android/iOS users can save or share.
- Show progress/failure without losing the selected report filters.
- Large-export limits and synchronous/background generation are architecture decisions.

### BAK-01: Google Drive connection and automatic backup

- In Setelan, users can connect their own Google Drive, enable/disable automatic backup, and disconnect.
- Authorization for Drive is separate from Monelog login.
- Automatic backup remains disabled until the user authorizes and enables it.
- Show connection state, schedule, last successful backup, next scheduled run, and failures.
- A backend scheduler runs while the frontend is closed.
- A manual Back up now action uses the same backup pipeline.
- Back up only the requesting user's categories, transactions, and relevant preferences from a consistent snapshot.
- Include schema version, creation time, record counts, and integrity metadata.
- Never include password hashes, JWTs, refresh-session secrets, or Google credentials in the backup.
- Protect stored Google credentials and handle expired/revoked authorization with a reconnect message.
- Retries must not produce overlapping jobs or delete the last valid backup.
- Apply retention only after a new backup succeeds.
- Disconnecting stops future backups and removes stored connection credentials; existing backup deletion requires a separate explicit decision.
- Proposed storage: Google Drive app-data folder, with backups listed in Monelog because they are not normal browsable Drive files.
- Backup encryption and key recovery must be designed before implementation.

### BAK-02: Restore

- List the connected user's available backups.
- Validate supported schema, integrity, and record relationships before mutation.
- Show backup date, record counts, and replacement behavior before confirmation.
- Restore only into the authenticated user's ledger; never trust ownership IDs in the imported file.
- Create a recoverable pre-restore snapshot.
- Replace the user's financial records transactionally; reject invalid backups without partial data changes.
- Define concurrent-write handling so edits cannot be silently lost during restore.
- Restore categories and transactions without restoring credentials or another user's identity.
- Verify restored record counts and totals and refresh the app.
- Test restoration, not merely successful upload.
- Drive backups supplement operational database backups; they do not replace them.

## 6. Screen references

The supplied screenshots are behavioral references, not implementation evidence.

| Screen | Observed reference | Planned adaptation |
|---|---|---|
| Utama | Today's totals, Add button, daily history | Responsive layout and readable currency |
| Daily detail | Date totals, transaction titles and amounts | Add edit/delete actions |
| Add transaction | Date, type, category, amount, title, Cancel/Save | Clear validation and preserved input on error |
| Laporan | Period, totals, category sections, additional menus | One period selector, simple totals/list, Excel/PDF |
| Setelan | No settings screenshot supplied | Categories, account, Google Drive backup/restore |

Template buttons and calculator icons shown in references are deferred proposals. The screenshot values are sample reference data, not production data or a migration request.

## 7. Repository responsibilities

| Area | monelog-api | monelog-app |
|---|---|---|
| Authentication | Credentials, sessions, authorization | Forms, session UX, platform storage |
| Transactions | Validation, persistence, totals | Entry, history, detail, editing |
| Reports | Queries and exact calculations | Period selection and simple lists |
| Export | Generate authorized reports | Download/save/share actions |
| Backup | OAuth, scheduler, upload, restore | Settings, status, authorization launch, confirmation |
| Contract | OpenAPI specification and generated Go bindings | Consume a known contract revision using JavaScript |
| Planning | Canonical shared requirements | Link shared requirements; own frontend architecture |

## 8. Quality requirements

- Enforce ownership at every data-access boundary.
- Use HTTPS in deployed environments; never commit secrets or log credentials.
- Define secure session handling separately for web and native clients.
- Use exact monetary arithmetic and consistent user-date semantics.
- Support small phone screens and desktop layouts without horizontal form overflow.
- Use labeled controls, keyboard navigation, and information beyond color alone.
- Provide loading, empty, success, and error states.
- Keep dependencies and generators reproducible.
- Test database constraints, cross-user access, date boundaries, export parity, duplicate submissions, and restore failure recovery.
- Validate web, Android, and iOS navigation, session handling, and file actions before claiming platform readiness.
- Define measurable performance and dataset targets during architecture planning.

## 9. Delivery order and GitHub workflow

1. Review and merge these requirements after resolving affected scope decisions.
2. Write architecture.md, database.md, api.md, and roadmap.md.
3. Define api/openapi.yaml as the authoritative interface contract.
4. Create scoped GitHub issues referencing requirement IDs from this document.
5. Implement foundation, authentication, categories, transactions, reports/exports, backup/restore, and release validation in dependency order.
6. Use one focused branch/PR per task with acceptance criteria and verification evidence.
7. Update requirements when an approved product decision changes.

Suggested issue groups (not yet created): foundation, AUTH-01, CAT-01, TXN-01/02/03, REP-01, EXP-01, BAK-01/02, and cross-platform validation. Backend and frontend implementation issues belong in their respective repositories.

## 10. Definition of done

A feature is complete when its acceptance criteria are implemented, relevant tests pass, the diff is reviewed, affected documentation is updated, and a PR is merged. A draft document or issue alone is not a completed feature.
