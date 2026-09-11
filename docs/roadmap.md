# Roadmap
Backend first, then web, then mobile. Estimates are planning ranges for one developer, not deadlines. Start after requirements confirmation.

| Milestone | Issues | Estimate | Exit gate |
| --- | --- | --- | --- |
| Foundation | 001–002 | 2–4 working days | Reproducible service and tested migrations |
| Secure contract | 003–004 | 3–5 days | Isolated auth + validated OpenAPI |
| Finance backend | 005–007 | 5–8 days | CRUD and reports pass integration tests |
| Browser MVP | 008 | 4–7 days | End-to-end flows on narrow/wide screens |
| Release reports | 009 | 2–4 days | Accurate downloadable XLSX/PDF |
| Mobile release | 010 | 3–6 days plus external signing/review time | Android/iOS device checks |
| Convenience | 011 | 2–3 days | Confirmed templates work |
| Backup | 012 | 3–6 days | Authorized scheduled backup + restore drill |

Backend MVP gate: issues 001–007. Browser MVP gate: add 008. No public production release before security tests, operational backup restore test and environment review.
Offline sync, wallets/transfers, graphs, calculator, budgets, bank feeds and Launlog integration are unestimated later work.
Before offline: separately design local schema, identity, pending queue, delete tombstones, conflict resolution and retry behavior. Do not attach it silently to mobile packaging.
If any assumption changes, revise requirements/database/API and dependent issue plans before implementation.

