# PLAN-001: Setup proyek backend

Status: Draf — sesuaikan dengan repositori sebelum implementasi.
Diperbarui: 14 September 2026 (v0.3)
Issue: [ISSUE-001](../issues/ISSUE-001-project-setup.md)
Repositori: monelog-api
Prasyarat: Tidak ada
Persyaratan: FR-01, FR-15, FR-16, FR-17

## Sebelum implementasi

Baca requirements.md, access-control.md, architecture.md, database.md, api.md, dan issue terkait.
Periksa instruksi, kode, dependensi, serta perubahan pengguna; pastikan tugas prasyarat selesai. Gunakan nama file dan perintah nyata saat refinement.
Kebijakan terkonfirmasi: pengguna biasa hanya CRUD data sendiri; admin dapat mengelola target mana pun yang dipilih. isDelete=false berarti aktif; true berarti soft delete.

## Urutan implementasi

1. Inspeksi instruksi/dependensi dan patok versi kompatibel.
2. Buat batas cmd/api dan internal config/handler/middleware/service/repository.
3. Tambahkan shutdown graceful, recovery, logging aman, timeout, dan health endpoint.
4. Dokumentasikan konfigurasi placeholder dan PostgreSQL lokal.
5. Tambahkan pemeriksaan config/health serta CI formatting, vet/lint, dan test.

## Area terdampak

Tentukan file nyata selama inspeksi repository dan jangan mengedit modul yang tidak terkait. Area utama disesuaikan dengan tujuan issue: handler, service, repository/query, kontrak API, UI, worker, migrasi, dan pengujian terkait.

## Validasi

Konfigurasi salah; liveness; readiness saat DB tidak tersedia; shutdown.
Jalankan go test ./..., go vet ./..., suite PostgreSQL/HTTP, atau perintah unit/component/E2E frontend yang benar-benar dikonfigurasi sesuai area. Periksa drift kode SQL/OpenAPI hasil generate dan catat perintah/hasil nyata. Jangan menyatakan otorisasi atau lifecycle runtime benar hanya dari review dokumentasi.

## Review otorisasi dan siklus hidup

Telusuri actor, owner terpilih, aksi, version, dan isDelete di setiap batas yang terdampak.
Operasi personal memakai actor sebagai owner; operasi admin memakai target terotorisasi. Pertahankan scope kategori/resource/cursor/job.
Untuk write admin, validasi dan audit harus sukses bersama transaksi data. Untuk job eksternal, simpan otorisasi/job/audit sebelum provider dan validasi ulang saat eksekusi.
Total finansial aktif mengecualikan baris true. Trash/restore mempertahankan batas owner. Jangan menyimpulkan penghapusan fisik atau perubahan role dari data impor.

## Batasan dan recovery

Pertahankan perubahan pengguna. Uji perubahan skema pada fixture yang boleh dibuang dan gunakan recovery maju yang direview; jangan menghapus baris bersama.
Sesudah implementasi, bandingkan setiap kriteria penerimaan dengan bukti dan perbarui status issue/index sesuai workflow pengguna.

## Titik review

Sajikan ruang lingkup file, langkah, keputusan yang belum selesai, dan pengujian yang telah diperjelas sebelum coding, kecuali implementasi telah diotorisasi pengguna.
