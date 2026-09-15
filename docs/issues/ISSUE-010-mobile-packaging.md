# ISSUE-010: Pengelolaan pemilik/admin Android/iOS

Status: Backlog
Diperbarui: 14 September 2026 (v0.3)
Repositori: monelog-app
Dependensi: 009
Remote issue: Belum dibuat
Persyaratan: FR-01, FR-09, FR-15, FR-17
Plan: [PLAN-010](../plans/PLAN-010.md)
Kebijakan: [Kontrol akses dan penghapusan lunak](../access-control.md)

## Tujuan

Mengemas aplikasi Vue yang telah diuji dengan Capacitor dan mempertahankan CRUD serta lifecycle boolean di perangkat.

## Kriteria penerimaan

- [ ] Kriteria fungsional issue tercapai dengan bukti pengujian nyata.
- [ ] Pengguna biasa hanya dapat memakai scope sendiri; admin dapat memakai target yang dipilih dan tetap mempertahankan owner/actor.
- [ ] isDelete, version, Trash/restore, dan validasi scope mengikuti kontrak bersama.
- [ ] Kegagalan otorisasi, versi lama, dan resource lintas owner menghasilkan status yang tepat.
- [ ] Audit, kode hasil generate, dokumentasi, dan implementasi tetap konsisten.

## Ruang lingkup dan area terdampak

Proyek Capacitor/native; adapter auth/API/download; view Vue bersama; smoke/E2E perangkat.
Area di atas adalah rencana; persempit menjadi file nyata saat inspeksi repositori. Jangan mengedit modul yang tidak terkait.

## Verifikasi

Login A/B/C; CRUD admin; delete/restore perangkat; form target lama; respons terlambat; resume/restart; hapus akun/demotion; berbagi file scoped; aksesibilitas.
Jalankan perintah yang dikonfigurasi untuk proyek (misalnya go test ./..., go vet ./..., suite PostgreSQL/HTTP, atau perintah frontend yang relevan). Catat perintah dan hasil sebenarnya; dokumentasi saja bukan bukti perilaku runtime.

## Batasan dan pemulihan

Persetujuan app store atau sinkronisasi offline tidak dijamin; infrastruktur platform/signing yang hilang harus dicatat sebagai blocker.
Pertahankan perubahan pengguna. Uji migrasi pada fixture yang boleh dibuang dan gunakan recovery maju yang telah direview; jangan menghapus baris bersama.

## Definisi selesai

Semua kriteria penerimaan memiliki bukti nyata; kontrak/skema/dokumentasi dan kode hasil generate konsisten; diff telah direview; regresi relevan lulus. Infrastruktur yang tidak tersedia harus dicatat secara eksplisit. Pembaruan plan atau dokumentasi saja tidak menyelesaikan issue.
