# PLAN-002: Migrasi basis data dan fondasi sqlc

Status: Draf — sesuaikan dengan repositori sebelum implementasi.
Diperbarui: 14 September 2026 (v0.3)
Issue: [ISSUE-002](../issues/ISSUE-002-database-foundation.md)
Repositori: monelog-api
Prasyarat: 001
Persyaratan: FR-01, FR-15, FR-16, FR-17

## Sebelum implementasi

Baca requirements.md, access-control.md, architecture.md, database.md, api.md, dan issue terkait.
Periksa instruksi, kode, dependensi, serta perubahan pengguna; pastikan tugas prasyarat selesai. Gunakan nama file dan perintah nyata saat refinement.
Kebijakan terkonfirmasi: pengguna biasa hanya CRUD data sendiri; admin dapat mengelola target mana pun yang dipilih. isDelete=false berarti aktif; true berarti soft delete.

## Urutan implementasi

1. Inspeksi skema; gunakan desain baru atau migrasi maju yang mempertahankan status lama.
2. Buat users, categories, transactions, sessions, dan admin_access_events sesuai urutan FK.
3. Tambahkan default false, constraint versi, atribusi actor, kepemilikan kategori, dan index aktif/Trash.
4. Generate query create/read/update/delete/restore berscope dengan lock akun dan insert audit.
5. Pisahkan privilege runtime dari kepemilikan migrasi dan uji fixture A/B/C di PostgreSQL.

## Area terdampak

Tentukan file nyata selama inspeksi repository dan jangan mengedit modul yang tidak terkait. Area utama disesuaikan dengan tujuan issue: handler, service, repository/query, kontrak API, UI, worker, migrasi, dan pengujian terkait.

## Validasi

Migrasi baru/legacy; default/constraint; owner/type/amount salah; idempotensi; query aktif/Trash; grant audit; drift generate.
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
