# Roadmap
Updated 14 September 2026 for user ownership and admin viewing. Backend first, then web, then mobile. Estimates are planning ranges for one developer, not deadlines. Start after requirements confirmation.

| Milestone | Issues | Estimate | Exit gate |
| --- | --- | --- | --- |
| Foundation | 001–002 | 2–5 working days | Reproducible service, role defaults and audit schema migrations |
| Secure contract | 003–004 | 4–7 days | Authentication, user/admin guards, operator provisioning and validated personal/admin API contract |
| Finance backend | 005–007 | 5–8 days | Owner-scoped CRUD/reports and cross-user denial tests pass |
| Admin backend | 013 (after 007) | 2–4 days | User directory, selected-user read endpoints, audit and complete authorization matrix pass |
| Browser MVP | 008 (after 013) | 5–8 days | Personal flows and admin selector/read-only flows on narrow/wide screens |
| Release reports | 009 | 2–4 days | Accurate owner-only XLSX/PDF, admin-view state cannot change export ownership |
| Mobile release | 010 | 3–6 days plus external signing/review time | Android/iOS personal and admin-view parity, no cached data leakage |
| Convenience | 011 | 2–3 days | Confirmed templates work |
| Backup | 012 | 3–6 days | Authorized scheduled backup + restore drill |

Backend MVP gate: issues 001–007 plus 013. Browser MVP gate: add 008. Existing IDs stay stable; implementation order is 001 → 002 → 003 → 004 → 005 → 006 → 007 → 013 → 008 → 009 → 010 → 011 → 012. No public production release before security tests, operational backup restore test and environment review.
Offline sync, wallets/transfers, graphs, calculator, budgets, bank feeds and Launlog integration are unestimated later work.
Before offline: separately design local schema, identity, pending queue, delete tombstones, conflict resolution and retry behavior. Do not attach it silently to mobile packaging.
If any assumption changes, revise requirements/database/API and dependent issue plans before implementation.

## Access-control release gate
FR-01/FR-15 are required for MVP. Review [access-control.md](access-control.md) and record tests for regular A/B and admin C before starting monelog-app.
Verify per-target totals, 403 for non-admin admin routes, 404 for mismatched records, current-role demotion, read-only enforcement, audit persistence and scope-bound pagination.
Role provisioning is operator-only and must be documented before staging admin tests; do not seed a publicly known admin credential.
Frontend release additionally needs delayed-response switching and logout/login isolation tests. Exports/templates/backups remain personal; neither UI selection nor worker payloads confer admin authority.
Estimates above include the new scope as draft ranges. This update changes planning, not issue completion status.
