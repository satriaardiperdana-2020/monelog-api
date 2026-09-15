# Peta jalan

Draf v0.3 • 14 September 2026
Backend dikerjakan lebih dahulu di monelog-api, lalu frontend Vue JavaScript di monelog-app.
Estimasi adalah rentang draf untuk satu pengembang, bukan tenggat. Hak admin penuh dan penghapusan lunak boolean telah dikonfirmasi.

| Milestone | Issue | Estimasi | Gerbang selesai |
| --- | --- | --- | --- |
| Fondasi | 001–002 | 3–5 hari kerja | Layanan dapat direproduksi; skema peran, flag boolean, atribusi aktor, dan audit |
| Autentikasi dan kontrak | 003–004 | 4–7 hari | Otorisasi akun/peran aktif, aturan siklus hidup, dan API pemilik/admin tervalidasi |
| Backend keuangan | 005–007 | 6–9 hari | CRUD/Trash/Restore milik sendiri dan total harian/laporan aktif yang benar |
| Backend admin | 013 setelah 007 | 4–7 hari | CRUD pengguna terpilih, pengelolaan akun/peran, audit, dan matriks A/B/C |
| MVP browser | 008 setelah 013 | 6–9 hari | Form pengelolaan pemilik/admin, keamanan pemilihan, dan Trash/Restore |
| Ekspor | 009 | 3–5 hari | XLSX/PDF pemilik/admin, job berdasarkan target, dan otorisasi unduhan |
| Mobile | 010 | 3–6 hari ditambah waktu signing/review | CRUD dan siklus hidup penghapusan Android/iOS berbasis peran |
| Templat | 011 | 2–4 hari | CRUD templat pemilik/admin, penghapusan boolean, dan penerapan berdasarkan scope |
| Pencadangan | 012 | 4–7 hari | Jadwal/job pemilik/admin, otorisasi provider, dan uji pemulihan data berflag |

Urutan ID dan pelaksanaan: 001 → 002 → 003 → 004 → 005 → 006 → 007 → 013 → 008 → 009 → 010 → 011 → 012.
013 mempertahankan nama file historis admin-viewing, tetapi kini berarti pengelolaan admin penuh. Jangan mengubah nomor tugas yang sudah ada.

## Gerbang backend

Issue 001–007 dan 013 harus selesai sebelum frontend dimulai. Verifikasi isolasi pengguna biasa; CRUD admin positif; atribusi pemilik terpilih; pengelolaan peran; pemblokiran akun terhapus; default dan migrasi flag true/false; versi/race; agregat aktif; Trash dan riwayat kategori; serta write/audit admin atomik.
Kontrak hasil generate harus mendokumentasikan mutasi admin yang didukung dan ejaan JSON `isDelete` yang tepat. Fitur tidak dianggap selesai hanya karena rencananya sudah ada.

## Gerbang frontend dan tahap berikutnya

Periksa identitas pemilik pada form/konfirmasi, simpan/buang saat target berubah, pembersihan respons terlambat, aksi admin yang terlihat, dan total Trash yang benar.
Ekspor/templat/pencadangan harus memiliki jalur personal dan admin saat fiturnya dirilis. Validasi ulang requester/pemilik/peran untuk job dan unduhan yang mengantre.
Pencadangan mempertahankan flag penghapusan tetapi mengecualikan peran/kredensial; kebijakan konflik restore, OAuth, jadwal, dan retensi harus dikonfirmasi sebelum Issue 012.
Rilis juga membutuhkan verifikasi backup/restore operasional, review lingkungan, serta pengujian browser/perangkat nyata.
Sinkronisasi offline, dompet/transfer, grafik, kalkulator, anggaran, feed bank, dan integrasi Launlog tetap menjadi ruang lingkup lanjutan yang belum diestimasi. Tinjau kembali tombstone dan aturan konflik sebelum sinkronisasi offline.
