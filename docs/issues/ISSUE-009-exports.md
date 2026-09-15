# ISSUE-009: Ekspor Excel dan PDF pemilik/admin

Status: Backlog
Diperbarui: 14 September 2026 (v0.3)
Repositori: monelog-api + monelog-app
Dependensi: 008
Remote issue: Belum dibuat
Persyaratan: FR-01, FR-08, FR-15, FR-16, FR-17
Plan: [PLAN-009](../plans/PLAN-009.md)
Kebijakan: [Kontrol akses dan penghapusan lunak](../access-control.md)

## Tujuan

Mengekspor seluruh transaksi aktif yang cocok untuk owner atau target admin dengan otorisasi job yang tepat.

## Kriteria penerimaan

- [ ] Kriteria fungsional issue tercapai dengan bukti pengujian nyata.
- [ ] Pengguna biasa hanya dapat memakai scope sendiri; admin dapat memakai target yang dipilih dan tetap mempertahankan owner/actor.
- [ ] isDelete, version, Trash/restore, dan validasi scope mengikuti kontrak bersama.
- [ ] Kegagalan otorisasi, versi lama, dan resource lintas owner menghasilkan status yang tepat.
- [ ] Audit, kode hasil generate, dokumentasi, dan implementasi tetap konsisten.

## Ruang lingkup dan area terdampak

Migrasi/worker export_jobs; service report; generator file; handler download; UI ekspor web.
Area di atas adalah rencana; persempit menjadi file nyata saat inspeksi repositori. Jangan mengedit modul yang tidak terkait.

## Verifikasi

Ekspor multi-halaman; parity report; baris true dikeluarkan; formula injection; judul PDF panjang; akses A/B/C; role lama; target berubah; retry/expiry job.
Jalankan perintah yang dikonfigurasi untuk proyek (misalnya go test ./..., go vet ./..., suite PostgreSQL/HTTP, atau perintah frontend yang relevan). Catat perintah dan hasil sebenarnya; dokumentasi saja bukan bukti perilaku runtime.

## Batasan dan pemulihan

Ekspor adalah artefak laporan, bukan format backup/import; Trash tidak pernah diekspor.
Pertahankan perubahan pengguna. Uji migrasi pada fixture yang boleh dibuang dan gunakan recovery maju yang telah direview; jangan menghapus baris bersama.

## Definisi selesai

Semua kriteria penerimaan memiliki bukti nyata; kontrak/skema/dokumentasi dan kode hasil generate konsisten; diff telah direview; regresi relevan lulus. Infrastruktur yang tidak tersedia harus dicatat secara eksplisit. Pembaruan plan atau dokumentasi saja tidak menyelesaikan issue.
