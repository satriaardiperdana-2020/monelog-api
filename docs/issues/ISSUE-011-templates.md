# ISSUE-011: Templat pemilik/admin dengan penghapusan lunak

Status: Backlog
Diperbarui: 14 September 2026 (v0.3)
Repositori: monelog-api + monelog-app
Dependensi: 010
Remote issue: Belum dibuat
Persyaratan: FR-01, FR-10, FR-15, FR-17
Plan: [PLAN-011](../plans/PLAN-011.md)
Kebijakan: [Kontrol akses dan penghapusan lunak](../access-control.md)

## Tujuan

Menyediakan templat yang dapat diedit owner atau admin untuk owner terpilih, dengan delete dan restore boolean.

## Kriteria penerimaan

- [ ] Kriteria fungsional issue tercapai dengan bukti pengujian nyata.
- [ ] Pengguna biasa hanya dapat memakai scope sendiri; admin dapat memakai target yang dipilih dan tetap mempertahankan owner/actor.
- [ ] isDelete, version, Trash/restore, dan validasi scope mengikuti kontrak bersama.
- [ ] Kegagalan otorisasi, versi lama, dan resource lintas owner menghasilkan status yang tepat.
- [ ] Audit, kode hasil generate, dokumentasi, dan implementasi tetap konsisten.

## Ruang lingkup dan area terdampak

Migrasi/query/handler/service templat; kontrak API; integrasi template dan form transaksi Vue.
Area di atas adalah rencana; persempit menjadi file nyata saat inspeksi repositori. Jangan mengedit modul yang tidak terkait.

## Verifikasi

CRUD owner dan admin C atas B; A ditolak; isDelete false/true/restore; version race; apply dua kali tidak membuat record; Save eksplisit; lifecycle kategori; ganti owner.
Jalankan perintah yang dikonfigurasi untuk proyek (misalnya go test ./..., go vet ./..., suite PostgreSQL/HTTP, atau perintah frontend yang relevan). Catat perintah dan hasil sebenarnya; dokumentasi saja bukan bukti perilaku runtime.

## Batasan dan pemulihan

Tidak ada pembuatan otomatis berulang atau bundle seharian kecuali persyaratan berubah.
Pertahankan perubahan pengguna. Uji migrasi pada fixture yang boleh dibuang dan gunakan recovery maju yang telah direview; jangan menghapus baris bersama.

## Definisi selesai

Semua kriteria penerimaan memiliki bukti nyata; kontrak/skema/dokumentasi dan kode hasil generate konsisten; diff telah direview; regresi relevan lulus. Infrastruktur yang tidak tersedia harus dicatat secara eksplisit. Pembaruan plan atau dokumentasi saja tidak menyelesaikan issue.
