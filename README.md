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
curl --fail http://127.0.0.1:8080/health/live
curl --fail http://127.0.0.1:8080/health/ready
```

## OpenAPI dan Swagger

Kontrak sumber ada di [api/openapi.yaml](api/openapi.yaml). Saat API berjalan, Swagger UI tersedia di `http://127.0.0.1:8080/swagger/` dan dokumen OpenAPI JSON tersedia di `http://127.0.0.1:8080/api/openapi.json`. Gunakan `make oapi-generate` setelah mengubah kontrak; kode hasil generate di `internal/api` harus di-commit.

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
