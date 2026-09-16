# PLAN-004: Kontrak API pemilik/admin dan penghapusan lunak

Status: Disempurnakan — siap direview sebelum perubahan kontrak atau kode generate.
Diperbarui: 16 September 2026 (v0.4)
Issue: [ISSUE-004](../issues/ISSUE-004-api-contract.md)
Repositori: monelog-api
Prasyarat: 003
Persyaratan: FR-01, FR-15, FR-16, FR-17

## Tujuan dan batas perubahan

ISSUE-004 menetapkan kontrak OpenAPI kanonis untuk route yang sudah berjalan dan seluruh endpoint MVP yang sudah direncanakan. Kontrak harus cukup presisi untuk menghasilkan model dan interface Go tanpa mengubah aturan owner/admin, uang, penghapusan lunak, atau konkurensi.

Refinement ini hanya mengubah dokumen rencana. Jangan mengubah `api/openapi.yaml`, `api/generate.yaml`, `internal/api/openapi.gen.go`, handler, service, query, migrasi, curl di README, atau route runtime tanpa konfirmasi pengguna. Jangan membuat handler kategori atau transaksi pada ISSUE-004. Jangan commit atau push hasil refinement ini.

## Temuan repositori saat ini

- ISSUE-003 telah mengimplementasikan health, register, login, refresh, logout, `/me`, JWT Bearer, refresh rotation, pemeriksaan actor aktif, dan role database saat ini.
- Curl register/login dan URL Swagger sudah ada di `README.md` serta `README.en.md`; pertahankan tanpa perubahan pada ISSUE-004 kecuali pengguna mengonfirmasi.
- Kontrak aktif berada di `api/openapi.yaml` dengan `openapi: 3.0.3`.
- Generator dipatok melalui `go generate ./api` ke `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0`.
- `api/generate.yaml` sudah memakai package `api`, output `../internal/api/openapi.gen.go`, `echo5-server`, `models`, `embedded-spec`, dan `strict-server`.
- Runtime memakai `github.com/labstack/echo/v5 v5.0.2`; karena itu `echo-server` untuk Echo v4 tidak boleh dipakai.
- Swagger UI tersedia pada `/swagger/`, redirect `/swagger`, dan dokumen runtime pada `/api/openapi.json`.
- Worktree memiliki perubahan README milik pengguna sebelum refinement ini. Perubahan tersebut harus dipertahankan dan tidak dimasukkan ke scope PLAN-004.

## Keputusan kontrak

### Versi dan format

- Gunakan OpenAPI `3.0.3` secara tepat. Jangan berpindah ke 3.1.x pada ISSUE-004 karena kontrak, generator, dan output saat ini sudah memakai 3.0.3.
- Base API adalah `/api/v1`; health dan dokumentasi berada di luar base tersebut.
- Semua operationId harus unik, lower camel case, dan stabil karena menjadi nama method hasil generate.
- JSON memakai `snake_case`, kecuali field boolean yang secara eksplisit bernama `isDelete`.
- Setiap object request memakai `additionalProperties: false`. Field `owner_user_id`, `user_id`, `role`, atau `isDelete` hanya boleh diterima bila operasi secara eksplisit mendefinisikannya.
- Operasi memakai `x-monelog-status: implemented` atau `x-monelog-status: planned`. Summary operasi planned diawali `PLANNED —` agar Swagger tidak menyatakannya sudah tersedia.
- Planned operation boleh didokumentasikan dalam kontrak, tetapi tidak boleh masuk generated Echo router sampai handler issue terkait tersedia. Request ke route yang belum diimplementasikan harus tetap 404, bukan placeholder 200/501.

### Bearer authentication

- Definisikan `components.securitySchemes.BearerAuth` sebagai HTTP `bearer` dengan `bearerFormat: JWT`.
- Jadikan BearerAuth default pada operasi `/api/v1` yang terlindungi. Beri `security: []` hanya pada health, register, login, refresh, logout, dan endpoint dokumentasi publik.
- JWT hanya menyimpan actor pada `sub`; role tidak menjadi sumber otorisasi token. Middleware memuat user aktif dan role terbaru dari database pada setiap request.
- Token hilang/tidak valid, actor terhapus, expiry, signature, algorithm, issuer, audience, atau claim role yang tidak sah menghasilkan `401 AUTHENTICATION_FAILED` yang sama.
- Respons terautentikasi dan respons auth memakai `Cache-Control: no-store`.

### Otorisasi owner/admin

- Route personal selalu memakai `owner_user_id = actor_user_id`, termasuk ketika actor adalah admin.
- Route `/admin/users/{user_id}/...` memerlukan actor dengan role `admin` terbaru; `{user_id}` adalah owner target dan tidak boleh dioverride dari body.
- Pengguna biasa yang memanggil route admin mendapat 403 sebelum target dicari.
- ID resource yang hilang atau bukan milik owner target menghasilkan 404 agar keberadaan resource lintas owner tidak bocor.
- Admin dapat CRUD, delete, restore, report, dan kelak export untuk owner target. Record baru tetap dimiliki target; `created_by`/`updated_by` dan audit menyimpan actor admin.
- Write finansial memerlukan target aktif. Data target terhapus boleh dibaca admin hanya pada operasi yang secara eksplisit mendukungnya; akun harus dipulihkan sebelum write finansial.
- Setiap respons admin bertubuh memuat `scope: {mode: "admin", owner_user_id}`. Respons personal bertubuh memuat `scope: {mode: "personal", owner_user_id}`. Endpoint auth yang belum memiliki owner scope mempertahankan envelope ISSUE-003.

## Inventaris endpoint

### Route yang sudah diimplementasikan

| Method | Path | operationId | Auth | Hasil utama |
| --- | --- | --- | --- | --- |
| GET | `/health/live` | `healthLive` | Publik | 200/503 `HealthResponse` |
| GET | `/health/ready` | `healthReady` | Publik | 200/503 `HealthResponse` |
| POST | `/api/v1/auth/register` | `register` | Publik | 201 `UserResponse` |
| POST | `/api/v1/auth/login` | `login` | Publik | 200 access token dan transport refresh |
| POST | `/api/v1/auth/refresh` | `refresh` | Publik dengan aturan refresh/CSRF | 200 sesi berotasi |
| POST | `/api/v1/auth/logout` | `logout` | Publik dengan aturan refresh/CSRF | 204 |
| GET | `/api/v1/me` | `getMe` | Bearer | 200 profil actor |
| PATCH | `/api/v1/me` | `updateMe` | Bearer | 200 profil terbaru |
| DELETE | `/api/v1/me` | `deleteMe` | Bearer + If-Match | 204 |

Route dokumentasi runtime yang sudah ada adalah `GET /swagger` (302 ke `/swagger/`), `GET /swagger/` (UI), dan `GET /api/openapi.json` (dokumen embedded). Route ini dicatat sebagai infrastruktur dokumentasi dan tidak menghasilkan interface domain.

### Endpoint akun/admin MVP yang direncanakan

| Method | Path | operationId | Hasil utama |
| --- | --- | --- | --- |
| GET | `/api/v1/admin/users` | `listAdminUsers` | 200 directory user berpaginasi |
| POST | `/api/v1/admin/users` | `createAdminUser` | 201 user aktif baru |
| GET | `/api/v1/admin/users/{user_id}` | `getAdminUser` | 200 metadata target |
| PATCH | `/api/v1/admin/users/{user_id}` | `updateAdminUser` | 200 timezone/role terbaru |
| DELETE | `/api/v1/admin/users/{user_id}` | `deleteAdminUser` | 204 soft delete dan cabut sesi |
| POST | `/api/v1/admin/users/{user_id}/restore` | `restoreAdminUser` | 200 user aktif terbaru |
| GET | `/api/v1/admin/audit-events` | `listAdminAuditEvents` | 200 event audit berpaginasi |

### Endpoint kategori MVP yang direncanakan

| Method | Personal | operationId | Admin target | operationId admin |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/categories` | `listCategories` | `/api/v1/admin/users/{user_id}/categories` | `listAdminUserCategories` |
| POST | `/api/v1/categories` | `createCategory` | `/api/v1/admin/users/{user_id}/categories` | `createAdminUserCategory` |
| GET | `/api/v1/categories/{category_id}` | `getCategory` | `/api/v1/admin/users/{user_id}/categories/{category_id}` | `getAdminUserCategory` |
| PATCH | `/api/v1/categories/{category_id}` | `updateCategory` | `/api/v1/admin/users/{user_id}/categories/{category_id}` | `updateAdminUserCategory` |
| DELETE | `/api/v1/categories/{category_id}` | `deleteCategory` | `/api/v1/admin/users/{user_id}/categories/{category_id}` | `deleteAdminUserCategory` |
| POST | `/api/v1/categories/{category_id}/restore` | `restoreCategory` | `/api/v1/admin/users/{user_id}/categories/{category_id}/restore` | `restoreAdminUserCategory` |

### Endpoint transaksi MVP yang direncanakan

| Method | Personal | operationId | Admin target | operationId admin |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/transactions` | `listTransactions` | `/api/v1/admin/users/{user_id}/transactions` | `listAdminUserTransactions` |
| POST | `/api/v1/transactions` | `createTransaction` | `/api/v1/admin/users/{user_id}/transactions` | `createAdminUserTransaction` |
| GET | `/api/v1/transactions/{transaction_id}` | `getTransaction` | `/api/v1/admin/users/{user_id}/transactions/{transaction_id}` | `getAdminUserTransaction` |
| PATCH | `/api/v1/transactions/{transaction_id}` | `updateTransaction` | `/api/v1/admin/users/{user_id}/transactions/{transaction_id}` | `updateAdminUserTransaction` |
| DELETE | `/api/v1/transactions/{transaction_id}` | `deleteTransaction` | `/api/v1/admin/users/{user_id}/transactions/{transaction_id}` | `deleteAdminUserTransaction` |
| POST | `/api/v1/transactions/{transaction_id}/restore` | `restoreTransaction` | `/api/v1/admin/users/{user_id}/transactions/{transaction_id}/restore` | `restoreAdminUserTransaction` |

Create transaction sukses pertama mengembalikan 201. Replay dengan `client_request_id` dan payload kanonis sama mengembalikan 200 record yang sama; replay berbeda atau replay record terhapus mengembalikan 409.

### Endpoint ringkasan/laporan MVP yang direncanakan

| Method | Personal | operationId | Admin target | operationId admin |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/daily-summaries` | `listDailySummaries` | `/api/v1/admin/users/{user_id}/daily-summaries` | `listAdminUserDailySummaries` |
| GET | `/api/v1/reports/summary` | `getReportSummary` | `/api/v1/admin/users/{user_id}/reports/summary` | `getAdminUserReportSummary` |
| GET | `/api/v1/reports/breakdown` | `getReportBreakdown` | `/api/v1/admin/users/{user_id}/reports/breakdown` | `getAdminUserReportBreakdown` |

### Endpoint export masa depan — planned, belum diimplementasikan

Operasi berikut milik ISSUE-009. Tandai tag dan summary sebagai `Planned exports (not implemented)`, beri `x-monelog-status: planned`, kecualikan operationId dari generation allow-list, dan pastikan runtime tetap 404 sampai ISSUE-009 menyediakan handler.

| Method | Personal | operationId | Admin target | operationId admin |
| --- | --- | --- | --- | --- |
| POST | `/api/v1/exports` | `createExport` | `/api/v1/admin/users/{user_id}/exports` | `createAdminUserExport` |
| GET | `/api/v1/exports/{export_id}` | `getExport` | `/api/v1/admin/users/{user_id}/exports/{export_id}` | `getAdminUserExport` |
| GET | `/api/v1/exports/{export_id}/download` | `downloadExport` | `/api/v1/admin/users/{user_id}/exports/{export_id}/download` | `downloadAdminUserExport` |

Create export kelak mengembalikan 202. Status job adalah `queued`, `running`, `succeeded`, `failed`, atau `canceled`; download mengembalikan file hanya untuk job sukses, 409 bila belum siap, dan 410 bila artifact kedaluwarsa.

Templat (ISSUE-011) dan Drive/backup (ISSUE-012) berada setelah MVP/release terkait dan belum memiliki bentuk request/state yang final. Jangan memasukkan operasi tersebut ke kontrak ISSUE-004; issue pemilik harus menetapkan endpoint lengkap sebelum generation.

## Skema request dan response

### Primitive dan enum bersama

| Schema | Bentuk |
| --- | --- |
| `UUID` | string, format `uuid` |
| `LocalDate` | string, format `date`, `YYYY-MM-DD` |
| `UTCDateTime` | string, format `date-time`, timestamp UTC RFC3339 |
| `Version` | integer int32, minimum 1 |
| `Cursor` | opaque string; client tidak boleh membongkar atau membuat sendiri |
| `Role` | string enum `user`, `admin` |
| `TransactionType` | string enum `income`, `expense` |
| `Currency` | string enum `IDR` |
| `RequestMode` | string enum `personal`, `admin` |
| `ExportFormat` | string enum `xlsx`, `pdf` |
| `ExportStatus` | string enum `queued`, `running`, `succeeded`, `failed`, `canceled` |

`isDelete` selalu required boolean pada response `User`, `Category`, dan `Transaction`. Nilai string seperti `"false"`, angka, dan null tidak valid. Field ini tidak boleh ada pada create/PATCH biasa; hanya DELETE dan `/restore` yang mengubahnya.

### Representasi uang

- Semua nilai uang API bertipe string, bukan `number`, `integer`, float32, atau float64.
- `Money` adalah string desimal non-negatif dengan tepat dua digit pecahan dan pattern `^[0-9]+\.[0-9]{2}$`; nol ditulis `"0.00"`.
- `PositiveMoney` memakai representasi yang sama dengan batas domain `0.01` sampai `999999999999.99`; validasi batas dilakukan sebagai decimal eksak di service, bukan konversi float.
- `SignedMoney` untuk `difference` memakai pattern `^-?[0-9]+\.[0-9]{2}$`.
- `Transaction.amount` dan input amount memakai `PositiveMoney`. Total income/expense/category memakai `Money`; difference memakai `SignedMoney`.

### Auth dan profil yang sudah berjalan

| Schema | Field required / aturan |
| --- | --- |
| `RegisterRequest` | `email`, `password` 12–128 byte UTF-8, `timezone` IANA/UTC |
| `LoginRequest` | `email`, `password`, `client_type` enum `native`/`web` |
| `NativeRefreshRequest` | `client_type: native`, `refresh_token` |
| `WebRefreshRequest` | `client_type: web`; secret dari cookie, bukan JSON |
| `ProfileUpdateRequest` | `timezone`, `version` |
| `User` | `id`, `email`, `role`, `timezone`, `currency`, `isDelete`, `version` |
| `SessionResponse` | `access_token`, `token_type: Bearer`, `expires_in`; `refresh_token` hanya native |

Refresh/logout browser mendokumentasikan cookie `__Host-monelog-refresh`, cookie `__Host-monelog-csrf`, header `X-CSRF-Token`, dan header `Origin`. Karena OpenAPI 3.0 tidak dapat menyatakan dependensi kondisional transport secara penuh, gunakan `oneOf` untuk native/web dan jelaskan syarat cookie/header dalam description serta examples.

### User admin dan audit

| Schema | Field required / aturan |
| --- | --- |
| `AdminUserCreateRequest` | `email`, `password`, `timezone`, optional `role` default `user`; tidak menerima `isDelete` |
| `AdminUserUpdateRequest` | required `version` dan minimal salah satu `timezone`/`role`; email immutable |
| `RestoreRequest` | `version` minimum 1 |
| `UserListResponse` | `data: User[]`, `page`, tanpa password/hash/session |
| `AdminAuditEvent` | `id`, `actor_user_id`, nullable `target_user_id`, `resource_type`, nullable `resource_id`, `action`, `outcome`, `request_id`, `safe_metadata`, `created_at` |
| `AdminAuditEventListResponse` | `data: AdminAuditEvent[]`, `page` |

`safe_metadata` hanya object JSON yang disanitasi dan tidak memuat password, token, payload finansial, atau secret provider.

### Category

| Schema | Field required / aturan |
| --- | --- |
| `CategoryCreateRequest` | `name` trimmed 1–80, `type`; tidak menerima owner/version/isDelete |
| `CategoryUpdateRequest` | `name`, `version`; type immutable |
| `Category` | `id`, `type`, `name`, `isDelete`, `version`, `created_at`, `updated_at` |
| `CategoryResponse` | `data: Category`, `scope` |
| `CategoryListResponse` | `data: Category[]`, `page`, `scope` |

### Transaction

| Schema | Field required / aturan |
| --- | --- |
| `TransactionCreateRequest` | `transaction_date`, `type`, `category_id`, `amount`, `title`, `client_request_id` |
| `TransactionUpdateRequest` | required `version` dan minimal satu dari `transaction_date`, `type`, `category_id`, `amount`, `title` |
| `Transaction` | `id`, `user_id`, `transaction_date`, `type`, `category_id`, `amount`, `title`, `client_request_id`, `created_by`, `updated_by`, `isDelete`, `version`, `created_at`, `updated_at` |
| `TransactionResponse` | `data: Transaction`, `scope` |
| `TransactionListResponse` | `data: Transaction[]`, `page`, `scope` |

Title di-trim, panjang 1–200. Tanggal masa depan ditolak untuk create/update MVP. Category harus aktif, memiliki owner yang sama, dan type yang sama pada create/update/restore.

### Daily summary dan report

| Schema | Field required / aturan |
| --- | --- |
| `DailySummary` | `date`, `income: Money`, `expense: Money`, `difference: SignedMoney` |
| `DailySummaryListResponse` | `data: DailySummary[]`, `page`, `scope` |
| `CategoryTotal` | `category_id`, `name`, `amount: Money` |
| `ReportSummary` | `start_date`, `end_date`, `income`, `expense`, `difference`, `top_income_categories`, `top_expense_categories` |
| `TimeBreakdownItem` | `period_start`, `income`, `expense`, `difference` untuk week/month |
| `CategoryBreakdownItem` | `category_id`, `name`, `type`, `amount` untuk category |
| `ReportBreakdownResponse` | `data` memakai `oneOf` sesuai `group_by`, ditambah `scope` |

List/report/export selalu mengecualikan transaksi `isDelete=true`. Empty report memakai `"0.00"`, signed zero `"0.00"`, dan array kosong.

### Export planned

| Schema | Field required / aturan |
| --- | --- |
| `ExportCreateRequest` | `start_date`, `end_date`, optional `type`, optional `category_id`, `format` |
| `ExportJob` | `id`, `owner_user_id`, `requested_by`, `request_mode`, filter immutable, `format`, `status`, nullable `expires_at`, `created_at`, `updated_at` |
| `ExportJobResponse` | `data: ExportJob`, `scope` |

Download response mendefinisikan media type XLSX/PDF, `Content-Disposition`, dan stream binary. Export tidak menerima `isDelete` dan selalu memakai transaksi aktif.

## Envelope response dan error

Response resource sukses memakai `{ "data": ..., "scope": ... }`; response list memakai `{ "data": [...], "page": {"next_cursor": string|null}, "scope": ... }`. Auth mempertahankan bentuk sesi ISSUE-003. Respons 204 tidak memiliki body.

Semua error API non-health memakai envelope yang sama:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request.",
    "fields": {
      "amount": "Must be between 0.01 and 999999999999.99."
    },
    "request_id": "opaque-request-id"
  }
}
```

- `error.code`, `error.message`, dan `error.request_id` required.
- `error.fields` optional object `additionalProperties: string`; gunakan hanya untuk kesalahan field yang aman ditampilkan.
- Pesan 5xx disanitasi dan tidak memuat SQL, DSN, password, token, cookie, hash, atau stack trace.
- Pemetaan: 400 `BAD_REQUEST`; 401 `AUTHENTICATION_FAILED`; 403 `FORBIDDEN`; 404 `NOT_FOUND`; 409 `CONFLICT`; 410 `GONE`; 422 `VALIDATION_ERROR`; 429 `RATE_LIMITED`; 500 `INTERNAL_ERROR`; 503 `SERVICE_UNAVAILABLE`.
- Login email tidak dikenal dan password salah tetap memakai body 401 identik.
- Implementasi ISSUE-003 saat ini hanya mengirim `code` dan `message`. Penambahan `request_id`/`fields` adalah drift yang harus direview dan dikonfirmasi sebelum kode handler diubah.

## Pagination, filter, dan rentang tanggal

- `limit`: query integer, default 30, minimum 1, maksimum 100.
- `cursor`: query string opaque. Cursor mengikat actor, mode, owner, operation, filter, sort, dan `isDelete`; cursor rusak atau lintas scope menghasilkan 400.
- List response memakai `page.next_cursor`, nullable; null menandai halaman terakhir.
- User directory: optional `q` trimmed 1–100 dan optional `isDelete` default false; urut `email ASC, id ASC`.
- Category list: optional `type`, optional `isDelete` default false; urut `name ASC, id ASC`.
- Transaction list: required `start_date`, `end_date`; optional `type`, `category_id`, `isDelete` default false, `limit`, `cursor`. Active sort `transaction_date DESC, created_at DESC, id DESC`; Trash `updated_at DESC, id DESC`.
- Daily summaries: required `start_date`, `end_date`, `limit`, `cursor`; urut `date DESC`.
- Report summary/breakdown: required `start_date`, `end_date`; breakdown juga required `group_by` enum `week`, `month`, `category`.
- Audit list: optional `actor_user_id`, `target_user_id`, `action`, `start_time`, `end_time`, `limit`, `cursor`; urut `created_at DESC, id DESC`.
- Rentang date inklusif, `start_date <= end_date`, maksimum 366 hari. Audit memakai UTC timestamps dengan rentang maksimum 366 hari. Preset UI 7/30 hari dihitung client memakai timezone owner, lalu mengirim tanggal eksplisit.
- `category_id` dari owner lain menghasilkan 404. Parameter/filter yang tidak dikenal atau tidak berlaku menghasilkan 400.

## Konkurensi dan lifecycle

- DELETE user/category/transaction personal maupun admin wajib membawa `If-Match` berisi ETag versi kuat, misalnya `If-Match: "3"`.
- Schema header adalah string pattern `^\"[1-9][0-9]*\"$`. Header hilang/rusak menghasilkan 400; versi stale atau lifecycle sudah berubah menghasilkan 409; owner/resource salah menghasilkan 404.
- PATCH tetap membawa `version` di body sesuai kontrak saat ini. `/restore` memakai `RestoreRequest {version}`. Jangan menerima versi dari dua tempat pada operasi yang sama.
- GET detail, create, PATCH, dan restore mengembalikan header `ETag: "<version>"` agar client dapat menyiapkan DELETE berikutnya.
- DELETE mengubah `isDelete` ke true dan menaikkan version; restore mengubahnya ke false dan menaikkan version. Tidak ada hard delete endpoint.
- Field `isDelete` tidak boleh ditulis langsung oleh client. Query `isDelete=true` hanya memilih Trash dan tidak mengubah scope owner.

## Konfigurasi oapi-codegen

Pertahankan generator `v2.8.0` dan Echo v5:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/v2.8.0/configuration-schema.json
package: api
output: ../internal/api/openapi.gen.go
generate:
  echo5-server: true
  models: true
  embedded-spec: true
  strict-server: true
output-options:
  include-operation-ids:
    - healthLive
    - healthReady
    - register
    - login
    - refresh
    - logout
    - getMe
    - updateMe
    - deleteMe
```

Allow-list di atas adalah batas runtime ISSUE-004. Tambahkan operationId planned hanya pada issue yang sekaligus menyediakan handler, service, otorisasi, dan test. Jangan mengganti `echo5-server` dengan `echo-server`. Jangan mengedit `internal/api/openapi.gen.go` secara manual.

## Urutan implementasi setelah konfirmasi

1. Pertahankan operasi ISSUE-003 beserta curl dan perilaku runtime; tandai sebagai implemented di kontrak.
2. Tambahkan komponen schema, parameter, header, response, security, dan contoh yang ditentukan plan ini.
3. Tambahkan operasi MVP planned dan export planned dengan status eksplisit; jangan membuat placeholder handler.
4. Tambahkan generation allow-list untuk sembilan operationId yang sudah berjalan agar route planned tidak terdaftar.
5. Regenerasi `internal/api/openapi.gen.go`; diff hasil generate harus hanya berasal dari kontrak/config.
6. Tambahkan contract test yang memuat embedded spec, menjalankan `document.Validate(context.Background())`, memeriksa operationId unik, status implemented/planned, security, `isDelete`, money-as-string, error envelope, dan seluruh pasangan personal/admin.
7. Pastikan route planned masih 404 dan route ISSUE-003 tidak berubah melalui handler tests yang sudah ada.
8. Review drift yang memerlukan perubahan runtime, terutama standard error `request_id`/`fields` dan ETag response. Jangan mengubah handler sampai pengguna mengonfirmasi scope tersebut.

## Validasi, lint, generation, dan drift check

Tambahkan target Make setelah implementasi disetujui:

```make
VACUUM_VERSION ?= v0.30.1

oapi-validate:
	VACUUM_NO_UPDATE_CHECK=true $(GO) run github.com/daveshanley/vacuum@$(VACUUM_VERSION) lint -d -e api/openapi.yaml

oapi-generate:
	$(GO) generate ./api

oapi-check: oapi-validate oapi-generate
	git diff --exit-code -- internal/api/openapi.gen.go
```

Perintah review nyata:

```bash
VACUUM_NO_UPDATE_CHECK=true go run github.com/daveshanley/vacuum@v0.30.1 lint -d -e api/openapi.yaml
go generate ./api
git diff --exit-code -- internal/api/openapi.gen.go
go test ./internal/api ./internal/handlers
go test ./...
go vet ./...
go build ./cmd/api
git diff --check
```

`make oapi-check` adalah drift gate; CI menjalankan dari checkout bersih setelah generated file committed. `make check` harus mencakup `oapi-check`. Kegagalan lint, parse, generation, compile, test, atau diff membuat gate gagal.

## Matriks validasi kontrak

- Semua path/operationId pada inventaris hadir tepat satu kali dan setiap path parameter required dideklarasikan.
- Semua operasi protected memiliki BearerAuth; public operation hanya yang terdaftar pada bagian auth/health/docs.
- Personal A tidak dapat memilih owner B; user A mendapat 403 pada route admin; admin C dapat menargetkan A/B; resource target lain tetap 404.
- `isDelete` required boolean pada response dan ditolak pada create/update; null/string/number tidak lolos schema.
- Semua field uang dihasilkan sebagai Go string atau named string type, tidak sebagai float32/float64.
- Missing/malformed If-Match → 400; stale version/lifecycle → 409; owner mismatch/missing → 404.
- Cursor rusak/lintas actor-mode-owner-filter → 400; list akhir memakai `next_cursor: null`.
- Date range invalid atau lebih dari 366 hari → 422; UUID/boolean/cursor syntax rusak → 400.
- Empty totals memakai string `"0.00"`; report, daily summary, dan export planned tidak menerima `isDelete=true`.
- Unknown request fields, owner override, role injection pada public/profile, dan lifecycle override ditolak.
- Planned category/transaction/report/admin/export route tetap tidak terdaftar pada runtime sampai issue implementasinya selesai.
- Generated code reproducible: generation kedua tidak menghasilkan diff.

## File yang boleh berubah setelah konfirmasi

| Path | Perubahan ISSUE-004 |
| --- | --- |
| `api/openapi.yaml` | Kontrak 3.0.3 lengkap, implemented/planned status, schemas, parameters, responses, security, examples |
| `api/generate.yaml` | Echo v5 config dan implemented operation allow-list |
| `internal/api/openapi.gen.go` | Hasil generate; tidak diedit manual |
| `internal/api/*_test.go` | Parse/validate dan invariant contract tests |
| `internal/handlers/*_test.go` | Bukti route current tetap bekerja dan planned tetap 404 |
| `Makefile` | `oapi-validate`, generation, lint, dan drift gate |
| `docs/api.md`, `docs/issues/ISSUE-004-api-contract.md`, `docs/issues.md` | Sinkronisasi keputusan dan bukti setelah implementasi benar-benar selesai |

Handler/service/repository/migration domain tidak termasuk scope. Bila standard error atau ETag memerlukan perubahan runtime, hentikan dan minta konfirmasi pengguna sebelum menyentuh code ISSUE-003.

## Definisi selesai ISSUE-004

- Kontrak OpenAPI 3.0.3 memuat semua operasi dan schema yang ditetapkan di sini dengan status implemented/planned yang benar.
- Hanya operationId implemented yang menghasilkan route Echo aktif.
- Owner/admin, money string, pagination/date range, `isDelete`, If-Match/version, error envelope, dan examples lulus contract tests.
- Lint, generation, drift check, unit test, vet, build, dan diff check lulus dengan output nyata yang dicatat.
- Tidak ada handler kategori/transaksi/admin/report/export, query, atau migrasi baru.
- Curl auth ISSUE-003 dan Swagger URL tidak diubah tanpa konfirmasi.
- Diff direview; commit/push hanya dilakukan bila pengguna meminta secara terpisah.
