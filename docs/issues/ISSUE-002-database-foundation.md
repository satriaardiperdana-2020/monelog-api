# ISSUE-002: Migrasi basis data dan fondasi sqlc

Status: Backlog
Diperbarui: 16 September 2026 (v0.4)
Repositori: monelog-api
Dependensi: 001
Remote issue: Belum dibuat
Persyaratan: FR-01, FR-15, FR-16, FR-17
Plan: [PLAN-002](../plans/PLAN-002.md)
Kebijakan: [Kontrol akses dan penghapusan lunak](../access-control.md)

## Tujuan

Membangun fondasi PostgreSQL dan sqlc untuk akun, kategori, transaksi, sesi, audit, lifecycle penghapusan lunak, versi, serta atribusi aktor. Issue ini menyediakan batas data yang diperlukan issue berikutnya; autentikasi, kontrak HTTP, dan alur pengelolaan admin lengkap dikerjakan pada issue masing-masing.

## Kriteria penerimaan

- [ ] Migrasi baru menghasilkan skema users, categories, transactions, refresh_sessions, dan admin_access_events yang konsisten dengan desain basis data.
- [ ] Batas data untuk ownership, tipe kategori/transaksi, nominal, is_delete, dan version menolak state yang tidak valid; semua data yang dapat dihapus lunak mulai aktif.
- [ ] Query sqlc dasar untuk scope owner, data aktif/Trash, dan lifecycle berversi dapat dihasilkan ulang secara deterministik.
- [ ] Fondasi actor dan audit mempertahankan perbedaan antara owner terpilih dan actor yang melakukan perubahan, tanpa memberi role aplikasi melalui migrasi atau query registrasi biasa.
- [ ] Migrasi dan query dibuktikan pada PostgreSQL disposable. Backfill legacy hanya diwajibkan bila inspeksi menemukan skema lama yang benar-benar digunakan.

## Ruang lingkup dan area terdampak

db/migrations; db/queries; konfigurasi sqlc; internal/repository/sqlc; fixture integrasi PostgreSQL; dokumentasi basis data yang terdampak.
Area di atas adalah rencana; persempit menjadi file nyata saat inspeksi repositori. Jangan mengedit modul yang tidak terkait.

## Verifikasi

Migrasi baru; boolean/default dan version; owner/type/amount salah; duplicate request key; query aktif versus Trash; atribusi owner/actor; akses audit runtime; generate deterministik. Uji backfill hanya bila ada skema legacy yang nyata.
Jalankan suite PostgreSQL yang dikonfigurasi serta pemeriksaan Go yang relevan. Catat perintah dan hasil sebenarnya; dokumentasi saja bukan bukti perilaku runtime.

## Batasan dan pemulihan

Jangan melakukan migrasi destruktif pada data bersama atau promosi akun otomatis.
Pertahankan perubahan pengguna. Uji migrasi pada fixture yang boleh dibuang dan gunakan recovery maju yang telah direview; jangan menghapus baris bersama.

## Definisi selesai

Semua kriteria penerimaan memiliki bukti nyata; skema, dokumentasi, dan kode sqlc hasil generate konsisten; diff telah direview; regresi relevan lulus. Infrastruktur yang tidak tersedia harus dicatat secara eksplisit. Pembaruan plan atau dokumentasi saja tidak menyelesaikan issue.

## Batas tanggung jawab

Otorisasi role, perilaku status HTTP, dan alur write-admin atomik diverifikasi pada ISSUE-003, ISSUE-004, dan ISSUE-013 dengan memakai fondasi ini.
