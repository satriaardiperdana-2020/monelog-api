# Desain basis data

Draf v0.3 • 14 September 2026
Desain logis; migrasi/query yang dapat dijalankan dibuat pada Issue 002. Pengelolaan admin penuh adalah Issue 013.

## Konvensi umum

Identifier UUID. created_at/updated_at bertipe TIMESTAMPTZ; timestamp diperbarui pada mutasi.
Entitas yang dapat dihapus lunak memiliki is_delete BOOLEAN NOT NULL DEFAULT FALSE dan version INTEGER NOT NULL DEFAULT 1 CHECK(version>0).
JSON mengekspos isDelete (ejaan tepat) dan Go memakai IsDelete. Field JSON lain tetap snake_case.
NULL bukan state penghapusan. Baris aktif hanya jika is_delete=false.

## Tabel

| Tabel | Field dan constraint utama |
| --- | --- |
| users | id PK; email dinormalisasi lowercase UNIQUE NOT NULL; password_hash NOT NULL; role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('user','admin')); timezone default Asia/Jakarta; currency default IDR CHECK(currency='IDR'); is_delete; version; created_at; updated_at |
| categories | id PK; user_id FK users NOT NULL; type CHECK IN ('income','expense'); name VARCHAR(80) NOT NULL; is_delete; version; created_at; updated_at; UNIQUE(id,user_id,type) |
| transactions | id PK; user_id FK users NOT NULL; category_id NOT NULL; type CHECK IN ('income','expense'); amount NUMERIC(14,2) CHECK(amount>0); transaction_date DATE; title VARCHAR(200); client_request_id UUID; request_hash TEXT; created_by/updated_by UUID FK users; is_delete; version; created_at; updated_at; UNIQUE(user_id,client_request_id) |
| refresh_sessions | id PK; user_id FK users; family_id UUID; token_hash UNIQUE; expires_at; revoked_at nullable; replaced_by self FK nullable; created_at |
| admin_access_events | id PK; actor_user_id FK users; target_user_id FK users nullable; resource_type; resource_id nullable; action; outcome; request_id; safe_metadata JSONB default '{}'; created_at UTC |
| transaction_templates (nanti) | id PK; user_id; category_id; type; name; amount; title; is_delete; version; created_at; updated_at |
| export_jobs (nanti, 009) | id PK; owner_user_id; requested_by; request_mode personal/admin; filter/format immutable; status; artifact locator; expires_at; created_at/updated_at |
| drive_connections (nanti, 012) | id PK; user_id UNIQUE; encrypted_refresh_token; provider_account_label; revoked_at nullable; created_at/updated_at |
| backup_schedules (nanti, 012) | id PK; owner_user_id; authorized_by; request_mode; schedule/timezone; state enabled/paused; version; created_at/updated_at |
| backup_jobs (nanti, 012) | id PK; owner_user_id; requested_by; request_mode; schedule_id nullable; scheduled_for; status; attempt; provider_file_id/checksum/error_code nullable; started_at/completed_at; UNIQUE(schedule_id,scheduled_for) untuk run terjadwal |

Tabel job berikutnya harus mendefinisikan nullability/status secara lengkap pada issue fiturnya. Definisi flag/version di atas berlaku untuk users, categories, transactions, dan templates.
Gunakan CHECK(length(btrim(name/title))>0) jika berlaku. Batas amount/title tetap di requirements.md.
Transactions menerapkan FOREIGN KEY(category_id,user_id,type) REFERENCES categories(id,user_id,type); templates memakai constraint yang sama.
Role menjelaskan kewenangan aplikasi, bukan hak superuser database. Edit admin mempertahankan user_id owner; created_by/updated_by menyimpan actor sebenarnya.
Keunikan email dan nama kategori per owner/type mencakup baris terhapus agar recovery tidak ambigu; restore baris lama, jangan membuat identitas baru.
Foreign key membatasi penghapusan fisik; tidak ada cascade erase untuk riwayat keuangan/audit.

## Index

- categories: UNIQUE(user_id,type,lower(name)), termasuk baris terhapus.
- transactions: (user_id,transaction_date DESC,id DESC) WHERE is_delete=false untuk list/total normal.
- transactions: (user_id,category_id,transaction_date) WHERE is_delete=false untuk laporan kategori.
- transactions: (user_id,updated_at DESC,id DESC) WHERE is_delete=true untuk Trash.
- refresh_sessions: (user_id,family_id) dan expires_at untuk cleanup.
- admin_access_events: (actor_user_id,created_at DESC) dan (target_user_id,created_at DESC).
- jobs: index owner/status dan status/waktu runnable, didefinisikan pada 009/012.

Jangan menambah index boolean dengan selektivitas rendah secara mandiri. Ukur EXPLAIN pada query dengan scope owner yang mewakili; jangan membuat index title/amount tanpa query nyata.

## Otorisasi dan konkurensi

Service membuat scope eksplisit {actor_user_id,owner_user_id,mode}. Personal owner=actor; admin owner=target path tervalidasi.
Semua baca/tulis berscope menyertakan owner ID, termasuk untuk admin. Resource terkait dari target lain menghasilkan 404.
Sebelum mutasi, lock baris akun yang terlibat dalam urutan UUID dengan FOR UPDATE, periksa ulang actor aktif/role admin dan state target, lalu mutasikan resource dengan version wajib.
Delete/restore/perubahan role akun memakai protokol lock yang sama. Jangan menahan lock saat memanggil Google atau merender file.
Aturan domain menerima owner biasa atau admin aktif; scope repository tidak boleh berasal langsung dari body request.

## Create, update, delete, dan restore

Create memasukkan is_delete=false, version=1. Transactions mencatat actor sebenarnya pada created_by dan updated_by.
Hash field create kanonis dan tegakkan UNIQUE(owner,client_request_id). Replay request yang sama mengembalikan hasil asli; field berbeda atau replay atas hasil terhapus mengembalikan 409. Kunci idempotensi tetap disimpan setelah delete.
Edit transaksi aktif memakai WHERE id=$id AND user_id=$owner AND version=$expected AND is_delete=false; menaikkan version dan updated_by/updated_at.
Soft delete memakai predicate sama, SET is_delete=true, version=version+1, updated_by=$actor, updated_at=now(). Tidak ada DELETE FROM pada tabel bisnis.
Restore memerlukan is_delete=true dan expected version; mengatur false dan menaikkan version. Validasi akun target serta kategori aktif/tipe/owner sebelum restore.
Percobaan lifecycle yang dimiliki tetapi stale/sudah terhapus/sudah dipulihkan mengembalikan 409. Baris yang benar-benar hilang atau owner salah mengembalikan 404.
Users/categories/templates juga memakai flag+version; atribusi mutasi admin ada di admin_access_events.
DTO PATCH/POST biasa menolak isDelete; route delete/restore adalah satu-satunya penulis siklus hidup.

## Perilaku kategori/akun

Soft delete kategori menyembunyikannya dari selector tanpa menghapus transaksi atau mengubah jumlah. Join historis mempertahankan label kategori owner yang sama meski category.is_delete=true.
Transaksi/templat baru/edit/restore harus memakai kategori aktif yang cocok; pulihkan kategori lebih dahulu atau pilih kategori aktif lain saat edit. Record historis tetap dapat dibaca dan dihapus.
Soft delete user mengatur flag akun dan mencabut sesi dalam transaksi yang sama, menjeda jadwal, dan memblokir job antre saat eksekusi. Flag anak tidak berubah.
Hanya admin yang dapat memulihkan akun terhapus karena pengguna itu tidak dapat autentikasi. Restore tidak menghidupkan sesi yang dicabut atau melanjutkan jadwal otomatis.
Perintah bootstrap/recovery operator tepercaya tetap tersedia; tidak ada promosi role otomatis.

## Pembacaan, laporan, dan ekspor

Detail/list transaksi aktif: WHERE user_id=$owner AND is_delete=false. Trash memakai predicate owner yang sama dengan is_delete=true dan updated_at DESC,id DESC.
Report/ekspor selalu memfilter transaksi aktif, apa pun state UI Trash. Gunakan transaction_date >= start AND <= end inklusif.
Jumlahkan income/expense secara terpisah dengan COALESCE(...,0); hitung difference. Filter tanggal sebelum pengelompokan minggu/bulan; bucket batas dapat parsial.
Riwayat harian hanya mengembalikan tanggal yang terisi; ringkasan hari ini dihitung terpisah dan dapat bernilai nol.
Jangan menambahkan predicate kategori terhapus pada join transaksi karena dapat menghilangkan jumlah historis.
Cursor mengikat actor, owner, mode, endpoint, filter tanggal/kategori/tipe, dan state penghapusan.

## Audit dan role

Admin access event mencakup read/create/update/delete/restore, ekspor/job, akun, dan operasi role. resource_type/action/outcome memakai whitelist service; safe_metadata boleh memuat versi/nama field berubah dan role lama/baru, tetapi tidak payload finansial/secret.
Write admin dan insert audit sukses commit secara atomik. Pembacaan menyimpan audit sukses sebelum mengembalikan data. Percobaan berwenang yang gagal mencatat denial/conflict jika memungkinkan tanpa menutupi error utama.
Runtime boleh insert/read record audit untuk service log admin terlindungi, tetapi tidak update/delete. Runtime bukan pemilik migrasi.
SQL registrasi mengecualikan role dan memakai default. SQL profil personal hanya memperbarui field yang diizinkan. SQL akun/role admin hanya dipanggil setelah otorisasi role dan persiapan audit; update diizinkan melalui operasi admin terlindungi.
Pembaruan role mencabut sesi refresh target. Lock/recheck mencegah transaksi yang diotorisasi sebelum demotion bersamaan melakukan commit setelah perubahan role yang berlawanan tanpa serialisasi.

## Job, backup, dan pemulihan

Simpan owner dan requester secara terpisah. Role admin tidak menjadikan requester sebagai owner keuangan.
Owner dapat mengakses job sendiri; admin aktif dapat mengelola job pada route target terverifikasi, termasuk job yang dibuat admin lain.
Worker memvalidasi ulang requester/owner/request_mode. Job admin memerlukan requester masih admin. Job tidak valid dibatalkan/dijeda sebelum eksekusi; ID/filter tidak berubah saat retry.
Data backup memuat isDelete dan hubungan baris; laporan mengecualikan baris terhapus tetapi backup mempertahankannya untuk recovery.
Kecualikan role/password/session/secret provider dan audit record dari backup data personal. Import memvalidasi skema ketat, menolak field privilege, dan hanya memetakan ke owner terotorisasi dengan flag penghapusan tetap.
Snapshot restore boleh mempertahankan transaksi historis aktif yang merujuk kategori owner sama yang terhapus; validasi owner/type FK tanpa menghapus transaksi atau mengaktifkan kategori diam-diam. Restore transaksi individual dan edit baru tetap memerlukan kategori aktif.
Kebijakan konflik version/restore dan retensi provider diselesaikan pada Issue 012.

## Migrasi dan verifikasi

Instalasi baru: users → categories → transactions → sessions → admin_access_events; templates/job menyusul sesuai issue.
Jika upgrade dari skema timestamp deletion, tambahkan is_delete=false lalu isi true saat deleted_at IS NOT NULL; pertahankan kolom lama hanya selama rollout bertahap dan hapus setelah verifikasi. Jangan mengaktifkan kembali baris terhapus.
Untuk desain archive kategori lama, petakan archived_at IS NOT NULL menjadi is_delete=true saat mengganti semantik archive. Repositori ini belum memiliki skema ter-deploy; gunakan skema baru kecuali inspeksi membuktikan sebaliknya.
Isi atribusi actor pada transaksi lama dengan owner yang diketahui hanya jika informasi actor historis tidak ada, dan dokumentasikan keterbatasannya.
Migrasi mempertahankan semua baris/owner ID. Pengujian memeriksa default, backfill true/false, constraint, penolakan eskalasi role, CRUD admin A/B/C, version lama, race delete/restore, perilaku report/Trash, histori kategori, serta audit admin atomik.
Jalankan migration down hanya pada fixture yang boleh dibuang sampai dampak data direview. Dokumentasi ini tidak menjalankan SQL produksi.
