# PLAN-006: CRUD transaksi, penghapusan lunak, dan ringkasan harian

Status: Draf — sesuaikan dengan repositori sebelum implementasi.
Diperbarui: 14 September 2026 (v0.3)
Issue: [ISSUE-006](../issues/ISSUE-006-transactions.md)
Repositori: monelog-api
Prasyarat: 005
Persyaratan: FR-01, FR-02, FR-03, FR-04, FR-15, FR-17

## Sebelum implementasi

Baca requirements.md, access-control.md, architecture.md, database.md, api.md, dan issue terkait.
Periksa instruksi, kode, dependensi, serta perubahan pengguna; pastikan tugas prasyarat selesai. Gunakan nama file dan perintah nyata saat refinement.
Kebijakan terkonfirmasi: pengguna biasa hanya CRUD data sendiri; admin dapat mengelola target mana pun yang dipilih. isDelete=false berarti aktif; true berarti soft delete.

## Urutan implementasi

1. Validasi amount/title/date/category dan lock akun/kategori.
2. Terapkan idempotensi per owner dan request hash serta catat actor.
3. Buat edit, SoftDeleteTransaction, dan RestoreTransaction berbasis expected version/state.
4. Buat query list/detail aktif dan Trash dengan cursor berscope.
5. Buat daily summary aktif dan siapkan handler untuk adapter admin.

## Area terdampak

Tentukan file nyata selama inspeksi repository dan jangan mengedit modul yang tidak terkait. Area utama disesuaikan dengan tujuan issue: handler, service, repository/query, kontrak API, UI, worker, migrasi, dan pengujian terkait.

## Validasi

Replay/conflict create; admin berscope; actor/owner; flag; versi; race; kategori; hari kosong; batas uang/tanggal/cursor.
Jalankan go test ./..., go vet ./..., suite PostgreSQL/HTTP, atau perintah unit/component/E2E frontend yang benar-benar dikonfigurasi sesuai area. Periksa drift kode SQL/OpenAPI hasil generate dan catat perintah/hasil nyata. Jangan menyatakan otorisasi atau lifecycle runtime benar hanya dari review dokumentasi.

## Review otorisasi dan siklus hidup

Telusuri actor, owner terpilih, aksi, version, dan isDelete di setiap batas yang terdampak.
Operasi personal memakai actor sebagai owner; operasi admin memakai target terotorisasi. Pertahankan scope kategori/resource/cursor/job.
Untuk write admin, validasi dan audit harus sukses bersama transaksi data. Untuk job eksternal, simpan otorisasi/job/audit sebelum provider dan validasi ulang saat eksekusi.
Total finansial aktif mengecualikan baris true. Trash/restore mempertahankan batas owner. Jangan menyimpulkan penghapusan fisik atau perubahan role dari data impor.

## Batasan dan recovery

Tidak ada wallet, transfer, transaksi otomatis berulang, atau sinkronisasi offline.
Sesudah implementasi, bandingkan setiap kriteria penerimaan dengan bukti dan perbarui status issue/index sesuai workflow pengguna.

## Titik review

Sajikan ruang lingkup file, langkah, keputusan yang belum selesai, dan pengujian yang telah diperjelas sebelum coding, kecuali implementasi telah diotorisasi pengguna.
