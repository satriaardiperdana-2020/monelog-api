# Indeks issue

Diperbarui 16 September 2026 (v0.4) untuk implementasi fondasi database ISSUE-002.
Status setiap tugas tercatat pada tabel. File tugas portabel dapat ditautkan ke GitHub Issue tanpa mengubah ID stabilnya.
Backend 001–007 dan 013 selesai sebelum Issue 008 monelog-app.

| ID | Tugas | Repositori | Dependensi | Status | Plan |
| --- | --- | --- | --- | --- | --- |
| ISSUE-001 | [Setup proyek backend](issues/ISSUE-001-project-setup.md) | monelog-api | Tidak ada | Backlog | [PLAN-001](plans/PLAN-001.md) |
| ISSUE-002 | [Migrasi basis data dan fondasi sqlc](issues/ISSUE-002-database-foundation.md) | monelog-api | 001 | Review | [PLAN-002](plans/PLAN-002.md) |
| ISSUE-003 | [Autentikasi, state akun, dan otorisasi peran](issues/ISSUE-003-authentication.md) | monelog-api | 002 | Backlog | [PLAN-003](plans/PLAN-003.md) |
| ISSUE-004 | [Kontrak API pemilik/admin dan penghapusan lunak](issues/ISSUE-004-api-contract.md) | monelog-api | 003 | Backlog | [PLAN-004](plans/PLAN-004.md) |
| ISSUE-005 | [CRUD kategori, Trash, dan restore](issues/ISSUE-005-categories.md) | monelog-api | 004 | Backlog | [PLAN-005](plans/PLAN-005.md) |
| ISSUE-006 | [CRUD transaksi, penghapusan lunak, dan ringkasan harian](issues/ISSUE-006-transactions.md) | monelog-api | 005 | Backlog | [PLAN-006](plans/PLAN-006.md) |
| ISSUE-007 | [Laporan transaksi aktif](issues/ISSUE-007-reports.md) | monelog-api | 006 | Backlog | [PLAN-007](plans/PLAN-007.md) |
| ISSUE-013 | [Pengelolaan penuh user dan data oleh admin](issues/ISSUE-013-admin-viewing.md) | monelog-api | 007 | Backlog | [PLAN-013](plans/PLAN-013.md) |
| ISSUE-008 | [UI pengelolaan pemilik/admin Vue](issues/ISSUE-008-web-mvp.md) | monelog-app | 013 | Backlog | [PLAN-008](plans/PLAN-008.md) |
| ISSUE-009 | [Ekspor Excel/PDF pemilik/admin](issues/ISSUE-009-exports.md) | monelog-api + monelog-app | 008 | Backlog | [PLAN-009](plans/PLAN-009.md) |
| ISSUE-010 | [Pengelolaan pemilik/admin Android/iOS](issues/ISSUE-010-mobile-packaging.md) | monelog-app | 009 | Backlog | [PLAN-010](plans/PLAN-010.md) |
| ISSUE-011 | [Templat pemilik/admin dengan penghapusan lunak](issues/ISSUE-011-templates.md) | monelog-api + monelog-app | 010 | Backlog | [PLAN-011](plans/PLAN-011.md) |
| ISSUE-012 | [Backup Drive dan restore tervalidasi pemilik/admin](issues/ISSUE-012-drive-backups.md) | monelog-api + monelog-app | 011 | Backlog | [PLAN-012](plans/PLAN-012.md) |

Issue 013 mempertahankan nama file asli agar tautan tetap berfungsi, tetapi cakupannya adalah pengelolaan admin penuh. ID tetap stabil; gunakan urutan dependensi pada tabel.

## Perubahan lintas proses

| Issue | Pekerjaan izin/siklus hidup |
| --- | --- |
| 001 | Batas service dan kebijakan yang telah dikonfirmasi |
| 002 | Boolean/default/backfill is_delete, role, lifecycle berversi, atribusi actor, dan skema audit |
| 003 | Autentikasi akun aktif, guard role saat ini untuk baca/tulis, bootstrap eksplisit |
| 004 | Kontrak CRUD owner/admin, field isDelete eksak, Trash/delete/restore, dan version |
| 005–007 | Kategori/transaksi berscope, delete/restore boolean, laporan aktif saja |
| 013 | Akun/role admin penuh, CRUD user terpilih, audit, dan suite otorisasi |
| 008 | Form/aksi admin aktif, Trash/restore sendiri/admin, target form immutable |
| 009 | Ekspor admin owner terpilih dan otorisasi ulang job/download |
| 010 | CRUD owner/admin Android/iOS dan parity lifecycle flag |
| 011 | Pengelolaan templat owner/admin, soft delete, dan restore |
| 012 | Pengelolaan Drive/backup owner/admin dan restore tervalidasi dengan flag |

Gunakan [access-control.md](access-control.md) sebagai matriks izin dan [workflow.md](workflow.md) untuk setiap tugas.
Issue duplikat tunggal issue.md tidak diperlukan. Jika issue remote dibuat nanti, tambahkan URL sebenarnya ke file/index terkait dan pertahankan ID portabel.
