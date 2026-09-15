# ISSUE-003: Autentikasi, state akun, dan otorisasi peran

Status: Backlog
Diperbarui: 14 September 2026 (v0.3)
Repositori: monelog-api
Dependensi: 002
Remote issue: Belum dibuat
Persyaratan: FR-01, FR-15, FR-16, FR-17
Plan: [PLAN-003](../plans/PLAN-003.md)
Kebijakan: [Kontrol akses dan penghapusan lunak](../access-control.md)

## Tujuan

Mengautentikasi akun browser/native dan memeriksa aktivitas actor serta role saat ini sebelum operasi personal/admin.

## Kriteria penerimaan

- [ ] Kriteria fungsional issue tercapai dengan bukti pengujian nyata.
- [ ] Pengguna biasa hanya dapat memakai scope sendiri; admin dapat memakai target yang dipilih dan tetap mempertahankan owner/actor.
- [ ] isDelete, version, Trash/restore, dan validasi scope mengikuti kontrak bersama.
- [ ] Kegagalan otorisasi, versi lama, dan resource lintas owner menghasilkan status yang tepat.
- [ ] Audit, kode hasil generate, dokumentasi, dan implementasi tetap konsisten.

## Ruang lingkup dan area terdampak

Bootstrap cmd/admin; middleware; handler/service auth/profile; query user/session.
Area di atas adalah rencana; persempit menjadi file nyata saat inspeksi repositori. Jangan mengedit modul yang tidak terkait.

## Verifikasi

Kredensial/JWT tidak valid; issuer/audience/expiry; replay/race refresh; CSRF/origin; injeksi role saat registrasi; akun terhapus; konflik versi self-delete; bootstrap; demotion dengan JWT lama.
Jalankan perintah yang dikonfigurasi untuk proyek (misalnya go test ./..., go vet ./..., suite PostgreSQL/HTTP, atau perintah frontend yang relevan). Catat perintah dan hasil sebenarnya; dokumentasi saja bukan bukti perilaku runtime.

## Batasan dan pemulihan

Jangan mengekspos password/session/provider secret. Endpoint pengelolaan target akun/peran lengkap berada pada 013.
Pertahankan perubahan pengguna. Uji migrasi pada fixture yang boleh dibuang dan gunakan recovery maju yang telah direview; jangan menghapus baris bersama.

## Definisi selesai

Semua kriteria penerimaan memiliki bukti nyata; kontrak/skema/dokumentasi dan kode hasil generate konsisten; diff telah direview; regresi relevan lulus. Infrastruktur yang tidak tersedia harus dicatat secara eksplisit. Pembaruan plan atau dokumentasi saja tidak menyelesaikan issue.
