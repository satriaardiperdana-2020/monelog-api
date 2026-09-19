# PLAN-006: CRUD transaksi, penghapusan lunak, dan ringkasan harian

Status: Refined — siap diimplementasikan setelah prasyarat Issue 005 selesai.
Diperbarui: 17 September 2026
Issue: [ISSUE-006](../issues/ISSUE-006-transactions.md)
Repositori: monelog-api
Prasyarat: 005
Persyaratan: FR-01, FR-02, FR-03, FR-04, FR-15, FR-17

## Hasil inspeksi repositori

- Skema dasar transaksi sudah ada di `db/migrations/000003_create_transactions.up.sql`. Skema memakai satu tabel `transactions` untuk `income` dan `expense`, `NUMERIC(14,2)`, `client_request_id`, `request_hash`, owner/actor terpisah, `is_delete`, dan `version`.
- Query awal ada di `db/queries/transactions.sql` dan kode sqlc sudah ada di `internal/repository/sqlc/transactions.sql.go`. Query tersebut belum memiliki filter tanggal/tipe/kategori, detail Trash, cursor keyset, `created_at` pada urutan list aktif, ringkasan harian, idempotent insert yang aman dari race, lock kategori, atau klasifikasi state untuk membedakan 404 dan 409.
- `api/openapi.yaml` dan `internal/api/openapi.gen.go` saat ini hanya memuat health, auth, dan `/me`; kontrak kategori maupun transaksi belum masuk ke kontrak kanonis.
- Implementasi kategori Issue 005 berada sebagai file untracked di `internal/handlers/categories.go`, `internal/repository/categories.go`, `internal/service/categories.go`, dan test terkait. Implementasi itu belum terhubung ke OpenAPI/routing dan saat ini mengacu pada tipe generated yang belum ada. File tersebut adalah perubahan pengguna dan tidak boleh ditimpa oleh Issue 006.
- Autentikasi sudah memasukkan `service.Actor` yang dimuat dari database ke request context melalui `middleware.Authenticate`. Route admin tetap harus memakai `middleware.RequireAdmin`, lalu service/repository memeriksa ulang actor dan target di dalam transaksi database untuk write.
- Harness integrasi PostgreSQL sudah tersedia di `internal/repository/sqlc/integration_test.go` dan `internal/service/auth_integration_test.go`. `Makefile` saat ini hanya menjalankan package sqlc untuk target `test-integration`; target itu perlu mencakup test integrasi service transaksi.
- Inspeksi awal tidak mengubah kode. Baseline `go test ./...` belum dapat dinilai dari shell biasa karena `go` tidak ada di `PATH`; toolchain proyek tersedia di `/usr/local/go/bin/go`. Selain itu, file kategori untracked belum dapat dikompilasi sampai kontrak Issue 005 digenerate.

## Gerbang sebelum implementasi

1. Selesaikan dan review Issue 005: kontrak kategori masuk ke `api/openapi.yaml`, generated code sinkron, route terpasang, seluruh file kategori untracked dikenali sebagai baseline pengguna, dan write kategori memakai urutan lock users → category yang sama agar tidak berlomba atau deadlock dengan create/update/restore transaksi.
2. Jalankan baseline dengan toolchain eksplisit dan cache di `/tmp`: `GO=/usr/local/go/bin/go GOPATH=/tmp/monelog-go-path GOMODCACHE=/tmp/monelog-go-modcache GOCACHE=/tmp/monelog-go-cache make test` serta test integrasi PostgreSQL yang dikonfigurasi. Catat kegagalan yang sudah ada sebelum perubahan Issue 006.
3. Jangan mengubah frontend, worker, laporan, ekspor, template, backup, wallet, transfer, recurring job, atau sinkronisasi offline.

## Kontrak HTTP dan OpenAPI

Ubah `api/openapi.yaml`, lalu generate ulang `internal/api/openapi.gen.go` melalui `make oapi-generate`. Tambahkan operasi konkret berikut; route personal selalu memperoleh owner dari actor terautentikasi, sedangkan route admin memperoleh owner hanya dari `{user_id}` pada path.

| operationId | Method dan path | Perilaku |
| --- | --- | --- |
| `listTransactions` | `GET /api/v1/transactions` | List aktif atau Trash milik actor dengan filter dan cursor. |
| `createTransaction` | `POST /api/v1/transactions` | Create milik actor; 201 untuk insert, 200 untuk replay idempoten. |
| `getTransaction` | `GET /api/v1/transactions/{id}` | Detail aktif/Trash milik actor. |
| `updateTransaction` | `PATCH /api/v1/transactions/{id}` | Update transaksi aktif dengan version. |
| `deleteTransaction` | `DELETE /api/v1/transactions/{id}` | Soft delete dengan `If-Match`. |
| `restoreTransaction` | `POST /api/v1/transactions/{id}/restore` | Restore dengan version pada body. |
| `listDailySummaries` | `GET /api/v1/daily-summaries` | Navigasi hari berisi transaksi aktif, terbaru dahulu. |
| `adminListTransactions` | `GET /api/v1/admin/users/{user_id}/transactions` | List owner yang dipilih admin. |
| `adminCreateTransaction` | `POST /api/v1/admin/users/{user_id}/transactions` | Create untuk owner yang dipilih, actor tetap admin. |
| `adminGetTransaction` | `GET /api/v1/admin/users/{user_id}/transactions/{id}` | Detail owner yang dipilih. |
| `adminUpdateTransaction` | `PATCH /api/v1/admin/users/{user_id}/transactions/{id}` | Update owner yang dipilih. |
| `adminDeleteTransaction` | `DELETE /api/v1/admin/users/{user_id}/transactions/{id}` | Soft delete owner yang dipilih. |
| `adminRestoreTransaction` | `POST /api/v1/admin/users/{user_id}/transactions/{id}/restore` | Restore owner yang dipilih. |
| `adminListDailySummaries` | `GET /api/v1/admin/users/{user_id}/daily-summaries` | Navigasi hari owner yang dipilih. |

Tambahkan schema/parameter reusable berikut:

- `TransactionType`: enum `income`, `expense`.
- `Money`: JSON string dengan pattern `^(0|[1-9][0-9]{0,11})\.[0-9]{2}$`; validasi service tetap menegakkan `0.01` sampai `999999999999.99`.
- `TransactionCreateRequest`: `transaction_date`, `type`, `category_id`, `amount`, `title`, dan `client_request_id`; `additionalProperties: false` dan tidak memiliki `user_id`, `owner_user_id`, `created_by`, `updated_by`, `isDelete`, atau `version`.
- `TransactionUpdateRequest`: seluruh field editable create kecuali `client_request_id`, ditambah `version`; `additionalProperties: false`.
- `VersionRequest`: hanya `version`; `additionalProperties: false`.
- `Transaction`: `id`, `user_id`, `category_id`, `category_name`, `transaction_date`, `type`, string `amount`, `title`, `client_request_id`, `created_by`, `updated_by`, `isDelete`, `version`, `created_at`, dan `updated_at`. `request_hash` tidak pernah diekspos.
- `TransactionResponse`, `TransactionListResponse`, `DailySummary`, `DailySummaryListResponse`, `Page`, dan `Scope`. Respons admin berisi `scope.mode=admin` dan `scope.owner_user_id`; respons personal tidak menerima owner dari client.
- List transaksi: `start_date` dan `end_date` wajib, `type`, `category_id`, `isDelete` default false, `limit` default 30 maksimum 100, dan `cursor` opsional. Rentang inklusif maksimal 366 hari.
- Ringkasan harian: `start_date`, `end_date`, `limit`, dan `cursor`; tidak menerima `isDelete`, tipe, atau kategori. Setiap item berisi `date`, string `income`, string `expense`, dan string `difference`. Hari kosong tidak dibuat sebagai bucket; rentang tanpa transaksi mengembalikan `data: []`.
- Navigasi hari memakai cursor `daily-summaries`; memilih satu tanggal memanggil list transaksi dengan `start_date=end_date=<tanggal>`, sehingga tidak diperlukan endpoint detail hari kedua.
- Dokumentasikan 400 untuk JSON/ID integer/query/cursor/`If-Match` yang rusak atau field terlarang; 401 untuk auth; 403 untuk route admin oleh non-admin; 404 untuk resource/kategori di luar owner; 409 untuk idempotency mismatch, lifecycle/version stale, target tidak aktif, atau kategori terhapus; 422 untuk title/amount/type/date/rentang yang tidak valid; 500 untuk kegagalan internal; 503 bila audit wajib tidak dapat disimpan.

## Model domain, uang, tanggal, dan cursor

Buat file berikut:

- `internal/service/money.go`: tipe `Money` berbasis integer sen, parser string desimal, formatter dua digit, serta konversi eksak ke/dari `pgtype.Numeric`. Dilarang memakai `float32`/`float64` pada DTO, service, repository, query arg, test fixture, atau agregasi.
- `internal/service/money_test.go`: batas `0.01`, `999999999999.99`, leading zero, digit desimal kurang/lebih, nilai nol/negatif/overflow, dan round-trip `pgtype.Numeric` tanpa pembulatan.
- `internal/service/transaction_cursor.go`: codec base64url JSON versioned untuk cursor aktif, Trash, dan daily summary.
- `internal/service/transaction_cursor_test.go`: round-trip dan penolakan cursor rusak, versi tak dikenal, endpoint/scope/filter berbeda, dan keyset tidak lengkap.
- `internal/service/transactions.go`: input/output domain, validasi, canonical hash, scope, dan orkestrasi repository.
- `internal/service/transactions_test.go`: validasi murni dan canonicalization tanpa database.

Aturan validasi service:

1. Trim title, wajib 1–200 karakter setelah trim.
2. Parse amount hanya dari string; format hasil selalu dua digit desimal dan nilainya 1–99.999.999.999.999 sen.
3. Parse `transaction_date` secara ketat sebagai `YYYY-MM-DD` tanpa timestamp.
4. Validasi timezone owner memakai `ValidateTimezone`, muat `time.Location`, lalu tolak tanggal transaksi setelah tanggal kalender `now` pada timezone owner. Inject clock ke service agar test tidak bergantung waktu nyata.
5. Type harus `income` atau `expense`.
6. Category harus berada pada owner yang sama, type harus sama, dan `is_delete` harus false untuk create/update/restore. Category owner lain atau tidak ada menjadi 404; type salah menjadi 422; category terhapus menjadi 409.
7. `version` harus positif. Create/update body menolak semua field tambahan sehingga owner dan lifecycle tidak dapat diinjeksi.
8. Canonical create payload terdiri dari version marker, owner ID, date canonical, lowercase type tervalidasi, category ID, amount dua digit, dan title hasil trim. Hash dengan SHA-256 hex; jangan masukkan actor atau urutan JSON. `client_request_id` adalah key terpisah dalam unique `(user_id, client_request_id)`.

Cursor menyimpan `v`, endpoint, actor, mode, owner, filter tanggal/tipe/kategori, state `isDelete`, dan keyset terakhir. Decoder harus membandingkan seluruh binding dengan request saat ini sebelum query. Cursor aktif menyimpan `(transaction_date, created_at, id)`, Trash menyimpan `(updated_at, id)`, dan daily summary menyimpan `date`. Cursor tidak pernah mengganti predicate owner pada SQL.

## Service dan repository

Tambahkan `internal/repository/transactions.go` dengan `Transactions` yang memegang pool dan generated queries. `internal/service/transactions.go` mengekspos metode berikut:

```text
List(ctx, TransactionScope, TransactionFilter) (TransactionPage, error)
Get(ctx, TransactionScope, id, deleted) (Transaction, error)
Create(ctx, TransactionScope, CreateTransactionInput) (CreateTransactionResult, error)
Update(ctx, TransactionScope, id, UpdateTransactionInput) (Transaction, error)
SoftDeleteTransaction(ctx, TransactionScope, id, expectedVersion) error
RestoreTransaction(ctx, TransactionScope, id, expectedVersion) (Transaction, error)
ListDailySummaries(ctx, TransactionScope, DailySummaryFilter) (DailySummaryPage, error)
```

`TransactionScope` berisi `Actor`, `Owner`, dan mode enum `personal|admin`. Scope personal wajib `Owner == Actor.UserID`; scope admin wajib actor role admin dan owner berasal dari path handler. Jangan menerima scope atau owner dalam body.

Repository menyediakan metode sepadan dengan nama `List`, `Get`, `Create`, `Update`, `SoftDelete`, `Restore`, dan `ListDailySummaries`. Repository mengembalikan sentinel yang cukup untuk service membedakan not found, stale lifecycle/version, category inactive/type mismatch, idempotent replay, dan idempotency conflict. Jangan mengubah error database mentah menjadi 404/409 di handler.

### Batas transaksi database

Semua create/update/delete/restore menjalankan satu transaksi PostgreSQL:

1. Lock actor dan owner melalui `LockUsersForUpdate`, yang sudah mengurutkan ID. Personal hanya perlu satu ID; admin memakai actor dan owner tanpa duplikasi.
2. Periksa actor masih aktif. Untuk mode admin, periksa role database saat ini masih `admin`; periksa owner ada dan aktif untuk semua write.
3. Untuk create/update/restore, lock kategori owner dengan query `FOR UPDATE`, lalu periksa type dan `is_delete=false`. Urutan selalu users dahulu, category kemudian, transaction row/insert terakhir agar race dengan delete category tidak melewati validasi.
4. Jalankan insert atau mutasi dengan predicate `id`, `user_id`, expected `version`, dan expected `is_delete`.
5. Untuk mode admin, insert `admin_access_events` dengan `resource_type=transaction`, action `create|update|delete|restore`, actor/target/resource/version/replay metadata aman, dan request ID yang diteruskan handler dari middleware Echo. Jangan membuat request ID baru bila request sudah memiliki `X-Request-ID`, serta jangan simpan title, amount, atau payload.
6. Commit mutasi dan audit bersama. Kegagalan audit membatalkan mutasi dan dipetakan ke 503.

Create memakai `INSERT ... ON CONFLICT (user_id, client_request_id) DO NOTHING RETURNING ...`. Jika insert tidak mengembalikan baris, ambil row `(owner, client_request_id) FOR UPDATE`: hash sama dan row aktif menghasilkan replay 200 terhadap resource hasil create asli dengan ID/`client_request_id` yang sama; hash berbeda atau row sudah terhapus menghasilkan 409. `request_hash` tetap hash payload create awal ketika transaksi kemudian diedit. Test idempotensi wajib mencakup dua create bersamaan agar hanya satu ID tersimpan.

Update harus mengunci/mengecek state resource owner sebelum mutasi agar owner salah/hilang menjadi 404 dan row milik owner dengan version atau lifecycle salah menjadi 409. Delete memakai pola yang sama tetapi tidak mensyaratkan kategori aktif sehingga transaksi historis tetap dapat dihapus. Restore mensyaratkan kategori terkait kembali aktif dan cocok sebelum `is_delete=false`.

Read personal mengandalkan actor database-backed dari middleware dan tetap memakai predicate owner. Read admin memeriksa actor/target saat ini, menjalankan read, menulis event audit `read|list|daily_summary`, lalu commit audit sebelum response. Resource owner lain tetap 404. Target terhapus boleh dibaca admin sesuai kontrak, tetapi tidak boleh ditulis sebelum account restore.

## Query SQL dan migrasi

Ubah `db/queries/transactions.sql` dan tambahkan `GetUserByID` (termasuk user terhapus) di `db/queries/users.sql` untuk validasi target pada read admin. Lalu generate ulang `internal/repository/sqlc/transactions.sql.go`, `internal/repository/sqlc/users.sql.go`, dan `internal/repository/sqlc/querier.go`. Pertahankan generated files hanya melalui `make sqlc-generate`.

Query yang harus tersedia:

- `LockScopedCategoryForTransaction`: category berdasarkan `(id,user_id) FOR UPDATE`, tanpa filter lifecycle/type agar service dapat membedakan 404, 409, dan 422.
- `InsertTransaction`: insert tervalidasi dengan `ON CONFLICT (user_id,client_request_id) DO NOTHING RETURNING *`.
- `GetTransactionByRequestIDForUpdate`: idempotency row owner-scoped termasuk baris terhapus.
- `GetTransactionStateForUpdate`: row berdasarkan `(id,user_id) FOR UPDATE` tanpa filter lifecycle untuk klasifikasi stale versus missing pada mutasi.
- `GetActiveTransaction` dan `GetDeletedTransaction`: join category berdasarkan owner/type tanpa memfilter `categories.is_delete`, sehingga `category_name` historis tetap tersedia.
- `ListActiveTransactionsPage`: predicate owner, `is_delete=false`, tanggal inklusif, filter type/category opsional, keyset `(transaction_date,created_at,id) < (...)`, urutan ketiganya DESC, dan `LIMIT limit+1`.
- `ListDeletedTransactionsPage`: predicate owner, `is_delete=true`, filter tanggal/type/category yang sama, keyset `(updated_at,id) < (...)`, urutan DESC, dan `LIMIT limit+1`.
- `UpdateTransaction`, `SoftDeleteTransaction`, dan `RestoreTransaction`: predicate owner, expected version/lifecycle, actor pada `updated_by`, version increment, dan `updated_at=CURRENT_TIMESTAMP`. Validasi kategori dilakukan setelah lock dan tetap dipertahankan sebagai predicate defensif pada update/restore.
- `ListDailySummariesPage`: owner dan rentang tanggal, `is_delete=false`, group per `transaction_date`, `COALESCE(SUM(amount) FILTER (WHERE type='income'),0)` dan expense, urutan tanggal DESC, keyset tanggal, `LIMIT limit+1`. Difference dihitung secara eksak dari numeric hasil query/service.

Tambahkan migrasi maju `db/migrations/000006_transaction_navigation_indexes.up.sql` dan pasangan `.down.sql`. Migrasi up mengganti `transactions_active_list_idx` menjadi `(user_id, transaction_date DESC, created_at DESC, id DESC) WHERE is_delete=false`; down mengembalikan index Issue 002. Index Trash saat ini `(user_id,updated_at DESC,id DESC) WHERE is_delete=true` sudah sesuai. Jangan membuat tabel income/expense terpisah atau kolom float.

## Handler dan wiring

Tambahkan `internal/handlers/transactions.go` dan `internal/handlers/transactions_test.go`.

- Handler memakai generated query/path/header types dan `decodeJSON` yang menolak unknown fields.
- `personalScope` selalu menyalin actor dari middleware menjadi actor sekaligus owner.
- `adminScope` selalu memakai actor middleware dan BIGINT `{user_id}` path sebagai owner; tidak ada fallback ke body/query.
- Handler mengambil request ID yang sudah dibuat middleware RequestID milik Echo dari context/header dan memasukkannya ke scope/call metadata untuk audit; repository tidak membuat identitas request pengganti.
- Map hasil create `Created=true` ke 201 dan replay ke 200; delete mengembalikan 204 tanpa body; list selalu mengembalikan array non-nil dan `page.next_cursor` null pada akhir.
- Format amount/total hanya melalui `Money.String()`. Timestamp audit/resource keluar sebagai UTC RFC3339 dan date sebagai `YYYY-MM-DD`.
- Map error sesuai kontrak OpenAPI, termasuk 503 untuk audit wajib. Semua route protected mewarisi `Cache-Control: no-store` dari middleware.

Ubah file wiring berikut:

- `internal/handlers/openapi.go`: tambah field transaction handler, constructor argument, dan seluruh adapter method generated personal/admin.
- `internal/handlers/routes.go`: tambah transaction handler pada `RegisterRoutes`; pasang `authenticate` pada semua operasi transaksi/daily summary dan `authenticate` lalu `middleware.RequireAdmin` pada semua operationId admin.
- `cmd/api/main.go`: buat `repository.NewTransactions`/`service.NewTransactions` sesuai constructor final, buat handler, dan injeksikan ke route registration.

Jangan memasang route manual di luar generated OpenAPI router.

## Pengujian

### Unit service

`internal/service/money_test.go`, `internal/service/transaction_cursor_test.go`, dan `internal/service/transactions_test.go` mencakup:

- parsing/format uang tanpa float dan seluruh batas;
- title trim/panjang, ID integer, type, strict date, future date pada `Asia/Jakarta` dan `UTC`, timezone invalid;
- canonical hash stabil untuk JSON/order/whitespace amount-title yang ekuivalen dan berbeda untuk perubahan field;
- cursor rusak serta reuse lintas actor/mode/owner/filter/active-Trash/endpoint;
- personal scope tidak dapat memilih owner lain dan mode admin memerlukan role admin.

### Unit handler

`internal/handlers/transactions_test.go` memakai fake service untuk membuktikan:

- field `user_id`, `owner_user_id`, `created_by`, `isDelete`, field uang numerik, dan unknown field ditolak sebelum service;
- personal create memakai actor sebagai owner;
- admin create memakai target path sebagai owner dan actor tetap admin;
- non-admin ditolak oleh middleware sebelum target/service dipanggil;
- 201 create versus 200 replay, `If-Match`, restore version, default `isDelete=false`, page cursor, envelope admin, dan pemetaan 400/404/409/422/503.

### Integrasi PostgreSQL/service

Tambahkan `internal/service/transactions_integration_test.go` dengan schema terisolasi dari helper integrasi yang sudah ada. Cakup:

1. Create income dan expense pada satu tabel; string amount round-trip dua digit dan agregasi daily tepat.
2. Category owner lain → 404; type salah → validation; category deleted → conflict pada create/update/restore.
3. Replay key+payload sama mengembalikan ID yang sama; payload beda → conflict; replay setelah delete → conflict; dua goroutine create menghasilkan tepat satu row.
4. User A tidak dapat get/list/update/delete/restore row B. Admin C dapat melakukan seluruh lifecycle B, owner tetap B, created/updated actor C benar, dan audit tersimpan.
5. Admin route target A dengan transaction/category B → 404; user biasa dengan scope admin → forbidden; owner terhapus menolak write.
6. Update stale, double delete, stale restore, double restore, dan race update/delete/restore menghasilkan paling banyak satu pemenang per expected version.
7. Delete menyimpan row, menaikkan version, mengeluarkannya dari list aktif dan daily total, memasukkannya ke Trash; restore memasukkan total tepat sekali.
8. Category yang dihapus setelah transaksi dibuat tidak menghilangkan detail/list label historis; delete transaksi tetap berhasil; update/restore gagal sampai category aktif atau kategori aktif lain dipilih saat update.
9. Pagination aktif dengan tanggal dan `created_at` sama serta pagination Trash dengan `updated_at` sama tidak melewatkan/menggandakan ID; filter date/type/category tetap terikat cursor.
10. Daily summary hanya berisi hari aktif dalam rentang, urut terbaru, cursor stabil, dan rentang kosong mengembalikan array kosong serta tidak memakai float.
11. Kegagalan insert audit admin me-rollback write. Read admin menghasilkan audit tanpa payload finansial.

Perluas `internal/repository/sqlc/integration_test.go` hanya untuk constraint/query yang tidak sudah dibuktikan pada test service: precision/range `NUMERIC(14,2)`, FK category owner/type, query keyset, dan migration index. Jangan menduplikasi lifecycle service secara mekanis.

Ubah target `test-integration` pada `Makefile` agar menjalankan `./internal/repository/sqlc ./internal/service`; handler HTTP tetap diuji oleh unit handler dengan middleware nyata.

## Urutan implementasi

1. Pastikan Issue 005 dan baseline bersih dari kegagalan yang tidak terkait.
2. Tambahkan kontrak OpenAPI transaksi/daily/admin dan generate code.
3. Tambahkan migrasi index dan rewrite query SQL; generate sqlc dan review tipe generated agar amount tetap `pgtype.Numeric`.
4. Implementasikan `Money`, validasi tanggal/timezone, canonical hash, dan cursor beserta unit test.
5. Implementasikan repository dengan transaksi DB, lock, idempotensi, optimistic lifecycle, dan audit atomik.
6. Implementasikan service personal/admin dan daily summaries.
7. Implementasikan handler, generated adapter, middleware operation map, dan composition root.
8. Jalankan test unit, integrasi PostgreSQL, race, vet, build, sqlc drift, dan OpenAPI drift; perbaiki hanya regresi dalam scope Issue 006.
9. Bandingkan bukti test dengan setiap kriteria ISSUE-006 sebelum mengubah status issue/index pada tahap implementasi. Refinement plan ini sendiri tidak menyelesaikan issue.

## Perintah verifikasi implementasi

Gunakan toolchain aktual proyek dan database fixture yang boleh dibuang:

```sh
GO=/usr/local/go/bin/go GOPATH=/tmp/monelog-go-path GOMODCACHE=/tmp/monelog-go-modcache GOCACHE=/tmp/monelog-go-cache make fmt-check test vet build verify
SQLC=/snap/bin/sqlc make sqlc-vet sqlc-check
GO=/usr/local/go/bin/go GOPATH=/tmp/monelog-go-path GOMODCACHE=/tmp/monelog-go-modcache GOCACHE=/tmp/monelog-go-cache make oapi-check
TEST_DATABASE_URL='postgresql://…' GO=/usr/local/go/bin/go GOPATH=/tmp/monelog-go-path GOMODCACHE=/tmp/monelog-go-modcache GOCACHE=/tmp/monelog-go-cache make test-integration
TEST_DATABASE_URL='postgresql://…' GO=/usr/local/go/bin/go GOPATH=/tmp/monelog-go-path GOMODCACHE=/tmp/monelog-go-modcache GOCACHE=/tmp/monelog-go-cache /usr/local/go/bin/go test -race -tags=integration ./internal/service
```

Catat command, database target non-produksi, hasil, dan kegagalan nyata. Jangan menjalankan migration down terhadap data bersama, serta jangan commit atau push tanpa instruksi terpisah.
