# PLAN-002: Migrasi basis data dan fondasi sqlc

Status: Diimplementasikan — siap direview.
Diperbarui: 16 September 2026 (v0.4)
Issue: [ISSUE-002](../issues/ISSUE-002-database-foundation.md)
Repositori: monelog-api
Prasyarat: 001
Persyaratan: FR-01, FR-15, FR-16, FR-17

## Keputusan implementasi

- Repositori belum memiliki skema deployed, sehingga ISSUE-002 memakai migrasi awal baru tanpa backfill legacy.
- Migrasi menggunakan pasangan up/down golang-migrate. Down hanya untuk fixture disposable.
- Query sqlc 1.31.1 menghasilkan package pgx/v5 di internal/repository/sqlc dan kode hasil generate disimpan di repository.
- Pemilik migrasi dan role runtime dipisahkan. Grant runtime berada di db/roles/runtime.sql; audit hanya dapat dibaca dan ditambahkan oleh runtime.
- Otorisasi role, respons HTTP, dan transaksi write-admin bersama audit tetap menjadi tanggung jawab ISSUE-003, ISSUE-004, dan ISSUE-013.

## Implementasi

1. db/migrations berisi satu pasangan migration up/down per tabel untuk users, categories, transactions, refresh_sessions, dan admin_access_events beserta constraint serta index yang diperlukan.
2. db/queries menyediakan query dasar berscope owner untuk akun, kategori, transaksi, sesi, audit, data aktif/Trash, dan lifecycle berversi.
3. sqlc.yaml menghasilkan internal/repository/sqlc secara deterministik dengan pgx/v5.
4. db/roles/runtime.sql memberi hak runtime minimum tanpa hak penghapusan fisik data bisnis atau perubahan audit.
5. integration_test.go menjalankan migrasi dalam schema acak pada PostgreSQL disposable dan memeriksa default, constraint, ownership, actor, idempotency key, lifecycle, audit, serta sesi.

## Area terdampak

db/migrations; db/queries; db/roles; sqlc.yaml; internal/repository/sqlc; Makefile; README; dokumentasi ISSUE-002 dan plan ini.

## Validasi

- Migrasi up, down satu versi, dan up ulang pada PostgreSQL 16 disposable: lulus.
- Suite integrasi PostgreSQL untuk skema dan query hasil generate: lulus.
- Role runtime: SELECT/INSERT audit diizinkan; UPDATE/DELETE audit dan DELETE transaksi ditolak.
- sqlc generate dan sqlc vet: lulus.
- Pemeriksaan Go, format, vet, build, race, dan drift hasil generate: lulus.

## Batasan dan recovery

Tidak ada migrasi legacy karena tidak ditemukan skema deployed. Jika skema lama ditemukan kemudian, buat migrasi maju terpisah yang mempertahankan owner dan state penghapusan.
Gunakan down hanya pada fixture yang boleh dibuang. Basis data bersama dipulihkan dengan migrasi maju yang direview.
