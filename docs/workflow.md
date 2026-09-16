# Alur kerja pengembangan langkah demi langkah

Persyaratan menjelaskan perilaku aplikasi. Issue mendefinisikan satu deliverable dan kriteria penerimaan. Plan menjelaskan langkah implementasi dan verifikasi. Prompt memilih tugas dan instruksinya.

1. Baca requirements.md dan access-control.md: CRUD milik sendiri, pengelolaan admin penuh, dan penghapusan `isDelete` telah dikonfirmasi.
2. Pertahankan dokumen kanonis di monelog-api. Tinjau instruksi repositori dan perubahan yang ada sebelum mengedit.
3. Gunakan docs/issues.md, satu file ISSUE-NNN, dan PLAN-NNN untuk setiap tugas. Pertahankan ID stabil di GitHub/Bitbucket.
4. Jika GitHub Issue dibuat, tautkan URL sebenarnya ke file tugas dan pertahankan ID portabelnya. ISSUE-002 dipetakan ke GitHub #6.
5. Kerjakan backend 001–007 → 013, lalu frontend 008 → 009–012 sesuai dependensi.
6. Inspeksi repositori dan sesuaikan plan terpilih dengan file, perintah, dan keputusan yang tersisa sebelum implementasi.
7. Implementasikan issue terpilih sesuai otorisasi pengguna saat ini. Batasi branch/commit/PR pada otorisasi tersebut.
8. Telusuri aktor, pemilik, dan operasi melalui handler, service, query, respons, cache, dan payload job.
9. Untuk setiap entitas, terapkan default `is_delete=false`, penghapusan lunak `true`, dan pemulihan `false` dengan pemeriksaan versi. Jangan menghapus baris bisnis secara fisik.
10. Jalankan pengujian issue yang bermakna, termasuk operasi admin positif dan penolakan lintas pemilik untuk pengguna biasa. Catat perintah/hasil nyata; infrastruktur yang tidak tersedia bukan tanda lulus.
11. Tinjau diff untuk predicate pemilik, pemeriksaan peran, filter penghapusan, riwayat kategori, target job, dan atomisitas audit. Validasi konsistensi kontrak/kode hasil generate.
12. Selesaikan review/merge sesuai otorisasi; perbarui issue dan indeks bersama-sama saat kriteria terpenuhi.
13. Lanjutkan ke tugas berikutnya yang tidak terblokir. Pembaruan dokumentasi tidak menandai implementasi sebagai Done.

## Prompt perencanaan

Baca docs/requirements.md, docs/access-control.md, docs/architecture.md, docs/database.md, docs/api.md, docs/issues/ISSUE-001-project-setup.md, dan docs/plans/PLAN-001.md. Inspeksi repositori tanpa membaca file rahasia. Sesuaikan plan issue ini dengan path nyata, langkah implementasi, perintah pengujian, dan keputusan yang belum selesai. Permintaan ini hanya untuk perencanaan.

## Prompt implementasi

Implementasikan issue terpilih menggunakan plan yang telah direview dan aturan akses saat ini. Pengguna biasa hanya melakukan CRUD atas datanya sendiri; admin dapat mengelola pemilik yang dipilih secara eksplisit. Gunakan API `isDelete` / Go `IsDelete` / SQL `is_delete`; penghapusan mempertahankan baris dan restore mengaktifkannya kembali. Pertahankan perubahan pengguna yang tidak terkait, jangan membaca file rahasia, jalankan pengujian relevan, dan laporkan hasil nyata. Ikuti instruksi terbaru untuk commit, push, dan PR.

## Prompt review

Review diff terhadap issue dan plan terpilih. Periksa isolasi pemilik dan CRUD admin yang berhasil, perubahan akun/peran, total uang yang tepat, race penghapusan lunak/restore, job pengguna terpilih, atomisitas audit, dan pengujian yang hilang. Laporkan temuan dengan referensi file dan tingkat keparahan. Permintaan ini hanya untuk review.

Ganti ISSUE-001 dan PLAN-001 dengan tugas yang sedang dikerjakan. Pekerjaan frontend menggunakan monelog-app dan spesifikasi API/izin backend yang dipatok.

## Daftar periksa otorisasi implementasi

Pemilik personal = aktor. Pemilik admin = target terverifikasi, dengan admin aktif dan operasi yang didukung. Write admin diizinkan; jangan mempertahankan penolakan lintas pengguna yang sudah tidak berlaku.
Setiap query dan mutasi mempertahankan satu pemilik keuangan. Data baru yang dibuat admin menjadi milik target, dengan atribusi aktor.
Siklus hidup adalah boolean, bukan penghapusan fisik. Total aktif menggunakan `transactions.is_delete=false` dan tidak hilang saat kategori historis dihapus.
Form, write tertunda, ekspor, templat, jadwal, dan job restore mengikat konteks target yang tidak berubah. Jangan mengalihkan operasi berjalan ke pemilik baru.
Gunakan akun biasa A/B dan admin C untuk jalur yang diizinkan maupun ditolak; sertakan Trash, restore, versi lama, otorisasi ulang job, akun terhapus, dan kegagalan audit.
