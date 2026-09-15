# ISSUE-002: Migrasi basis data dan fondasi sqlc

Status: Backlog
Diperbarui: 14 September 2026 (v0.3)
Repositori: monelog-api
Dependensi: 001
Remote issue: Belum dibuat
Persyaratan: FR-01, FR-15, FR-16, FR-17
Plan: [PLAN-002](../plans/PLAN-002.md)
Kebijakan: [Kontrol akses dan penghapusan lunak](../access-control.md)

## Tujuan

Membuat skema PostgreSQL dan query sqlc untuk akun, kategori, transaksi, sesi, audit, penghapusan lunak, versi, dan atribusi aktor.

## Kriteria penerimaan

- [ ] Kriteria fungsional issue tercapai dengan bukti pengujian nyata.
- [ ] Pengguna biasa hanya dapat memakai scope sendiri; admin dapat memakai target yang dipilih dan tetap mempertahankan owner/actor.
- [ ] isDelete, version, Trash/restore, dan validasi scope mengikuti kontrak bersama.
- [ ] Kegagalan otorisasi, versi lama, dan resource lintas owner menghasilkan status yang tepat.
- [ ] Audit, kode hasil generate, dokumentasi, dan implementasi tetap konsisten.

## Ruang lingkup dan area terdampak

db/migrations; db/queries; konfigurasi sqlc; internal/repository/sqlc; fixture integrasi PostgreSQL.
Area di atas adalah rencana; persempit menjadi file nyata saat inspeksi repositori. Jangan mengedit modul yang tidak terkait.

## Verifikasi

Migrasi baru/legacy; boolean NOT NULL/default; owner/type/amount salah; duplicate request key; query aktif versus Trash; grant audit; generate deterministik.
Jalankan perintah yang dikonfigurasi untuk proyek (misalnya go test ./..., go vet ./..., suite PostgreSQL/HTTP, atau perintah frontend yang relevan). Catat perintah dan hasil sebenarnya; dokumentasi saja bukan bukti perilaku runtime.

## Batasan dan pemulihan

Jangan melakukan migrasi destruktif pada data bersama atau promosi akun otomatis.
Pertahankan perubahan pengguna. Uji migrasi pada fixture yang boleh dibuang dan gunakan recovery maju yang telah direview; jangan menghapus baris bersama.

## Definisi selesai

Semua kriteria penerimaan memiliki bukti nyata; kontrak/skema/dokumentasi dan kode hasil generate konsisten; diff telah direview; regresi relevan lulus. Infrastruktur yang tidak tersedia harus dicatat secara eksplisit. Pembaruan plan atau dokumentasi saja tidak menyelesaikan issue.
