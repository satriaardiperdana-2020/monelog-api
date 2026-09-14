# Roadmap
Draft v0.3 • 14 September 2026
Backend first in monelog-api, then Vue JavaScript frontend in monelog-app.
Estimates are draft ranges for one developer, not deadlines. Full admin permissions and boolean soft deletion are confirmed.

| Milestone | Issues | Estimate | Exit gate |
| --- | --- | --- | --- |
| Foundation | 001–002 | 3–5 working days | Reproducible service; role, boolean flag, actor attribution and audit schema |
| Auth and contract | 003–004 | 4–7 days | Current account/role authorization, lifecycle rules and validated owner/admin API |
| Finance backend | 005–007 | 6–9 days | Own CRUD/Trash/Restore and correct active-only daily/report totals |
| Admin backend | 013 after 007 | 4–7 days | Full selected-user CRUD, account/role management, audit and A/B/C matrix |
| Browser MVP | 008 after 013 | 6–9 days | Owner and admin management forms, selection safety and Trash/Restore |
| Exports | 009 | 3–5 days | Owner/admin XLSX/PDF, target-scoped jobs and download authorization |
| Mobile | 010 | 3–6 days plus signing/review time | Android/iOS role-aware CRUD and deletion/restore parity |
| Templates | 011 | 2–4 days | Own/admin template CRUD, boolean deletion and scoped apply |
| Backups | 012 | 4–7 days | Own/admin schedules/jobs, provider authorization and flagged-data restore drill |

Stable issue IDs and execution order: 001 → 002 → 003 → 004 → 005 → 006 → 007 → 013 → 008 → 009 → 010 → 011 → 012.
013 keeps its historical admin-viewing filename but now means full admin management. Do not renumber existing tasks.

## Backend gate
Issues 001–007 and 013 pass before starting frontend work.
Verify regular-user isolation; positive admin create/edit/delete/restore; selected-owner attribution; role management; deleted-account blocking; true/false defaults and migrated flags; versions/races; active-only aggregates; Trash and category history; atomic admin writes/audits.
The generated contract must document supported admin mutations and exact isDelete JSON spelling.
No feature is complete merely because its plan exists.

## Frontend and later gates
Check owner identity on forms/confirmations, save/discard on target change, late response cleanup, visible admin actions and correct Trash totals.
Exports/templates/backups have both personal and admin operation paths when each feature ships. Revalidate requester/owner/role for queued jobs and downloads.
Backups preserve deletion flags while excluding roles/credentials; confirm restore conflicts, OAuth flow, schedule and retention before Issue 012.
Release additionally requires operational database backup/restore verification, environment review and actual browser/device tests.
Offline sync, wallets/transfers, graphs, calculator, budgets, bank feeds and Launlog integration remain unestimated later scope. Revisit tombstones and conflict rules before offline synchronization.
