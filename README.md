# Paket perencanaan Monelog

Draf v0.3 • 14 September 2026 • Spesifikasi dan rencana tugas; implementasi aplikasi belum dimulai.

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

Mulai [ISSUE-001](docs/issues/ISSUE-001-project-setup.md) dengan [PLAN-001](docs/plans/PLAN-001.md).
Selesaikan backend 001–007 dan [ISSUE-013](docs/issues/ISSUE-013-admin-viewing.md) sebelum frontend Issue 008.
Issue 013 kini mencakup pengelolaan admin penuh; nama file lama dipertahankan agar tautan tetap berfungsi.
Selanjutnya kerjakan ekspor, Android/iOS, templat, dan pencadangan Drive sesuai milestone.
File Markdown tugas adalah sumber kebenaran yang portabel dan belum memiliki GitHub Issue tertaut. Pertahankan dokumentasi kanonis di monelog-api; monelog-app merujuk pada commit dokumentasi/API yang dipatok.
