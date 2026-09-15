# ISSUE-012: Backup Drive dan restore tervalidasi pemilik/admin

Status: Backlog
Diperbarui: 14 September 2026 (v0.3)
Repositori: monelog-api + monelog-app
Dependensi: 011
Remote issue: Belum dibuat
Persyaratan: FR-01, FR-12, FR-15, FR-16, FR-17
Plan: [PLAN-012](../plans/PLAN-012.md)
Kebijakan: [Kontrol akses dan penghapusan lunak](../access-control.md)

## Tujuan

Mengelola backup Drive per pengguna yang telah diotorisasi serta memulihkan data dengan batas owner dan delete boolean tetap terjaga.

## Kriteria penerimaan

- [ ] Kriteria fungsional issue tercapai dengan bukti pengujian nyata.
- [ ] Pengguna biasa hanya dapat memakai scope sendiri; admin dapat memakai target yang dipilih dan tetap mempertahankan owner/actor.
- [ ] isDelete, version, Trash/restore, dan validasi scope mengikuti kontrak bersama.
- [ ] Kegagalan otorisasi, versi lama, dan resource lintas owner menghasilkan status yang tepat.
- [ ] Audit, kode hasil generate, dokumentasi, dan implementasi tetap konsisten.

## Ruang lingkup dan area terdampak

OAuth/service/worker Drive; migrasi connection/schedule/job; skema backup/restore; API berscope; UI settings; pengujian restore terisolasi.
Area di atas adalah rencana; persempit menjadi file nyata saat inspeksi repositori. Jangan mengedit modul yang tidak terkait.

## Verifikasi

Operasi positif dan penolakan A/B; OAuth kedaluwarsa/dicabut; demotion requester; owner terhapus; run duplikat; tampering target; backup rusak; field role/owner berbahaya; parity flag/count/total.
Jalankan perintah yang dikonfigurasi untuk proyek (misalnya go test ./..., go vet ./..., suite PostgreSQL/HTTP, atau perintah frontend yang relevan). Catat perintah dan hasil sebenarnya; dokumentasi saja bukan bukti perilaku runtime.

## Batasan dan pemulihan

Tidak ada overwrite produksi tanpa pengawasan, kredensial plaintext, atau asumsi role admin aplikasi memberi consent Google.
Pertahankan perubahan pengguna. Uji migrasi pada fixture yang boleh dibuang dan gunakan recovery maju yang telah direview; jangan menghapus baris bersama.

## Definisi selesai

Semua kriteria penerimaan memiliki bukti nyata; kontrak/skema/dokumentasi dan kode hasil generate konsisten; diff telah direview; regresi relevan lulus. Infrastruktur yang tidak tersedia harus dicatat secara eksplisit. Pembaruan plan atau dokumentasi saja tidak menyelesaikan issue.
