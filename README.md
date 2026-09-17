# Monelog API

Fondasi backend Go untuk Monelog. ISSUE-001 menyediakan konfigurasi tervalidasi, koneksi PostgreSQL, health check, logging HTTP aman, dan graceful shutdown.

## Bahasa / Language

- [Bahasa Indonesia](README.md)
- [English](README.en.md)

## Perilaku yang telah dikonfirmasi

- Pengguna biasa hanya dapat melakukan CRUD atas datanya sendiri.
- Admin dapat menjalankan seluruh operasi aplikasi pada data pengguna yang dipilih, termasuk CRUD, peran, laporan, ekspor, templat, dan pencadangan.
- Penghapusan lunak menggunakan `isDelete=false` untuk data aktif dan `true` untuk data terhapus. Penghapusan mempertahankan baris; pemulihan mengubah flag kembali menjadi `false`.

## Dokumen

1. [Persyaratan](docs/requirements.md)
2. [Kontrol akses dan penghapusan lunak](docs/access-control.md)
3. [Arsitektur](docs/architecture.md)
4. [Basis data](docs/database.md)
5. [API](docs/api.md)
6. [Peta jalan](docs/roadmap.md)
7. [Indeks issue](docs/issues.md)
8. [Alur kerja](docs/workflow.md)

Lihat [struktur proyek](docs/architecture.md#struktur-proyek) untuk susunan direktori backend yang direncanakan.

Mulai [ISSUE-001](docs/issues/ISSUE-001-project-setup.md) dengan [PLAN-001](docs/plans/PLAN-001.md).
Selesaikan backend 001–007 dan [ISSUE-013](docs/issues/ISSUE-013-admin-viewing.md) sebelum frontend Issue 008.
Issue 013 kini mencakup pengelolaan admin penuh; nama file lama dipertahankan agar tautan tetap berfungsi.
Selanjutnya kerjakan ekspor, Android/iOS, templat, dan pencadangan Drive sesuai milestone.
File Markdown tugas adalah sumber kebenaran yang portabel dan dapat ditautkan ke GitHub Issue tanpa mengubah ID. Pertahankan dokumentasi kanonis di monelog-api; monelog-app merujuk pada commit dokumentasi/API yang dipatok.

## Menjalankan API

Prasyarat:

- Go 1.27.1
- PostgreSQL yang dapat diakses melalui URL koneksi lokal
- sqlc 1.31.1 untuk regenerasi query
- golang-migrate 4.18 atau kompatibel untuk migrasi

Salin nilai dari `.env.example` ke `.env.development` dan ganti placeholder dengan kredensial PostgreSQL lokal. Saat `APP_ENV` tidak ditetapkan, aplikasi memuat `.env.development`; saat ditetapkan, aplikasi memuat `.env.<APP_ENV>`. Environment variable dari shell atau deployment selalu mengesampingkan nilai dari file.

```bash
go run ./cmd/api
```

Konfigurasi opsional dan nilai default tersedia di [.env.example](.env.example). Jangan commit `.env`, password, token, atau key.

Health endpoint:

- `GET /health/live` mengembalikan `200` ketika proses API hidup.
- `GET /health/ready` mengembalikan `200` ketika PostgreSQL tersedia dan `503` ketika belum siap.

```bash
# Memastikan proses HTTP API hidup.
curl --fail http://127.0.0.1:8080/health/live
# Memastikan PostgreSQL dan dependensi runtime siap menerima request.
curl --fail http://127.0.0.1:8080/health/ready
```

## OpenAPI dan Swagger

Kontrak sumber ada di [api/openapi.yaml](api/openapi.yaml). Saat API berjalan, buka [Swagger UI](http://127.0.0.1:8080/swagger/) atau [OpenAPI JSON](http://127.0.0.1:8080/api/openapi.json). Pilih `Authorize` di Swagger UI untuk mengisi access token setelah login.

Contoh registrasi untuk membuat akun pengguna baru:

```bash
curl --location 'http://127.0.0.1:8080/api/v1/auth/register' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "email": "you@example.com",
    "password": "correct-horse-battery-staple",
    "timezone": "Asia/Jakarta"
  }'
```

Contoh login untuk Android/iOS. API mengembalikan refresh token native di JSON:

```bash
curl --location 'http://127.0.0.1:8080/api/v1/auth/login' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "email": "you@example.com",
    "password": "correct-horse-battery-staple",
    "client_type": "native"
  }'
```

Contoh login browser. API memerlukan `Origin` terdaftar dan mengirim refresh token melalui cookie HttpOnly:

```bash
curl --location 'http://127.0.0.1:8080/api/v1/auth/login' \
  --header 'Origin: https://app.example.com' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "email": "you@example.com",
    "password": "correct-horse-battery-staple",
    "client_type": "web"
  }'
```

`client_type` wajib ditulis dengan underscore dan hanya menerima `native` atau `web`. Gunakan `make oapi-generate` setelah mengubah kontrak; kode hasil generate di `internal/api` harus di-commit.

Contoh transaksi setelah mendapatkan `access_token` dari login. Header `Authorization` dipakai untuk semua endpoint aplikasi berikut:

```bash
ACCESS_TOKEN='<access-token-dari-login>'

# Membuat kategori income yang dapat dipilih saat mencatat pemasukan.
curl --location 'http://127.0.0.1:8080/api/v1/categories' \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data '{"name":"Gaji","type":"income"}'

# Mencatat satu transaksi. amount adalah string uang eksak dan client_request_id membuat retry aman.
curl --location 'http://127.0.0.1:8080/api/v1/transactions' \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data '{"transaction_date":"2026-09-17","type":"income","category_id":"<category-uuid>","amount":"15000000.00","title":"Gaji September","client_request_id":"<request-uuid>"}'

# Menampilkan history transaksi aktif dalam rentang tanggal tertentu; ulangi dengan page.next_cursor untuk halaman berikutnya.
curl --location 'http://127.0.0.1:8080/api/v1/transactions?start_date=2026-09-01&end_date=2026-09-30&limit=30' \
  --header "Authorization: Bearer $ACCESS_TOKEN"

# Menampilkan total income, expense, dan difference per hari untuk kartu ringkasan main page atau navigasi tanggal.
curl --location 'http://127.0.0.1:8080/api/v1/daily-summaries?start_date=2026-09-01&end_date=2026-09-30' \
  --header "Authorization: Bearer $ACCESS_TOKEN"

# Membaca transaksi yang sudah dihapus lunak dari Trash.
curl --location 'http://127.0.0.1:8080/api/v1/transactions?start_date=2026-09-01&end_date=2026-09-30&isDelete=true' \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

Swagger UI menjelaskan kegunaan setiap endpoint, parameter, scope akses, format amount, cursor pagination, optimistic locking, dan aturan idempotensi. Endpoint admin memakai pola `/api/v1/admin/users/{user_id}/...`; `user_id` target harus dipilih eksplisit pada URL.

## Pengembangan

```bash
make fmt
make test
make vet
make build
make check
```

## Basis data

Jalankan migrasi dengan akun pemilik skema. Gunakan `migrate-down` hanya pada basis data disposable setelah dampak datanya direview.

```bash
DATABASE_URL="$DATABASE_URL" make migrate-up
make sqlc-generate
make sqlc-vet
TEST_DATABASE_URL="$TEST_DATABASE_URL" make test-integration
```

Runtime aplikasi harus memakai role yang berbeda dari pemilik migrasi. Setelah tabel tersedia, berikan hak runtime minimum dengan:

```bash
DATABASE_URL="$DATABASE_URL" RUNTIME_DB_ROLE=monelog_runtime make runtime-grants
```

Role runtime dapat membaca dan menulis data aplikasi yang dibutuhkan, tetapi tidak dapat menghapus fisik baris bisnis atau mengubah/menghapus audit. Otorisasi pengguna dan admin tetap ditegakkan oleh service pada issue berikutnya.
