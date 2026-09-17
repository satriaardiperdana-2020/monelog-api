# PLAN-007: Laporan transaksi aktif

Status: Refined — siap diimplementasikan sesudah review plan.
Diperbarui: 17 September 2026
Issue: [ISSUE-007](../issues/ISSUE-007-reports.md)
Prasyarat: 006 (sudah terintegrasi)
Persyaratan: FR-01, FR-06, FR-07, FR-15, FR-17

## Hasil inspeksi dan keputusan

- Issue 006 sudah memiliki stack transaksi lengkap: source sqlc `db/queries/transactions.sql`, service `internal/service/transactions.go`, handler `internal/handlers/transactions.go`, adapter `internal/handlers/openapi.go`, dan route `internal/handlers/routes.go`. Service langsung memakai `internal/repository/sqlc`; tidak ada repository handwritten yang perlu atau sebaiknya ditambah.
- `transactions.amount` adalah `NUMERIC(14,2)`. Report memakai `COALESCE(... )::numeric(14,2)`, lalu `moneyFromNumeric` dan `formatCents`; tidak boleh memakai `float64`.
- Endpoint Reports baru menyediakan summary dan breakdown. Panel daftar transaksi Reports memakai `GET /api/v1/transactions` yang sudah ada, atau `/api/v1/admin/users/{user_id}/transactions`, dengan filter tanggal/tipe/kategori, cursor, dan pagination; jangan membuat endpoint list kedua.
- Parameter `range=last_7_days|last_30_days|custom` menjadi kontrak report. Preset dihitung server memakai timezone owner dan `Transactions.now`: tujuh hari mencakup hari ini + enam hari sebelumnya, 30 hari mencakup hari ini + 29 hari sebelumnya. `custom` mewajibkan `start_date` dan `end_date`; keduanya inklusif dan maksimum 366 hari. Parameter tanggal bersama preset ditolak. Rentang kosong/masa depan valid dan menghasilkan nol.
- Minggu dimulai Senin dengan `date_trunc('week', transaction_date)::date`; bulan memakai `date_trunc('month', transaction_date)::date`. Filter rentang diterapkan sebelum grouping sehingga bucket batas parsial. Join kategori historis tidak memfilter `categories.is_delete`; transaksi selalu `t.is_delete=FALSE`.
- Index aktif yang ada (`transactions_active_list_idx` dan `transactions_active_category_idx`) sudah mendukung predicate owner, active, date, dan list kategori. Tidak ada migration/index pada implementasi awal; index baru hanya boleh ditambah setelah bukti `EXPLAIN (ANALYZE, BUFFERS)` dan benchmark.

## File yang diubah

| File | Perubahan |
| --- | --- |
| `api/openapi.yaml` | Tambahkan empat operasi report, parameter range/group, schema dan errors. |
| `internal/api/openapi.gen.go` | Hasil `make oapi-generate`; jangan edit manual. |
| `db/queries/transactions.sql` | Tambahkan empat query sqlc agregasi report. |
| `internal/repository/sqlc/transactions.sql.go`, `internal/repository/sqlc/querier.go` | Hasil `make sqlc-generate`; jangan edit manual. |
| `internal/service/transactions.go` | Filter/DTO report, resolver range timezone owner, metode summary/breakdown dan mapping numeric. |
| `internal/handlers/transactions.go` | Interface, handler personal/admin, decoding parameter, dan mapping response. |
| `internal/handlers/openapi.go` | Empat delegasi `api.ServerInterface` baru. |
| `internal/handlers/routes.go` | Middleware untuk operation ID report personal/admin. |
| `internal/handlers/transactions_test.go` | Fake service dan test HTTP/handler report. |
| `internal/service/transactions_test.go` | Unit resolver preset/custom dan format money. |
| `internal/service/transactions_integration_test.go` | Fixture/report lifecycle, scope, aggregasi dan determinisme. |
| `internal/service/transactions_benchmark_test.go` | Benchmark integration report 100k transaksi. |
| `docs/api.md`, `docs-en/api.md` | Sinkronkan kontrak API dengan OpenAPI final. |

`cmd/api/main.go` tidak berubah: `service.NewTransactions(pool)` dan `handlers.NewTransactions(...)` sudah merupakan dependency wiring yang benar. Tidak ada migration baseline.

## OpenAPI dan generated handler wiring

Tambahkan route Bearer-authenticated berikut:

| Operation ID | Route | Hasil |
| --- | --- | --- |
| `getReportSummary` | `GET /api/v1/reports/summary` | `ReportSummaryResponse` |
| `getReportBreakdown` | `GET /api/v1/reports/breakdown` | `ReportBreakdownResponse` |
| `adminGetReportSummary` | `GET /api/v1/admin/users/{user_id}/reports/summary` | `ReportSummaryResponse` |
| `adminGetReportBreakdown` | `GET /api/v1/admin/users/{user_id}/reports/breakdown` | `ReportBreakdownResponse` |

Summary menerima `range` dan custom dates. Breakdown menerima parameter sama plus `group_by=week|month|category`. Tambahkan enum `ReportRange` dan `ReportGroupBy`, serta schema berikut:

- `ReportPeriod`: `start_date`, `end_date`.
- `ReportCategoryTotal`: `category_id`, `name`, `type`, `amount`.
- `ReportSummary`: period, income, expense, difference, `top_income_categories`, `top_expense_categories`.
- `ReportPeriodTotal`: `period_start`, income, expense, difference.
- `ReportBreakdown`: period dan array `periods`/`categories`; array yang tak dipakai tetap `[]`.

Money tetap string dua desimal; difference memakai pola signed DailySummary. Top-five diurut `amount DESC, category_id ASC`, maksimal lima per tipe. Week/month diurut `period_start ASC`; category diurut `type ASC, amount DESC, category_id ASC`. Tambahkan 422 untuk input rentang/group tidak valid dan 403 pada route admin.

Setelah `make oapi-generate`, compiler akan meminta empat metode pada `openAPIServer`. Teruskan ke `Transactions.Summary`, `AdminSummary`, `Breakdown`, dan `AdminBreakdown`. Route map memasang `authenticate` untuk personal dan `authenticate, middleware.RequireAdmin` untuk admin, memakai operation ID persis di atas.

## Query sqlc dan metode repository

Tambahkan empat named query ke `db/queries/transactions.sql`, semuanya menerima `user_id`, `start_date`, `end_date`, selalu `is_delete=FALSE`, dan tidak menerima parameter lifecycle:

1. `GetReportSummary`: `COALESCE(SUM(amount) FILTER (WHERE type='income'),0)::numeric(14,2)` dan expense dalam satu row.
2. `ListTopReportCategories`: join `transactions t` ke `categories c` pada `(id,user_id,type)` tanpa filter category deleted; group tipe/category; gunakan `ROW_NUMBER() OVER (PARTITION BY t.type ORDER BY SUM(t.amount) DESC,c.id ASC) <= 5`.
3. `ListReportPeriodBreakdown`: group week atau month setelah filter; parameter group tervalidasi di service dan `CASE` literal dipakai untuk `date_trunc`, tanpa interpolasi SQL.
4. `ListReportCategoryBreakdown`: join historis, group id/name/type, urut deterministik.

`sqlc` menghasilkan metode `GetReportSummary`, `ListTopReportCategories`, `ListReportPeriodBreakdown`, dan `ListReportCategoryBreakdown` pada `*sqlc.Queries` serta interface `Querier`. Itulah metode repository untuk issue ini; tidak buat `internal/repository/reports.go`.

## Service

Tambahkan `ReportFilter`, `ReportSummary`, `ReportCategoryTotal`, `ReportPeriodTotal`, dan `ReportBreakdown` ke `internal/service/transactions.go`, dengan uang sebagai string. Tambahkan `GetReportSummary(ctx, scope, filter)` dan `GetReportBreakdown(ctx, scope, filter)`.

Kedua metode memanggil `validateReadScope` sebelum query, lalu `resolveReportRange`. Resolver memuat timezone owner untuk preset memakai `GetUserByID`, menjalankan `ValidateTimezone`, dan memakai `s.now`; custom memakai `validateDateRange`. USER selalu `actor==owner`; admin aktif selalu memakai owner path eksplisit; target hilang adalah `ErrNotFound`. Setelah query berhasil, operasi admin memanggil `auditRead` dengan action `report_summary` atau `report_breakdown`, resource nil.

Mapping aggregate memakai `moneyFromNumeric`. Difference selalu `formatCents(income.Cents()-expense.Cents())`. Summary kosong menghasilkan tiga `"0.00"`; semua list kosong adalah slices `[]`, bukan null.

## Pengujian dan benchmark

- Unit: preset Jakarta/UTC di sekitar tengah malam, last-7/last-30 inklusif, custom satu hari/leap day/tahun boundary, range terbalik atau >366 hari, kombinasi parameter salah, dan future-only zero report.
- Service integration: income 1000000.00 minus expense 43500.00 = `956500.00`; top-five dan tie `category_id ASC`; buckets week Monday/month/year parsial; deleted category tetap terhitung; deleted transaction menghilang lalu restore kembali sekali; report kosong memberi `"0.00"` dan `[]`.
- Scope integration: A tidak dapat membaca B dari personal scope; admin membaca B hanya dari admin scope; total B tidak memuat A; admin access event memuat actor dan target; role/target error tepat.
- Handler: scope personal/admin, `RequireAdmin`, response money strings, invalid enum/custom range, serta list transaksi Reports mempertahankan cursor/filter terikat scope.
- Jalankan `go test ./...`, `go vet ./...`, `make build`, `make sqlc-vet`, `make sqlc-check`, dan `make oapi-check`. Dengan `TEST_DATABASE_URL`, jalankan `make test-integration`.

Tambahkan `BenchmarkReportsActiveOwner100k` build-tag `integration`. Seed satu owner dengan 100,000 transaksi aktif selama 366 hari dan beberapa rows deleted; ukur summary, top-category, week/month/category breakdown, dan list terfilter setelah `b.ResetTimer()`. Laporkan `ns/op`, allocations, database version, fixture, query count, dan lingkungan; benchmark tidak menjadi assertion CI.

## Inspeksi query plan dan migration gate

Pada PostgreSQL disposable yang telah `ANALYZE` (dan `VACUUM` untuk menguji index-only scan), jalankan `EXPLAIN (ANALYZE, BUFFERS)` untuk empat query pada owner 100k transaksi dan rentang 7, 30, dan 366 hari. Simpan actual rows, execution time, shared hit/read, dan scan terpilih. Pastikan predicate `user_id`, `is_delete=false`, dan date range memakai `transactions_active_list_idx`; periksa list kategori memakai `transactions_active_category_idx`.

Hanya bila plan 7/30 hari melakukan sequential scan atau heap I/O yang terbukti melanggar target staging, tambahkan migration baru dengan index partial covering owner/date, misalnya `INCLUDE (type,category_id,amount)`, berikut down migration. Bandingkan benchmark/plan sebelum-sesudah dan biaya ukuran/write. Jangan mengubah migration 000003 atau 000006 yang telah diterapkan, dan jangan menambah index spekulatif.

## Batasan

Tidak ada Vue/mobile UI, chart, export, stored balance, wallet, atau transfer. Reports selalu active-only dan menolak `isDelete`; Trash tetap memakai API transaksi Issue 006.
