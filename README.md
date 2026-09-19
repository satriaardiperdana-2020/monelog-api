# Monelog API

Backend REST API Monelog untuk autentikasi, profil pengguna, kategori, transaksi, Trash/restore, ringkasan harian, laporan, dan akses admin. Aplikasi dibangun dengan Go, Echo, PostgreSQL, sqlc, dan kontrak OpenAPI.

Dokumentasi ini khusus backend. Versi English tersedia di [README.en.md](README.en.md).

## Fitur yang tersedia

- Registrasi, login, refresh token rotation, logout, dan JWT access token.
- Profil pengguna dengan optimistic locking melalui `version`.
- Kategori `income` dan `expense`, termasuk soft delete dan restore.
- Transaksi dengan nilai uang eksak, idempotensi, pagination berbasis cursor, soft delete, dan restore.
- Ringkasan harian serta laporan berdasarkan minggu, bulan, atau kategori.
- Endpoint admin untuk membaca dan mengubah kategori, transaksi, dan laporan user tertentu.
- PostgreSQL migration, generated query sqlc, OpenAPI, Swagger UI, health check, dan graceful shutdown.

Semua primary key entitas memakai `BIGSERIAL`. Identifier dan foreign key memakai `BIGINT`. Nilai uang disimpan sebagai `NUMERIC(14,2)` dan dikirim melalui API sebagai string dua angka desimal.

## 1. Prasyarat

Siapkan:

- Go `1.27.1`, sesuai [go.mod](go.mod).
- PostgreSQL.
- GNU Make.
- `golang-migrate` dengan driver PostgreSQL.
- sqlc `1.31.1` bila ingin meregenerasi query.
- `curl`; `jq` bersifat opsional untuk mengambil nilai dari response.

Periksa instalasi:

```bash
go version
psql --version
migrate -version
sqlc version
make --version
```

## 2. Ambil source code dan dependency Go

```bash
git clone <repository-url> monelog-api
cd monelog-api
go mod download
go mod verify
```

Repository sudah menyimpan hasil generate sqlc dan OpenAPI. sqlc hanya dibutuhkan ketika file di `db/queries`, migration, atau konfigurasi sqlc berubah.

## 3. Buat database dan role PostgreSQL

Contoh berikut memisahkan role pemilik schema untuk migration dan role runtime dengan hak minimum. Jalankan sebagai administrator PostgreSQL:

```bash
sudo -u postgres psql
```

Di dalam `psql`:

```sql
CREATE ROLE monelog_owner LOGIN PASSWORD 'change-owner-password';
CREATE ROLE monelog_runtime LOGIN PASSWORD 'change-runtime-password';
CREATE DATABASE monelog OWNER monelog_owner;
CREATE DATABASE monelog_test OWNER monelog_owner;
GRANT CONNECT ON DATABASE monelog TO monelog_runtime;
\q
```

Database `monelog_test` digunakan oleh integration test pada langkah 16. Role runtime tidak memerlukan akses ke database tersebut.

Gunakan password yang berbeda dan aman. Jangan menyimpan password nyata di repository.

URL koneksi pemilik schema:

```text
postgres://monelog_owner:change-owner-password@127.0.0.1:5432/monelog?sslmode=disable
```

URL koneksi runtime:

```text
postgres://monelog_runtime:change-runtime-password@127.0.0.1:5432/monelog?sslmode=disable
```

`sslmode=disable` hanya sesuai untuk PostgreSQL lokal. Gunakan TLS sesuai penyedia database pada staging dan production.

## 4. Terapkan DDL dan migration

File DDL kanonis berada di [db/migrations](db/migrations). Jangan membuat tabel satu per satu secara manual; jalankan seluruh migration berurutan agar constraint, index, sequence, dan `schema_migrations` konsisten.

```bash
export MIGRATION_DATABASE_URL='postgres://monelog_owner:change-owner-password@127.0.0.1:5432/monelog?sslmode=disable'

DATABASE_URL="$MIGRATION_DATABASE_URL" make migrate-up
```

Migration membuat tabel berikut:

| Tabel | Kegunaan |
|---|---|
| `users` | Akun, role, timezone, currency, soft delete, dan version. |
| `categories` | Kategori income/expense milik user. |
| `transactions` | Transaksi, idempotency key, actor, soft delete, dan version. |
| `refresh_sessions` | Refresh token rotation dan revocation family. |
| `admin_access_events` | Audit akses dan mutasi oleh admin. |
| `schema_migrations` | Versi migration yang dikelola `golang-migrate`. |

Verifikasi DDL:

```bash
psql "$MIGRATION_DATABASE_URL" -c 'SELECT version, dirty FROM schema_migrations;'
psql "$MIGRATION_DATABASE_URL" -c '\dt public.*'
psql "$MIGRATION_DATABASE_URL" -c '\d+ public.users'
psql "$MIGRATION_DATABASE_URL" -c '\d+ public.transactions'
```

Versi terbaru harus berstatus `dirty = false`. Migration `000007` mengonversi instalasi lama dari UUID ke BIGINT dan tidak dapat mengembalikan UUID asli, sehingga migration tersebut sengaja tidak mendukung rollback.

Berikan hak minimum kepada role runtime setelah DDL tersedia:

```bash
DATABASE_URL="$MIGRATION_DATABASE_URL" \
RUNTIME_DB_ROLE=monelog_runtime \
make runtime-grants
```

Perintah ini memberikan hak tabel dan sequence yang dibutuhkan aplikasi tanpa memberi izin menghapus fisik baris bisnis atau mengubah/menghapus audit.

## 5. Konfigurasi environment

Salin template:

```bash
cp .env.example .env.development
```

Buat key JWT acak minimal 32 byte dalam format Base64:

```bash
openssl rand -base64 32
```

Isi `.env.development`:

```dotenv
APP_ENV=development
HTTP_ADDR=127.0.0.1:8080
HTTP_READ_HEADER_TIMEOUT=5s
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=15s
HTTP_IDLE_TIMEOUT=60s
SHUTDOWN_TIMEOUT=10s
HEALTH_READY_TIMEOUT=2s
DATABASE_URL=postgres://monelog_runtime:change-runtime-password@127.0.0.1:5432/monelog?sslmode=disable
AUTH_JWT_HMAC_KEY=<hasil-openssl-rand-base64-32>
AUTH_JWT_ISSUER=monelog-api
AUTH_JWT_AUDIENCE=monelog-app
AUTH_ACCESS_TOKEN_TTL=24h
AUTH_ALLOWED_ORIGINS=http://localhost:5173
```

Aturan konfigurasi:

- Tanpa `APP_ENV`, aplikasi membaca `.env.development`.
- `APP_ENV=staging` membuat aplikasi membaca `.env.staging`.
- Environment variable dari shell selalu mengalahkan nilai file.
- `AUTH_JWT_HMAC_KEY` wajib Base64 dan hasil decode minimal 32 byte.
- `AUTH_ACCESS_TOKEN_TTL` harus positif dan maksimal 24 jam.
- `AUTH_ALLOWED_ORIGINS` menerima satu atau beberapa origin yang dipisahkan koma.
- Jangan commit `.env.*`, password database, access token, refresh token, atau signing key.

## 6. Buat admin pertama

Registrasi HTTP selalu membuat role `user`. Admin pertama dibuat melalui CLI interaktif dan hanya dapat dibuat jika belum ada admin aktif:

```bash
go run ./cmd/admin -email admin@example.com -timezone Asia/Jakarta
```

CLI akan meminta password tanpa menampilkannya. Password harus memiliki panjang 12–128 karakter dan memenuhi validasi service.

## 7. Jalankan backend

Mode langsung:

```bash
go run ./cmd/api
```

Atau build binary:

```bash
make build
./bin/monelog-api
```

Server default berjalan di `http://127.0.0.1:8080`.

Periksa health endpoint:

```bash
curl -i http://127.0.0.1:8080/health/live
curl -i http://127.0.0.1:8080/health/ready
```

Response berhasil:

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ok"}
```

`/health/live` memeriksa proses HTTP. `/health/ready` juga memeriksa koneksi PostgreSQL dan mengembalikan `503` dengan `{"status":"unavailable"}` ketika database belum siap.

## 8. OpenAPI dan Swagger

- Kontrak sumber: [api/openapi.yaml](api/openapi.yaml)
- OpenAPI JSON saat server berjalan: <http://127.0.0.1:8080/api/openapi.json>
- Swagger UI: <http://127.0.0.1:8080/swagger/>

Di Swagger UI, jalankan login, salin `access_token`, pilih **Authorize**, lalu tempel token tanpa prefix `Bearer`.

## 9. Konvensi request API

Contoh berikut menggunakan:

```bash
export BASE_URL='http://127.0.0.1:8080'
export ACCESS_TOKEN='<access-token>'
export REFRESH_TOKEN='<refresh-token>'
export CATEGORY_ID=1
export TRANSACTION_ID=1
export USER_ID=1
```

Ketentuan umum:

- Endpoint terproteksi memakai `Authorization: Bearer <access_token>`.
- Semua ID adalah integer positif 64-bit.
- `amount` adalah string, misalnya `"43500.00"`.
- Tanggal memakai `YYYY-MM-DD`; timestamp response memakai RFC3339 UTC.
- `type` hanya menerima `income` atau `expense`.
- Update/delete/restore memakai `version` untuk optimistic locking.
- DELETE memakai header `If-Match`, misalnya `If-Match: "2"`.
- `client_request_id` adalah integer positif yang unik per user. Retry dengan ID dan payload sama mengembalikan transaksi lama dengan status `200`; payload berbeda menghasilkan `409`.
- JSON menolak field yang tidak dikenal dan lebih dari satu JSON value.

Format error konsisten:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request."
  }
}
```

Kode umum: `400` syntax/cursor salah, `401` autentikasi gagal, `403` tidak diizinkan, `404` resource tidak ditemukan dalam scope, `409` konflik state/version/idempotensi, `422` validasi gagal, `429` login rate limited, dan `500`/`503` untuk kegagalan server/dependency.

## 10. Autentikasi dan profil

### Registrasi user

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@example.com",
    "password": "correct-horse-battery-staple",
    "timezone": "Asia/Jakarta"
  }'
```

Response `201 Created`:

```json
{
  "data": {
    "id": 1,
    "email": "user@example.com",
    "role": "user",
    "timezone": "Asia/Jakarta",
    "currency": "IDR",
    "isDelete": false,
    "version": 1
  }
}
```

Registrasi juga membuat kategori default user.

### Login native

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@example.com",
    "password": "correct-horse-battery-staple",
    "client_type": "native"
  }'
```

Response `200 OK`:

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<opaque-refresh-token>",
  "token_type": "Bearer",
  "expires_in": 86399
}
```

Simpan token dari response, atau gunakan `jq`:

```bash
LOGIN_RESPONSE=$(curl -sS -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"correct-horse-battery-staple","client_type":"native"}')

export ACCESS_TOKEN=$(printf '%s' "$LOGIN_RESPONSE" | jq -r '.access_token')
export REFRESH_TOKEN=$(printf '%s' "$LOGIN_RESPONSE" | jq -r '.refresh_token')
```

### Refresh native

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/refresh" \
  -H 'Content-Type: application/json' \
  -d "{\"client_type\":\"native\",\"refresh_token\":\"$REFRESH_TOKEN\"}"
```

Response sama seperti login native dan berisi access token serta refresh token baru. Setelah rotasi, gunakan refresh token yang baru.

### Logout native

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/logout" \
  -H 'Content-Type: application/json' \
  -d "{\"client_type\":\"native\",\"refresh_token\":\"$REFRESH_TOKEN\"}"
```

Response `204 No Content` tidak memiliki body.

### Login dan refresh web

Client web harus memakai HTTPS di luar localhost, origin yang terdaftar, cookie jar, dan CSRF token. Login tidak mengembalikan refresh token di JSON:

```bash
curl -i -c cookies.txt -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Origin: http://localhost:5173' \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"correct-horse-battery-staple","client_type":"web"}'
```

Response body:

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_in": 86399
}
```

Server menyimpan refresh token dalam cookie `__Host-monelog-refresh` yang HttpOnly dan CSRF token dalam cookie `__Host-monelog-csrf`. Kirim nilai cookie CSRF melalui `X-CSRF-Token` saat refresh/logout:

```bash
export CSRF_TOKEN='<nilai-cookie-__Host-monelog-csrf>'

curl -i -b cookies.txt -c cookies.txt \
  -X POST "$BASE_URL/api/v1/auth/refresh" \
  -H 'Origin: http://localhost:5173' \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"client_type":"web"}'
```

### Baca profil

```bash
curl -i "$BASE_URL/api/v1/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Response `200` memakai bentuk user yang sama dengan response registrasi.

### Update timezone profil

```bash
curl -i -X PATCH "$BASE_URL/api/v1/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"timezone":"UTC","version":1}'
```

Response `200` mengembalikan profil dengan `timezone: "UTC"` dan `version: 2`.

### Soft delete akun

```bash
curl -i -X DELETE "$BASE_URL/api/v1/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'If-Match: "2"'
```

Response `204 No Content`. Seluruh refresh session user ikut dicabut.

## 11. Kategori

### List kategori aktif

```bash
curl -i "$BASE_URL/api/v1/categories?type=expense&isDelete=false" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Response `200`:

```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "type": "expense",
      "name": "Makanan",
      "isDelete": false,
      "version": 1
    }
  ]
}
```

### Buat kategori

```bash
curl -i -X POST "$BASE_URL/api/v1/categories" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Langganan","type":"expense"}'
```

Response `201` berisi satu object kategori dalam `data`. Nama kategori bersifat unik tanpa membedakan huruf besar/kecil untuk kombinasi user dan type.

### Baca dan update kategori

```bash
curl -i "$BASE_URL/api/v1/categories/$CATEGORY_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i -X PATCH "$BASE_URL/api/v1/categories/$CATEGORY_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Langganan Digital","version":1}'
```

Response update `200` menaikkan `version` menjadi `2`.

### Archive dan restore kategori

```bash
curl -i -X DELETE "$BASE_URL/api/v1/categories/$CATEGORY_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'If-Match: "2"'

curl -i "$BASE_URL/api/v1/categories/$CATEGORY_ID?isDelete=true" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i -X POST "$BASE_URL/api/v1/categories/$CATEGORY_ID/restore" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"version":3}'
```

DELETE mengembalikan `204`. Response restore `200` mengembalikan kategori aktif dan menaikkan version kembali.

## 12. Transaksi

### Buat transaksi

Gunakan kategori aktif yang dimiliki user dan memiliki type yang sama:

```bash
curl -i -X POST "$BASE_URL/api/v1/transactions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{
    \"transaction_date\": \"2026-09-17\",
    \"type\": \"expense\",
    \"category_id\": $CATEGORY_ID,
    \"amount\": \"43500.00\",
    \"title\": \"Makan siang\",
    \"client_request_id\": 1001
  }"
```

Response `201 Created`:

```json
{
  "data": {
    "id": 1,
    "user_id": 1,
    "category_id": 1,
    "category_name": "Makanan",
    "transaction_date": "2026-09-17",
    "type": "expense",
    "amount": "43500.00",
    "title": "Makan siang",
    "client_request_id": 1001,
    "created_by": 1,
    "updated_by": 1,
    "isDelete": false,
    "version": 1,
    "created_at": "2026-09-17T05:00:00Z",
    "updated_at": "2026-09-17T05:00:00Z"
  }
}
```

### List dan pagination transaksi

```bash
curl -i "$BASE_URL/api/v1/transactions?start_date=2026-09-01&end_date=2026-09-30&type=expense&category_id=$CATEGORY_ID&limit=30" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Response `200`:

```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "category_id": 1,
      "category_name": "Makanan",
      "transaction_date": "2026-09-17",
      "type": "expense",
      "amount": "43500.00",
      "title": "Makan siang",
      "client_request_id": 1001,
      "created_by": 1,
      "updated_by": 1,
      "isDelete": false,
      "version": 1,
      "created_at": "2026-09-17T05:00:00Z",
      "updated_at": "2026-09-17T05:00:00Z"
    }
  ],
  "page": {"next_cursor": null}
}
```

Jika `next_cursor` tidak null, kirim nilainya tanpa dimodifikasi:

```bash
curl -i --get "$BASE_URL/api/v1/transactions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  --data-urlencode 'start_date=2026-09-01' \
  --data-urlencode 'end_date=2026-09-30' \
  --data-urlencode 'limit=30' \
  --data-urlencode 'cursor=<next_cursor>'
```

### Baca dan update transaksi

```bash
curl -i "$BASE_URL/api/v1/transactions/$TRANSACTION_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i -X PATCH "$BASE_URL/api/v1/transactions/$TRANSACTION_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{
    \"transaction_date\": \"2026-09-17\",
    \"type\": \"expense\",
    \"category_id\": $CATEGORY_ID,
    \"amount\": \"50000.00\",
    \"title\": \"Makan siang dan kopi\",
    \"version\": 1
  }"
```

Response update `200` mengembalikan transaksi dengan `version: 2` dan `updated_by` sesuai actor.

### Soft delete, Trash, dan restore transaksi

```bash
curl -i -X DELETE "$BASE_URL/api/v1/transactions/$TRANSACTION_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'If-Match: "2"'

curl -i "$BASE_URL/api/v1/transactions?start_date=2026-09-01&end_date=2026-09-30&isDelete=true" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i "$BASE_URL/api/v1/transactions/$TRANSACTION_ID?isDelete=true" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i -X POST "$BASE_URL/api/v1/transactions/$TRANSACTION_ID/restore" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"version":3}'
```

DELETE menghasilkan `204`. Restore menghasilkan `200` dan transaksi kembali masuk history serta laporan aktif.

## 13. Ringkasan dan laporan

### Ringkasan harian

```bash
curl -i "$BASE_URL/api/v1/daily-summaries?start_date=2026-09-01&end_date=2026-09-30&limit=30" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Response `200`:

```json
{
  "data": [
    {
      "date": "2026-09-17",
      "income": "0.00",
      "expense": "50000.00",
      "difference": "-50000.00"
    }
  ],
  "page": {"next_cursor": null}
}
```

### Report summary

Preset yang tersedia adalah `last_7_days` dan `last_30_days`:

```bash
curl -i "$BASE_URL/api/v1/reports/summary?range=last_30_days" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Rentang custom wajib memiliki kedua tanggal:

```bash
curl -i "$BASE_URL/api/v1/reports/summary?range=custom&start_date=2026-09-01&end_date=2026-09-30" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Contoh response:

```json
{
  "data": {
    "period": {"start_date":"2026-09-01","end_date":"2026-09-30"},
    "income": "15000000.00",
    "expense": "2500000.00",
    "difference": "12500000.00",
    "top_income_categories": [
      {"category_id":2,"name":"Gaji","type":"income","amount":"15000000.00"}
    ],
    "top_expense_categories": [
      {"category_id":1,"name":"Makanan","type":"expense","amount":"2500000.00"}
    ]
  }
}
```

### Report breakdown

`group_by` menerima `week`, `month`, atau `category`:

```bash
curl -i "$BASE_URL/api/v1/reports/breakdown?range=custom&start_date=2026-09-01&end_date=2026-09-30&group_by=week" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Response memiliki `period`, `group_by`, array `periods`, dan array `categories`. Untuk grouping waktu, nilai berada di `periods`; untuk grouping kategori, nilai berada di `categories`.

## 14. Endpoint admin

Login menggunakan akun yang dibuat oleh CLI admin, lalu isi `ADMIN_ACCESS_TOKEN` dan target `USER_ID`:

```bash
export ADMIN_ACCESS_TOKEN='<admin-access-token>'
export USER_ID=1
```

Endpoint admin memakai request body dan response yang sama dengan endpoint personal, tetapi owner ditentukan oleh `{user_id}` pada URL. Semua aksi admin dicatat pada `admin_access_events`.

Contoh list kategori dan transaksi target:

```bash
curl -i "$BASE_URL/api/v1/admin/users/$USER_ID/categories?isDelete=false" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"

curl -i "$BASE_URL/api/v1/admin/users/$USER_ID/transactions?start_date=2026-09-01&end_date=2026-09-30" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

Contoh membuat kategori dan transaksi untuk target:

```bash
curl -i -X POST "$BASE_URL/api/v1/admin/users/$USER_ID/categories" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Bonus Tahunan","type":"income"}'

curl -i -X POST "$BASE_URL/api/v1/admin/users/$USER_ID/transactions" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"transaction_date\":\"2026-09-17\",\"type\":\"income\",\"category_id\":$CATEGORY_ID,\"amount\":\"1000000.00\",\"title\":\"Koreksi admin\",\"client_request_id\":2001}"
```

Contoh laporan target:

```bash
curl -i "$BASE_URL/api/v1/admin/users/$USER_ID/reports/summary?range=last_30_days" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"

curl -i "$BASE_URL/api/v1/admin/users/$USER_ID/reports/breakdown?range=last_30_days&group_by=category" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

Pemetaan endpoint admin lengkap:

| Operasi | Endpoint |
|---|---|
| List/create kategori | `GET/POST /api/v1/admin/users/{user_id}/categories` |
| Get/update/delete kategori | `GET/PATCH/DELETE /api/v1/admin/users/{user_id}/categories/{id}` |
| Restore kategori | `POST /api/v1/admin/users/{user_id}/categories/{id}/restore` |
| List/create transaksi | `GET/POST /api/v1/admin/users/{user_id}/transactions` |
| Get/update/delete transaksi | `GET/PATCH/DELETE /api/v1/admin/users/{user_id}/transactions/{id}` |
| Restore transaksi | `POST /api/v1/admin/users/{user_id}/transactions/{id}/restore` |
| Ringkasan harian | `GET /api/v1/admin/users/{user_id}/daily-summaries` |
| Report summary | `GET /api/v1/admin/users/{user_id}/reports/summary` |
| Report breakdown | `GET /api/v1/admin/users/{user_id}/reports/breakdown` |

Token user biasa pada endpoint tersebut menghasilkan `403 Forbidden`:

```json
{"error":{"code":"FORBIDDEN","message":"Forbidden."}}
```

## 15. Daftar endpoint

| Method | Path | Auth |
|---|---|---|
| `GET` | `/health/live` | Tidak |
| `GET` | `/health/ready` | Tidak |
| `POST` | `/api/v1/auth/register` | Tidak |
| `POST` | `/api/v1/auth/login` | Tidak |
| `POST` | `/api/v1/auth/refresh` | Refresh token/cookie |
| `POST` | `/api/v1/auth/logout` | Refresh token/cookie |
| `GET/PATCH/DELETE` | `/api/v1/me` | Bearer |
| `GET/POST` | `/api/v1/categories` | Bearer |
| `GET/PATCH/DELETE` | `/api/v1/categories/{id}` | Bearer |
| `POST` | `/api/v1/categories/{id}/restore` | Bearer |
| `GET/POST` | `/api/v1/transactions` | Bearer |
| `GET/PATCH/DELETE` | `/api/v1/transactions/{id}` | Bearer |
| `POST` | `/api/v1/transactions/{id}/restore` | Bearer |
| `GET` | `/api/v1/daily-summaries` | Bearer |
| `GET` | `/api/v1/reports/summary` | Bearer |
| `GET` | `/api/v1/reports/breakdown` | Bearer |
| Beragam | `/api/v1/admin/users/{user_id}/...` | Bearer admin |

## 16. Regenerasi kode dan quality checks

Setelah mengubah migration atau query SQL:

```bash
make sqlc-generate
make sqlc-vet
```

Setelah mengubah [api/openapi.yaml](api/openapi.yaml):

```bash
make oapi-generate
```

Jalankan pemeriksaan backend:

```bash
make fmt
make test
make test-race
make vet
make build
make check
```

Integration test membuat schema sementara pada database test:

```bash
export TEST_DATABASE_URL='postgres://monelog_owner:change-owner-password@127.0.0.1:5432/monelog_test?sslmode=disable'
make test-integration
```

Jangan arahkan `TEST_DATABASE_URL` ke database production.

## 17. Struktur penting backend

```text
api/                    kontrak OpenAPI dan konfigurasi generator
cmd/api/                entrypoint HTTP API
cmd/admin/              CLI bootstrap admin pertama
db/migrations/          DDL migration up/down
db/queries/             sumber query sqlc
db/roles/               grant role runtime
internal/api/           kode OpenAPI hasil generate
internal/auth/          JWT, password hashing, refresh secret
internal/handlers/      transport HTTP
internal/middleware/    auth, CORS, logging, rate limiting
internal/repository/    akses PostgreSQL dan generated sqlc
internal/service/       aturan bisnis
```

Dokumentasi tambahan tersedia di [docs](docs), terutama [database](docs/database.md), [API](docs/api.md), [arsitektur](docs/architecture.md), dan [kontrol akses](docs/access-control.md).
