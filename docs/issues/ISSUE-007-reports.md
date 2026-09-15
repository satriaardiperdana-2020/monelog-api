# ISSUE-007: Laporan transaksi aktif

Status: Backlog
Diperbarui: 14 September 2026 (v0.3)
Repositori: monelog-api
Dependensi: 006
Remote issue: Belum dibuat
Persyaratan: FR-01, FR-06, FR-07, FR-15, FR-17
Plan: [PLAN-007](../plans/PLAN-007.md)
Kebijakan: [Kontrol akses dan penghapusan lunak](../access-control.md)

## Tujuan

Menghasilkan laporan berscope owner yang konsisten mengecualikan transaksi terhapus.

## Kriteria penerimaan

- [ ] Kriteria fungsional issue tercapai dengan bukti pengujian nyata.
- [ ] Pengguna biasa hanya dapat memakai scope sendiri; admin dapat memakai target yang dipilih dan tetap mempertahankan owner/actor.
- [ ] isDelete, version, Trash/restore, dan validasi scope mengikuti kontrak bersama.
- [ ] Kegagalan otorisasi, versi lama, dan resource lintas owner menghasilkan status yang tepat.
- [ ] Audit, kode hasil generate, dokumentasi, dan implementasi tetap konsisten.

## Ruang lingkup dan area terdampak

Handler/service laporan; SQL agregasi; helper tanggal/desimal; fixture integrasi.
Area di atas adalah rencana; persempit menjadi file nyata saat inspeksi repositori. Jangan mengedit modul yang tidak terkait.

## Verifikasi

Rentang 7/30/kustom dan bucket parsial; income/expense/difference; delete/restore; histori kategori; total A/B/admin; zona waktu target; laporan kosong.
Jalankan perintah yang dikonfigurasi untuk proyek (misalnya go test ./..., go vet ./..., suite PostgreSQL/HTTP, atau perintah frontend yang relevan). Catat perintah dan hasil sebenarnya; dokumentasi saja bukan bukti perilaku runtime.

## Batasan dan pemulihan

Tidak ada UI grafik, kolom saldo tersimpan, atau total yang memasukkan Trash.
Pertahankan perubahan pengguna. Uji migrasi pada fixture yang boleh dibuang dan gunakan recovery maju yang telah direview; jangan menghapus baris bersama.

## Definisi selesai

Semua kriteria penerimaan memiliki bukti nyata; kontrak/skema/dokumentasi dan kode hasil generate konsisten; diff telah direview; regresi relevan lulus. Infrastruktur yang tidak tersedia harus dicatat secara eksplisit. Pembaruan plan atau dokumentasi saja tidak menyelesaikan issue.
