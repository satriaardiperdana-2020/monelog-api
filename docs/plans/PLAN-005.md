# PLAN-005: CRUD kategori, Trash, dan restore

Status: Draf — sesuaikan dengan repositori sebelum implementasi.
Diperbarui: 14 September 2026 (v0.3)
Issue: [ISSUE-005](../issues/ISSUE-005-categories.md)
Repositori: monelog-api
Prasyarat: 004
Persyaratan: FR-01, FR-05, FR-15, FR-17

## Sebelum implementasi

Baca requirements.md, access-control.md, architecture.md, database.md, api.md, dan issue terkait.
Periksa instruksi, kode, dependensi, serta perubahan pengguna; pastikan tugas prasyarat selesai. Gunakan nama file dan perintah nyata saat refinement.
Kebijakan terkonfirmasi: pengguna biasa hanya CRUD data sendiri; admin dapat mengelola target mana pun yang dipilih. isDelete=false berarti aktif; true berarti soft delete.

## Urutan implementasi

1. Implementasikan list/detail/create/update/delete/restore dan validasi service.
2. Gunakan default aktif false, filter Trash true, dan pemeriksaan versi.
3. Pertahankan keunikan nama pada baris aktif/terhapus; konflik lebih baik daripada identitas ganda.
4. Pertahankan join historis owner yang sama meski kategori dihapus.
5. Expose handler inti agar dapat dipakai adapter admin 013.

## Area terdampak

Tentukan file nyata selama inspeksi repository dan jangan mengedit modul yang tidak terkait. Area utama disesuaikan dengan tujuan issue: handler, service, repository/query, kontrak API, UI, worker, migrasi, dan pengujian terkait.

## Validasi

Ownership, tipe, nama duplikat/case, baris Trash, restore, versi lama, dan race kategori/transaksi.
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
