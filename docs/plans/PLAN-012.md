# PLAN-012: Backup Drive pemilik/admin dan restore tervalidasi

Status: Draf — sesuaikan dengan repositori sebelum implementasi.
Diperbarui: 14 September 2026 (v0.3)
Issue: [ISSUE-012](../issues/ISSUE-012-drive-backups.md)
Repositori: monelog-api + monelog-app
Prasyarat: 011
Persyaratan: FR-01, FR-12, FR-15, FR-16, FR-17

## Sebelum implementasi

Baca requirements.md, access-control.md, architecture.md, database.md, api.md, dan issue terkait.
Periksa instruksi, kode, dependensi, serta perubahan pengguna; pastikan tugas prasyarat selesai. Gunakan nama file dan perintah nyata saat refinement.
Kebijakan terkonfirmasi: pengguna biasa hanya CRUD data sendiri; admin dapat mengelola target mana pun yang dipilih. isDelete=false berarti aktif; true berarti soft delete.

## Urutan implementasi

1. Rancang format berversi, jadwal/zona waktu, retensi, dan kebijakan konflik.
2. Periksa kebutuhan OAuth provider dan ikat state actor/target/mode.
3. Buat koneksi terenkripsi, schedule/job dengan recheck, retry, cancel/pause/resume, audit.
4. Sertakan data aktif/terhapus, validasi checksum/skema, dan map hanya ke target sah.
5. Lakukan drill restore di DB terisolasi dan bandingkan flag, jumlah, serta total aktif.
6. Buat UI progress/settings dan dokumentasikan recovery operasional.

## Area terdampak

Tentukan file nyata selama inspeksi repository dan jangan mengedit modul yang tidak terkait. Area utama disesuaikan dengan tujuan issue: handler, service, repository/query, kontrak API, UI, worker, migrasi, dan pengujian terkait.

## Validasi

Operasi positif/penolakan; OAuth; demotion; owner terhapus; duplicate run; tamper; backup rusak; field role/owner.
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
