# ISSUE-013: Pengelolaan penuh user dan data oleh admin

Status: Backlog
Diperbarui: 14 September 2026 (v0.3)
Repositori: monelog-api
Dependensi: 007
Remote issue: Belum dibuat
Persyaratan: FR-01, FR-15, FR-16, FR-17
Plan: [PLAN-013](../plans/PLAN-013.md)
Kebijakan: [Kontrol akses dan penghapusan lunak](../access-control.md)

## Tujuan

Mengizinkan admin aktif menjalankan seluruh operasi aplikasi yang didukung pada data pengguna terpilih, termasuk CRUD dan pemulihan soft delete.

## Kriteria penerimaan

- [ ] Kriteria fungsional issue tercapai dengan bukti pengujian nyata.
- [ ] Pengguna biasa hanya dapat memakai scope sendiri; admin dapat memakai target yang dipilih dan tetap mempertahankan owner/actor.
- [ ] isDelete, version, Trash/restore, dan validasi scope mengikuti kontrak bersama.
- [ ] Kegagalan otorisasi, versi lama, dan resource lintas owner menghasilkan status yang tepat.
- [ ] Audit, kode hasil generate, dokumentasi, dan implementasi tetap konsisten.

## Ruang lingkup dan area terdampak

Handler admin; middleware/peran; service otorisasi; query user/category/transaction/report/audit; OpenAPI; suite PostgreSQL/HTTP.
Area di atas adalah rencana; persempit menjadi file nyata saat inspeksi repositori. Jangan mengedit modul yang tidak terkait.

## Verifikasi

A/B ditolak dan C berhasil; record C menjadi milik B; promosi/demotion; delete/restore akun; active/Trash; version race; mismatch kategori; rollback audit; total target; actor kedaluwarsa/terhapus.
Jalankan perintah yang dikonfigurasi untuk proyek (misalnya go test ./..., go vet ./..., suite PostgreSQL/HTTP, atau perintah frontend yang relevan). Catat perintah dan hasil sebenarnya; dokumentasi saja bukan bukti perilaku runtime.

## Batasan dan pemulihan

Tidak ada perubahan role produksi otomatis, penghapusan baris bisnis fisik, atau respons kredensial mentah. Nama file historis dipertahankan agar tautan tetap berfungsi.
Pertahankan perubahan pengguna. Uji migrasi pada fixture yang boleh dibuang dan gunakan recovery maju yang telah direview; jangan menghapus baris bersama.

## Definisi selesai

Semua kriteria penerimaan memiliki bukti nyata; kontrak/skema/dokumentasi dan kode hasil generate konsisten; diff telah direview; regresi relevan lulus. Infrastruktur yang tidak tersedia harus dicatat secara eksplisit. Pembaruan plan atau dokumentasi saja tidak menyelesaikan issue.
