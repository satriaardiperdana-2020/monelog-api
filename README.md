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
File Markdown tugas adalah sumber kebenaran yang portabel dan belum memiliki GitHub Issue tertaut. Pertahankan dokumentasi kanonis di monelog-api; monelog-app merujuk pada commit dokumentasi/API yang dipatok.

## Menjalankan API

Prasyarat:

- Go 1.27.1
- PostgreSQL yang dapat diakses melalui URL koneksi lokal

Salin nilai dari `.env.example` ke environment shell Anda dan ganti placeholder dengan kredensial PostgreSQL lokal. Aplikasi membaca environment variable secara langsung dan tidak memuat file `.env` secara otomatis.

```bash
export DATABASE_URL='postgres://<user>:<password>@127.0.0.1:5432/monelog?sslmode=disable'
make run
```

Konfigurasi opsional dan nilai default tersedia di [.env.example](.env.example). Jangan commit `.env`, password, token, atau key.

Health endpoint:

- `GET /health/live` mengembalikan `200` ketika proses API hidup.
- `GET /health/ready` mengembalikan `200` ketika PostgreSQL tersedia dan `503` ketika belum siap.

```bash
curl --fail http://127.0.0.1:8080/health/live
curl --fail http://127.0.0.1:8080/health/ready
```

## Pengembangan

```bash
make fmt
make test
make vet
make build
make check
```
